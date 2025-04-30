package actionandtrigger

//go:generate protoc --go_out=../../../../../.. --go_opt=paths=source_relative  --proto_path=../../../../../.. --plugin=protoc-gen-cre=../../../protoc-gen-cre --cre_out=. capabilities/v2/protoc/pkg/test_capabilities/actionandtrigger/action_and_trigger.proto

//go:generate protoc --go_out=../../../../../.. --go_opt=paths=source_relative  --proto_path=../../../../../.. --plugin=protoc-gen-cre=../../../protoc-gen-cre --cre_out=. loop/internal/pb/evm/evm.proto
