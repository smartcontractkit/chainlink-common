package sdk

const (
	IdLen                       = 36              // IdLen is 36 bytes to match a UUID's string length
	DefaultMaxResponseSizeBytes = 5 * 1024 * 1024 // 5 MB
	ResponseBufferToSmall       = "response buffer too small"
)
