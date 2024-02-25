package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/paralleltree/mastoshield/rule"
	"github.com/rs/xid"
)

func Handler(
	upstream http.Handler, denyResponseCode int, rulesets []rule.RuleSet,
	onProcessing func(string, *http.Request), onAllowed func(string, *http.Request), onDenied func(string, *http.Request),
	onError func(string, error),
) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		reqID := xid.New().String()

		if onProcessing != nil {
			onProcessing(reqID, r)
		}

		testRequest := func(r *rule.ProxyRequest, ruleset rule.RuleSet) (bool, error) {
			for _, matcher := range ruleset.Matchers {
				matched, err := matcher.Test(r)
				if err != nil {
					return false, fmt.Errorf("test request: %w", err)
				}
				if !matched {
					return false, nil
				}
			}
			return true, nil
		}

		allowAction := func(w http.ResponseWriter, r *http.Request) {
			if onAllowed != nil {
				defer onAllowed(reqID, r)
			}
			upstream.ServeHTTP(w, r)
		}
		denyAction := func(w http.ResponseWriter, r *http.Request) {
			if onDenied != nil {
				defer onDenied(reqID, r)
			}
			w.WriteHeader(denyResponseCode)
			w.Write([]byte{})
		}
		errAction := func(w http.ResponseWriter, r *http.Request, err error) {
			if onError != nil {
				defer onError(reqID, err)
			}
			allowAction(w, r)
		}

		req := rule.NewProxyRequest(r)
		b, _ := req.Body()
		fmt.Fprintln(os.Stderr, string(b))

		for _, ruleset := range rulesets {
			matched, err := testRequest(req, ruleset)
			if err != nil {
				errAction(w, r, err)
				return
			}
			if matched {
				switch ruleset.Action {
				case rule.ACTION_ALLOW:
					allowAction(w, r)
				case rule.ACTION_DENY:
					denyAction(w, r)
				default:
					errAction(w, r, fmt.Errorf("unexpected action: %v", ruleset.Action))
				}
				return
			}
		}

		// default action is allow
		allowAction(w, r)
	}
}
