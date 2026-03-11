package token_source

import (
	"sync"

	"github.com/Motmedel/utils_go/pkg/oauth2/types/token"
)

type TokenSource interface {
	Token() (*token.Token, error)
}

type staticTokenSource struct {
	t *token.Token
}

func (s *staticTokenSource) Token() (*token.Token, error) {
	return s.t, nil
}

// NewStatic returns a TokenSource that always returns the same token.
func NewStatic(t *token.Token) TokenSource {
	return &staticTokenSource{t: t}
}

type reuseTokenSource struct {
	new TokenSource
	mu  sync.Mutex
	t   *token.Token
}

func (s *reuseTokenSource) Token() (*token.Token, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.t.Valid() {
		return s.t, nil
	}

	t, err := s.new.Token()
	if err != nil {
		return nil, err
	}

	s.t = t
	return t, nil
}

// NewReusable returns a TokenSource that caches the token from src
// and refreshes it when expired.
func NewReusable(t *token.Token, src TokenSource) TokenSource {
	return &reuseTokenSource{
		t:   t,
		new: src,
	}
}
