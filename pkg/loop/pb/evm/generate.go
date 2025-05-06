//go:generate protoc --proto_path=.:..:../../../../ --go_out=../../../../capabilities/v2/protoc/pkg/evm --go_opt=paths=source_relative --go-grpc_out=../../../../capabilities/v2/protoc/pkg/evm --go-grpc_opt=paths=source_relative evm.proto

//go:generate protoc --proto_path=.:..:../../../../ --go_out=. --go_opt=paths=source_relative --go-grpc_out=../../../../capabilities/v2/protoc/pkg/evm --go-grpc_opt=paths=source_relative --plugin=protoc-gen-cre=../../../../capabilities/v2/protoc/protoc-gen-cre --cre_out=. evm.proto
package evmpb
