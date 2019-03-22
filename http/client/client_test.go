package client

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// When no User-Agent header is provided, newRequest causes the default value to be used.
func TestNewRequestDefaultsUserAgentWhenNoneIsSpecified(t *testing.T) {
	client, err := New("http://www.example.com")
	require.NoError(t, err)
	req, err := client.newRequest("my-method", "/api/test", nil)
	require.NoError(t, err)
	ua := req.Header.Get("User-Agent")
	require.Equal(t, userAgent, ua)
}

// When we provide a User-Agent header, newRequest uses this value instead of the default.
func TestNewRequestPrefersSpecifiedUserAgentHeaderToDefault(t *testing.T) {
	expUA := "Regula/Test Go"
	hO := Header("User-Agent", expUA)
	client, err := New("http://www.example.com", hO)
	require.NoError(t, err)
	req, err := client.newRequest("my-method", "/api/test", nil)
	require.NoError(t, err)
	ua := req.Header.Get("User-Agent")
	require.Equal(t, expUA, ua)
}
