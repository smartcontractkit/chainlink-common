package main

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/smartcontractkit/chainlink-common/pkg/values/installer/pkg"
)

func Generate(config *CapabilityConfig) error {
	return GenerateMany(map[string]*CapabilityConfig{".": config})
}

func GenerateMany(dirToConfig map[string]*CapabilityConfig) error {
	gen := createGenerator()

	fileToFrom := map[string]string{}
	for from, config := range dirToConfig {
		for _, file := range config.FullProtoFiles() {
			fileToFrom[file] = from
		}
		link(gen, config)
	}

	fmt.Println("Generating capabilities")
	errMap := gen.GenerateMany(fileToFrom)
	if len(errMap) > 0 {
		var errStrings []string
		for file, err := range errMap {
			if err != nil {
				errStrings = append(errStrings, fmt.Sprintf("file %s\n%v\n", file, err))
			}
		}

		return errors.New(strings.Join(errStrings, ""))
	}

	fmt.Println("Moving generated files to correct locations")
	for from, config := range dirToConfig {
		for i, file := range config.FullProtoFiles() {
			file = strings.Replace(file, ".proto", ".pb.go", 1)
			to := strings.Replace(config.Files[i], ".proto", ".pb.go", 1)
			if err := os.Rename(path.Join(from, file), path.Join(from, to)); err != nil {
				return fmt.Errorf("failed to move generated file %s: %w", file, err)
			}
		}

		if err := os.RemoveAll(path.Join(from, "capabilities")); err != nil {
			return fmt.Errorf("failed to remove capabilities directory %w", err)
		}
	}

	return nil
}

func createGenerator() *pkg.ProtocGen {
	gen := &pkg.ProtocGen{}
	gen.LinkPackage(pkg.Packages{Go: "github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb", Proto: "tools/generator/v1alpha/cre_metadata.proto"})
	gen.LinkPackage(pkg.Packages{Go: "github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb", Proto: "sdk/v1alpha/sdk.proto"})
	return gen
}

func link(gen *pkg.ProtocGen, config *CapabilityConfig) {
	for _, file := range config.FullProtoFiles() {
		gen.LinkPackage(pkg.Packages{Go: config.FullGoPackageName(), Proto: file})
	}
}
