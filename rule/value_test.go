package rule

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValue(t *testing.T) {
	v1 := NewBoolValue(true)
	require.True(t, v1.Equal(v1))
	require.True(t, v1.Equal(NewBoolValue(true)))
	require.False(t, v1.Equal(NewBoolValue(false)))
	require.False(t, v1.Equal(NewStringValue("true")))
}
