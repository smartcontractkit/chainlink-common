//go:generate protoc --go_out=../../../../ --go_opt=paths=source_relative  --proto_path=../../../../ capabilities/pb/capabilities.proto values/pb/values.proto workflows/sdk/v2/pb/sdk.proto
package pb
