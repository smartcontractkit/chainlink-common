.PHONY: gomodtidy
gomodtidy:
	go mod tidy
	cd ./ops && go mod tidy

.PHONY: godoc
godoc:
	go install golang.org/x/tools/cmd/godoc@latest
	# http://localhost:6060/pkg/github.com/smartcontractkit/chainlink-relay/
	godoc -http=:6060

.PHONY: mockery
mockery: $(mockery) ## Install mockery.
	go install github.com/vektra/mockery/v2@v2.28.1