package cmd

import "strings"

type Struct struct {
	Name    string
	Outputs map[string]Field
	Ref     *string
}

func (s Struct) RefPkg() string {
	if s.Ref == nil {
		return ""
	}
	return strings.Split(*s.Ref, ".")[0]
}
