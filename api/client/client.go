package client

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	ppath "path"
	"time"

	"github.com/heetch/regula"
	"github.com/heetch/regula/api"
	"github.com/heetch/regula/version"
	"golang.org/x/net/context/ctxhttp"
)

const (
	userAgent = "RulesEngine/" + version.Version + " Go"
	timeout   = 5 * time.Second
)

// A Client manages communication with the Rules Engine API using HTTP.
type Client struct {
	baseURL    *url.URL
	userAgent  string
	httpClient *http.Client
}

// NewClient creates an HTTP client that uses a base url to communicate with the api server.
func NewClient(baseURL string, opts ...Option) (*Client, error) {
	var c Client
	var err error

	c.baseURL, err = url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(&c)
	}

	if c.userAgent == "" {
		c.userAgent = userAgent
	}

	if c.httpClient == nil {
		c.httpClient = &http.Client{
			Timeout: timeout,
		}
	}

	return &c, nil
}

// ListRulesets fetches all the rulesets.
func (c *Client) ListRulesets(ctx context.Context, prefix string) ([]api.Ruleset, error) {
	req, err := c.newRequest("GET", ppath.Join("/rulesets/", prefix), nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("list", "")
	req.URL.RawQuery = q.Encode()

	var rl []api.Ruleset

	_, err = c.do(ctx, req, &rl)
	return rl, err
}

// EvalRuleset evaluates the given ruleset with the given params.
// Each parameter must be encoded to string before being passed to the params map.
func (c *Client) EvalRuleset(ctx context.Context, path string, params map[string]string) (*api.Value, error) {
	req, err := c.newRequest("GET", ppath.Join("/rulesets/", path), nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("eval", "")
	for k, v := range params {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()

	var resp api.Value

	_, err = c.do(ctx, req, &resp)
	return &resp, nil
}

func (c *Client) newRequest(method, path string, body interface{}) (*http.Request, error) {
	rel := url.URL{Path: path}
	u := c.baseURL.ResolveReference(&rel)

	var r io.Reader

	if body != nil {
		var buf bytes.Buffer

		err := json.NewEncoder(&buf).Encode(body)
		if err != nil {
			return nil, err
		}

		r = &buf
	}

	req, err := http.NewRequest(method, u.String(), r)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)

	return req, nil
}

func (c *Client) do(ctx context.Context, req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := ctxhttp.Do(ctx, c.httpClient, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		var apiErr api.Error
		err = dec.Decode(&apiErr)
		if err != nil {
			return resp, err
		}

		return resp, apiErr
	}

	return resp, dec.Decode(v)
}

// Option allows Client customization.
type Option func(*Client) error

// HTTPClient specifies the http client to be used.
func HTTPClient(httpClient *http.Client) Option {
	return func(c *Client) error {
		c.httpClient = httpClient
		return nil
	}
}

// UserAgent specifies which user agent to sent alongside the request.
func UserAgent(userAgent string) Option {
	return func(c *Client) error {
		c.userAgent = userAgent
		return nil
	}
}

// NewGetter uses the given client to fetch all the rulesets from the server
// and returns a Getter that holds the results in memory.
// No subsequent round trips are performed after this function returns.
func NewGetter(ctx context.Context, client *Client, prefix string) (*rules.MemoryGetter, error) {
	ls, err := client.ListRulesets(ctx, prefix)
	if err != nil {
		return nil, err
	}

	var m rules.MemoryGetter

	for _, re := range ls {
		m.AddRuleset(re.Path, re.Version, re.Ruleset)
	}

	return &m, nil
}
