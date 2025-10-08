package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"go/format"
	"os"
	"runtime/debug"
	"strings"
	"text/template"
)

func main() {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		panic("build info unavailable")
	}
	location := strings.TrimPrefix(bi.Path, "github.com/smartcontractkit/")

	for _, gen := range []genInfo{
		{template: intsTemplate, fileName: "int_gen.go"},
		{template: intsTestTemplate, fileName: "int_gen_test.go"},
	} {
		t, err := template.New(gen.fileName).
			Funcs(map[string]any{"div": func(a, b int) int { return a / b }}).
			Parse(gen.template)
		if err != nil {
			panic(err)
		}

		br := bytes.Buffer{}
		if err = t.Execute(&br, []int{8, 16, 32, 64}); err != nil {
			panic(err)
		}

		res, err := format.Source(br.Bytes())
		if err != nil {
			panic(err)
		}

		res =
			fmt.Appendf(nil,
				"// DO NOT MODIFY: automatically generated from %s/main.go using the template %s\n\n%s",
				location,
				gen.fileName,
				string(res),
			)

		if err = os.WriteFile(gen.fileName, res, 0600); err != nil {
			panic(err)
		}
	}
}

type genInfo struct {
	template string
	fileName string
}

//go:embed ints.go.tmpl
var intsTemplate string

//go:embed ints_test.go.tmpl
var intsTestTemplate string
