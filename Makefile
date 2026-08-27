.PHONY: build test vet docker-up

build:
	go build -o bin/udb-mysql-mcp-server .

test:
	go test -race -count=1 ./...

vet:
	go vet ./...

docker-up:
	docker build -t udb-mysql-mcp-server:latest .
	docker run --rm -it --env-file .env -p 127.0.0.1:9000:9000 udb-mysql-mcp-server:latest
