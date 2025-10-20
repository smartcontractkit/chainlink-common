package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/generator"
)

func main() {
	var (
		pkgPath      = flag.String("pkg", "", "Go package path containing the interface (module-qualified)")
		ifaceName    = flag.String("interface", "", "Interface name (e.g. ExampleService)")
		serviceName  = flag.String("service", "Example", "Proto service name (e.g. Example)")
		protoOut     = flag.String("proto-out", "", ".proto output path")
		goOutDir     = flag.String("go-out", "", "Directory for generated Go wrappers")
		goPkg        = flag.String("go-pkg", "", "Go package import path for generated wrappers")
		protoPkg     = flag.String("proto-pkg", "loop.solana", "Proto package name")
		goPackageOpt = flag.String("proto-go-package", "github.com/smartcontractkit/chainlink-common/pkg/loop/solana", "option go_package value for the .proto")
		configPath   = flag.String("config", "", "YAML config for externals/aliases/enums")
	)
	flag.Parse()

	if *pkgPath == "" || *ifaceName == "" || *protoOut == "" || *goOutDir == "" || *goPkg == "" {
		log.Fatalf("missing required flags: --pkg, --interface, --proto-out, --go-out, --go-pkg")
	}

	cfg, err := generator.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	svc, err := generator.ParseInterface(*pkgPath, *ifaceName, cfg)
	if err != nil {
		log.Fatalf("parse: %v", err)
	}
	svc.ServiceName = *serviceName
	svc.ProtoPkg = *protoPkg
	svc.OptionGoPackage = *goPackageOpt
	svc.WrapGoPackage = *goPkg

	if err := os.MkdirAll(*goOutDir, 0o755); err != nil {
		log.Fatalf("mkdir go-out: %v", err)
	}

	if err := generator.RenderAll(*protoOut, *goOutDir, svc); err != nil {
		log.Fatalf("render: %v", err)
	}

	fmt.Printf("Generated:\n %s\n %s/server.go\n %s/client.go\n", *protoOut, *goOutDir, *goOutDir)
}
