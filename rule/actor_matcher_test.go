package rule_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/paralleltree/mastoshield/rule"
)

func TestActorMatcher_Test(t *testing.T) {
	cases := []struct {
		name          string
		actor         string
		prefixPattern string
		wantResult    bool
	}{
		{
			name:          "actor is under specified instance",
			actor:         "https://example.com/users/bob",
			prefixPattern: "https://example.com",
			wantResult:    true,
		},
		{
			name:          "actor is not under specified instance",
			actor:         "https://example.com/users/bob",
			prefixPattern: "https://apapap.net",
			wantResult:    false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			m, err := rule.NewActorMatcher(tt.prefixPattern)
			if err != nil {
				t.Fatalf("create matcher: %v", err)
			}

			payload := struct {
				Actor string `json:"actor"`
			}{
				Actor: tt.actor,
			}

			body, err := json.Marshal(payload)
			if err != nil {
				t.Fatalf("marshal json: %v", err)
			}

			req, err := http.NewRequest("POST", "/inbox", bytes.NewBuffer(body))
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
