#!/bin/bash -e

if [ -e .env ]; then
  source .env
fi

docker login docker.pkg.github.com --username ${GITHUB_USERNAME} -p ${GITHUB_ACCESS_TOKEN}

containers=("ngrokd" "ngrok" "subdir2subdomain")
for container in ${containers[@]}
do
  echo "push ${container} to github package registry"
  docker tag ngrok_${container} docker.pkg.github.com/enrapt/ngrok/${container}:latest
  docker push docker.pkg.github.com/enrapt/ngrok/${container}:latest
done
