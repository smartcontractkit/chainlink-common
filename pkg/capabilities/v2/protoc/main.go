package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg"
)

func main() {
	args := map[string]string{}
	protogen.Options{ParamFunc: func(name, value string) error {
		args[strings.ToLower(name)] = value
		return nil
	}}.Run(func(plugin *protogen.Plugin) error {
		goLang := pkg.ServerLangaugeGo
		serverLanguage, err := parseArg(args, "server_language", func(value string) (pkg.ServerLanguage, error) {
			serverLanguage := pkg.ServerLanguage(strings.ToLower(value))
			return serverLanguage, serverLanguage.Validate()
		}, &goLang)
		if err != nil {
			return err
		}

		for _, file := range plugin.Files {
			if !file.Generate {
				continue
			}
			if err := pkg.GenerateServer(plugin, file, serverLanguage); err != nil {
				log.Printf("failed to generate for %s: %v", file.Desc.Path(), err)
				os.Exit(1)
			}
		}
		return nil
	})
}

func parseArg[T any](args map[string]string, name string, parse func(string) (T, error), defaultValue *T) (T, error) {
	arg, ok := args[name]
	if !ok {
		if defaultValue != nil {
			return *defaultValue, nil
		}

		var t T
		return t, fmt.Errorf("missing required argument %v", name)
	}

	return parse(arg)
}
