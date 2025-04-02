package cmd

type Field struct {
	Type        string
	NumSlice    int
	IsPrimitive bool
	ConfigName  string
	SkipCap     bool
}

func (f Field) WrapCap() bool {
	return !f.SkipCap && !f.IsPrimitive && f.NumSlice == 0
}
