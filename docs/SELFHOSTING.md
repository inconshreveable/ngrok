# How to run your own ngrokd server

Running your own ngrok server is really easy! The instructions below will guide you along your way!

## 1. Get an SSL certificate
ngrok provides secure tunnels via TLS, so you'll need an SSL certificate. Assuming you want to create
tunnels on *.example.com, buy a wildcard SSL certificate for *.example.com.

## 2. Modify your DNS
You need to use the DNS management tools given to you by your provider to create an A
record which points *.example.com to the IP address of the server where you will run ngrokd.

## 3. Compile it
You can compile an ngrokd server with the following command:

	make release-server

Make sure you compile it with the GOOS/GOARCH environment variables set to the platform of
your target server. Then copy the binary over to your server.

## 4. Run the server
You'll run the server with the following command.


	./ngrokd -tlsKey="/path/to/tls.key" -tlsCrt="/path/to/tls.crt" -domain="example.com"

### Specifying your TLS certificate and key
ngrok only makes TLS-encrypted connections. When you run ngrokd, you'll need to instruct it
where to find your TLS certificate and private key. Specify the paths with the following switches:

	-tlsKey="/path/to/tls.key" -tlsCert="/path/to/tls.crt"

### Setting the server's domain
When you run your own ngrokd server, you need to tell ngrokd the domain it's running on so that it
knows what URLs to issue to clients.

	-domain="example.com"

## 5. Configure the client
In order to connect with a client, you'll need to set two options in ngrok's configuration file.
The ngrok configuration file is a simple YAML file that is read from ~/.ngrok by default. You may specify
a custom configuration file path with the -config switch. Your config file must contain the following two
options.

	server_addr: example.com:4443
	trust_host_root_certs: true

Subsitute the address of your ngrokd server for "example.com:4443". The "trust_host_root_certs" parameter instructs
ngrok to trust the root certificates on your computer when establishing TLS connections to the server. By default, ngrok
only trusts the root certificate for ngrok.com.

## 6. Connect with a client
Then, just run ngrok as usual to connect securely to your own ngrokd server!

	ngrok 80

## FAQ
#### Do I really need a wildcard certificate?
If you don't need to run https tunnels, then you don't need a wildcard cert.

#### I don't want to pay for an SSL certificate, can I use a self-signed one?
Yes, it's possible to use a self-signed certificate, but you'll need to recompile ngrok with your signing CA.

