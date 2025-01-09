package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd"
)

var localPrefix = flag.String(
	"local_prefix",
	"github.com/smartcontractkit",
	"The local prefix to use when formatting go files.",
)

var types = flag.String(
	"types",
	"",
	"Comma separated list of types to generate for.  If empty, all types created in the package will be generated."+
		" if set, other types in the same package will automatically be added to the skip_cap list",
)

var skipCap = flag.String(
	"skip_cap",
	"",
	"Comma separated list of types (including the import name), or impute to not expect a capability definition to exist for"+
		" By default, this generator assumes that all types referenced (aside from primitives) will either be generated with this call or already have Cap type",
)

var dir = flag.String("dir", ".", "The input directory, defaults to the running directory")

func main() {
	flag.Parse()
	templates := map[string]cmd.TemplateAndCondition{}
	cmd.AddDefaultGoTemplates(templates, false)
	helpers := []cmd.WorkflowHelperGenerator{&cmd.TemplateWorkflowGeneratorHelper{Templates: templates}}

	info := cmd.UserGenerationInfo{
		Dir:          *dir,
		LocalPrefix:  *localPrefix,
		Helpers:      helpers,
		GenForStruct: genForStruct(),
	}

	if err := cmd.GenerateUserTypes(info); err != nil {
		panic(err)
	}
}

func genForStruct() func(string) bool {
	skipGen := buildSkipGen()
	genPackageType := buildGenPkgType()
	return func(s string) bool {
		if skipGen[s] {
			return false
		}

		pkgAndStruct := strings.Split(s, ".")

		switch len(pkgAndStruct) {
		case 1:
			return genPackageType(pkgAndStruct[0])

		case 2:
			if skipGen[pkgAndStruct[0]] {
				return false
			}
		default:
			panic(fmt.Sprintf("invalid type %s", s))
		}

		return true
	}
}

func buildSkipGen() map[string]bool {
	skipGen := map[string]bool{}
	for _, skip := range strings.Split(*skipCap, ",") {
		skipGen[skip] = true
	}
	return skipGen
}

func buildGenPkgType() func(string) bool {
	genPkgType := func(_ string) bool { return true }
	if *types != "" {
		genPkg := map[string]bool{}
		for _, t := range strings.Split(*types, ",") {
			genPkg[t] = true
		}
		genPkgType = func(s string) bool {
			return genPkg[s]
		}
	}

	return genPkgType
}
