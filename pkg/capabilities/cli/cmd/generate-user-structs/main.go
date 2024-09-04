package main

import (
	"flag"
	"go/types"
	"log"
	"maps"
	"strings"

	"golang.org/x/tools/go/packages"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd"
)

var localPrefix = flag.String("local_prefix", "github.com/smartcontractkit", "The local prefix to use when formatting go files")
var structs = flag.String("structs", "", "Comma separated list of structs to generate capability wrappers for")

func main() {
	// Define configuration for loading packages
	cfg := &packages.Config{Mode: packages.NeedTypes | packages.NeedImports}

	parsedStructs := map[string]bool{}
	for _, s := range strings.Split(*structs, ",") {
		parsedStructs[s] = true
	}

	pkgs, err := packages.Load(cfg, ".")
	if err != nil {
		log.Fatalf("Error loading package: %v\n", err)
	}

	if len(pkgs) != 1 {
		log.Fatalf("Expected exactly one package, got %d\n", len(pkgs))
	}

	generatedInfos := walk(pkgs[0], parsedStructs)

}

func walk(pkg *packages.Package, structs map[string]bool) []cmd.GeneratedInfo {
	genInfos := make([]cmd.GeneratedInfo, 0, len(structs))
	missingStructs := maps.Clone(structs)
	for _, def := range pkg.TypesInfo.Defs {
		if def == nil || !structs[def.Name()] {
			continue
		}

		missingStructs[def.Name()] = false

		if _, ok := def.Type().Underlying().(*types.Struct); ok {
			genInfos = append(genInfos, genInfoFor(pkg.Name, def))
		} else {
			log.Fatalf("Expected %s to be a struct, but it is a %T\n", def.Name(), def.Type)
		}
	}

	if len(missingStructs) > 0 {
		missing := make([]string, 0, len(missingStructs))
		for s := range missingStructs {
			missing = append(missing, s)
		}
		log.Fatalf("Could not find structs: %s\n", strings.Join(missing, ", "))
	}

	return genInfos
}

func genInfoFor(pkg string, def types.Object) cmd.GeneratedInfo {
	return cmd.GeneratedInfo{
		Package:      pkg,
		Config:       cmd.Struct{},
		Types:        nil,
		RootOutput:   "",
		ExtraImports: nil,
		ID:           nil,
		FullPackage:  "",
	}
}
