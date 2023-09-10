test:
	@echo "Running tests"
	@go test -timeout 30s ./...
.PHONY: test

vet:
	@echo "Vetting package"
	@go vet ./...
.PHONY: vet

race:
	@echo "Testing for race conditions"
	@go test -race -timeout 30s ./...
.PHONY: race

lint:
	@echo "Installing linter"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.54.2
	
	@echo "Linting code"
	@golangci-lint run ./...
.PHONY: lint

fmt:
	@echo "Checking code format"
	@go fmt ./...