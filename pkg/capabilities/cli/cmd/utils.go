package cmd

import (
	"os"
	"path"
	"strings"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
)

func printFiles(dir string, files map[string]string) error {
	for file, content := range files {
		if !strings.HasPrefix(file, dir) {
			file = dir + "/" + file
		}

		if err := os.MkdirAll(path.Dir(file), os.ModePerm); err != nil {
			return err
		}

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			return err
		}
	}

	return nil
}

func capabilityTypeFromString(capabilityTypeRaw string) capabilities.CapabilityType {
	var capabilityType capabilities.CapabilityType
	for ; capabilityType.IsValid() == nil; capabilityType++ {
		if capabilityType.String() == capabilityTypeRaw {
			return capabilityType
		}
	}

	return capabilityType
}
