.PHONY: build test benzhi-docker
build:
	go build -o bin/chillrack ./cmd/chillrack
test:
	go test ./... -count=1
benzhi-docker:
	sh build_benzhi_docker.sh