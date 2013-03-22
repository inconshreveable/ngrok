# ngrok

## What is ngrok?
ngrok is a tool that makes it easier to develop networked services (HTTP/TCP)
It is a man-in-the-middle proxy server that you run locally on your development box. It allows you to do the following things:
    - Inspect all http requests/responses sent to/from the proxied application
    - Expose a local http server to the internet on a subdomain of ngrok.com


## Compiling ngrok
Binaries get placed in ./bin

    git clone git@github.com:inconshreveable/ngrok.git
    cd ngrok
    go get -d ./...
    make
    bin/ngrok [LOCAL PORT]

