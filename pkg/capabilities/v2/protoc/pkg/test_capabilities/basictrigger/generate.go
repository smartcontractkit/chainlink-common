package basictrigger

//go:generate protoc --go_out=../../../../../.. --go_opt=paths=source_relative  --proto_path=../../../../../.. --plugin=protoc-gen-cre=../../../protoc-gen-cre --cre_out=. capabilities/v2/protoc/pkg/test_capabilities/basictrigger/basic_trigger.proto
