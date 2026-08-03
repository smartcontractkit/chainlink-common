package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/smartcontractkit/chainlink-protos/cre/go/installer/pkg"
)

func parseImportPaths(s string) (proto, goPkg string) {
	proto, goPkg, _ = strings.Cut(s, "=")
	return
}

func generateFlat(gen *pkg.ProtocGen, proto, outDir string) error {
	if err := gen.GenerateFile(proto, "."); err != nil {
		return err
	}

	pb := strings.TrimSuffix(proto, ".proto") + ".pb.go"
	dst := filepath.Join(outDir, filepath.Base(pb))

	if err := os.Rename(pb, dst); err != nil {
		return err
	}

	topDir := strings.Split(proto, string(os.PathSeparator))[0]
	return os.RemoveAll(topDir)
}

func main() {
	capDir := flag.String("pkg", "", "the go package to generate in")
	file := flag.String("file", "", "the proto file to generate from")
	defaultPathToV2 := filepath.Join("..", "..")
	pathToV2 := flag.String("pathToV2", defaultPathToV2, "path to the v2 directory")
	importedProto := flag.String("import", "", "path to proto to be imported by the main proto")
	withMonitoring := flag.Bool("with-monitoring", false, "generate v2 monitoring lifecycle hooks in the server dispatch layer")
	flag.Parse()

	if *withMonitoring {
		os.Setenv("CL_PROTOC_WITH_MONITORING", "true")
	}

	gen := &pkg.ProtocGen{Plugins: []pkg.Plugin{pkg.GoPlugin, {Name: "cre", Path: filepath.Join(*pathToV2, "protoc")}}}
	gen.LinkPackage(pkg.Packages{Go: *capDir, Proto: *file})

	if *importedProto != "" {
		proto, goPkg := parseImportPaths(*importedProto)
		// link imported proto to the main file
		gen.LinkPackage(pkg.Packages{Proto: proto, Go: goPkg})

		// create a new dir for the imported proto so that it can be imported from other places
		relPkg := strings.TrimPrefix(goPkg, *capDir+"/")
		if err := os.MkdirAll(relPkg, 0o755); err != nil {
			log.Fatal("Error creating package dir:", err)
		}

		subGen := &pkg.ProtocGen{Plugins: []pkg.Plugin{pkg.GoPlugin}}
		subGen.LinkPackage(pkg.Packages{Proto: proto, Go: goPkg})
		if err := generateFlat(subGen, proto, relPkg); err != nil {
			log.Fatal("Error generating linked file:", err)
		}
	}

	if err := generateFlat(gen, *file, "."); err != nil {
		log.Fatal("Error generating main file:", err)
	}
}
