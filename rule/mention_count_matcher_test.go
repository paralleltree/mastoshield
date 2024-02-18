package rule_test

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/paralleltree/mastoshield/rule"
)

func TestMentionCountMatcher(t *testing.T) {
	testCreateBody := `
	{
		"id": "https://example.com/users/aaa/statuses/111950688941481832/activity",
		"type": "Create",
		"actor": "https://example.com/users/erma9t4m2c",
		"published": "2024-02-18T04:52:27Z",
		"to": [
			"https://www.w3.org/ns/activitystreams#Public"
		],
		"cc": [
			"https://example.com/users/aaa/followers",
			"https://example.com/users/test1",
			"https://example.com/users/test2",
			"https://example.com/users/test3"
		],
		"object": {
			"id": "https://example.com/users/aaa/statuses/111950688941481832",
			"type": "Note",
			"content": "",
			"tag": [
				{
					"type": "Mention",
					"href": "https://example.com/users/test1",
					"name": "@test1@example.com"
				},
				{
					"type": "Mention",
					"href": "https://example.com/users/test2",
					"name": "@test2@example.com"
				},
				{
					"type": "Mention",
					"href": "https://example.com/users/test3",
					"name": "@test3@example.com"
				}
			]
		}
	}`
	testDeleteBody := `
	{
		"@context": [
			"https://www.w3.org/ns/activitystreams"
		],
		"id": "https://example.com/users/test/statuses/111947119289823313#delete",
		"type": "Delete",
		"actor": "https://example.com/users/test",
		"to": [
			"https://www.w3.org/ns/activitystreams#Public"
		],
		"object": {
			"id": "https://example.com/users/test/statuses/111947119289823313",
			"type": "Tombstone",
			"atomUri": "https://example.com/users/test/statuses/111947119289823313"
		}
	}
	`
	cases := []struct {
		name          string
		moreThanCount int
		requestBody   string
		wantResult    bool
	}{
		{
			name:          "mention count is larger than threshold",
			moreThanCount: 2,
			requestBody:   testCreateBody,
			wantResult:    true,
		},
		{
			name:          "mention count equals to threshold",
			moreThanCount: 3,
			requestBody:   testCreateBody,
			wantResult:    false,
		},
		{
			name:          "mention count is less than threshold",
			moreThanCount: 4,
			requestBody:   testCreateBody,
			wantResult:    false,
		},
		{
			name:          "not Create Note activity",
			moreThanCount: 1,
			requestBody:   testDeleteBody,
			wantResult:    false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			m, err := rule.NewMentionCountMatcher(tt.moreThanCount)
			if err != nil {
				t.Fatalf("create matcher: %v", err)
			}

			req, err := http.NewRequest("POST", "/inbox", bytes.NewBuffer([]byte(tt.requestBody)))
			if err != nil {
				t.Fatalf("create request: %v", err)
			}

			gotResult, err := m.Test(rule.NewProxyRequest(req))
			if err != nil {
				t.Fatalf("test: %v", err)
			}
			if tt.wantResult != gotResult {
				t.Errorf("unexpected result: want %v, but got %v", tt.wantResult, gotResult)
			}
		})
	}
}
