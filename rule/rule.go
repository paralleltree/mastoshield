package rule

type ActionType int

const (
	ACTION_ALLOW ActionType = 0
	ACTION_DENY  ActionType = 1
)

type RuleMatcher interface {
	Test(req *ProxyRequest) (bool, error)
}

type RuleSet struct {
	Action   ActionType
	Matchers []RuleMatcher
}
