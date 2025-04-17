//go:generate protoc --go_out=../../../../ --go_opt=paths=source_relative  --proto_path=../../../../ capabilities/pb/capabilities.proto values/pb/values.proto workflows/wasm/v2/pb/wasm.proto
package pb
