build:
	go build -o agent cmd/agent/main.go
	go build -o server cmd/server/main.go
	go build -o server cmd/befwctl/*.go

test:
	go test -v befw/internal/...
	
statictest:
	go vet -vettool=$(shell which statictest) ./...