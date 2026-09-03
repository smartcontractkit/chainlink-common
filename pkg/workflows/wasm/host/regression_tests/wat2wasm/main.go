// Command wat2wasm compiles a WebAssembly text-format file to a binary wasm
// module using the wasmtime engine that the host already depends on, so
// regression tests can hand-write modules that no Go compiler would emit
// (e.g. a module that omits its linear memory export) without requiring an
// external toolchain to be installed.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/bytecodealliance/wasmtime-go/v48"
)

func main() {
	in := flag.String("in", "", "path to the .wat source file")
	out := flag.String("out", "", "path to the .wasm file to write")
	flag.Parse()

	if err := run(*in, *out); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(in, out string) error {
	if in == "" || out == "" {
		return fmt.Errorf("both -in and -out are required")
	}

	wat, err := os.ReadFile(in)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", in, err)
	}

	binary, err := wasmtime.Wat2Wasm(string(wat))
	if err != nil {
		return fmt.Errorf("failed to compile %s: %w", in, err)
	}

	if err := os.WriteFile(out, binary, 0600); err != nil {
		return fmt.Errorf("failed to write %s: %w", out, err)
	}

	return nil
}
