# kodi_jsonrpc
[![GoDoc](https://godoc.org/github.com/StreamBoat/xbmc_jsonrpc?status.svg)](http://godoc.org/github.com/StreamBoat/xbmc_jsonrpc) ![License-MIT](http://img.shields.io/badge/license-MIT-red.svg)
```go
import "github.com/StreamBoat/xbmc_jsonrpc"
```

Package `kodi_jsonrpc` provides an interface for communicating with a Kodi/XBMC
server via the raw JSON-RPC socket

Extracted from the kodi-callback-daemon.

Released under the terms of the MIT License (see LICENSE).

## Usage

```go
const (
	VERSION = `1.0.0`

	// Minimum Kodi/XBMC API version
	KODI_MIN_VERSION = 6

	LogDebugLevel = log.DebugLevel
	LogInfoLevel  = log.InfoLevel
	LogWarnLevel  = log.WarnLevel
	LogErrorLevel = log.ErrorLevel
	LogFatalLevel = log.FatalLevel
	LogPanicLevel = log.PanicLevel
)
```

#### func  SetLogLevel

```go
func SetLogLevel(level log.Level)
```
SetLogLevel adjusts the level of logger output, level must be one of:

LogDebugLevel LogInfoLevel LogWarnLevel LogErrorLevel LogFatalLevel
LogPanicLevel

#### type Connection

```go
type Connection struct {
	Notifications chan Notification
}
```

Main type for interacting with Kodi

#### func  New

```go
func New(address string, timeout time.Duration) (conn Connection, err error)
```
New returns a Connection to the specified address. If timeout (seconds) is
greater than zero, connection will fail if initial version query is not returned
within this time.

User must ensure Close() is called on returned Connection when finished with it,
to avoid leaks.

#### func (\*Connection) Close

```go
func (c *Connection) Close()
```
Close Kodi connection

#### func (\*Connection) Send

```go
func (c *Connection) Send(req Request, want_response bool) Response
```
Send an RPC Send to the Kodi server. Returns a Response, but does not attach a
channel for it if `want_response` is false (for fire-and-forget commands that
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

Notification stores Kodi server->client notifications.

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

#### func (\*Response) Read

```go
func (rchan *Response) Read(timeout time.Duration) (result map[string]interface{}, err error)
```
Return the result and any errors from the response channel
