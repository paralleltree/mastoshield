package rule_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/paralleltree/mastoshield/rule"
)

func TestRemoteIPAddressMatcher_ForRemoteAddr(t *testing.T) {
	cases := []struct {
		name        string
		targetRange string
		remoteAddr  string
		wantResult  bool
	}{
		{
			name:        "target range contains remote addr",
			targetRange: "192.168.1.0/24",
			remoteAddr:  "192.168.1.1:30000",
			wantResult:  true,
		},
		{
			name:        "target range does not contain remote addr",
			targetRange: "192.168.1.0/24",
			remoteAddr:  "192.168.2.1:30000",
			wantResult:  false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			m, err := rule.NewRemoteIPAddressMatcher(tt.targetRange)
			if err != nil {
				t.Fatalf("create matcher: %v", err)
			}

			req, err := http.NewRequest("POST", "/", nil)
			if err != nil {
				t.Fatalf("create request: %v", err)
			}
			req.RemoteAddr = tt.remoteAddr

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

func TestRemoteIPAddressMatcher_ForXForwardedFor(t *testing.T) {
	cases := []struct {
		name        string
		targetRange string
		remoteIP    string
		wantResult  bool
	}{
		{
			name:        "target range contains remote addr",
			targetRange: "192.168.1.0/24",
			remoteIP:    "192.168.1.1",
			wantResult:  true,
		},
		{
			name:        "target range does not contain remote addr",
			targetRange: "192.168.1.0/24",
			remoteIP:    "192.168.2.1",
			wantResult:  false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			m, err := rule.NewRemoteIPAddressMatcher(tt.targetRange)
			if err != nil {
				t.Fatalf("create matcher: %v", err)
			}

			req, err := http.NewRequest("POST", "/", nil)
			if err != nil {
				t.Fatalf("create request: %v", err)
			}
			req.Header.Set("X-Forwarded-For", fmt.Sprintf("%s, %s", tt.remoteIP, "127.0.0.1"))

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
