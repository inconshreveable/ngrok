# ngrock 

## build ngrokd/ngrok docker images

```sh
$ docker-compose build build 
$ docker-compose up build
$ docker-compose build ngrokd ngrok
```

## push docker images to github package registry

```sh
$ ./scripts/docker/push.sh
```
