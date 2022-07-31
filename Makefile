build:
	go build -o bin/main src/*.go

build-linux:
	GOOS=linux go build -o bin/main src/*.go

build-windows:
	GOOS=windows go build -o bin/main src/*.go

run:
	go run src/*.go

flush-cache:
	go run src/*.go --flush-cache

all: build