package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/smartcontractkit/chainlink-common/pkg/values/installer/pkg"
)

func main() {
	gen := &pkg.ProtocGen{
		ProtocHelper: ProtocHelper{},
		Plugins:      []pkg.Plugin{{Name: "cre", Path: "../.."}},
	}

	internalProtos := []*pkg.CapabilityConfig{
		{
			Category:      "internal",
			Pkg:           "consensus",
			MajorVersion:  1,
			PreReleaseTag: "alpha",
			Files:         []string{"consensus.proto"},
		},
		{
			Category:     "internal",
			Pkg:          "actionandtrigger",
			MajorVersion: 1,
			Files:        []string{"action_and_trigger.proto"},
		},
		{
			Category:     "internal",
			Pkg:          "basicaction",
			MajorVersion: 1,
			Files:        []string{"basic_action.proto"},
		},
		{
			Category:     "internal",
			Pkg:          "basictrigger",
			MajorVersion: 1,
			Files:        []string{"basic_trigger.proto"},
		},
		{
			Category:     "internal",
			Pkg:          "nodeaction",
			MajorVersion: 1,
			Files:        []string{"node_action.proto"},
		},
		{
			Category:     "internal",
			Pkg:          "importclash",
			MajorVersion: 1,
			Files:        []string{"clash.proto"},
		},
		{
			Category:     "internal/importclash",
			Pkg:          "p1",
			MajorVersion: 1,
			Files:        []string{"import.proto"},
		},
		{
			Category:     "internal/importclash",
			Pkg:          "p2",
			MajorVersion: 1,
			Files:        []string{"import.proto"},
		},
	}

	internalProtosToDir := map[string]*pkg.CapabilityConfig{}

	for _, proto := range internalProtos {
		key := proto.Pkg
		categoryParts := strings.Split(proto.Category, string(filepath.Separator))
		if len(categoryParts) > 1 {
			prefix := filepath.Join(categoryParts[1:]...)
			key = filepath.Join(prefix, key)
		}

		internalProtosToDir[key] = proto
		if err := os.MkdirAll(key, os.ModePerm); err != nil {
			panic(err)
		}
	}

	if err := gen.GenerateMany(internalProtosToDir); err != nil {
		panic(err)
	}
}
