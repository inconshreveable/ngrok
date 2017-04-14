if [ $# == 1 ]; then
  SUBJ="/CN="$1
  rm -f rootCA.*
  rm -f device.*
  openssl genrsa -out rootCA.key 2048
  openssl req -x509 -new -nodes -key rootCA.key -subj $SUBJ -days 5000 -out rootCA.pem
  openssl genrsa -out device.key 2048
  openssl req -new -key device.key -subj $SUBJ -out device.csr
  openssl x509 -req -in device.csr -CA rootCA.pem -CAkey rootCA.key -CAcreateserial -out device.crt -days 5000
  cp -f rootCA.pem assets/client/tls/ngrokroot.crt
  cp -f device.crt assets/server/tls/snakeoil.crt
  cp -f device.key assets/server/tls/snakeoil.key
else
  echo "1 parameter is needed. It must be a string of domain. "
fi
