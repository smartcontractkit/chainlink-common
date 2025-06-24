package main

import "github.com/smartcontractkit/chainlink-common/pkg/values/installer/pkg"

func main() {
	gen := &pkg.ProtocGen{}
	gen.AddSourceDirectories(".")
	if err := gen.Generate("ocr3_types.proto", "."); err != nil {
		panic(err)
	}
}
