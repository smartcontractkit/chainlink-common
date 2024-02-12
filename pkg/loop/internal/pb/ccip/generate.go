//go:generate protoc --proto_path=.:.. --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative models.proto
//go:generate protoc --proto_path=.:.. --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative onramp.proto
//go:generate protoc --proto_path=.:.. --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative offramp.proto
//go:generate protoc --proto_path=.:.. --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative commitstore.proto
//go:generate protoc --proto_path=.:.. --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative gaspriceestimator.proto

package ccippb
