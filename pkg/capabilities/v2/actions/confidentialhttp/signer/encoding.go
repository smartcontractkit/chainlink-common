package signer

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"io"
	"net/http"
)

func toHex(b []byte) string {
	return hex.EncodeToString(b)
}

func toBase64(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

// readBodyForHashing reads and rewinds req.Body so it can be signed and then
// re-sent by http.Client. Returns the raw body bytes.
//
// Go sets req.GetBody automatically for bodies created via bytes.NewBuffer
// and similar; we rely on that to restore the body for the second read. If
// GetBody is unset and Body is consumed, subsequent Do() would send an empty
// body. Callers must ensure GetBody is set before invoking any HMAC/SigV4
// signer.
func readBodyForHashing(req *http.Request) ([]byte, error) {
	if req.Body == nil || req.Body == http.NoBody {
		return nil, nil
	}
	b, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	if closer, ok := req.Body.(io.Closer); ok {
		_ = closer.Close()
	}
	// Restore body for the actual send.
	if req.GetBody != nil {
		nb, gerr := req.GetBody()
		if gerr != nil {
			return nil, gerr
		}
		req.Body = nb
	}
	return b, nil
}

func sha256Hex(b []byte) string {
	h := sha256.Sum256(b)
	return toHex(h[:])
}
