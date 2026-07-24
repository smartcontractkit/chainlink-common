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

// reservedExtensionNames holds every CloudEvent extension name that NewEvent sets internally,
// plus the CloudEvents core context attribute names (id, source, type, specversion, time,
// subject, dataschema, datacontenttype) and the spec-forbidden "data" name. WithResourceAttributeExtensions
// consults this set so that a resource attribute can never silently overwrite event-lifecycle
// metadata or collide with a CloudEvents core attribute.
var reservedExtensionNames = map[string]struct{}{
	IdempotencyKeyAttr: {},
	"recordedtime":     {},
	"id":               {},
	"source":           {},
	"type":             {},
	"specversion":      {},
	"time":             {},
	"subject":          {},
	"dataschema":       {},
	"datacontenttype":  {},
	"data":             {},
}

// reservedMetadataKeys holds gRPC-reserved header names that could otherwise be reached by
// sanitizeExtensionName's [a-z0-9] sanitization. Verified against grpc-go v1.79.1's
// isReservedHeader: every other reserved header (pseudo-headers, "content-type", "grpc-*")
// contains a ':' or '-' that sanitization strips, so "te" is the only one actually reachable.
// SanitizeMetadataHeaders consults this set so that edge case is handled deterministically
// rather than relying on grpc's own (silent) handling of a reserved header.
var reservedMetadataKeys = map[string]struct{}{
	"te": {},
}

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
