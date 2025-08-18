package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/smartcontractkit/chainlink-protos/cre/go/installer/pkg"
)

func main() {
	gen := &pkg.ProtocGen{Plugins: []pkg.Plugin{pkg.GoPlugin}}
	capDir := flag.String("pkg", "", "the go package to generate in")
	file := flag.String("file", "", "the go file to generate from")
	defaultPathToV2 := filepath.Join("..", "..")
	pathToV2 := flag.String("pathToV2", defaultPathToV2, "How to get to the ")
	flag.Parse()

	gen.Plugins = []pkg.Plugin{pkg.GoPlugin, {Name: "cre", Path: filepath.Join(*pathToV2, "protoc")}}

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
