package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDedupe(t *testing.T) {
	t.Parallel()

	in := strings.Join([]string{
		"mode: set",
		"github.com/example/pkg/foo.go:10.2,12.1 1 0",
		"github.com/example/pkg/foo.go:10.2,12.1 1 1",
		"github.com/example/pkg/foo.go:10.2,12.1 1 0",
		"github.com/example/pkg/bar.go:1.1,2.1 2 0",
	}, "\n") + "\n"

	var out bytes.Buffer
	require.NoError(t, dedupe(strings.NewReader(in), &out))

	require.Equal(t, strings.Join([]string{
		"mode: set",
		"github.com/example/pkg/foo.go:10.2,12.1 1 1",
		"github.com/example/pkg/bar.go:1.1,2.1 2 0",
	}, "\n")+"\n", out.String())
}
