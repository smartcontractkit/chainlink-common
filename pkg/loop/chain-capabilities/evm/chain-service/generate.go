//go:generate protoc --proto_path=../../../../ --go_out=../../../.. --go_opt=paths=source_relative --go-grpc_out=../../../.. --go-grpc_opt=paths=source_relative loop/chain-capabilities/evm/chain-service/evm.proto
package evmpb
