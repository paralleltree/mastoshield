package rule

import "net/http"

type ActionType int

const (
	ACTION_ALLOW ActionType = 0
	ACTION_DENY  ActionType = 1
)

type RuleMatcher interface {
	Test(req *http.Request, reqBody func() ([]byte, error)) (bool, error)
}

type RuleSet struct {
	Action   ActionType
	Matchers []RuleMatcher
}
