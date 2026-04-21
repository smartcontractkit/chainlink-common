package signer

import "errors"

// Error sentinels emitted by signers. These are matched by the framework
// executor's string-sentinel-to-caperrors mapping to classify the error's
// visibility and code (see confidential-compute/capabilities/framework/
// executor.go).
//
// Sentinels here represent *signer-layer* failures. Validation-layer failures
// (e.g. missing required fields, secret-name not in vault_don_secrets) are
// returned from the confidentialhttp validator.
var (
	// ErrSecretIdentifierNil means an auth variant referenced a nil
	// SecretIdentifier. The validator should have rejected the request
	// before it reached the signer.
	ErrSecretIdentifierNil = errors.New("confidentialhttp signer: secret identifier is nil")

	// ErrSecretNameEmpty means an auth variant referenced a secret name
	// that was the empty string. This should never happen if the validator
	// ran first.
	ErrSecretNameEmpty = errors.New("confidentialhttp signer: secret name is empty")

	// ErrSecretNotProvided means the workflow did not include a value for
	// a secret name that the auth config requires. The validator should
	// have rejected the request earlier.
	ErrSecretNotProvided = errors.New("confidentialhttp signer: required secret not provided")

	// ErrSecretEmpty means the Vault-DON returned an empty value for a
	// required secret.
	ErrSecretEmpty = errors.New("confidentialhttp signer: required secret is empty")

	// ErrStringOrSecretNil means a required StringOrSecret field was nil.
	ErrStringOrSecretNil = errors.New("confidentialhttp signer: StringOrSecret is nil")

	// ErrStringOrSecretUnset means a StringOrSecret oneof had no variant set.
	ErrStringOrSecretUnset = errors.New("confidentialhttp signer: StringOrSecret value not set")

	// ErrSigV4Sign indicates the AWS SigV4 signer failed. Wrapped with the
	// underlying error from aws-sdk-go-v2.
	ErrSigV4Sign = errors.New("confidentialhttp signer: AWS SigV4 signing failed")

	// ErrHmacTemplateParse means a user-supplied canonical_template did not
	// parse as a valid Go text/template. Surfaced as a user error.
	ErrHmacTemplateParse = errors.New("confidentialhttp signer: HMAC canonical template parse failed")

	// ErrHmacTemplateExec means template execution failed at runtime
	// (unknown field access, type mismatch, etc.).
	ErrHmacTemplateExec = errors.New("confidentialhttp signer: HMAC canonical template execution failed")

	// ErrHmacTemplateTimeout means template execution exceeded the time
	// budget (guards against pathological user templates).
	ErrHmacTemplateTimeout = errors.New("confidentialhttp signer: HMAC canonical template timed out")

	// ErrOAuth2TokenEndpointUnreachable is returned when the IdP could not
	// be reached at all (network error, TLS handshake failure, etc.).
	ErrOAuth2TokenEndpointUnreachable = errors.New("confidentialhttp signer: OAuth2 token endpoint unreachable")

	// ErrOAuth2TokenEndpointHTTPError is returned when the IdP returned a
	// non-2xx response. The body is NOT surfaced in the error message to
	// avoid leaking secrets that some IdPs echo back.
	ErrOAuth2TokenEndpointHTTPError = errors.New("confidentialhttp signer: OAuth2 token endpoint returned non-2xx")

	// ErrOAuth2TokenResponseInvalid means the 2xx response from the IdP did
	// not parse as a valid OAuth2 access-token response.
	ErrOAuth2TokenResponseInvalid = errors.New("confidentialhttp signer: OAuth2 token response invalid")

	// ErrOAuth2TokenURLInvalid means token_url was not https:// or otherwise
	// malformed. Validator should have caught this earlier; this is a
	// defense-in-depth check.
	ErrOAuth2TokenURLInvalid = errors.New("confidentialhttp signer: OAuth2 token URL invalid")

	// ErrUnsupportedHashAlgorithm is returned when HmacCustom specifies a
	// hash enum value the signer does not implement.
	ErrUnsupportedHashAlgorithm = errors.New("confidentialhttp signer: unsupported hash algorithm")

	// ErrUnsupportedEncoding is returned for an unrecognized "hex"/"base64"
	// encoding string.
	ErrUnsupportedEncoding = errors.New("confidentialhttp signer: unsupported encoding")
)
