package config

import (
	"fmt"
	"io"
	"strings"

	"github.com/paralleltree/mastoshield/rule"
	"gopkg.in/yaml.v3"
)

type accessControlConfig struct {
	RuleSets []ruleSetConfig `yaml:"rulesets"`
}

type ruleSetConfig struct {
	Action string       `yaml:"action"`
	Rules  []ruleConfig `yaml:"rules"`
}

type ruleConfig struct {
	Source     string `yaml:"source"`
	Contains   string `yaml:"contains"`
	StartsWith string `yaml:"starts_with"`
	MoreThan   int    `yaml:"more_than"`
}

func LoadAccessControlConfig(f io.Reader) ([]rule.RuleSet, error) {
	body, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("read stream: %w", err)
	}

	configBody := accessControlConfig{}
	if err := yaml.Unmarshal(body, &configBody); err != nil {
		return nil, fmt.Errorf("unmarshal yaml: %w", err)
	}

	rulesets, err := buildRuleSets(configBody.RuleSets)
	if err != nil {
		return nil, fmt.Errorf("build rule sets: %w", err)
	}
	if err := validateRuleSets(rulesets); err != nil {
		return nil, fmt.Errorf("validate rule sets: %w", err)
	}
	return rulesets, nil
}

func buildRuleSets(rulesetsConfig []ruleSetConfig) ([]rule.RuleSet, error) {
	rulesets := make([]rule.RuleSet, 0, len(rulesetsConfig))
	for _, rulesetConfig := range rulesetsConfig {
		ruleset := rule.RuleSet{}

		switch strings.ToLower(rulesetConfig.Action) {
		case "allow":
			ruleset.Action = rule.ACTION_ALLOW
		case "deny":
			ruleset.Action = rule.ACTION_DENY
		default:
			return nil, fmt.Errorf("unexpected action type: %s", rulesetConfig.Action)
		}

		for _, ruleConfig := range rulesetConfig.Rules {
			matcher, err := buildRuleMatcher(ruleConfig)
			if err != nil {
				return nil, fmt.Errorf("build rule matcher: %w", err)
			}
			ruleset.Matchers = append(ruleset.Matchers, matcher)
		}
		rulesets = append(rulesets, ruleset)
	}
	return rulesets, nil
}

func buildRuleMatcher(ruleConfig ruleConfig) (rule.RuleMatcher, error) {
	switch strings.ToLower(ruleConfig.Source) {
	case "note_body":
		return rule.NewNoteContentMatcher(ruleConfig.Contains)
	case "mention_count":
		return rule.NewMentionCountMatcher(ruleConfig.MoreThan)
	case "actor":
		return rule.NewActorMatcher(ruleConfig.StartsWith)
	case "user_agent", "useragent":
		return rule.NewUserAgentMatcher(ruleConfig.Contains)
	case "remote_ip":
		return rule.NewRemoteIPAddressMatcher(ruleConfig.Contains) // Containedが適当な気はするけど...
	}
	return nil, fmt.Errorf("no matcher resolved: %s", ruleConfig.Source)
}

func validateRuleSets(rulesets []rule.RuleSet) error {
	for _, ruleset := range rulesets {
		if len(ruleset.Matchers) == 0 {
			return fmt.Errorf("empty matchers in ruleset")
		}
	}
	return nil
}
