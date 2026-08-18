//go:generate protoc --go_out=../../ --go_opt=paths=source_relative --proto_path=../../ beholder/pb/example.proto
//go:generate protoc --go_out=../../ --go_opt=paths=source_relative --proto_path=../../ beholder/pb/base_message.proto

package pb
