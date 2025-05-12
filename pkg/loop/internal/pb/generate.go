//go:generate protoc --proto_path=../../../ --go_out=../../../ --go_opt=paths=source_relative --go-grpc_out=../../../ --go-grpc_opt=paths=source_relative loop/internal/pb/relayer.proto
//go:generate protoc --proto_path=../../../ --go_out=../../../ --go_opt=paths=source_relative --go-grpc_out=../../../ --go-grpc_opt=paths=source_relative loop/internal/pb/reporting.proto
//go:generate protoc --proto_path=../../../ --go_out=../../../ --go_opt=paths=source_relative --go-grpc_out=../../../ --go-grpc_opt=paths=source_relative loop/internal/pb/median.proto
//go:generate protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative reporting_plugin_service.proto
//go:generate protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative telemetry.proto
//go:generate protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative pipeline_runner.proto
//go:generate protoc --proto_path=../../../ --go_out=../../../ --go_opt=paths=source_relative --go-grpc_out=../../../ --go-grpc_opt=paths=source_relative loop/internal/pb/contract_reader.proto

//go:generate protoc --proto_path=../../../ --go_out=../../../ --go_opt=paths=source_relative --go-grpc_out=../../../ --go-grpc_opt=paths=source_relative loop/internal/pb/codec.proto

//go:generate protoc --proto_path=../../../ --go_out=../../../  --go_opt=paths=source_relative --go-grpc_out=../../../ --go-grpc_opt=paths=source_relative loop/internal/pb/contract_writer.proto

//go:generate protoc --proto_path=../../../ --go_out=../../../ --go_opt=paths=source_relative --go-grpc_out=../../../ --go-grpc_opt=paths=source_relative loop/internal/pb/median_datasource.proto
//go:generate protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative keyvalue_store.proto
//go:generate protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative validate_config.proto
//go:generate protoc --go_out=../../../ --go_opt=paths=source_relative --go-grpc_out=../../../ --go-grpc_opt=paths=source_relative --proto_path=../../../ capabilities/pb/capabilities.proto capabilities/pb/registry.proto values/pb/values.proto loop/internal/pb/capabilities_registry.proto
package pb
