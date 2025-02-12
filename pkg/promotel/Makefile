help: ## Print this help text
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-40s\033[0m %s\n", $$1, $$2}'

.PHONY: \
	tidy \
	fmt \
	lint \
	checks \
	test

tidy: fmt ## run go mod tidy
	go mod tidy

fmt: ## run go fmt
	go fmt ./...

lint: ## run golangci-lint
	golangci-lint run ./... --config=../../.golangci.yml

checks: ## run go checks
	go mod verify
	go vet ./...
	go run honnef.co/go/tools/cmd/staticcheck@latest -checks=all,-ST1000,-U1000,-ST1003 ./...
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

test: ## run unit tests
	go test -v ./...
