package client

import (
	"context"
	"net/http"
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

// RoundTripCounter will count the number of times RoundTrip is
// called, and store this value in its Count member.  For each time it
// is called it will return a nil http.Response and the value of its
// Err member. It implements the http.RoundTripper interface, and thus
// can be used as the value for a http.Client.Transport.
type RoundTripCounter struct {
	Err   error
	Count int
}

// RoundTrip makes RoundTripCounter implement the http.RoundTripper interface.
func (r *RoundTripCounter) RoundTrip(req *http.Request) (*http.Response, error) {
	r.Count++
	return nil, r.Err
}

// timeoutErr returns an error that will always be treated as a timeout if returned by the http.Client.Transport.RoundTrip
type timeoutErr struct{}

// Timeout makes timeoutErr comply with the private interface url.timeout
func (te *timeoutErr) Timeout() bool {
	return true
}

// Error makes timeoutErr comply with the error interface.
func (te *timeoutErr) Error() string {
	return "net/http: request canceled while waiting for connection (Client.Timeout exceeded while awaiting headers)"
}

func TestTimeout(t *testing.T) {
	rtc := &RoundTripCounter{Err: &timeoutErr{}}

	// Note, we can achieve the effect of having this timeout by
	// simply creating a http.Client with its Timeout field set to
	// 1. However, this wouldn't allow us to count the number of
	// attempts that were made.  We specifically don't want the
	// client to retry when timeouts occur, so we need to make
	// assertions about that.  See: https://github.com/heetch/regula/issues/19
	impatientClient := HTTPClient(
		&http.Client{
			Transport: rtc})
	c, err := New("http://www.example.com", impatientClient)
	require.NoError(t, err)
	req, err := c.newRequest("foo", "/api/test", nil)
	require.NoError(t, err)
	_, err = c.tryN(context.Background(), req, nil, 3)
	// // This *must* error
	require.EqualError(t, err, "foo http://www.example.com/api/test: net/http: request canceled while waiting for connection (Client.Timeout exceeded while awaiting headers")
	// Explicitly check we don't retry
	require.Equal(t, 1, rtc.Count)
}
