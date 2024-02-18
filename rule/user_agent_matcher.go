package rule

import (
	"fmt"
	"strings"
)

type userAgentMatcher struct {
	pattern string
}

func NewUserAgentMatcher(pattern string) (*userAgentMatcher, error) {
	if len(pattern) == 0 {
		return nil, fmt.Errorf("empty pattern text")
	}
	return &userAgentMatcher{
		pattern: pattern,
	}, nil
}

func (m *userAgentMatcher) Test(req *ProxyRequest) (bool, error) {
	return strings.Contains(req.Request.Header.Get("User-Agent"), m.pattern), nil
}
