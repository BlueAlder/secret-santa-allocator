BINARY_NAME=secret-santa-allocator

build:
	go build -o bin/

clean:
	go clean
	rm bin/${BINARY_NAME}