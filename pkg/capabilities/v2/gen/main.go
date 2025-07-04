package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/smartcontractkit/chainlink-common/pkg/values/installer/pkg"
)

func main() {
	gen := &pkg.ProtocGen{}
	gen.LinkPackage(pkg.Packages{Go: "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/pb", Proto: "tools/generator/v1alpha/cre_metadata.proto"})
	gen.LinkPackage(pkg.Packages{Go: "github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb", Proto: "sdk/v1alpha/sdk.proto"})

	capDir := flag.String("pkg", "", "the go package to generate in")
	file := flag.String("file", "", "the go file to generate from")
	defaultPathToV2 := filepath.Join("..", "..")
	pathToV2 := flag.String("pathToV2", defaultPathToV2, "How to get to the ")
	flag.Parse()

	gen.Plugins = []pkg.Plugin{{Name: "cre", Path: filepath.Join(*pathToV2, "protoc")}}

	gen.LinkPackage(pkg.Packages{Go: *capDir, Proto: *file})

	if err := gen.GenerateFile(*file, "."); err != nil {
		log.Fatal("Error generating file:", err)
	}

	pb := strings.Replace(*file, ".proto", ".pb.go", 1)
	if err := os.Rename(pb, filepath.Base(pb)); err != nil {
		log.Fatal("Error renaming file:", err)
	}

	pathParts := strings.Split(*file, string(os.PathSeparator))
	if err := os.RemoveAll(pathParts[0]); err != nil {
		log.Fatal("Error removing path:", err)
	}
}
