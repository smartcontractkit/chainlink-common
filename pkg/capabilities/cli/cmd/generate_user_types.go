package cmd

import (
	"errors"
	"os"
	"path"
	"strings"
)

func GenerateUserTypes(info UserGenerationInfo) error {
	dir, err := os.ReadDir(info.Dir)
	if err != nil {
		return err
	}

	var errs []error
	for _, file := range dir {
		fileName := file.Name()
		if file.IsDir() || !strings.HasSuffix(fileName, ".go") {
			continue
		}

		rawContent, err2 := os.ReadFile(path.Join(info.Dir, fileName))
		if err2 != nil {
			return err2
		}

		content := string(rawContent)
		if strings.HasPrefix(content, "// Code generated by github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli, DO NOT EDIT.") {
			continue
		}

		typeInfo := TypeInfo{CapabilityType: "common"}

		err2 = generateFromGoSrc(
			info.Dir, info.LocalPrefix, content, typeInfo, info.Helpers, map[string]string{}, "", info.GenForStruct)
		if err2 != nil {
			errs = append(errs, err2)
		}
	}

	return errors.Join(errs...)
}

type UserGenerationInfo struct {
	Dir          string
	LocalPrefix  string
	Helpers      []WorkflowHelperGenerator
	GenForStruct func(string) bool
}