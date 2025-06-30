package pkg

import (
	"errors"
	"fmt"
	"path"
	"strconv"
)

type CapabilityConfig struct {
	Category      string
	Pkg           string
	MajorVersion  int
	PreReleaseTag string
	Files         []string
}

func (c *CapabilityConfig) FullProtoFiles() []string {
	protoDir := path.Join("capabilities", c.Category, c.Pkg, "v"+strconv.Itoa(c.MajorVersion)+c.PreReleaseTag)
	fullFiles := make([]string, len(c.Files))
	for i, file := range c.Files {
		fullFiles[i] = path.Join(protoDir, file)
	}
	return fullFiles
}

func (c *CapabilityConfig) Validate() error {
	if c.Category == "" {
		return errors.New("category must not be empty")
	}
	if c.Pkg == "" {
		return errors.New("pkg must not be empty")
	}
	if c.MajorVersion < 1 {
		return fmt.Errorf("major-version must be >= 1, got %d", c.MajorVersion)
	}
	if len(c.Files) == 0 {
		return errors.New("files must not be empty")
	}
	return nil
}
