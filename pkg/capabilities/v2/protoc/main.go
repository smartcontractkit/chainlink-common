package main

import (
	"log"
	"os"

	"google.golang.org/protobuf/compiler/protogen"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg"
)

func main() {
	protogen.Options{}.Run(func(plugin *protogen.Plugin) error {
		for _, file := range plugin.Files {
			if !file.Generate {
				continue
			}
			if err := pkg.GenerateClient(plugin, file); err != nil {
				log.Printf("failed to generate for %s: %v", file.Desc.Path(), err)
				os.Exit(1)
			}
		}
		return nil
	})
}
