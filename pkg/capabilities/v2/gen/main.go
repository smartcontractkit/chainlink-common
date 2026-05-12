package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/smartcontractkit/chainlink-protos/cre/go/installer/pkg"
)

type links []string

func (l *links) String() string { return strings.Join(*l, ",") }
func (l *links) Set(v string) error {
	*l = append(*l, v)
	return nil
}

func parseLink(s string) (proto, goPkg string) {
	proto, goPkg, _ = strings.Cut(s, "=")
	return
}

func main() {
	gen := &pkg.ProtocGen{Plugins: []pkg.Plugin{pkg.GoPlugin}}
	capDir := flag.String("pkg", "", "the go package to generate in")
	file := flag.String("file", "", "the go file to generate from")
	defaultPathToV2 := filepath.Join("..", "..")
	pathToV2 := flag.String("pathToV2", defaultPathToV2, "path to the v2 directory")
	var extraLinks links
	var genLinks links
	flag.Var(&extraLinks, "link", "proto=gopkg mapping added to the main generation (no separate generation)")
	flag.Var(&genLinks, "link-and-generate", "proto=gopkg; add mapping to main generation and also generate the proto into its sub-package dir")
	flag.Parse()

	gen.Plugins = []pkg.Plugin{pkg.GoPlugin, {Name: "cre", Path: filepath.Join(*pathToV2, "protoc")}}

	gen.LinkPackage(pkg.Packages{Go: *capDir, Proto: *file})
	for _, l := range extraLinks {
		proto, goPkg := parseLink(l)
		gen.LinkPackage(pkg.Packages{Proto: proto, Go: goPkg})
	}
	for _, l := range genLinks {
		proto, goPkg := parseLink(l)
		gen.LinkPackage(pkg.Packages{Proto: proto, Go: goPkg})
	}

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

	for _, l := range genLinks {
		proto, goPkg := parseLink(l)
		subGen := &pkg.ProtocGen{Plugins: []pkg.Plugin{pkg.GoPlugin}}
		subGen.LinkPackage(pkg.Packages{Proto: proto, Go: goPkg})
		if err := subGen.GenerateFile(proto, "."); err != nil {
			log.Fatal("Error generating linked file:", err)
		}
		relPkg := strings.TrimPrefix(goPkg, *capDir+"/")
		subPb := strings.Replace(proto, ".proto", ".pb.go", 1)
		if err := os.MkdirAll(relPkg, 0o755); err != nil {
			log.Fatal("Error creating sub-package dir:", err)
		}
		if err := os.Rename(subPb, filepath.Join(relPkg, filepath.Base(subPb))); err != nil {
			log.Fatal("Error renaming linked file:", err)
		}
		subPathParts := strings.Split(proto, string(os.PathSeparator))
		if err := os.RemoveAll(subPathParts[0]); err != nil {
			log.Fatal("Error removing path:", err)
		}
	}
}
