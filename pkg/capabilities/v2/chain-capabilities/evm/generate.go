//go:generate protoc --proto_path=../../../../ --go_out=../../../.. --go_opt=paths=source_relative --plugin=protoc-gen-cre=../../../../capabilities/v2/protoc/protoc-gen-cre --cre_out=../../../../capabilities/v2/chain-capabilities/evm capabilities/v2/chain-capabilities/evm/capability.proto
package evm
