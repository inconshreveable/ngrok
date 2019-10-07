#!/bin/bash -e

source .env

docker login docker.pkg.github.com --username ${GITHUB_USERNAME} -p ${GITHUB_ACCESS_TOKEN}

docker tag ngrok_ngrokd docker.pkg.github.com/enrapt/ngrok/ngrokd:latest
docker push docker.pkg.github.com/enrapt/ngrok/ngrokd:latest

docker tag ngrok_ngrok docker.pkg.github.com/enrapt/ngrok/ngrok:latest
docker push docker.pkg.github.com/enrapt/ngrok/ngrok:latest

docker tag ngrok_subdir2subdomain docker.pkg.github.com/enrapt/ngrok/subdir2subdomain:latest
docker push docker.pkg.github.com/enrapt/ngrok/subdir2subdomain:latest
