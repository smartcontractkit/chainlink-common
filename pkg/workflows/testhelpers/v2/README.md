This package is used to allow consistent testing across the SDK for:
- SDK implementation in testutils 
- SDK implementation in WASM (aside from the host callbacks)
- The host Module struct

Though the host test is sufficient to test the WASM SDK, it would be unclear if the error is in the guest or host.