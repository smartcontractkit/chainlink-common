PROTO_IMPORTS := \
	Mvalues/v1/values.proto=github.com/smartcontractkit/chainlink-common/pkg/values/pb \
	paths=source_relative \

define gen
--$1_out=. \
$(foreach m,$(PROTO_IMPORTS),--$1_opt=$(m))
endef

.PHONY: generate
generate:
	protoc -I../../../../proto_vendor/chainlink-protos/cre \
		  -I../../.. \
		  -I. \
	       $(call gen,go) \
	       $(call gen,go-grpc) \
	       $(FILE)