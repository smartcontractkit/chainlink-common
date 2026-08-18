//go:generate protoc --proto_path=../../../../../ --go_out=../../../../../ --go-grpc_out=../../../../../ --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative pkg/loop/internal/pb/keystore/keystore.proto
package keystorepb
