//go:generate protoc --go_out=. --go_opt=paths=source_relative  -I../../.. -I../../../../proto_vendor/chainlink-protos/cre -I. --go_opt=Mvalues/v1/values.proto=github.com/smartcontractkit/chainlink-common/pkg/values/pb  wasm.proto
package pb
