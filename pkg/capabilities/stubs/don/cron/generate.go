package cron

//go:generate protoc --go_out=../../../.. --go_opt=paths=source_relative  --proto_path=../../../.. --plugin=protoc-gen-cre=../../../v2/protoc/protoc-gen-cre --cre_out=. capabilities/stubs/don/cron/cron.proto
