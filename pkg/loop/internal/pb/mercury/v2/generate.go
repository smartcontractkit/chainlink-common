// NOTE: the relative paths in the proto_path are to ensure we find common utilities, like BigInt
//go:generate protoc --proto_path=.:../.. --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative reportcodec.proto

package mercury_v2_pb
