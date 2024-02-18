package rule

import (
	"encoding/json"
	"fmt"
	"strings"
)

type mentionCountMatcher struct {
	moreThan int
}

func NewMentionCountMatcher(moreThan int) (*mentionCountMatcher, error) {
	if moreThan < 0 {
		return nil, fmt.Errorf("invalid count: %d", moreThan)
	}
	return &mentionCountMatcher{
		moreThan: moreThan,
	}, nil
}

func (m *mentionCountMatcher) Test(req *ProxyRequest) (bool, error) {
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
			Type string `json:"type"`
			Tag  []struct {
				Type string `json:"type"`
				HRef string `json:"href"`
				Name string `json:"name"`
			} `json:"tag"`
		} `json:"object"`
	}{}

	if err := json.Unmarshal(body, &createActivityPayload); err != nil {
		return false, fmt.Errorf("unmarshal json: %w", err)
	}

	if createActivityPayload.Object.Type != "Note" {
		return false, nil
	}

	mentionCount := 0
	for _, tag := range createActivityPayload.Object.Tag {
		if tag.Type == "Mention" {
			mentionCount += 1
		}
	}

	return mentionCount > m.moreThan, nil
}
