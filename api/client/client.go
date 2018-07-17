package client

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	ppath "path"
	"time"

	"github.com/heetch/regula"
	"github.com/heetch/regula/api"
	"github.com/heetch/regula/rule"
	"github.com/heetch/regula/version"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"golang.org/x/net/context/ctxhttp"
)

const (
	userAgent  = "RulesEngine/" + version.Version + " Go"
	timeout    = 5 * time.Second
	watchDelay = 1 * time.Second
)

// A Client manages communication with the Rules Engine API using HTTP.
type Client struct {
	Logger     zerolog.Logger
	WatchDelay time.Duration // Time between failed watch requests. Defaults to 1s

	baseURL    *url.URL
	userAgent  string
	httpClient *http.Client
}

// New creates an HTTP client that uses a base url to communicate with the api server.
func New(baseURL string, opts ...Option) (*Client, error) {
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

	c.Logger = zerolog.New(os.Stderr).With().Timestamp().Logger()
	c.WatchDelay = watchDelay

	return &c, nil
}

// ListRulesets fetches all the rulesets starting with the given prefix.
func (c *Client) ListRulesets(ctx context.Context, prefix string) (*api.RulesetList, error) {
	req, err := c.newRequest("GET", ppath.Join("/rulesets/", prefix), nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("list", "")
	req.URL.RawQuery = q.Encode()

	var rl api.RulesetList

	_, err = c.do(ctx, req, &rl)
	return &rl, err
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
	return &resp, err
}

// PutRuleset creates a ruleset version on the given path.
func (c *Client) PutRuleset(ctx context.Context, path string, rs *rule.Ruleset) (*api.Ruleset, error) {
	req, err := c.newRequest("PUT", ppath.Join("/rulesets/", path), rs)
	if err != nil {
		return nil, err
	}

	var resp api.Ruleset

	_, err = c.do(ctx, req, &resp)
	return &resp, err
}

// WatchResponse contains a list of events occured on group of rulesets.
// If an error occurs during the watching, the Err field will be populated.
type WatchResponse struct {
	Events *api.Events
	Err    error
}

// WatchRulesets watchs the given path for changes and send the events in the returned channel.
// The given context must be used to stop the watcher.
func (c *Client) WatchRulesets(ctx context.Context, prefix string) <-chan WatchResponse {
	ch := make(chan WatchResponse)

	go func() {
		defer close(ch)

		var revision string

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			req, err := c.newRequest("GET", ppath.Join("/rulesets/", prefix), nil)
			if err != nil {
				ch <- WatchResponse{Err: errors.Wrap(err, "failed to create watch request")}
				return
			}

			q := req.URL.Query()
			q.Add("watch", "")
			if revision != "" {
				q.Add("revision", revision)
			}
			req.URL.RawQuery = q.Encode()

			var events api.Events
			_, err = c.do(ctx, req, &events)
			if err != nil {
				if e, ok := err.(*api.Error); ok {
					switch e.Response.StatusCode {
					case http.StatusNotFound:
						ch <- WatchResponse{Err: err}
						return
					case http.StatusRequestTimeout:
						c.Logger.Debug().Err(err).Msg("watch request timed out")
					case http.StatusInternalServerError:
						c.Logger.Debug().Err(err).Msg("watch request failed: internal server error")
					default:
						c.Logger.Error().Err(err).Int("status", e.Response.StatusCode).Msg("watch request returned unexpected status")
					}
				} else {
					switch err {
					case context.Canceled:
						fallthrough
					case context.DeadlineExceeded:
						c.Logger.Debug().Msg("watch context done")
					default:
						c.Logger.Error().Err(err).Msg("watch request failed")
					}
				}

				// avoid too many requests on errors.
				time.Sleep(c.WatchDelay)
				continue
			}

			ch <- WatchResponse{Events: &events}
			revision = events.Revision
		}
	}()

	return ch
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

	c.Logger.Debug().Str("url", req.URL.String()).Int("status", resp.StatusCode).Msg("request sent")

	dec := json.NewDecoder(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		var apiErr api.Error

		_ = dec.Decode(&apiErr)

		apiErr.Response = resp

		return resp, &apiErr
	}

	err = dec.Decode(v)
	if err != nil && err != io.EOF {
		return nil, err
	}

	return resp, nil
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

	for _, re := range ls.Rulesets {
		m.AddRuleset(re.Path, re.Version, re.Ruleset)
	}

	return &m, nil
}
