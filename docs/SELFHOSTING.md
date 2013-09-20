# How to run your own ngrokd server

Running your own ngrok server is really easy.

## Compile it
You can compile an ngrokd server with the following command:

	make release-server

## Run the server
You'll run the server with the following command. There's more information on the


	./ngrokd -tlsKey="/path/to/tls.key" -tlsCert="/path/to/tls.crt" -domain="example.com"

### Specifying your TLS certificate and key
ngrok only makes TLS-encrypted connections. To run your own ngrokd server, you'll need your own
TLS certificates for your domain. You need a *wild card* certificate for *.example.com.
When you run ngrokd, you'll need to instruct it where to find your TLS certificate and private
key. Specify the paths with the following switches:

	-tlsKey="/path/to/tls.key" -tlsCert="/path/to/tls.crt"

### Setting the server's domain
When you run your own ngrokd server, you need to tell ngrokd the domain it's running on so that it
knows what URLs to issue to clients.

	-domain="example.com"

### Modifying your DNS
You need to use the DNS management tools given to you by your provider to create an A
record which points *.example.com to the IP address of the server running ngrokd.

## Connect with a client
In order to connect with a client, you'll need to set two options in ngrok's configuration file.
The ngrok configuration file is a simple YAML file that is read from ~/.ngrok by default. You may specify
a custom configuration file path with the -config switch. Your config file must contain the following two
options.

	server_addr: example.com:4443
	trust_host_root_certs: true

Then, just run ngrok as usual to connect securely to your own ngrokd server!

	ngrok 80
