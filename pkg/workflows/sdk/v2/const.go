package sdk

const (
	DefaultMaxResponseSizeBytes = 5 * 1024 * 1024 // 5 MB
	ResponseBufferTooSmall      = "response buffer too small"
	// proto encoder outputs a map with these keys so that user payload can be easily extracted
	ConsensusResponseMapKeyMetadata = "metadata"
	ConsensusResponseMapKeyPayload  = "payload"
)
