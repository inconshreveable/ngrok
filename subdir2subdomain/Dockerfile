FROM nginx:latest

RUN apt-get update
RUN apt-get install -y dnsutils

ADD ./nginx.conf /etc/nginx/
# ADD ./bin/.htpasswd /etc/nginx/
ADD ./bin/dev.enrapt.jp.2048.combined.cer /etc/nginx/
ADD ./bin/dev.enrapt.jp.2048.nopass.key /etc/nginx/
