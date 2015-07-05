
<a name="v2.0.2"></a>
# v2.0.2 (2015-07-05)

## :house: Housekeeping

- **logs**:
  - Include full notification content in debug output ([bff25145](https://github.com/StreamBoat/kodi_jsonrpc/commit/bff25145c2766ff2877feee2b56b6d4415fccbb8))


<a name="v2.0.1"></a>
# v2.0.1 (2015-07-05)

## :house: Housekeeping

- **style**:
  - Fix lint that won't break existing APIs ([54af862a](https://github.com/StreamBoat/kodi_jsonrpc/commit/54af862af1928000f87bcbb3092cf1ac59746960))


<a name="v2.0.0"></a>
# v2.0.0 (2015-03-17)

## :bug: Bug Fixes

- **core**:
  - Implement ring buffer for notification writes ([a9e60882](https://github.com/StreamBoat/kodi_jsonrpc/commit/a9e60882ddab062ca7fced3cb56d5586b9e1ad1f))  <br>Clients that don't process notifications could result in blocked reader
    goroutine

## :house: Housekeeping

- **core**:
  - Return error trying to send on closed connection ([2ad70a41](https://github.com/StreamBoat/kodi_jsonrpc/commit/2ad70a415c661980a3001245d3757ee483461c09))

## Breaking Changes

- due to [2ad70a41](https://github.com/StreamBoat/kodi_jsonrpc/commit/2ad70a415c661980a3001245d3757ee483461c09), `Send` now returns `(Response, error)`


<a name="v1.0.3"></a>
# v1.0.3 (2014-11-25)

## :house: Housekeeping

- **concurrency**:
  - Improve concurrency support for Read operation ([380de829](https://github.com/StreamBoat/kodi_jsonrpc/commit/380de829d0eeadcaf5d457daa80d79c0404a3c6c))  
    <br>Also, add timeout support to initial connection


<a name="v1.0.2"></a>
# v1.0.2 (2014-11-25)

## :bug: Bug Fixes

- **reconnect**:
  - Add reconnect locking and actually update `Connected` ([32b5b63e](https://github.com/StreamBoat/kodi_jsonrpc/commit/32b5b63e3840c122abdc787e34e9f7c6ace16702))


<a name="v1.0.1"></a>
# v1.0.1 (2014-11-23)

## :bug: Bug Fixes

- **reconnect**:
  - Less painful reconnect ([a0f91e9d](https://github.com/StreamBoat/kodi_jsonrpc/commit/a0f91e9d89b2536700fe17d4c5709849a844203c))  <br>Vastly reduce the number of things touched by a reconnect, we only need
    to mess with the connection itself, and the attached encoders/decoders


<a name="v1.0.0"></a>
# v1.0.0 (2014-11-23)

## :house: Housekeeping

- **core**:
  - Rename from XBMC to Kodi, best to get this out of the way ([6e307398](https://github.com/StreamBoat/kodi_jsonrpc/commit/6e30739875014414562eb6ae11e7a30bc85e792c))
- **build**:
  - Add config for abe33/changelog-gen ([1502885c](https://github.com/StreamBoat/kodi_jsonrpc/commit/1502885c4d32f38850fc07b15215c0d29e0c23a2))  
    <br>TODO: Add contribution guidelines for commit messages
- **logging**:
  - Replace go-logging with logrus for structured logs ([0c4273a0](https://github.com/StreamBoat/kodi_jsonrpc/commit/0c4273a01011b2ca871ab7dfea61e7f8b123565e))

## Breaking Changes

- due to [6e307398](https://github.com/StreamBoat/kodi_jsonrpc/commit/6e30739875014414562eb6ae11e7a30bc85e792c), renaming the package represents a significant break for all existing users, which is why I wanted to get it done now, rather than at some future time.
- due to [0c4273a0](https://github.com/StreamBoat/kodi_jsonrpc/commit/0c4273a01011b2ca871ab7dfea61e7f8b123565e), `SetLogLevel` now takes a constant - one of: `LogDebugLevel`, `LogInfoLevel`, `LogWarnLevel`, `LogErrorLevel`, `LogFatalLevel`, `LogPanicLevel`

