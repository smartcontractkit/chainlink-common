module github.com/smartcontractkit/chainlink-common

go 1.23.3

require (
	github.com/andybalholm/brotli v1.1.1
	github.com/atombender/go-jsonschema v0.16.1-0.20240916205339-a74cd4e2851c
	github.com/bytecodealliance/wasmtime-go/v23 v23.0.0
	github.com/confluentinc/confluent-kafka-go/v2 v2.3.0
	github.com/dominikbraun/graph v0.23.0
	github.com/fxamacker/cbor/v2 v2.5.0
	github.com/go-json-experiment/json v0.0.0-20231102232822-2e55bd4e08b0
	github.com/go-playground/validator/v10 v10.4.1
	github.com/go-viper/mapstructure/v2 v2.1.0
	github.com/google/go-cmp v0.6.0
	github.com/google/uuid v1.6.0
	github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus v1.0.1
	github.com/hashicorp/consul/sdk v0.16.0
	github.com/hashicorp/go-hclog v1.5.0
	github.com/hashicorp/go-plugin v1.6.2
	github.com/iancoleman/strcase v0.3.0
	github.com/invopop/jsonschema v0.12.0
	github.com/jackc/pgx/v4 v4.18.3
	github.com/jmoiron/sqlx v1.4.0
	github.com/jonboulle/clockwork v0.4.0
	github.com/jpillora/backoff v1.0.0
	github.com/lib/pq v1.10.9
	github.com/linkedin/goavro/v2 v2.12.0
	github.com/marcboeker/go-duckdb v1.8.3
	github.com/pelletier/go-toml/v2 v2.2.0
	github.com/prometheus/client_golang v1.17.0
	github.com/riferrei/srclient v0.5.4
	github.com/santhosh-tekuri/jsonschema/v5 v5.2.0
	github.com/scylladb/go-reflectx v1.0.1
	github.com/shopspring/decimal v1.4.0
	github.com/smartcontractkit/grpc-proxy v0.0.0-20240830132753-a7e17fec5ab7
	github.com/smartcontractkit/libocr v0.0.0-20241007185508-adbe57025f12
	github.com/stretchr/testify v1.9.0
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.52.0
	go.opentelemetry.io/otel v1.30.0
	go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc v0.0.0-20240823153156-2a54df7bffb9
	go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp v0.6.0
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v1.28.0
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp v1.30.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.28.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.30.0
	go.opentelemetry.io/otel/exporters/stdout/stdoutlog v0.4.0
	go.opentelemetry.io/otel/exporters/stdout/stdoutmetric v1.28.0
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.28.0
	go.opentelemetry.io/otel/log v0.6.0
	go.opentelemetry.io/otel/metric v1.30.0
	go.opentelemetry.io/otel/sdk v1.30.0
	go.opentelemetry.io/otel/sdk/log v0.6.0
	go.opentelemetry.io/otel/sdk/metric v1.30.0
	go.opentelemetry.io/otel/trace v1.30.0
	go.uber.org/goleak v1.3.0
	go.uber.org/multierr v1.11.0
	go.uber.org/zap v1.27.0
	golang.org/x/crypto v0.28.0
	golang.org/x/exp v0.0.0-20240909161429-701f63a606c0
	golang.org/x/tools v0.26.0
	gonum.org/v1/gonum v0.15.1
	google.golang.org/grpc v1.67.1
	google.golang.org/protobuf v1.35.1
	sigs.k8s.io/yaml v1.4.0
)

require (
	github.com/apache/arrow-go/v18 v18.0.0 // indirect
	github.com/bahlo/generic-list-go v0.2.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/buger/jsonparser v1.1.1 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/fatih/color v1.17.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-playground/locales v0.13.0 // indirect
	github.com/go-playground/universal-translator v0.17.0 // indirect
	github.com/goccy/go-json v0.10.3 // indirect
	github.com/goccy/go-yaml v1.12.0 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/golang/snappy v0.0.5-0.20220116011046-fa5810519dcb // indirect
	github.com/google/flatbuffers v24.3.25+incompatible // indirect
	github.com/grpc-ecosystem/go-grpc-middleware/v2 v2.1.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.22.0 // indirect
	github.com/hashicorp/yamux v0.1.1 // indirect
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgconn v1.14.3 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.3.3 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/pgtype v1.14.0 // indirect
	github.com/klauspost/compress v1.17.11 // indirect
	github.com/klauspost/cpuid/v2 v2.2.8 // indirect
	github.com/leodido/go-urn v1.2.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mr-tron/base58 v1.2.0 // indirect
	github.com/oklog/run v1.0.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.21 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/client_model v0.4.1-0.20230718164431-9a2bf3000d16 // indirect
	github.com/prometheus/common v0.44.0 // indirect
	github.com/prometheus/procfs v0.11.1 // indirect
	github.com/sanity-io/litter v1.5.5 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/wk8/go-ordered-map/v2 v2.1.8 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/zeebo/xxh3 v1.0.2 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.30.0 // indirect
	go.opentelemetry.io/proto/otlp v1.3.1 // indirect
	golang.org/x/mod v0.21.0 // indirect
	golang.org/x/net v0.30.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/sys v0.26.0 // indirect
	golang.org/x/text v0.19.0 // indirect
	golang.org/x/xerrors v0.0.0-20240903120638-7835f813f4da // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240903143218-8af14fe29dc1 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240903143218-8af14fe29dc1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
