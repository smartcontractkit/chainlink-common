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
	go test -v ./...

.PHONY: regenerate
regenerate:
	go test -v ./dashboards/... -generate=1