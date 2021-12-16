# How to run your own pgrokd server

Running your own pgrok server is really easy! The instructions below will guide you along your way!

## 1. Get an SSL certificate

pgrok provides secure tunnels via TLS, so you'll need an SSL certificate. Assuming you want to create
tunnels on _.ejemplo.me, buy a wildcard SSL certificate for _.ejemplo.me. Note that if you
don't need to run https tunnels that you don't need a wildcard certificate. (In fact, you can
just use a self-signed cert at that point, see the section on that later in the document).

## 2. Modify your DNS

You need to use the DNS management tools given to you by your provider to create an A
record which points \*.ejemplo.me to the IP address of the server where you will run pgrokd.

## 3. Compile it

You can compile an pgrokd server with the following command:

    make server

Or you can download it from release section https://github.com/jerson/pgrok/releases you need **pgrokd** file

Then copy the binary over to your server.

## 4. Run the server


### Some important options

#### Specifying your TLS certificate and key

pgrok only makes TLS-encrypted connections. When you run pgrokd, you'll need to instruct it
where to find your TLS certificate and private key. Specify the paths with the following switches:

    -tlsKey="/path/to/tls.key" -tlsCrt="/path/to/tls.crt"

#### Setting the server's domain

When you run your own pgrokd server, you need to tell pgrokd the domain it's running on so that it
knows what URLs to issue to clients.

    -domain="ejemplo.me"

#### Protect you client(pgrok) to server(pgrokd) connection with a CA

if you include this, all pgrok client will need two aditional arguments to connect to this server, that needs to be a client certificate for this CA

     -tunnelTLSClientCA=./ca.crt

to see which parameters pgrok client needs, is in "Connect with a client" section

#### Protect you exposed subdomain with a CA

ej: sample.ejemplo.me will only be accesible when the user has a client certificate allowed by this CA and also needs to be installed on his device

     -tlsClientCA=./ca.crt

### (Option 1) Command line
You'll run the server with the following command.

    ./pgrokd -domain ejemplo.me -httpAddr=:80 -httpsAddr=:443 -tunnelAddr=:4443 -tlsCrt=./tls.crt -tlsKey=./tls.key


### (Option 2) Supervidord
you can use supervisor to run in background, here is a sample config file

```ini
[program:pgrokd]
directory=/root/pgrok
autostart=true
autorestart=true
command=/root/pgrok/pgrokd -domain ejemplo.me -log-level=WARNING -httpAddr=:80 -httpsAddr=:443 -tunnelAddr=:4443 -tlsCrt=./certs/tls.crt -tlsKey=./certs/tls.key

```

### (Option 3) Docker Compose

or also you can use docker-compose

```yaml
version: "3.7"

services:
  pgrokd:
    image: jerson/pgrok
    entrypoint: pgrokd
    command: -domain ejemplo.me -httpAddr=:80 -httpsAddr=:443 -tunnelAddr=:4443 -tlsCrt=/certs/tls.crt -tlsKey=/certs/tls.key
    volumes:
      - /home/certs:/certs
    ports:
      - 80:80
      - 443:443
      - 4443:4443
```



## 5. Configure client

In order to Configure client, you'll need to set two options in pgrok's configuration file.
The pgrok configuration file is a simple YAML file that is read from ~/.pgrok by default. You may specify
a custom configuration file path with the -config switch. Your config file must contain the following two
options.

    server_addr: ejemplo.me:4443
    trust_host_root_certs: true

Substitute the address of your pgrokd server for "ejemplo.me:4443". The "trust_host_root_certs" parameter instructs
pgrok to trust the root certificates on your computer when establishing TLS connections to the server. By default, pgrok
only trusts the root certificate for ejemplo.me.

## 6. Connect with a client

Then, just run pgrok as usual to connect securely to your own pgrokd server!

    pgrok -subdomain=customsubdomain 127.0.0.1:3000

or you can specify a custom server here too

    pgrok -log=stdout -serveraddr=ejemplo.me:4443 -subdomain=customsubdomain 127.0.0.1:3000

or if you are using "Protect you client(pgrok) to server(pgrokd) connection with a CA
" here is where you need to include to arguments ` -tlsClientCrt=./client_crt -tlsClientKey=./client_key`

    pgrok -log=stdout -log-level=WARNING -serveraddr=ejemplo:4443 -tlsClientCrt=./client_crt -tlsClientKey=./client_key -subdomain=customsubdomain 127.0.0.1:3000
# pgrokd with a self-signed SSL certificate

It's possible to run pgrokd with a a self-signed certificate, but you'll need to recompile pgrok with your signing CA.
If you do choose to use a self-signed cert, please note that you must either remove the configuration value for
trust_host_root_certs or set it to false:

    trust_host_root_certs: false
