package rule

import (
	"encoding/json"
	"fmt"
	"net/http"
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

func (m *noteContentMatcher) Test(req *http.Request, bodyFetcher func() ([]byte, error)) (bool, error) {
	if !strings.HasSuffix(req.URL.Path, "/inbox") {
		return false, nil
	}

	body, err := bodyFetcher()
	if err != nil {
		return false, fmt.Errorf("fetch body: %w", err)
	}

	payload := struct {
		Type   string `json:"type"`
		Object struct {
			Type    string `json:"type"`
			Content string `json:"content"`
		} `json:"object"`
	}{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return false, fmt.Errorf("unmarshal json: %w", err)
	}

	return payload.Type == "Create" &&
			(payload.Object.Type == "Note" || payload.Object.Type == "Question") &&
			strings.Contains(payload.Object.Content, m.pattern),
		nil
}
