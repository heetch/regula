package etcd

import (
	"testing"

	"github.com/heetch/regula"
	"github.com/heetch/regula/rule"
	"github.com/heetch/regula/store"
	"github.com/stretchr/testify/require"
)

func TestValidation(t *testing.T) {
	t.Run("OK - ruleset name", func(t *testing.T) {
		names := []string{
			"path/to/my-ruleset",
			"path/to/my-awesome-ruleset",
			"path/to/my-123-ruleset",
			"path/to/my-ruleset-123",
		}

		for _, n := range names {
			err := validateRulesetName(n)
			require.NoError(t, err)
		}
	})

	t.Run("NOK - ruleset name", func(t *testing.T) {
		names := []string{
			"PATH/TO/MY-RULESET",
			"/path/to/my-ruleset",
			"path/to/my-ruleset/",
			"path/to//my-ruleset",
			"path/to/my_ruleset",
			"1path/to/my-ruleset",
			"path/to/my--ruleset",
		}

		for _, n := range names {
			err := validateRulesetName(n)
			require.True(t, store.IsValidationError(err))
		}
	})

	t.Run("OK - param names", func(t *testing.T) {
		names := []string{
			"a",
			"abc",
			"abc-xyz",
			"abc-123",
			"abc-123-xyz",
		}

		for _, n := range names {
			rs, _ := regula.NewBoolRuleset(
				rule.New(
					rule.BoolParam(n),
					rule.BoolValue(true),
				),
			)

			for _, r := range rs.Rules {
				params := r.Params()
				err := validateParamNames(params)
				require.NoError(t, err)
			}
		}
	})

	t.Run("NOK - param names", func(t *testing.T) {
		names := []string{
			"ABC",
			"abc-",
			"abc_",
			"abc--xyz",
			"abc_xyz",
			"0abc",
		}

		names = append(names, reservedWords...)

		for _, n := range names {
			rs, _ := regula.NewBoolRuleset(
				rule.New(
					rule.BoolParam(n),
					rule.BoolValue(true),
				),
			)

			for _, r := range rs.Rules {
				params := r.Params()
				err := validateParamNames(params)
				require.True(t, store.IsValidationError(err))
			}
		}
	})
}

// Limit should be set to 50 if the given one is <= 0 or > 100.
func TestComputeLimit(t *testing.T) {
	l := computeLimit(0)
	require.Equal(t, 50, l)
	l = computeLimit(-10)
	require.Equal(t, 50, l)
	l = computeLimit(110)
	require.Equal(t, 50, l)
	l = computeLimit(70)
	require.Equal(t, 70, l)
}
