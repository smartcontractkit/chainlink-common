package pkg

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
)

type Generator struct {
	GeneratorHelper
}

func (g *Generator) Generate(config *CapabilityConfig) error {
	return g.GenerateMany(map[string]*CapabilityConfig{".": config})
}

func (g *Generator) GenerateMany(dirToConfig map[string]*CapabilityConfig) error {
	if err := installProtocGenToDir(g.PluginName(), g.HelperName()); err != nil {
		return err
	}

	gen := g.createGenerator()

	fileToFrom := map[string]string{}
	for from, config := range dirToConfig {
		for _, file := range config.FullProtoFiles() {
			fileToFrom[file] = from
		}
		g.link(gen, config)
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

func (g *Generator) createGenerator() *ProtocGen {
	// protoc plugin names are in the form protoc-gen-<plugin-name>
	pluginParts := strings.Split(g.PluginName(), "-")
	pluginShortName := pluginParts[len(pluginParts)-1]

	gen := &ProtocGen{Plugins: []Plugin{{Name: pluginShortName, Path: ".tools"}}}
	gen.LinkPackage(Packages{Go: g.SdkPgk(), Proto: "tools/generator/v1alpha/cre_metadata.proto"})
	gen.LinkPackage(Packages{Go: g.SdkPgk(), Proto: "sdk/v1alpha/sdk.proto"})
	return gen
}

func (g *Generator) link(gen *ProtocGen, config *CapabilityConfig) {
	for _, file := range config.FullProtoFiles() {
		goPkg := g.FullGoPackageName(config)
		gen.LinkPackage(Packages{Go: goPkg, Proto: file})
	}
}
