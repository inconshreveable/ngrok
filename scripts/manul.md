GOOS=linux GOARCH=amd64 make release-server release-client
GOOS=darwin GOARCH=amd64 make release-server release-client

scp -P 28286 -r root@io3oo.com:/root/ngrok/bin/darwin_amd64/ .

nohup sh /root/ngrok/serv_run.sh &>/dev/null

scp -P 28286 -r root@io3oo.com:~/test/sample/ ~/Documents/devel/ngrok/