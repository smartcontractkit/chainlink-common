package main

import (
	"bytes"
	_ "embed"
	"log"
	"text/template"

	"github.com/smartcontractkit/chainlink-common/pkg/utils/codegen"
)

//go:embed compute.go.templ
var computeGo string

const toolName = "github.com/smartcontractkit/chainlink-common/pkg/workflows/gen"

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

	files := map[string]string{"compute_generated.go": results.String()}
	if err = codegen.WriteFiles(".", "github.com/smartcontractkit", toolName, files); err != nil {
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
