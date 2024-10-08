build:
	@go built -o bin/BankApi

run: build
	@./bin/BankApi

test:
	@go test -v ./...