package mismatchedpb

// Generate an example where the package name of the generated sdk does not match the directory name which can mess with the import path in mock and server gen
//go:generate protoc --go_out=../../../../../.. --go_opt=paths=source_relative  --proto_path=../../../../../.. --plugin=protoc-gen-cre=../../../protoc-gen-cre --cre_out=. capabilities/v2/protoc/pkg/test_capabilities/mismatched/mismatched.proto
