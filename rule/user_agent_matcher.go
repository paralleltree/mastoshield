package rule

import "strings"

type userAgentMatcher struct {
	pattern string
}

func NewUserAgentMatcher(pattern string) (*userAgentMatcher, error) {
	return &userAgentMatcher{
		pattern: pattern,
	}, nil
}

func (m *userAgentMatcher) Test(req *ProxyRequest) (bool, error) {
	return strings.Contains(req.Request.Header.Get("User-Agent"), m.pattern), nil
}
