package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"go/format"
	"os"
	"text/template"
)

func main() {
	for _, gen := range []genInfo{
		{template: intsTemplate, fileName: "int_gen.go"},
		{template: intsTestTemplate, fileName: "int_gen_test.go"},
	} {
		t, err := template.New(gen.fileName).Parse(gen.template)
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

		res = []byte(
			fmt.Sprintf(
				"// DO NOT MODIFY: automatically generated from pkg/codec/raw/types/main.go using the template %s\n\n%s",
				gen.fileName,
				string(res),
			))

		if err = os.WriteFile(gen.fileName, res, 0600); err != nil {
			panic(err)
		}
	}
}

type genInfo struct {
	template string
	fileName string
}

//go:embed ints.tmpl
var intsTemplate string

//go:embed ints_test.tmpl
var intsTestTemplate string
