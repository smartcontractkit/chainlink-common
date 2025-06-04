package bindings

// This function is not EVM specific, it's generic and should be provided by CRE
func GenerateReport(chainID uint32, userData []byte) commonReport {
	return commonReport{}
}

// This is not EVM specific, it's generic
type commonReport struct {
	RawReport     []byte
	ReportContext []byte
	Signatures    [][]byte
	ID            []byte
}
