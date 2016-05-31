#!/bin/bash

# 编译安装ngrok
#

rm -rf ngrok &>/dev/null
echo "Download source .... from : https://github.com/***/ngrok.git"
#git clone https://github.com/inconshreveable/ngrok.git ngrok
git clone https://github.com/tutumcloud/ngrok.git ngrok
echo -e "\n\n...... Downloading Completely!"
cd ngrok

echo "Clean build cached files..."
make clean &>/dev/null
rm -rf bin/

echo "clean CA Files..."
rm -rf *.crt *.srl *.csr *.key *.pem
rm -rf assets/client/tls/*.crt assets/client/tls/*.key

echo "Remake CA Certificate Files"
NGROK_BASE_DOMAIN="io3oo.com"

openssl genrsa -out base.key 2048
openssl req -new -x509 -nodes -key base.key -days 10000 -subj "/CN=io3oo.com" -out base.pem
openssl genrsa -out server.key 2048
openssl req -new -key server.key -subj "/CN=io3oo.com" -out server.csr
openssl x509 -req -in server.csr -CA base.pem -CAkey base.key -CAcreateserial -days 10000 -out server.crt

cp base.pem assets/client/tls/ngrokroot.crt
cp server.crt assets/server/tls/snakeoil.crt
cp server.key assets/server/tls/snakeoil.key

echo "..................new CA Files...................."
ls -l assets/client/tls/ | grep -v "^total" | grep -v "^$"
ls -l assets/server/tls/ | grep -v "^total" | grep -v "^$"

#echo "......................................"
#echo "make mac os x & amd64 release..."
#GOOS=darwin GOARCH=amd64 make release-server release-client

echo "......................................"
echo "make linux & amd64 release..."
#make clean &>/dev/null
GOOS=linux GOARCH=amd64 make release-server release-client

chmod +x bin/ngrokd
chmod +x bin/ngrok
mv server.* bin/
mv base.* bin/

echo "#!/bin/bash" > bin/server.sh
echo "basepath=\$(cd \`dirname \$0\`; pwd)" >> bin/server.sh
echo "cd \$basepath" >> bin/server.sh
echo "./ngrokd -tlsKey=server.key -tlsCrt=server.crt -domain=\"io3oo.com\" -httpAddr=\":8081\" -httpsAddr=\":8082\"" >> bin/server.sh

echo "server_addr: io3oo.com:4443" > bin/ngrok.cfg
echo "trust_host_root_certs: false" >> bin/ngrok.cfg

echo "#!/bin/bash" > bin/client.sh
echo "basepath=\$(cd \`dirname \$0\`; pwd)" >> bin/client.sh
echo "cd \$basepath" >> bin/client.sh
echo "./ngrok -subdomain pub -config=ngrok.cfg -proto=http 127.0.0.1:80" >> bin/client.sh

echo "................Build Result...................."
tree bin/