const (
	keyNamePrefix = "EVMOCROnchainKeyRing"
)

func NewEVMOCROnchainKeyRing(name string, ks keystore.Keystore) *EVMOCROnchainKeyRing {
	return EVMOCROnchainKeyRing{ks: ks, name: name}
}

func (EVMOCROnchainKeyRing) KeyName(name string) string {
	return fmt.Sprintf("%s_%s", keyNamePrefix, name)
}

func (EVMOCROnchainKeyRing) CreateKey(ctx context.Context, name string) error {
	return ks.CreateKey(ctx, fmt.Sprintf("%s_%s", KeyName, name), keystore.Ed25519)
}

func (EVMOCROnchainKeyRing) DeleteKey(ctx context.Context, name string) error {
	return ks.DeleteKey(ctx, fmt.Sprintf("%s_%s", KeyName, name), keystore.Ed25519)
}

func (EVMOCROnchainKeyRing) PublicKey() OnchainPublicKey {
	if key, err := ks.GetKey(ctx, fmt.Sprintf("%s_%s", KeyName, name)); err == nil {
		return key.PublicKey
	}
	return nil
}

func ReportToSigData(reportCtx ocrtypes.ReportContext, report ocrtypes.Report) []byte {
	rawReportContext := evmutil.RawReportContext(reportCtx)
	sigData := crypto.Keccak256(report)
	sigData = append(sigData, rawReportContext[0][:]...)
	sigData = append(sigData, rawReportContext[1][:]...)
	sigData = append(sigData, rawReportContext[2][:]...)
	return crypto.Keccak256(sigData)
}

func (EVMOnchainKeyRing) Sign(ctx ReportContext, rep Report) (signature []byte, err error) {
	hash, err := ReportToSigData(ctx, rep)
	if err != nil {
		return nil, err
	}
	return = ks.Sign(ctx, fmt.Sprintf("%s_%s", KeyName, name), hash)
}

func (EVMOnchainKeyRing) Verify(_ OnchainPublicKey, _ ReportContext, _ Report, signature []byte) bool {
	// TODO
	return false
}