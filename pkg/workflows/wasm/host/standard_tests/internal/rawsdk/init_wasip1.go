package rawsdk

//go:wasmimport env version_v2
func versionV2()

// Allow the host to link against the right version of the SDK.
// without the call, versionV2 is optimized out by the compiler.
func init() {
	versionV2()
}
