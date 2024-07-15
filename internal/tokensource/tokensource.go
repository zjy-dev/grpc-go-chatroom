package tokensource

import "context"

type Auth struct {
	token string
}

func New(token string) Auth {
	return Auth{
		token: token,
	}
}

func (a Auth) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": "bearer " + a.token,
	}, nil
}

func (a Auth) RequireTransportSecurity() bool {
	return false
}
