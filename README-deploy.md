# ngrok deploy guidline

## build ngrokd/ngrok docker images

```sh
$ docker-compose build build 
$ docker-compose up build
$ docker-compose build ngrokd ngrok
```

## build subdir2subdomain

```sh
$ ./scripts/subdir2subdomain/prepare.sh
$ docker-compose build subdir2subdomain
```

## Run in local environment

```sh
$ docker-compose up -d ngrokd ngrok subdir2subdomain stub-server dnsmasq
```

Access to https://localhost/subdir2subdomain/omura.ngrok-dev/

## push docker images to github package registry

```sh
$ ./scripts/docker/push.sh
```
