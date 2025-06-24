package main

import "github.com/smartcontractkit/chainlink-common/pkg/values/installer/pkg"

func main() {
	if err := (&pkg.ProtocGen{}).Generate("values/v1/values.proto", "."); err != nil {
		panic(err)
	}
}
