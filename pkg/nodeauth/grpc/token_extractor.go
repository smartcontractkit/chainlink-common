package grpc

import (
	"context"
	"errors"
	"strings"

	"google.golang.org/grpc/metadata"
)

const (
	// AuthorizationHeader is the lowercase header key for authorization
	AuthorizationHeader = "authorization"
	// BearerPrefix is the prefix for Bearer tokens
	BearerPrefix = "Bearer "
)

var (
	ErrMissingMetadata   = errors.New("missing metadata")
	ErrMissingAuthHeader = errors.New("missing authorization header")
	ErrInvalidAuthFormat = errors.New("invalid authorization header format")
)

// ExtractBearerToken extracts a Bearer token from gRPC incoming context metadata.
// Used by servers requiring the JWT authentication this package provides.
func ExtractBearerToken(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", ErrMissingMetadata
	}

	authHeaders := md.Get(AuthorizationHeader)
	if len(authHeaders) == 0 {
		return "", ErrMissingAuthHeader
	}

	authHeader := authHeaders[0]
	if !strings.HasPrefix(authHeader, BearerPrefix) {
		return "", ErrInvalidAuthFormat
	}

	return strings.TrimPrefix(authHeader, BearerPrefix), nil
}
