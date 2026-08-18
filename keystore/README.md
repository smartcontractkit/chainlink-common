[![Go Reference](https://pkg.go.dev/badge/github.com/smartcontractkit/chainlink-common/keystore.svg)](https://pkg.go.dev/github.com/smartcontractkit/chainlink-common/keystore)


## Keystore

Dependency minimized chain family agnostic key storage library. Supports the following key types:
- Digital Signatures: ECDSA on secp256k1 and Ed25519. 
Note KMS is supported for these key types.
- Hybrid Encryption: X25519 (nacl/box) and ECDH on P256

Warning: ECDH on P256 is pending audit do not use in 
production.

Family specific logic is layered on top in the following locations:
- [EVM](https://github.com/smartcontractkit/chainlink-evm/tree/develop/pkg/keys/v2)
    - Note this also holds a full e2e example of using the keystore with [libocr](https://github.com/smartcontractkit/libocr/tree/master). 
- [Solana](https://github.com/smartcontractkit/chainlink-solana/tree/develop/pkg/solana/keys)
- More coming soon


### Examples 

#### Signatures 
```go
package main

import (
	"context"
	"crypto/sha256"

	"github.com/smartcontractkit/chainlink-common/keystore"
)

func main() {
	ctx := context.Background()

    // Note postgres and file based storage also supported. 
	ks, err := keystore.LoadKeystore(ctx, keystore.NewMemoryStorage(), "password")

	createResp, _ := ks.CreateKeys(ctx, keystore.CreateKeysRequest{
		Keys: []keystore.CreateKeyRequest{
			{KeyName: "my-key", KeyType: keystore.ECDSA_S256},
		},
	})
	// Sign data 
	data := []byte("hello world")
	hash := sha256.Sum256(data)
	
	signResp, _ := ks.Sign(ctx, keystore.SignRequest{
		KeyName: "my-key",
		Data:    hash[:],
	})

	// Verify the signature
	verifyResp, _ := ks.Verify(ctx, keystore.VerifyRequest{
		KeyType:   keystore.ECDSA_S256,
		PublicKey: createResp.Keys[0].KeyInfo.PublicKey,
		Data:      hash[:],
		Signature: signResp.Signature,
	})
    // verifyResp.Valid == true
}
```

#### KMS Signatures 
```go
package main

import (
	"context"
	"crypto/sha256"

	"github.com/smartcontractkit/chainlink-common/keystore"
	"github.com/smartcontractkit/chainlink-common/keystore/kms"
)

func main() {
	ctx := context.Background()

	// Create a KMS backed keystore. 
	kmsClient, _ := kms.NewClient(ctx, kms.ClientOptions{
		Profile: "my-profile", // Optional: omit for default credential chain
	})
	ks, _ := kms.NewKeystore(kmsClient)

	data := []byte("hello world")
	hash := sha256.Sum256(data)
    // Same signer interface but uses AWS allocated
    // key names.
	signResp, _ := ks.Sign(ctx, keystore.SignRequest{
		KeyName: "AWSkeyID",  
		Data:    hash[:], 
	})

	// Verify the signature
	verifyResp, _ := keystore.Verify(ctx, keystore.VerifyRequest{
		KeyType:   keysResp.Keys[0].KeyInfo.KeyType,
		PublicKey: keysResp.Keys[0].KeyInfo.PublicKey,
		Data:      hash[:],
		Signature: signResp.Signature,
	})
	// verifyResp.Valid == true
}
```


#### Encryption
```go
package main

import (
	"context"

	"github.com/smartcontractkit/chainlink-common/keystore"
)

func main() {
	ctx := context.Background()

	ks, _ := keystore.LoadKeystore(ctx, keystore.NewMemoryStorage(), "password")

	// Create an X25519 key for encryption/decryption
	createResp, _ := ks.CreateKeys(ctx, keystore.CreateKeysRequest{
		Keys: []keystore.CreateKeyRequest{
			{KeyName: "encrypt-key", KeyType: keystore.X25519},
		},
	})

	// Encrypt data using the public key
	data := []byte("secret message")
	encryptResp, _ := ks.Encrypt(ctx, keystore.EncryptRequest{
		RemoteKeyType: keystore.X25519,
		RemotePubKey:  createResp.Keys[0].KeyInfo.PublicKey,
		Data:          data,
	})

	// Decrypt using the key name
	decryptResp, _ := ks.Decrypt(ctx, keystore.DecryptRequest{
		KeyName:       "encrypt-key",
		EncryptedData: encryptResp.EncryptedData,
	})
	// decryptResp.Data == data
}
```


#### CLI
```bash
# Set up environment variables
export KEYSTORE_PASSWORD="my-secure-password"
export KEYSTORE_FILE_PATH="./keystore.json"

# Create a new keystore file (if using file storage)
touch ./keystore.json

# Create an ECDSA key
keys create -d '{"Keys": [{"KeyName": "my-key", "KeyType": "ECDSA_S256"}]}'

# List all keys
keys list

# Get a specific key
keys get -d '{"KeyNames": ["my-key"]}'

# Sign data (data must be base64-encoded, 32 bytes for ECDSA_S256)
echo -n "hello world" | shasum -a 256 | cut -d' ' -f1 | xxd -r -p | base64
# Use the output in the sign command:
keys sign -d '{"KeyName": "my-key", "Data": "<base64-hash>"}'

# Verify a signature
keys verify -d '{"KeyType": "ECDSA_S256", "PublicKey": "<base64-public-key>", "Data": "<base64-hash>", "Signature": "<base64-signature>"}'
```

For KMS usage, set `KEYSTORE_KMS_PROFILE` instead:
```bash
export KEYSTORE_KMS_PROFILE="my-aws-profile"
keys list  # Lists KMS keys
keys sign -d '{"KeyName": "arn:aws:kms:us-west-2:123456789012:key/abc123", "Data": "<base64-hash>"}'
```

### Design Principles
- **Embeddable CLI** The cli package is designed to support 
embedding in downstream applications so a consistent CLI 
can be shared across them.
- **Typed extensibility**: Use structs for requests/responses that are extensible and easy to wrap via a network layer if needed.
- **Storage abstract**: Keystore interfaces can be implemented with memory, file, database, etc. for storage to be useable in a variety of contexts. Uses write-through caching to maintain synchronization between in-memory keys and stored keys.
- **Admin interface for mutations**: Only the Admin interface mutates the keystore; all other interfaces are read-only. Admin interface is plural/batched to support atomic batched mutations.
- **Client side key naming**: Keystore doesn't impose specific key algorithms/curves for specific contexts. It supports a minimum viable set of algorithms/curves for chainlink-wide use cases. Clients define a name for each key representing the context in which they wish to use it.
- **Common serialization/encryption**: Protobuf serialization (compact, versioned) for key material, then encrypted before persistence with a passphrase.