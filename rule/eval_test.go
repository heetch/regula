package rule

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEval(t *testing.T) {
	tests := []struct {
		node Node
		ctx  map[string]string
	}{
		{Eq(ValStr("foo"), ValStr("foo")), nil},
		{Eq(ValStr("foo"), VarStr("bar")), map[string]string{"bar": "foo"}},
		{
			Eq(
				Eq(ValStr("bar"), ValStr("bar")),
				Eq(ValStr("foo"), ValStr("foo")),
			),
			nil,
		},
	}

	for _, test := range tests {
		r, err := New(test.node, ReturnsStr("matched"))
		require.NoError(t, err)
		res, err := r.Eval(test.ctx)
		require.NoError(t, err)
		require.Equal(t, "matched", res.Value)
		require.Equal(t, "string", res.Type)
	}
}
