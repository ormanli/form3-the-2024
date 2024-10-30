.PHONY: lint
lint:
	golangci-lint run

.PHONY: generate-mocks
generate-mocks:
	mockery

.PHONY: test
test:
	go clean -testcache
	go test ./... -race