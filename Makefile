DOCKER_TAG=$(shell git rev-parse HEAD)
DOCKER_IMAGE=parkr/mypod:$(DOCKER_TAG)

all: build test

build:
	go install ./...

test:
	go test ./...

docker-build:
	docker build -t $(DOCKER_IMAGE) .

docker-release: docker-build
	docker push $(DOCKER_IMAGE)

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
