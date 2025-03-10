//go:generate protoc --go_out=../../../../ --go_opt=paths=source_relative  --proto_path=../../../../ capabilities/pb/capabilities.proto values/pb/values.proto workflows/wasm/legacy/pb/wasm.proto
package pb
