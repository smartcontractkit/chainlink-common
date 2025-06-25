//go:generate go run ./generate
//go:generate protoc --go_out=. --go_opt=paths=source_relative -I. ocr3_config_types.proto
package types
