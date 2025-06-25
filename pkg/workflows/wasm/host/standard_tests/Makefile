# Declare all matching test.wasm targets as phony
.PHONY: FORCE
FORCE:

# Pattern rule to build test.wasm in any subdirectory
%/test.wasm: FORCE
	cd $* && GOOS=wasip1 GOARCH=wasm CGO_ENABLED=0 go build -o test.wasm .
