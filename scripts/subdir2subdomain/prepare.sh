#!/bin/bash -e

mkdir -p ./subdir2subdomain/bin
rm -rf ./subdir2subdomain/bin/*

htpasswd -c -b ./subdir2subdomain/bin/.htpasswd nginxuser superstrongpassword1

scp enrapt@dev.enrapt.jp:/etc/nginx/ssl/dev.enrapt.jp.2048.combined.cer ./subdir2subdomain/bin/
scp enrapt@dev.enrapt.jp:/etc/nginx/ssl/dev.enrapt.jp.2048.nopass.key ./subdir2subdomain/bin/
