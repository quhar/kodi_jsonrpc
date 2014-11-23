
<a name="v1.0.0"></a>
# v1.0.0 (2014-11-23)

## :house: Housekeeping

- **core**:
  - Rename from XBMC to Kodi, best to get this out of the way ([6e307398](https://github.com/StreamBoat/xbmc_jsonrpc/commit/6e30739875014414562eb6ae11e7a30bc85e792c))
- **build**:
  - Add config for abe33/changelog-gen ([1502885c](https://github.com/StreamBoat/xbmc_jsonrpc/commit/1502885c4d32f38850fc07b15215c0d29e0c23a2))  
    <br>TODO: Add contribution guidelines for commit messages
- **logging**:
  - Replace go-logging with logrus for structured logs ([0c4273a0](https://github.com/StreamBoat/xbmc_jsonrpc/commit/0c4273a01011b2ca871ab7dfea61e7f8b123565e))

## Breaking Changes

- due to [6e307398](https://github.com/StreamBoat/xbmc_jsonrpc/commit/6e30739875014414562eb6ae11e7a30bc85e792c), renaming the package represents a significant break for all existing users, which is why I wanted to get it done now, rather than at some future time.
- due to [0c4273a0](https://github.com/StreamBoat/xbmc_jsonrpc/commit/0c4273a01011b2ca871ab7dfea61e7f8b123565e), `SetLogLevel` now takes a constant - one of: `LogDebugLevel`, `LogInfoLevel`, `LogWarnLevel`, `LogErrorLevel`, `LogFatalLevel`, `LogPanicLevel`

