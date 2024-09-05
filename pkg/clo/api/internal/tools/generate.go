package main

import (
	"fmt"
	"go/types"
	"os"
	"strings"

	"github.com/99designs/gqlgen/api"
	"github.com/99designs/gqlgen/codegen/config"
	"github.com/99designs/gqlgen/plugin/modelgen"
	"github.com/Khan/genqlient/generate"

	"github.com/smartcontractkit/feeds-manager/api/models"
)

// Define a modelgen plugin hook to mutate generated models and interfaces
func mutateHook(b *modelgen.ModelBuild) *modelgen.ModelBuild {
	tmpModels := b.Models[:0]
	for _, m := range b.Models {
		// Skip generating default error models
		if !isError(m) {
			tmpModels = append(tmpModels, m)
		}
		// Set any Error fields to custom Error type
		for _, f := range m.Fields {
			if f.GoName == "Errors" {
				f.Type = types.NewSlice(&models.Error{})
			}
			// Add omitempty to json tags
			if !strings.Contains(f.Tag, "omitempty") {
				tag := strings.TrimSuffix(f.Tag, `"`)
				f.Tag = fmt.Sprintf(`%v,omitempty"`, tag)
			}
		}
	}
	b.Models = tmpModels

	// Skip generating default error interfaces
	interfaces := b.Interfaces[:0]
	for _, i := range b.Interfaces {
		if !strings.Contains(i.Name, "Error") {
			interfaces = append(interfaces, i)
		}
	}
	b.Interfaces = interfaces
	return b
}

func isError(m *modelgen.Object) bool {
	// Models explicitly named Error
	if strings.Contains(m.Name, "Error") {
		return true
	}
	// Models that only contain a single field, named Message
	if len(m.Fields) == 1 && m.Fields[0].GoName == "Message" {
		return true
	}
	return false
}

func main() {
	cfg, err := config.LoadConfig("api/internal/tools/gqlgen.yml")
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to load config", err.Error())
		os.Exit(2)
	}

	// Attach the mutation function onto modelgen plugin
	p := modelgen.Plugin{
		MutateHook: mutateHook,
	}

	// Generate client models using gqlgen, recover after panic due to models no longer matching schema
	func() {
		defer func() {
			_ = recover()
		}()
		_ = api.Generate(cfg, api.ReplacePlugin(&p))
	}()

	// Generate client functions using genqlient (point it to the genqlient.yml config)
	os.Args = append(os.Args, "api/internal/tools/genqlient.yml")
	generate.Main()
}
