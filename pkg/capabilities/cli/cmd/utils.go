package cmd

import (
	"os"
	"path"
	"strings"
)

func printFiles(dir string, files map[string]string) error {
	for file, content := range files {
		if !strings.HasPrefix(file, dir) {
			file = dir + "/" + file
		}

		if err := os.MkdirAll(path.Dir(file), 0600); err != nil {
			return err
		}

		if err := os.WriteFile(file, []byte(content), 0600); err != nil {
			return err
		}
	}

	return nil
}

func capitalize(s string) string {
	return strings.ToUpper(string(s[0])) + s[1:]
}
