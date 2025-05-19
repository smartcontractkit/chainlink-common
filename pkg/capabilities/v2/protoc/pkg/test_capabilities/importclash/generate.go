package importclash

//go:generate protoc --go_out=../../../../../.. --go_opt=paths=source_relative  --proto_path=../../../../../.. --plugin=protoc-gen-cre=../../../protoc-gen-cre --cre_out=. capabilities/v2/protoc/pkg/test_capabilities/importclash/p1/pb/import.proto
//go:generate protoc --go_out=../../../../../.. --go_opt=paths=source_relative  --proto_path=../../../../../.. --plugin=protoc-gen-cre=../../../protoc-gen-cre --cre_out=. capabilities/v2/protoc/pkg/test_capabilities/importclash/p2/pb/import.proto
//go:generate protoc --go_out=../../../../../.. --go_opt=paths=source_relative  --proto_path=../../../../../.. --plugin=protoc-gen-cre=../../../protoc-gen-cre --cre_out=. capabilities/v2/protoc/pkg/test_capabilities/importclash/clash.proto
