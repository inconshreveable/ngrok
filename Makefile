SHELL = /usr/bin/env bash

DOMAIN = ngrok.bnotions.com
EMAIL = softwareadmin@bnotions.com
CLIENT_NAME = intersect
PROJECT_NAME = ngrok
DOCKER_NAME = $(CLIENT_NAME)/$(PROJECT_NAME)

ifndef GIT_COMMIT
	GIT_COMMIT = $(shell git rev-parse HEAD)
endif

ifndef BRANCH_NAME
	BRANCH_NAME = $(shell git rev-parse --abbrev-ref HEAD)
endif

ifndef TAG
	TAG = $(subst /,-,$(BRANCH_NAME))
endif

ifndef REPOSITORY
	ifdef AWS_ECR_REGISTRY_ID
		REPOSITORY = $(AWS_ECR_REGISTRY_ID).dkr.ecr.$(AWS_ECR_REGION).amazonaws.com
	else
		REPOSITORY = localhost
	endif
endif

ifndef STACK_NAME
	STACK_NAME = $(CLIENT_NAME)-$(PROJECT_NAME)-$(TAG)
endif

ifndef STACK_FILE
	STACK_FILE = deploy-stack.yml
endif

ifndef SONAR_HOST
	SONAR_HOST = https://sonar.bnotions.com
endif

ifndef SONAR_LOGIN
	SONAR_LOGIN = nil
endif

.PHONY: default server client deps fmt clean all release-all assets client-assets server-assets contributors

export GOPATH:=$(shell pwd)
export CLIENT_NAME
export PROJECT_NAME
export DOCKER_NAME
export GIT_COMMIT
export TAG
export REPOSITORY
export BUILD_NUMBER
export STACK_FILE
export STACK_NAME
export SONAR_HOST
export SONAR_LOGIN

BUILDTAGS=debug
default: all

deps: assets
	go get -tags '$(BUILDTAGS)' -d -v ngrok/...

server: deps
	go install -tags '$(BUILDTAGS)' ngrok/main/ngrokd

fmt:
	go fmt ngrok/...

client: deps
	go install -tags '$(BUILDTAGS)' ngrok/main/ngrok

assets: client-assets server-assets

bin/go-bindata:
	GOOS="" GOARCH="" go get github.com/jteeuwen/go-bindata/go-bindata

client-assets: bin/go-bindata
	bin/go-bindata -nomemcopy -pkg=assets -tags=$(BUILDTAGS) \
		-debug=$(if $(findstring debug,$(BUILDTAGS)),true,false) \
		-o=src/ngrok/client/assets/assets_$(BUILDTAGS).go \
		assets/client/...

server-assets: bin/go-bindata
	bin/go-bindata -nomemcopy -pkg=assets -tags=$(BUILDTAGS) \
		-debug=$(if $(findstring debug,$(BUILDTAGS)),true,false) \
		-o=src/ngrok/server/assets/assets_$(BUILDTAGS).go \
		assets/server/...

release-client: BUILDTAGS=release
release-client: client

release-server: BUILDTAGS=release
release-server: server

release-all: fmt release-client release-server

all: fmt client server

letsencrypt:
	docker run -it --rm \
		-v "$$PWD/letsencrypt:/etc/letsencrypt" \
		certbot/certbot certonly \
			--manual \
			-d *.$(DOMAIN) \
			-d $(DOMAIN) \
			--preferred-challenges dns \
			--email $(EMAIL) \
			--agree-tos

letsencrypt-root:
	wget -qO assets/client/tls/ngrokroot.crt https://letsencrypt.org/certs/letsencryptauthorityx3.pem

docker-release:
	docker run --rm \
		--userns host \
		--user $$(id -u $$USER) \
		--volume "$$PWD":/usr/src/app \
		--workdir /usr/src/app \
		-e HOME=/usr/src/app \
		108026493146.dkr.ecr.us-east-1.amazonaws.com/intersect/golang:latest \
			make release-server


docker-image:
	docker build \
		--build-arg BUILD_NUMBER \
		--build-arg GIT_COMMIT \
		--build-arg DOMAIN=$(DOMAIN) \
		--tag "$$DOCKER_NAME:$$TAG" .

docker-tag:
	docker tag "$$DOCKER_NAME:$$TAG" "$$REPOSITORY/$$DOCKER_NAME:$$GIT_COMMIT"
	docker tag "$$DOCKER_NAME:$$TAG" "$$REPOSITORY/$$DOCKER_NAME:$$TAG"
	@echo ---
	@echo Tagged: "$$REPOSITORY/$$DOCKER_NAME:$$GIT_COMMIT"
	@echo Tagged: "$$REPOSITORY/$$DOCKER_NAME:$$TAG"

docker-push:
	docker push "$$REPOSITORY/$$DOCKER_NAME:$$GIT_COMMIT"
	docker push "$$REPOSITORY/$$DOCKER_NAME:$$TAG"
	@echo ---
	@echo Pushed: "$$REPOSITORY/$$DOCKER_NAME:$$GIT_COMMIT"
	@echo Pushed: "$$REPOSITORY/$$DOCKER_NAME:$$TAG"

docker-sonar:
	docker pull 108026493146.dkr.ecr.us-east-1.amazonaws.com/intersect/sonar-scanner:latest
	docker run --rm \
		--userns host \
		--user $$(id -u $$USER) \
		--volume "$$PWD":/usr/src/app \
		--workdir /usr/src/app \
		108026493146.dkr.ecr.us-east-1.amazonaws.com/intersect/sonar-scanner:latest \
			-D sonar.host.url="$$SONAR_HOST" \
			-D sonar.login="$$SONAR_LOGIN" \
			-D sonar.projectKey="$$CLIENT_NAME:$$PROJECT_NAME" \
			-D sonar.projectName="$$PROJECT_NAME" \
			-D sonar.projectVersion=1.0."$$BUILD_NUMBER" \
			-D sonar.branch.name="$$BRANCH_NAME" \
			-D sonar.sources=.

ecr-login:
	eval $$(aws ecr get-login | sed 's/-e none//')

ecr-repository:
	if ! aws ecr describe-repositories --repository-names "$$DOCKER_NAME"; then \
		aws ecr create-repository --repository-name "$$DOCKER_NAME"; \
	fi

clean:
	go clean -i -r ngrok/...
	rm -rf src/ngrok/client/assets/ src/ngrok/server/assets/

contributors:
	echo "Contributors to ngrok, both large and small:\n" > CONTRIBUTORS
	git log --raw | grep "^Author: " | sort | uniq | cut -d ' ' -f2- | sed 's/^/- /' | cut -d '<' -f1 >> CONTRIBUTORS
