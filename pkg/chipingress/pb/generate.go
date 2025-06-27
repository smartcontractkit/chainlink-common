//go:generate protoc --proto_path . --proto_path ../../.. --proto_path ../../../vendor --go_out . --go_opt paths=source_relative --go-grpc_out . --go-grpc_opt paths=source_relative chip_ingress.proto
package pb
