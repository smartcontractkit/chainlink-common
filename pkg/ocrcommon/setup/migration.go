// Package internal holds the migration-generation logic and template used by
// the setup CLI (../migrate). It is internal so the only supported entrypoint
// is the command, while the logic stays independently testable.
package main

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/template"
)

//go:embed migration.sql.tmpl
var migrationTemplate string

// tableNameRe restricts table names to a safe, unquoted SQL identifier. The
// name is interpolated directly into DDL, so we must not allow anything that
// could break out of the identifier.
var tableNameRe = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// migrationFileRe matches goose-style migration filenames, capturing the
// leading zero-padded sequence number (e.g. "0007_create_foo.sql" -> "0007").
var migrationFileRe = regexp.MustCompile(`^(\d+)_.*\.sql$`)

// CreateMigration writes a new goose migration that creates tableName into dir,
// choosing the next sequential migration number based on the .sql files already
// present. The directory is created if it does not exist. It returns the path
// of the file that was written.
func CreateMigration(dir, tableName string) (string, error) {
	if !tableNameRe.MatchString(tableName) {
		return "", fmt.Errorf("invalid table name %q: must match %s", tableName, tableNameRe.String())
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create migrations dir %q: %w", dir, err)
	}

	next, err := nextMigrationNumber(dir)
	if err != nil {
		return "", err
	}

	body, err := renderMigration(tableName)
	if err != nil {
		return "", err
	}

	filename := fmt.Sprintf("%04d_create_%s.sql", next, tableName)
	path := filepath.Join(dir, filename)
	if _, err := os.Stat(path); err == nil {
		return "", fmt.Errorf("migration %q already exists", path)
	}

	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		return "", fmt.Errorf("failed to write migration %q: %w", path, err)
	}
	return path, nil
}

// nextMigrationNumber returns the migration sequence number to use next: one
// greater than the highest existing number, or 1 when dir has no migrations.
func nextMigrationNumber(dir string) (int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, fmt.Errorf("failed to read migrations dir %q: %w", dir, err)
	}

	var numbers []int
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		m := migrationFileRe.FindStringSubmatch(e.Name())
		if m == nil {
			continue
		}
		n, err := strconv.Atoi(m[1])
		if err != nil {
			return 0, fmt.Errorf("failed to parse migration number from %q: %w", e.Name(), err)
		}
		numbers = append(numbers, n)
	}

	if len(numbers) == 0 {
		return 1, nil
	}
	sort.Ints(numbers)
	return numbers[len(numbers)-1] + 1, nil
}

func renderMigration(tableName string) (string, error) {
	tmpl, err := template.New("migration").Parse(migrationTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse migration template: %w", err)
	}
	var sb strings.Builder
	if err := tmpl.Execute(&sb, struct{ TableName string }{TableName: tableName}); err != nil {
		return "", fmt.Errorf("failed to render migration template: %w", err)
	}
	return sb.String(), nil
}
