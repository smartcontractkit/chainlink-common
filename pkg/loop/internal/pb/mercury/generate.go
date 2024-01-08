//go:generate protoc --proto_path=.:..:./v1:./v2:./v3 --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative mercury.proto
package mercury_pb
