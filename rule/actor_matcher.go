package rule

import (
	"encoding/json"
	"fmt"
	"strings"
)

type actorMatcher struct {
	prefix string
}

func NewActorMatcher(prefix string) (*actorMatcher, error) {
	if len(prefix) == 0 {
		return nil, fmt.Errorf("empty prefix pattern: %s", prefix)
	}
	return &actorMatcher{
		prefix: prefix,
	}, nil
}

func (m *actorMatcher) Test(req *ProxyRequest) (bool, error) {
	if !strings.HasSuffix(req.Request.URL.Path, "/inbox") {
		return false, nil
	}

	body, err := req.Body()
	if err != nil {
		return false, fmt.Errorf("read body: %w", err)
	}

	payload := struct {
		Actor string `json:"actor"`
	}{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return false, fmt.Errorf("unmarshal json: %w", err)
	}

	return strings.HasPrefix(payload.Actor, m.prefix), nil
}
