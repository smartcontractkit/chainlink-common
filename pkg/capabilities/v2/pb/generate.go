//go:generate protoc --proto_path=../../.. --go_out=../../.. --go_opt=paths=source_relative --go-grpc_out=../../.. --go-grpc_opt=paths=source_relative capabilities/v2/pb/capabilities_v2.proto
package pb
