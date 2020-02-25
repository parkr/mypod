all: build test

build:
	go install ./...

test:
	go test ./...

docker-build:
	docker build -t mypod .

docker-server: docker-build
	docker run \
		-it \
		--user $(shell id -u):$(shell id -g) \
		-v $(shell pwd)/example:/storage \
		-p 5312:5312 \
		mypod \
		-http=:5312 \
		-debug=true


