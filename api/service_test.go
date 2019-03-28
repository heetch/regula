package api_test

import (
	"testing"

	"github.com/heetch/regula/api"
	"github.com/stretchr/testify/require"
)

// Limit should be set to 50 if the given one is <= 0 or > 100.
func TestListOptionsGetLimit(t *testing.T) {
	tests := []struct {
		in, out int
	}{
		{0, 50},
		{-10, 50},
		{110, 50},
		{70, 70},
	}

	for _, test := range tests {
		opt := api.ListOptions{Limit: test.in}
		require.Equal(t, test.out, opt.GetLimit())
	}
}
