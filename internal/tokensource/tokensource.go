package tokensource

import (
	"context"
)

// Auth is an implementation of grpc.PerRPCCredentials interface
type Auth struct {
	token string
}

// New creates a new Auth object with the given token
func New(token string) Auth {
	return Auth{
		token: token,
	}
}

// GetRequestMetadata returns the metadata for the request
func (a Auth) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {

	// Create a map to store the metadata
	metadataMap := make(map[string]string)

	// Add the authorization header to the metadata map
	metadataMap["authorization"] = "bearer " + a.token

	// Return the metadata map
	return metadataMap, nil
}

// RequireTransportSecurity returns whether or not the transport security is required
func (a Auth) RequireTransportSecurity() bool {

	// Return false as the transport security is not required
	return false
}
