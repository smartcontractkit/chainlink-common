//go:generate protoc --go_out=../../ --go_opt=paths=source_relative --go-grpc_out=../../ --go-grpc_opt=paths=source_relative --proto_path=../../ billing/pb/billing_service.proto capabilities/pb/capabilities.proto
package pb
