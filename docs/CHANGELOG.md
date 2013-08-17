# Changelog
## 0.21 - 08/17/2013
- IMPROVEMENT: The ngrok web ui can now be disabled with -webport=-1

## 0.20 - 08/17/2013
- BUGFIX: Fixed a bug where ngrok would not stop its autoupdate loop even after it should stop

## 0.19 - 08/17/2013
- BUGFIX: Fixed a bug where ngrok's would loop infinitely trying to checking for updates after the second update check
- BUGFIX: Fixed a race condition in ngrokd's metrics logging immediately after start up

## 0.18 - 08/15/2013
- BUGFIX: Fixed a bug where ngrok would compare the Host header for virtual hosting using case-sensitive comparisons
- BUGFIX: Fixed a bug where ngrok would not include the port number in the virtual host when not serving on port 80
- BUGFIX: Fixed a bug where ngrok would crash when trying to replay a request
- IMPROVEMENT: ngrok can now indicate manual updates again
- IMPROVEMENT: ngrok can now supports update channels
- IMPROVEMENT: ngrok can now detect some updates that will fail before downloading

## 0.17 - 07/30/2013
- BUGFIX: Fixed an issue where ngrok's registry cache would return a URL from a different protocol

## 0.16 - 07/30/2013
- BUGFIX: Fixed an issue where ngrok would crash when parsing bad XML that wasn't a syntax error
- BUGFIX: Fixed an issue where ngrok would crash when parsing bad JSON that wasn't a syntax error
- BUGFIX: Fixed an issue where the web ui would sometimes not update the request body when changing requests
- BUGFIX: Fixed an issue where ngrokd's registry cache would not load from file
- BUGFIX: Fixed an issue where ngrokd's registry cache would not save to file
- BUGFIX: Fixed an issue where ngrok would refuse requests with an Authorization header if no HTTP auth was specified.
- BUGFIX: Fixed a bug where ngrok would fail to cross-compile in you hadn't compiled natively first
- IMPROVEMENT: ngrok's registry cache now handles and attempts to restore TCP URLs
- IMPROVEMENT: Added simple Travis CI integration to make sure ngrok compiles

## 0.15 - 07/27/2013
- FEATURE: ngrok can now update itself automatically

## 0.14 - 07/03/2013
- BUGFIX: Fix an issue where ngrok could never save/load the authtoken file on linux
- BUGFIX: Fix an issue where ngrok wouldn't emit log messages while loading authtokens

## 0.13 - 07/02/2013
- FEATURE: -hostname switch on client allows you to run tunnels over custom domains (requires you CNAME your DNS)
- IMPROVEMENT: ngrok client UI now shows the client IP address for a request
- IMPROVEMENT: ngrok client UI now shows how long ago a request was made (uservoice request 4127487)
- IMPROVEMENT: ngrokd now uses and LRU cache for tunnel affinity data
- IMPROVEMENT: ngrokd can now save and restore its tunnel affinity cache to a file to preserve across restarts

## 0.12 - 06/30/2013
- IMPROVEMENT: Improved developer documentation
- IMPROVEMENT: Simplified build process with custom version of go-bindata that compiles assets into binary releases
- BUGFIX: Github issue #4: Raw/Binary requests bodies are no longer truncated at 8192 bytes.
