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

.PHONY: update
update:
	go test ./dashboards/... -update=1
