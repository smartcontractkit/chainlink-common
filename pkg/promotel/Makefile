help: ## Print this help text
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-40s\033[0m %s\n", $$1, $$2}'

.PHONY: \
	tidy \
	fmt \
	lint \
	test \
	run-example

tidy: fmt ## run go mod tidy
	go mod tidy

fmt: ## run go fmt
	go fmt ./...

lint: ## run golangci-lint
	golangci-lint run ./... --config=../../.golangci.yml

test: ## run unit tests
	go test -v ./...

build: tidy ## build the demo
	go build ./cmd/example.go

run-example: ## run the example
	go run ./cmd/example.go
