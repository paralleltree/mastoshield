package rule_test

import (
	"net/http"
	"testing"

	"github.com/paralleltree/mastoshield/rule"
)

func TestUserAgentMatcher(t *testing.T) {
	cases := []struct {
		name       string
		pattern    string
		userAgent  string
		wantResult bool
	}{
		{
			name:       "user agent contains target pattern",
			userAgent:  "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36",
			pattern:    "Chrome",
			wantResult: true,
		},
		{
			name:       "user agent does not contain target pattern",
			userAgent:  "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36",
			pattern:    "xyz",
			wantResult: false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			m, err := rule.NewUserAgentMatcher(tt.pattern)
			if err != nil {
				t.Fatalf("create matcher: %v", err)
			}

			req, err := http.NewRequest("GET", "http://example.com", nil)
			if err != nil {
				t.Fatalf("create request: %v", err)
			}
			req.Header.Set("User-Agent", tt.userAgent)

			gotResult, err := m.Test(rule.NewProxyRequest(req))
			if err != nil {
				t.Fatalf("test: %v", err)
			}
			if tt.wantResult != gotResult {
				t.Errorf("unexpected test result: want %v, but got %v", tt.wantResult, gotResult)
			}
		})
	}
}
