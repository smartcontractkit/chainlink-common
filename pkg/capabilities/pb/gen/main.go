package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/smartcontractkit/chainlink-protos/cre/go/installer/pkg"
)

func main() {
	gen := &pkg.ProtocGen{Plugins: []pkg.Plugin{pkg.GoPlugin, {Name: "go-grpc"}}}
	gen.AddSourceDirectories(".", "../../")
	entries, err := os.ReadDir(".")
	if err != nil {
		panic(fmt.Sprintf("failed to read directory: %v\n", err))
	}

	errors := map[string]error{}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".proto") {
			if err = gen.GenerateFile(entry.Name(), ""); err != nil {
				errors[entry.Name()] = err
			}
		}
	}

	if len(errors) > 0 {
		fmt.Println("Errors encountered during generation:")
		for file, err := range errors {
			fmt.Printf("File: %s, Error: %v\n", file, err)
		}
		os.Exit(1)
	}
}
