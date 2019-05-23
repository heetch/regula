package etcd

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/heetch/regula"
	"github.com/heetch/regula/api"
	"github.com/heetch/regula/rule"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	_ api.RulesetService = new(RulesetService)
	_ regula.Evaluator   = new(RulesetService)
)

var (
	dialTimeout = 5 * time.Second
	endpoints   = []string{"localhost:2379", "etcd:2379"}
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func newEtcdRulesetService(t *testing.T) (*RulesetService, func()) {
	t.Helper()

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
	})
	require.NoError(t, err)

	s := RulesetService{
		Client:    cli,
		Namespace: fmt.Sprintf("regula-api-tests-%d/", rand.Int()),
	}

	return &s, func() {
		cli.Delete(context.Background(), s.Namespace, clientv3.WithPrefix())
		cli.Close()
	}
}

func createRuleset(t *testing.T, s *RulesetService, path string, rules ...*rule.Rule) *regula.Ruleset {
	t.Helper()

	_, err := s.Put(context.Background(), path, rules)
	if err != nil && err != api.ErrRulesetNotModified {
		require.NoError(t, err)
	}

	rs, err := s.Get(context.Background(), path, "")
	assert.NoError(t, err)
	return rs
}

func createBoolRuleset(t *testing.T, s *RulesetService, path string, rules ...*rule.Rule) *regula.Ruleset {
	t.Helper()

	err := s.Create(context.Background(), path, &regula.Signature{ReturnType: "bool"})
	assert.False(t, err != nil && err != api.ErrAlreadyExists)
	return createRuleset(t, s, path, rules...)
}
