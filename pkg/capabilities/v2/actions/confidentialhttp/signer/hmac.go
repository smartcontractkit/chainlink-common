package signer

import (
	"errors"
	"fmt"

	confhttppb "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/actions/confidentialhttp"
)

func newHmacSigner(cfg *confhttppb.HmacAuth) (Signer, error) {
	if cfg == nil {
		return nil, errors.New("hmac auth config is nil")
	}
	switch v := cfg.GetVariant().(type) {
	case *confhttppb.HmacAuth_Sha256:
		return newHmacSha256Signer(v.Sha256)
	case *confhttppb.HmacAuth_AwsSigV4:
		return newAwsSigV4Signer(v.AwsSigV4)
	case *confhttppb.HmacAuth_Custom:
		return newHmacCustomSigner(v.Custom)
	case nil:
		return nil, errors.New("hmac auth variant not set")
	default:
		return nil, fmt.Errorf("unsupported hmac variant %T", v)
	}
}

// encode formats a raw MAC digest as hex or base64.
// encoding may be "", "hex" (default), or "base64".
func encodeMAC(mac []byte, encoding string) (string, error) {
	switch encoding {
	case "", "hex":
		return toHex(mac), nil
	case "base64":
		return toBase64(mac), nil
	default:
		return "", fmt.Errorf("%w: %q", ErrUnsupportedEncoding, encoding)
	}
}
