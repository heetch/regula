package client

import (
	"context"

	"github.com/heetch/rules-engine"
	"github.com/heetch/rules-engine/rule"
)

// Getter is an in-memory getter with http prefetch.
type Getter struct {
	rulesets map[string]*rule.Ruleset
}

// NewGetter uses the given client to fetch all the rulesets from the server
// and returns a Getter that holds the results in memory.
// No subsequent round trips are performed after this function returns.
func NewGetter(ctx context.Context, client *Client) (*Getter, error) {
	ls, err := client.ListRulesets(ctx)
	if err != nil {
		return nil, err
	}

	g := Getter{
		rulesets: make(map[string]*rule.Ruleset),
	}

	for _, re := range ls {
		g.rulesets[re.Name] = re.Ruleset
	}

	return &g, nil
}

// Get returns the selected ruleset from the memory or returns rules.ErrRulesetNotFound.
func (g *Getter) Get(_ context.Context, key string, params *rule.Params) (*rule.Ruleset, error) {
	r, ok := g.rulesets[key]
	if !ok {
		return nil, rules.ErrRulesetNotFound
	}

	return r, nil
}
