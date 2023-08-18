test:
	@echo "Running tests"
	@go test -timeout 1s ./...
.PHONY: test

vet:
	@echo "Vetting package"
	@go vet ./...
.PHONY: vet

race:
	@echo "Testing for race conditions"
	@go test -race -timeout 1s ./...
.PHONY: race
