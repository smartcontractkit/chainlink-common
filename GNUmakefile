.PHONY: generate
generate: mockery 
	go generate -x ./...

.PHONY: mockery
mockery: $(mockery) ## Install mockery.
	go install github.com/vektra/mockery/v2@v2.38.0
