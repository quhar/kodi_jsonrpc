# xbmc_jsonrpc
[![GoDoc](https://godoc.org/github.com/streamboat/xbmc_jsonrpc?status.svg)](http://godoc.org/github.com/streamboat/xbmc_jsonrpc) ![License-MIT](http://img.shields.io/badge/license-MIT-red.svg)

```go
import "github.com/streamboat/xbmc_jsonrpc"
```

Package xbmc_jsonrpc provides an interface for communicating with an XBMC server
via the raw JSON-RPC socket

Extracted from the xbmc-callback-daemon.

Released under the terms of the MIT License (see LICENSE).

## Usage

```go
const (
	VERSION = `0.0.1`

	// Minimum XBMC API version
	XBMC_MIN_VERSION = 6

	// Logger properties
	LOG_FORMAT = `[%{color}%{level}%{color:reset}] %{message}`
	LOG_MODULE = `xbmc_jsonrpc`
)
```

#### func  SetLogLevel

```go
func SetLogLevel(level string) error
```
SetLogLevel adjusts the level of logger output

#### type Connection

```go
type Connection struct {
	Notifications chan Notification
}
```

Main type for interacting with XBMC

#### func  New

```go
func New(address string, timeout time.Duration) (conn Connection, err error)
```
New returns a Connection to the specified address. If timeout (seconds) is
greater than zero, connection will fail if initial version query is not returned
within this time.

User must ensure Close() is called on returned Connection when finished with it,
to avoid leaks.

#### func (*Connection) Close

```go
func (c *Connection) Close()
```
Close XBMC connection

#### func (*Connection) Send

```go
func (c *Connection) Send(req Request, want_response bool) Response
```
Send an RPC Send to the XBMC server. Returns a Response, but does not attach a
channel for it if want_response is false (for fire-and-forget commands that
don't return any useful response).

#### type Notification

```go
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
```

Notification stores XBMC server->client notifications.

#### type Request

```go
type Request struct {
	Id      *uint32                 `json:"id,omitempty"`
	Method  string                  `json:"method"`
	Params  *map[string]interface{} `json:"params,omitempty"`
	JsonRPC string                  `json:"jsonrpc"`
}
```

RPC Request type

#### type Response

```go
type Response struct {
	Pending bool // If Pending is false, Response is unwanted, or been consumed
}
```

RPC Response provides a reader for returning responses

#### func (*Response) Read

```go
func (rchan *Response) Read(timeout time.Duration) (result map[string]interface{}, err error)
```
Return the result and any errors from the response channel
