package ccipocr3

// AttestationStatus is a struct holding all the necessary information to build payload to
// mint USDC on the destination chain. Valid AttestationStatus always contains MessageHash and Attestation.
// In case of failure, Error is populated with more details.
type AttestationStatus struct {
	// MessageHash is the hash of the message that the attestation was fetched for. It's going to be MessageSent event hash
	MessageHash Bytes
	// Attestation is the attestation data fetched from the API, encoded in bytes
	Attestation Bytes
	// Error is the error that occurred during fetching the attestation data
	Error error
}
