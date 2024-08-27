package codegen

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"golang.org/x/tools/imports"
)

func WriteFiles(dir, localPrefix string, files map[string]string) error {
	var errs []error
	for file, content := range files {
		if !strings.HasPrefix(file, dir) {
			file = dir + "/" + file
		}

		if strings.HasSuffix(file, ".go") {
			imports.LocalPrefix = localPrefix
			rawContent, err := imports.Process(file, []byte(content), nil)
			if err != nil {
				// print an error, but also write the file so debugging the generator isn't a pain.
				fmt.Printf("Error formatting file %s: %s\n", file, err)
				errs = append(errs, err)
			} else {
				content = string(rawContent)
			}
		}

		if err := os.MkdirAll(path.Dir(file), 0600); err != nil {
			errs = append(errs, err)
		} else if err = os.WriteFile(file, []byte(content), 0600); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) == 0 {
		return nil
	}

	return errors.Join(errs...)
}
