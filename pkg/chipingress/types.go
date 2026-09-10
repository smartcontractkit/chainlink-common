package chipingress

import (
	cepb "github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	ce "github.com/cloudevents/sdk-go/v2"

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress/pb"
)

// IdempotencyKeyAttr is the CloudEvent extension attribute name for a per-event idempotency key.
// Set it via the attributes map of NewEvent.
// When the event is emitted over Kafka using the CloudEvents Kafka binding, extensions become
// Kafka headers named "ce_<name>" (e.g., ce_idempotencykey), enabling downstream deduplication.
const IdempotencyKeyAttr = "idempotencykey"

// ResourceHeaderPrefix namespaces producer resource attributes sent as outgoing gRPC metadata.
// SanitizeMetadataHeaders applies it to every key it emits.
//
// It is the wire contract with chip-ingress, which forwards metadata carrying this prefix onto every
// Kafka record a request produces and emits the key unchanged. Requiring the prefix inbound and
// preserving it outbound keeps the namespace closed, which is what makes the forwarding safe: a
// client can only cause a header beginning with this prefix to be written, so a resource attribute
// cannot shadow a "ce_" header, an identity header the server derives from the verified auth token,
// or — on this side of the wire — a reserved gRPC metadata key such as the CSA auth token's.
//
// The same constant exists in chip-ingress as constants.ResourceHeaderPrefix. Duplicating it across
// repositories is deliberate, matching how authHeaderKey is already spelled in both pkg/beholder and
// pkg/chipingress; the two must stay byte-identical or forwarding silently stops.
const ResourceHeaderPrefix = "resource_"

type (
	// Cloudevents types
	CloudEvent   = ce.Event
	CloudEventPb = cepb.CloudEvent

	// Client
	ChipIngressClient              = pb.ChipIngressClient
	ChipIngress_StreamEventsClient = pb.ChipIngress_StreamEventsClient

	// Message types
	CloudEventBatch      = pb.CloudEventBatch
	EmptyRequest         = pb.EmptyRequest
	PublishErrorCode     = pb.PublishErrorCode
	PingResponse         = pb.PingResponse
	PublishOptions       = pb.PublishOptions
	PublishResponse      = pb.PublishResponse
	PublishResult        = pb.PublishResult
	PublishError         = pb.PublishError
	StreamEventsRequest  = pb.StreamEventsRequest
	StreamEventsResponse = pb.StreamEventsResponse
)
