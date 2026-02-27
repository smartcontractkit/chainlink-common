// After changing the Aptos capability proto (e.g. adding the View RPC), run:
//   go generate .
// from this directory so that client.pb.go and server/client_server_gen.go are regenerated
// (same codegen path as EVM/Solana; do not hand-edit the generated server).
//go:generate go run ../../gen --pkg=github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/chain-capabilities/aptos --file=capabilities/blockchain/aptos/v1alpha/client.proto
package aptos
