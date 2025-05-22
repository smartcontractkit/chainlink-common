//go:generate protoc --proto_path=../../../../ --go_out=../../../../ --go_opt=paths=source_relative --go-grpc_out=../../../../ --go-grpc_opt=paths=source_relative loop/internal/pb/ccip/commitstore.proto
//go:generate protoc --proto_path=../../../../ --go_out=../../../../ --go_opt=paths=source_relative --go-grpc_out=../../../../ --go-grpc_opt=paths=source_relative loop/internal/pb/ccip/gaspriceestimator.proto
//go:generate protoc --proto_path=../../../../ --go_out=../../../../ --go_opt=paths=source_relative --go-grpc_out=../../../../ --go-grpc_opt=paths=source_relative loop/internal/pb/ccip/models.proto
//go:generate protoc --proto_path=../../../../ --go_out=../../../../ --go_opt=paths=source_relative --go-grpc_out=../../../../ --go-grpc_opt=paths=source_relative loop/internal/pb/ccip/offramp.proto
//go:generate protoc --proto_path=../../../../ --go_out=../../../../ --go_opt=paths=source_relative --go-grpc_out=../../../../ --go-grpc_opt=paths=source_relative loop/internal/pb/ccip/onramp.proto
//go:generate protoc --proto_path=../../../../ --go_out=../../../../ --go_opt=paths=source_relative --go-grpc_out=../../../../ --go-grpc_opt=paths=source_relative loop/internal/pb/ccip/pricegetter.proto
//go:generate protoc --proto_path=../../../../ --go_out=../../../../ --go_opt=paths=source_relative --go-grpc_out=../../../../ --go-grpc_opt=paths=source_relative loop/internal/pb/ccip/priceregistry.proto
//go:generate protoc --proto_path=../../../../ --go_out=../../../../ --go_opt=paths=source_relative --go-grpc_out=../../../../ --go-grpc_opt=paths=source_relative loop/internal/pb/ccip/tokendata.proto
//go:generate protoc --proto_path=../../../../ --go_out=../../../../ --go_opt=paths=source_relative --go-grpc_out=../../../../ --go-grpc_opt=paths=source_relative loop/internal/pb/ccip/tokenpool.proto
//go:generate protoc --proto_path=.:.. --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative factories.proto

package ccippb
