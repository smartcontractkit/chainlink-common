package rawsdk

//go:wasmimport env version_v2
func versionV2()

// Allow the host to link against the right version of the SDK.
func init() {
	versionV2()
}
