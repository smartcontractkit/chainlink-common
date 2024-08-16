package main

import (
	"bytes"
	_ "embed"
	"go/format"
	"log"
	"os"
	"text/template"
)

//go:embed compute.go.templ
var computeGo string

func main() {
	computes := rangeNum(11)[1:]
	t, err := template.New("go compute").Funcs(template.FuncMap{"RangeNum": rangeNum}).Parse(computeGo)
	if err != nil {
		log.Fatal(err)
	}

	results := bytes.Buffer{}
	if err = t.Execute(&results, computes); err != nil {
		log.Fatal(err)
	}

	formatted, err := format.Source(results.Bytes())
	if err != nil {
		if err2 := os.WriteFile("compute.gen.go", results.Bytes(), 0644); err2 != nil {
			log.Fatalf("error formatting source and writing to file\n%v\n%v", err, err2)
		}
		log.Fatalf("eror fromatting go file still written to %s, but this tool must be fixed", "compute.gen.go")
	}

	if err = os.WriteFile("compute_generated.go", formatted, 0644); err != nil {
		log.Fatal(err)
	}
}

func rangeNum(num int) []int {
	computes := make([]int, num)
	for i := range num {
		computes[i] = i
	}

	return computes
}
