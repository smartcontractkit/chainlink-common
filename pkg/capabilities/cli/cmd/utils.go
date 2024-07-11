package cmd

import (
	"os"
	"strings"
)

func printFiles(dir string, files map[string]string) error {
	for file, content := range files {
		if !strings.HasPrefix(file, dir) {
			file = dir + "/" + file
		}

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			return err
		}
	}

	return nil
}
