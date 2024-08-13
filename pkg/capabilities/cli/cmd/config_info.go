package cmd

import "github.com/atombender/go-jsonschema/pkg/generator"

type ConfigInfo struct {
	generator.Config
	SchemaToTypeInfo map[string]TypeInfo
}
