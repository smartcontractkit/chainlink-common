package http

//go:generate protoc --go_out=../../../.. --go_opt=paths=source_relative  --proto_path=../../../.. --plugin=protoc-gen-cre=../../protoc/protoc-gen-cre --cre_out=. capabilities/v2/triggers/http/http_trigger.proto
