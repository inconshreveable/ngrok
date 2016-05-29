#!/bin/bash

basepath=$(cd `dirname $0`; pwd)
cd $basepath

rm -rf ngrok
mkdir ngrok
scp -P 28286 -r root@io3oo.com:~/ngrok/bin/ ngrok/

chmod -R +x ngrok/*