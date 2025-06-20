VALUES_PKG=Mvalues/v1/values.proto=github.com/smartcontractkit/chainlink-common/pkg/values/pb
.PHONY: generate
generate:
	protoc -I../../../proto_vendor/chainlink-protos/cre \
			--go_out=. \
			--go_opt=paths=source_relative \
			--go_opt=$(VALUES_PKG) \
			values/v1/values.proto
	mv values/v1/values.pb.go values.pb.go
	rm -rf values