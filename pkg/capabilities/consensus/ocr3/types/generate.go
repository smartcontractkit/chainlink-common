//go:generate protoc --go_out=. --go_opt=paths=source_relative -I. -I../../../../../proto_vendor/chainlink-protos/cre --go_opt=Mvalues/v1/values.proto=github.com/smartcontractkit/chainlink-common/pkg/values/pb ocr3_types.proto
//go:generate protoc --go_out=. --go_opt=paths=source_relative -I. ocr3_config_types.proto
package types
