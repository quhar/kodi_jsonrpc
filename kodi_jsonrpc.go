// Package kodi_jsonrpc provides an interface for communicating with a Kodi/XBMC
// server via the raw JSON-RPC socket
//
// Extracted from the kodi-callback-daemon.
//
// Released under the terms of the MIT License (see LICENSE).
package kodi_jsonrpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/mitchellh/mapstructure"
)

// Main type for interacting with Kodi
type Connection struct {
	conn             net.Conn
	write            chan interface{}
	Notifications    chan Notification
	enc              *json.Encoder
	dec              *json.Decoder
	responseLock     sync.Mutex
	writeWait        sync.WaitGroup
	notificationWait sync.WaitGroup
	requestId        uint32
	responses        map[uint32]*chan *rpcResponse

	Connected bool
	Closed    bool

	address string
	timeout time.Duration
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

// Notification stores Kodi server->client notifications.
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

const (
	VERSION = `1.0.1`

	// Minimum Kodi/XBMC API version
	KODI_MIN_VERSION = 6

	LogDebugLevel = log.DebugLevel
	LogInfoLevel  = log.InfoLevel
	LogWarnLevel  = log.WarnLevel
	LogErrorLevel = log.ErrorLevel
	LogFatalLevel = log.FatalLevel
	LogPanicLevel = log.PanicLevel
)

func init() {
	// Initialize logger, default to level Info
	log.SetLevel(LogInfoLevel)
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

// SetLogLevel adjusts the level of logger output, level must be one of:
//
// LogDebugLevel
// LogInfoLevel
// LogWarnLevel
// LogErrorLevel
// LogFatalLevel
// LogPanicLevel
func SetLogLevel(level log.Level) {
	log.SetLevel(level)
}

// Return the result and any errors from the response channel
func (rchan *Response) Read(timeout time.Duration) (result map[string]interface{}, err error) {
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
	close(*rchan.channel)

	return result, err
}

// Unpack the result and any errors from the Response
func (res *rpcResponse) unpack() (result map[string]interface{}, err error) {
	if res.Error != nil {
		err = errors.New(fmt.Sprintf(
			`Kodi error (%v): %v`, res.Error.Code, res.Error.Message,
		))
	} else if res.Result != nil {
		result = *res.Result
	} else {
		log.WithField(`response`, res).Debug(`Received unknown response type from Kodi`)
	}
	return result, err
}

// init brings up an instance of the Kodi Connection
func (c *Connection) init(address string, timeout time.Duration) (err error) {

	if c.address == `` {
		c.address = address
	}
	if c.timeout == 0 && timeout != 0 {
		c.timeout = timeout
	}

	if err = c.connect(); err != nil {
		return err
	}

	c.write = make(chan interface{}, 16)
	c.Notifications = make(chan Notification, 16)

	c.responses = make(map[uint32]*chan *rpcResponse)

	go c.reader()
	go c.writer()

	rchan := c.Send(Request{Method: `JSONRPC.Version`}, true)

	res, err := rchan.Read(c.timeout)
	if err != nil {
		log.WithField(`error`, err).Error(`Kodi responded`)
		return err
	}
	if version := res[`version`].(map[string]interface{}); version != nil {
		if version[`major`].(float64) < KODI_MIN_VERSION {
			return errors.New(`Kodi version too low, upgrade to Frodo or later`)
		}
	}

	return
}

// Send an RPC Send to the Kodi server.
// Returns a Response, but does not attach a channel for it if want_response is
// false (for fire-and-forget commands that don't return any useful response).
func (c *Connection) Send(req Request, want_response bool) Response {
	req.JsonRPC = `2.0`
	res := Response{}

	c.writeWait.Add(1)
	if want_response == true {
		c.responseLock.Lock()
		id := c.requestId
		ch := make(chan *rpcResponse)
		c.responses[id] = &ch
		c.requestId++
		c.responseLock.Unlock()
		req.Id = &id

		log.WithField(`request`, req).Debug(`Sending Kodi Request (response desired)`)
		c.write <- req
		res.channel = &ch
		res.Pending = true
	} else {
		log.WithField(`request`, req).Debug(`Sending Kodi Request (response undesired)`)
		c.write <- req
		res.Pending = false
	}
	c.writeWait.Done()

	return res
}

// set whether we're connected or not
func (c *Connection) connected(status bool) {
}

// connect establishes a TCP connection
func (c *Connection) connect() (err error) {
	c.connected(false)
	defer c.connected(true)

	c.conn, err = net.Dial(`tcp`, c.address)
	for err != nil {
		log.WithField(`error`, err).Error(`Connecting to Kodi`)
		log.Info(`Attempting reconnect...`)
		time.Sleep(time.Second)
		c.conn, err = net.Dial(`tcp`, c.address)
	}
	err = nil

	c.enc = json.NewEncoder(c.conn)
	c.dec = json.NewDecoder(c.conn)

	log.Info(`Connected to Kodi`)

	return
}

// writer loop processes outbound requests
func (c *Connection) writer() {
	for {
		var req interface{}
		req = <-c.write
		err := c.enc.Encode(req)
		if _, ok := err.(net.Error); ok {
			err = c.connect()
			c.enc.Encode(req)
		} else if err != nil {
			log.WithField(`error`, err).Warn(`Failed encoding request for Kodi`)
			err = c.connect()
			c.enc.Encode(req)
		}
	}
}

// reader loop processes inbound responses and notifications
func (c *Connection) reader() {
	for {
		res := new(rpcResponse)
		err := c.dec.Decode(res)
		if _, ok := err.(net.Error); err == io.EOF || ok {
			log.WithField(`error`, err).Error(`Reading from Kodi`)
			log.Error(`If this error persists, make sure you are using the JSON-RPC port, not the HTTP port!`)
			err = c.connect()
		} else if err != nil {
			log.WithField(`error`, err).Error(`Decoding response from Kodi`)
			continue
		}
		if res.Id == nil && res.Method != nil {
			c.notificationWait.Add(1)
			log.WithField(`response.Method`, *res.Method).Debug(`Received notification from Kodi`)
			n := Notification{}
			n.Method = *res.Method
			mapstructure.Decode(res.Params, &n.Params)
			c.Notifications <- n
			c.notificationWait.Done()
		} else if res.Id != nil {
			if ch := c.responses[uint32(*res.Id)]; ch != nil {
				if res.Result != nil {
					log.WithField(`response.Result`, *res.Result).Debug(`Received response from Kodi`)
				}
				*ch <- res
			} else {
				log.WithField(`response.Id`, *res.Id).Warn(`Received Kodi response for unknown request`)
				log.WithField(`connection.responses`, c.responses).Debug(`Current response channels`)
			}
		} else {
			if res.Error != nil {
				log.WithField(`response.Error`, *res.Error).Warn(`Received unparseable Kodi response`)
			} else {
				log.WithField(`response`, res).Warn(`Received unparseable Kodi response`)
			}
		}
	}
}

// Close Kodi connection
func (c *Connection) Close() {
	if c.Closed {
		return
	}
	c.Closed = true

	if c.write != nil {
		c.writeWait.Wait()
		close(c.write)
	}
	if c.Notifications != nil {
		c.notificationWait.Wait()
		close(c.Notifications)
	}
	if c.conn != nil {
		_ = c.conn.Close()
	}

	log.Info(`Disconnected from Kodi`)
}
