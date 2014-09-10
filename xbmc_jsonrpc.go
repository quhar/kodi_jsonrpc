// Package xbmc_jsonrpc provides an interface for communicating with an XBMC
// server via the raw JSON-RPC socket
//
// Extracted from the xbmc-callback-daemon.
//
// Released under the terms of the MIT License (see LICENSE).
package xbmc_jsonrpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"runtime"
	"sync"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/op/go-logging"
	"github.com/stefantalpalaru/pool"
)

const (
	VERSION = `0.0.1`

	// Minimum XBMC API version
	XBMC_MIN_VERSION = 6

	// Logger properties
	LOG_FORMAT = `[%{color}%{level}%{color:reset}] %{message}`
	LOG_MODULE = `xbmc_jsonrpc`
)

var logger *logging.Logger

// Main type for interacting with XBMC
type Connection struct {
	conn          net.Conn
	write         chan interface{}
	Notifications chan Notification
	enc           *json.Encoder
	dec           *json.Decoder
	lock          sync.Mutex
	requestId     uint32
	responses     map[uint32]*chan *rpcResponse
	pool          *pool.Pool

	address string
}

// RPC Request type
type Request struct {
	Id      *uint32                 `json:"id,omitempty"`
	Method  string                  `json:"method"`
	Params  *map[string]interface{} `json:"params,omitempty"`
	JsonRPC string                  `json:"jsonrpc"`
}

// RPC response error type
type rpcError struct {
	Code    float64                 `json:"code"`
	Message string                  `json:"message"`
	Data    *map[string]interface{} `json:"data"`
}

// RPC Response provides a reader for returning responses
type Response struct {
	channel *chan *rpcResponse
	Pending bool // If Pending is false, Response is unwanted, or been consumed
}

// RPC response type
type rpcResponse struct {
	Id      *float64                `json:"id"`
	JsonRPC string                  `json:"jsonrpc"`
	Method  *string                 `json:"method"`
	Params  *map[string]interface{} `json:"params"`
	Result  *map[string]interface{} `json:"result"`
	Error   *rpcError               `json:"error"`
}

// Notification stores XBMC server->client notifications.
type Notification struct {
	Method string `json:"method" mapstructure:"method"`
	Params struct {
		Data struct {
			Item *struct {
				Type string `json:"type" mapstructure:"type"`
			} `json:"item" mapstructure:"item"` // Optional
		} `json:"data" mapstructure:"data"`
	} `json:"params" mapstructure:"params"`
}

func init() {
	// Initialize logger, default to level `info`
	logging.SetFormatter(logging.MustStringFormatter(LOG_FORMAT))
	logger = logging.MustGetLogger(LOG_MODULE)
	logging.SetLevel(logging.INFO, LOG_MODULE)
}

// New returns a Connection to the specified address.
// If timeout (seconds) is greater than zero, connection will fail if initial
// version query is not returned within this time.
//
// User must ensure Close() is called on returned Connection when finished with
// it, to avoid leaks.
func New(address string, timeout time.Duration) (conn Connection, err error) {
	conn = Connection{}
	err = conn.init(address, timeout)

	return conn, err
}

// SetLogLevel adjusts the level of logger output
func SetLogLevel(level string) error {
	switch level {
	case `debug`, `DEBUG`:
		logging.SetLevel(logging.DEBUG, LOG_MODULE)
	case `notice`, `NOTICE`:
		logging.SetLevel(logging.NOTICE, LOG_MODULE)
	case `info`, `INFO`:
		logging.SetLevel(logging.INFO, LOG_MODULE)
	case `warning`, `WARNING`, `warn`, `WARN`:
		logging.SetLevel(logging.WARNING, LOG_MODULE)
	case `error`, `ERROR`:
		logging.SetLevel(logging.ERROR, LOG_MODULE)
	case `critical`, `CRITICAL`, `crit`, `CRIT`:
		logging.SetLevel(logging.CRITICAL, LOG_MODULE)
	default:
		return errors.New(fmt.Sprintf(`Unknown log level: %s`, level))
	}

	return nil
}

// Return the result and any errors from the response channel
func (rchan *Response) Read(timeout time.Duration) (result map[string]interface{}, err error) {
	//if timeout == 0

	if rchan.Pending != true {
		return result, errors.New(`No pending responses!`)
	}
	if rchan.channel == nil {
		return result, errors.New(`Expected response channel, but got nil!`)
	}

	res := new(rpcResponse)
	if timeout > 0 {
		select {
		case res = <-*rchan.channel:
		case <-time.After(timeout * time.Second):
			err = errors.New(`Timeout waiting on response channel`)
		}
	} else {
		res = <-*rchan.channel
	}
	if err == nil {
		result, err = res.unpack()
	}

	return result, err
}

// Unpack the result and any errors from the Response
func (res *rpcResponse) unpack() (result map[string]interface{}, err error) {
	if res.Error != nil {
		err = errors.New(fmt.Sprintf(
			`XBMC error (%v): %v`, res.Error.Code, res.Error.Message,
		))
	} else if res.Result != nil {
		result = *res.Result
	} else {
		logger.Debug(`Received unknown response type from XBMC: %v`, res)
	}
	return result, err
}

// init brings up an instance of the XBMC Connection
func (c *Connection) init(address string, timeout time.Duration) (err error) {
	if c.address == `` {
		c.address = address
	}

	if err = c.connect(); err != nil {
		return err
	}

	c.write = make(chan interface{}, 4)
	c.Notifications = make(chan Notification, 4)

	c.responses = make(map[uint32]*chan *rpcResponse)

	c.enc = json.NewEncoder(c.conn)
	c.dec = json.NewDecoder(c.conn)

	c.pool = pool.New(runtime.NumCPU() * 3)
	c.pool.Run()

	go c.reader()
	go c.writer()

	rchan := c.Send(Request{Method: `JSONRPC.Version`}, true)

	res, err := rchan.Read(timeout)
	if err != nil {
		logger.Error(`XBMC responded: %v`, err)
		return err
	}
	if version := res[`version`].(map[string]interface{}); version != nil {
		if version[`major`].(float64) < XBMC_MIN_VERSION {
			return errors.New(`XBMC version too low, upgrade to Frodo or later`)
		}
	}

	logger.Info(`Connected to XBMC`)

	return
}

// Send an RPC Send to the XBMC server.
// Returns a Response, but does not attach a channel for it if want_response is
// false (for fire-and-forget commands that don't return any useful response).
func (c *Connection) Send(req Request, want_response bool) Response {
	req.JsonRPC = `2.0`
	res := Response{}

	if want_response == true {
		c.lock.Lock()
		id := c.requestId
		ch := make(chan *rpcResponse)
		c.responses[id] = &ch
		c.requestId++
		c.lock.Unlock()
		req.Id = &id

		logger.Debug(`Sending XBMC Request (response requested): %v`, req)
		c.write <- req
		res.channel = &ch
		res.Pending = true
	} else {
		logger.Debug(`Sending XBMC Request (no response requested): %v`, req)
		c.write <- req
		res.Pending = false
	}

	return res
}

// connect establishes a TCP connection
func (c *Connection) connect() (err error) {
	c.conn, err = net.Dial(`tcp`, c.address)
	for err != nil {
		logger.Error(`Connecting to XBMC: %v`, err)
		logger.Info(`Attempting reconnect...`)
		time.Sleep(time.Second)
		c.conn, err = net.Dial(`tcp`, c.address)
	}
	err = nil

	return
}

// writer loop processes outbound requests
func (c *Connection) writer() {
	for {
		var req interface{}
		req = <-c.write
		if err := c.enc.Encode(req); err != nil {
			logger.Warning(`Failed encoding request for XBMC: %v`, err)
			break
		}
	}
}

// reader loop processes inbound responses and notifications
func (c *Connection) reader() {
	for {
		res := new(rpcResponse)
		err := c.dec.Decode(res)
		if err == io.EOF {
			logger.Error(`Reading from XBMC: %v`, err)
			logger.Error(`If this error persists, make sure you are using the JSON-RPC port, not the HTTP port!`)
			time.Sleep(time.Second)
			if err = c.connect(); err != nil {
				logger.Error(`Reconnecting to XBMC: %v`, err)
			}
		} else if err != nil {
			logger.Error(`Decoding response from XBMC: %v`, err)
		} else {
			if res.Id == nil && res.Method != nil {
				logger.Debug(`Received notification from XBMC: %v`, *res.Method)
				n := Notification{}
				n.Method = *res.Method
				mapstructure.Decode(res.Params, &n.Params)
				c.Notifications <- n
			} else if res.Id != nil {
				if ch := c.responses[uint32(*res.Id)]; ch != nil {
					if res.Result != nil {
						logger.Debug(`Received response from XBMC: %v`, *res.Result)
					}
					*ch <- res
				} else {
					logger.Warning(
						`Received XBMC response for unknown request: %v`,
						*res.Id,
					)
					logger.Debug(
						`Current response channels: %v`, c.responses,
					)
				}
			} else {
				if res.Error != nil {
					logger.Warning(`Received unparseable XBMC response: %v`, res.Error)
				} else {
					logger.Warning(`Received unparseable XBMC response: %v`, res)
				}
			}
		}
	}
}

// Close XBMC connection
func (c *Connection) Close() {
	for _, v := range c.responses {
		close(*v)
	}
	close(c.write)
	close(c.Notifications)
	c.pool.Stop()
	c.conn.Close()
}
