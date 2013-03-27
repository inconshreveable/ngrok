# ngrok

## What is ngrok?
ngrok is a tool that makes it easier to develop networked services (HTTP/TCP)
It is a man-in-the-middle proxy server that you run locally on your development box. It allows you to do the following things:

- Inspect all http requests/responses sent to/from the proxied application
- Replay any previously observed http request to the proxied application
- Expose a local http server to the internet on a subdomain of ngrok.com
- Expose a local tcp server to the internet on a random port on ngrok.com


## Downloading and installing ngrok
ngrok has _no_ dependencies. Just download a single binary for your platform and
run it.

[Linux](http://ngrok.com/dl/linux_386/ngrok")
[Mac OSX](http://ngrok.com/dl/darwin_386/ngrok")


## Compiling ngrok
Binaries get placed in ./bin

    git clone git@github.com:inconshreveable/ngrok.git
    cd ngrok && make
    bin/ngrok [LOCAL PORT]

