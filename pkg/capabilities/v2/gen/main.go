package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/smartcontractkit/chainlink-common/pkg/values/installer/pkg"
)

func main() {
	gen := &pkg.ProtocGen{Plugins: []pkg.Plugin{{Name: "cre", Path: "protoc"}}}
	// Note the second directory is for chain-capabilities/evm.
	// Once the dependencies are inverted, it will be removed.
	gen.AddSourceDirectories(".", "../../../..")
	gen.LinkPackage(pkg.Packages{Go: "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/pb", Proto: "tools/generator/v1alpha/cre_metadata.proto"})
	gen.LinkPackage(pkg.Packages{Go: "github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb", Proto: "sdk/v1alpha/sdk.proto"})

	errors := map[string]error{}

	err := filepath.WalkDir(".", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Proto directory doesn't use itself as a plugin
		if d.IsDir() && filepath.Clean(path) == filepath.Clean("protoc/pkg/pb") {
			return filepath.SkipDir
		}

		if !d.IsDir() && strings.HasSuffix(d.Name(), ".proto") {
			if genErr := gen.Generate(d.Name(), filepath.Dir(path)); genErr != nil {
				errors[path] = genErr
			}
		}
		return nil
	})
	if err != nil {
		panic(fmt.Errorf("error walking directory: %w", err))
	}

	if len(errors) > 0 {
		fmt.Println("Errors encountered during generation:")
		for file, err := range errors {
			fmt.Printf("File: %s, Error: %v\n", file, err)
		}
		os.Exit(1)
	}
}
