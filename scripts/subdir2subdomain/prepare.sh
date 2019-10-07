#!/bin/bash -e

if [ -e .env ]; then
  source .env
fi

mkdir -p ./subdir2subdomain/bin
rm -rf ./subdir2subdomain/bin/*

htpasswd -c -b ./subdir2subdomain/bin/.htpasswd nginxuser superstrongpassword1

scp -o StrictHostKeyChecking=no ${SSL_CERTIFICATE_PATH}/dev.enrapt.jp.2048.combined.cer ./subdir2subdomain/bin/
scp -o StrictHostKeyChecking=no ${SSL_CERTIFICATE_PATH}/dev.enrapt.jp.2048.nopass.key ./subdir2subdomain/bin/
