package kms

import "encoding/asn1"

// SPKI represents the SubjectPublicKeyInfo structure as defined in [RFC 5280] in ASN.1 format.
//
// The public key that AWS KMS returns is a DER-encoded X.509 public key, also known as
// SubjectPublicKeyInfo (SPKI). This structure is used to unpack the public key returned by the KMS
// GetPublicKey API call.
//
// For more details: see the AWS KMS documentation on [GetPublicKey response syntax].
//
// [RFC 5280]: https://datatracker.ietf.org/doc/html/rfc5280
// [GetPublicKey response syntax]: https://docs.aws.amazon.com/kms/latest/APIReference/API_GetPublicKey.html#API_GetPublicKey_ResponseSyntax
type SPKI struct {
	AlgorithmIdentifier SPKIAlgorithmIdentifier
	SubjectPublicKey    asn1.BitString
}

// SPKIAlgorithmIdentifier represents the AlgorithmIdentifier structure for the
// SubjectPublicKeyInfo (SPKI) in ASN.1 format.
type SPKIAlgorithmIdentifier struct {
	Algorithm  asn1.ObjectIdentifier
	Parameters asn1.ObjectIdentifier
}

// ECDSASig represents the ECDSA signature structure as defined in [RFC 3279] in ASN.1 format.
// This structure is used to unpack the ECDSA signature returned by AWS KMS when signing data.
//
// [RFC 3279] https://datatracker.ietf.org/doc/html/rfc3279#section-2.2.3
type ECDSASig struct {
	R asn1.RawValue
	S asn1.RawValue
}
