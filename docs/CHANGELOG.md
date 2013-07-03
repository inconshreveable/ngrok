# Changelog

## 0.14 - 07/03/2013
- BUGFIX Fix an issue where ngrok could never save/load the authtoken file on linux
- BUGFIX Fix an issue where ngrok wouldn't emit log messages while loading authtokens

## 0.13 - 07/02/2013
- FEATURE -hostname switch on client allows you to run tunnels over custom domains (requires you CNAME your DNS)
- IMPROVEMENT ngrok client UI now shows the client IP address for a request
- IMPROVEMENT ngrok client UI now shows how long ago a request was made (uservoice request 4127487)
- IMPROVEMENT ngrokd now uses and LRU cache for tunnel affinity data
- IMPROVEMENT ngrokd can now save and restore its tunnel affinity cache to a file to preserve across restarts

## 0.12 - 06/30/2013
- IMPROVEMENT Improved developer documentation
- IMPROVEMENT Simplified build process with custom version of go-bindata that compiles assets into binary releases
- BUGFIX Github issue #4: Raw/Binary requests bodies are no longer truncated at 8192 bytes.
