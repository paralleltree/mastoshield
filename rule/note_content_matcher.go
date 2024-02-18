package rule

import (
	"encoding/json"
	"fmt"
	"strings"
)

type noteContentMatcher struct {
	pattern string
}

func NewNoteContentMatcher(pattern string) (*noteContentMatcher, error) {
	return &noteContentMatcher{
		pattern: pattern,
	}, nil
}

func (m *noteContentMatcher) Test(req *ProxyRequest) (bool, error) {
	if !strings.HasSuffix(req.Request.URL.Path, "/inbox") {
		return false, nil
	}

	body, err := req.Body()
	if err != nil {
		return false, fmt.Errorf("fetch body: %w", err)
	}

	payload := struct {
		Type string `json:"type"`
	}{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return false, fmt.Errorf("unmarshal json: %w", err)
	}

	if payload.Type != "Create" {
		return false, nil
	}

	createActivityPayload := struct {
		Object struct {
			Type    string `json:"type"`
			Content string `json:"content"`
		} `json:"object"`
	}{}

	if err := json.Unmarshal(body, &createActivityPayload); err != nil {
		return false, fmt.Errorf("unmarshal json: %w", err)
	}

	return createActivityPayload.Object.Type == "Note" &&
			strings.Contains(createActivityPayload.Object.Content, m.pattern),
		nil
}
