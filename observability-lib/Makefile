.PHONY: build
build:
	go build -o observability-lib .

.PHONY: run
run: build
	./observability-lib

.PHONY: lint
lint:
	golangci-lint run

.PHONY: test
test:
	go test ./...

.PHONY: regenerate
regenerate:
	go test ./dashboards/... -generate=1