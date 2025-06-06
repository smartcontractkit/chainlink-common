package gateway

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMessage_Validate(t *testing.T) {
	msg := &Message{
		Body: MessageBody{
			MessageId: "abcd",
			Method:    "request",
			DonId:     "donA",
			Receiver:  "0x0000000000000000000000000000000000000000",
			Payload:   []byte("datadata"),
		},
	}

	signMessage(t, msg)
	// valid
	require.NoError(t, msg.Validate())

	// missing message ID
	msg.Body.MessageId = ""
	require.Error(t, msg.Validate())
	// message ID ending with null bytes
	msg.Body.MessageId = "myid\x00\x00"
	require.Error(t, msg.Validate())
	msg.Body.MessageId = "abcd"
	require.NoError(t, msg.Validate())

	// missing DON ID
	msg.Body.DonId = ""
	require.Error(t, msg.Validate())
	// DON ID ending with null bytes
	msg.Body.DonId = "mydon\x00\x00"
	require.Error(t, msg.Validate())
	msg.Body.DonId = "donA"
	require.NoError(t, msg.Validate())

	// method name too long
	msg.Body.Method = string(bytes.Repeat([]byte("a"), MessageMethodMaxLen+1))
	require.Error(t, msg.Validate())
	// empty method name
	msg.Body.Method = ""
	require.Error(t, msg.Validate())
	// method name ending with null bytes
	msg.Body.Method = "method\x00"
	require.Error(t, msg.Validate())
	msg.Body.Method = "request"
	require.NoError(t, msg.Validate())

	// incorrect receiver
	msg.Body.Receiver = "blah"
	require.Error(t, msg.Validate())
	msg.Body.Receiver = "0x0000000000000000000000000000000000000000"
	require.NoError(t, msg.Validate())

	// invalid signature
	msg.Signature = "0x00"
	require.Error(t, msg.Validate())
}

func flatten(data ...[]byte) []byte {
	var result []byte
	for _, d := range data {
		result = append(result, d...)
	}
	return result
}

func signMessage(t *testing.T, msg *Message) {
	// Generate ECDSA key pair
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	rawData := GetRawMessageBody(&msg.Body)
	hash := sha256.Sum256(flatten(rawData...))

	// Sign the hash
	r, s, err := ecdsa.Sign(rand.Reader, privKey, hash[:])
	require.NoError(t, err)

	// Encode signature as hex string (r || s || v)
	sigBytes := append(r.Bytes(), s.Bytes()...)

	// Ethereum signatures include a recovery id (v). For testing, set v = 0.
	v := byte(0)
	sigBytes = append(sigBytes, v)
	msg.Signature = "0x" + hex.EncodeToString(sigBytes)
}
