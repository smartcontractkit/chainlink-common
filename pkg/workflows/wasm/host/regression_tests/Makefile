# Declare all matching test.wasm targets as phony
.PHONY: FORCE
FORCE:

# Pattern rule to build test.wasm in any subdirectory.
# A subdirectory holding a main.wat is compiled from WebAssembly text by the
# local wat2wasm helper (Go + wasmtime, no external toolchain needed); any
# other subdirectory is a Go package built for wasip1.
%/test.wasm: FORCE
	@if [ -f "$*/main.wat" ]; then \
		go run ./wat2wasm -in "$*/main.wat" -out "$@"; \
	else \
		cd $* && GOOS=wasip1 GOARCH=wasm CGO_ENABLED=0 go build -o test.wasm .; \
	fi
