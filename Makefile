DOCKER_TAG=$(shell git rev-parse HEAD)
DOCKER_IMAGE=parkr/mypod:$(DOCKER_TAG)

all: build test

build:
	go install ./...

test:
	go test ./...

docker-buildx-info:
	docker buildx version
	docker buildx ls

docker-buildx-create: docker-buildx-info
	docker buildx create --platform linux/amd64,linux/arm64 --use

docker-build: docker-buildx-info
	docker buildx build -t $(DOCKER_IMAGE) --platform linux/arm64,linux/amd64 .

docker-release: docker-build
	docker buildx build -t $(DOCKER_IMAGE) --platform linux/arm64,linux/amd64 --push .

docker-server: docker-build
	docker run \
		-it \
		--user $(shell id -u):$(shell id -g) \
		-v $(shell pwd)/example:/storage \
		-p 5312:5312 \
		$(DOCKER_IMAGE) \
		-http=:5312 \
		-debug=true

docker-debug: docker-build
	docker run \
		-it \
		--net=host \
		--user $(shell id -u):$(shell id -g) \
		-v $(shell pwd)/example:/storage \
		-p 5312:5312 \
		--entrypoint=/bin/sh \
		$(DOCKER_IMAGE)
