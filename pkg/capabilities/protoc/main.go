package main

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/protoc/pkg"
	wasmpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/pb"
)

func main() {
	args := map[string]string{}
	protogen.Options{ParamFunc: func(name, value string) error {
		args[strings.ToLower(name)] = value
		return nil
	}}.Run(func(plugin *protogen.Plugin) error {
		mode, err := parseArg(args, "mode", parseMode, nil)
		if err != nil {
			return err
		}

		capabilityId, ok := args["id"]
		if !ok {
			return fmt.Errorf("missing required argument capability_id")
		}

		f := false
		trigger, err := parseArg(args, "trigger", func(value string) (bool, error) {
			switch strings.ToLower(value) {
			case "true", "t":
				return true, nil
			case "false", "f":
				return false, nil
			default:
				return false, fmt.Errorf("invalid value for trigger, expected true or false")
			}
		}, &f)

		if err != nil {
			return err
		}

		goLang := pkg.ServerLangaugeGo
		serverLanguage, err := parseArg(args, "server_language", func(value string) (pkg.ServerLanguage, error) {
			serverLanguage := pkg.ServerLanguage(strings.ToLower(value))
			return serverLanguage, serverLanguage.Validate()
		}, &goLang)
		if err != nil {
			return err
		}

		for _, file := range plugin.Files {
			if err = pkg.GenerateClient(plugin, mode, trigger, capabilityId, file); err != nil {
				return err
			}

			if err = pkg.GenerateServer(plugin, capabilityId, serverLanguage, file); err != nil {
				return err
			}
		}

		return nil
	})
}

func parseMode(value string) (*wasmpb.Mode, error) {
	mode, ok := wasmpb.Mode_value[strings.ToUpper(value)]
	if !ok {
		return nil, fmt.Errorf("unknown mode %s", value)
	}

	tmp := wasmpb.Mode(mode)
	return &tmp, nil
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
