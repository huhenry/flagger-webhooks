TAG?=latest
VERSION?=$(shell grep 'const VERSION' main.go | awk '{ print $$4 }' | tr -d '"' | head -n1)
NAME:=grife
DOCKER_REPOSITORY:=huhenry
DOCKER_IMAGE_NAME:=$(DOCKER_REPOSITORY)/$(NAME)

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/grife main.go

test:
	go test -v -race ./...

fmt:
	gofmt -l * | grep ".*\.go"; if [ "$$?" = "0" ]; then exit 1; fi;

run:
	go run main.go 

image: build
	docker build -t $(DOCKER_IMAGE_NAME):v$(VERSION) .

push-container: build-container
	docker push $(DOCKER_IMAGE_NAME):v$(VERSION)