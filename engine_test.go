package regula_test

import (
	"context"
	"testing"
	"time"

	"github.com/heetch/regula"
	"github.com/heetch/regula/errors"
	"github.com/heetch/regula/rule"
	"github.com/stretchr/testify/require"
)

func TestEngine(t *testing.T) {
	ctx := context.Background()

	buf := regula.NewRulesetBuffer()

	buf.Add("match-string-a", "1", &regula.Ruleset{
		Rules: []*rule.Rule{
			rule.New(rule.Eq(rule.StringParam("foo"), rule.StringValue("bar")), rule.StringValue("matched a v1")),
		},
	})
	buf.Add("match-string-a", "2", &regula.Ruleset{
		Rules: []*rule.Rule{
			rule.New(rule.Eq(rule.StringParam("foo"), rule.StringValue("bar")), rule.StringValue("matched a v2")),
		},
	})
	buf.Add("match-string-b", "1", &regula.Ruleset{
		Rules: []*rule.Rule{
			rule.New(rule.True(), rule.StringValue("matched b")),
		},
	})
	buf.Add("type-mismatch", "1", &regula.Ruleset{
		Rules: []*rule.Rule{
			rule.New(rule.True(), &rule.Value{Type: "int", Data: "5"}),
		},
	})
	buf.Add("no-match", "1", &regula.Ruleset{
		Rules: []*rule.Rule{
			rule.New(rule.Eq(rule.StringValue("foo"), rule.StringValue("bar")), rule.StringValue("matched d")),
		},
	})
	buf.Add("match-bool", "1", &regula.Ruleset{
		Rules: []*rule.Rule{
			rule.New(rule.True(), &rule.Value{Type: "bool", Data: "true"}),
		},
	})
	buf.Add("match-int64", "1", &regula.Ruleset{
		Rules: []*rule.Rule{
			rule.New(rule.True(), &rule.Value{Type: "int64", Data: "-10"}),
		},
	})
	buf.Add("match-float64", "1", &regula.Ruleset{
		Rules: []*rule.Rule{
			rule.New(rule.True(), &rule.Value{Type: "float64", Data: "-3.14"}),
		},
	})
	buf.Add("match-duration", "1", &regula.Ruleset{
		Rules: []*rule.Rule{
			rule.New(rule.True(), rule.StringValue("3s")),
		},
	})

	e := regula.NewEngine(buf)

	t.Run("LowLevel", func(t *testing.T) {
		res, err := e.Get(ctx, "match-string-a", regula.Params{
			"foo": "bar",
		})
		require.NoError(t, err)
		require.Equal(t, "matched a v2", res.Value.Data)
		require.Equal(t, "2", res.Version)

		res, err = e.Get(ctx, "match-string-a", regula.Params{
			"foo": "bar",
		}, regula.Version("1"))
		require.NoError(t, err)
		require.Equal(t, "matched a v1", res.Value.Data)
		require.Equal(t, "1", res.Version)

		str, err := e.GetString(ctx, "match-string-b", nil)
		require.NoError(t, err)
		require.Equal(t, "matched b", str)

		b, err := e.GetBool(ctx, "match-bool", nil)
		require.NoError(t, err)
		require.True(t, b)

		i, err := e.GetInt64(ctx, "match-int64", nil)
		require.NoError(t, err)
		require.Equal(t, int64(-10), i)

		f, err := e.GetFloat64(ctx, "match-float64", nil)
		require.NoError(t, err)
		require.Equal(t, -3.14, f)

		_, err = e.GetString(ctx, "match-bool", nil)
		require.NoError(t, err)

		_, err = e.GetString(ctx, "type-mismatch", nil)
		require.NoError(t, err)

		_, err = e.GetString(ctx, "no-match", nil)
		require.Equal(t, errors.ErrNoMatch, err)

		_, err = e.GetString(ctx, "not-found", nil)
		require.Equal(t, errors.ErrRulesetNotFound, err)
	})

	t.Run("StructLoading", func(t *testing.T) {
		to := struct {
			StringA  string        `ruleset:"match-string-a"`
			Bool     bool          `ruleset:"match-bool"`
			Int64    int64         `ruleset:"match-int64"`
			Float64  float64       `ruleset:"match-float64"`
			Duration time.Duration `ruleset:"match-duration"`
		}{}

		err := e.LoadStruct(ctx, &to, regula.Params{
			"foo": "bar",
		})

		require.NoError(t, err)
		require.Equal(t, "matched a v2", to.StringA)
		require.Equal(t, true, to.Bool)
		require.Equal(t, int64(-10), to.Int64)
		require.Equal(t, -3.14, to.Float64)
		require.Equal(t, 3*time.Second, to.Duration)
	})

	t.Run("StructLoadingWrongKey", func(t *testing.T) {
		to := struct {
			StringA string `ruleset:"match-string-a,required"`
			Wrong   string `ruleset:"no-exists,required"`
		}{}

		err := e.LoadStruct(ctx, &to, regula.Params{
			"foo": "bar",
		})

		require.Error(t, err)
	})

	t.Run("StructLoadingMissingParam", func(t *testing.T) {
		to := struct {
			StringA string `ruleset:"match-string-a"`
		}{}

		err := e.LoadStruct(ctx, &to, nil)

		require.Error(t, err)
	})
}
