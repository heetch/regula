package client

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	reghttp "github.com/heetch/regula/http"
	"github.com/heetch/regula/version"
	"github.com/rs/zerolog"
	"golang.org/x/net/context/ctxhttp"
)

const (
	userAgent  = "Regula/" + version.Version + " Go"
	watchDelay = 1 * time.Second
	retryDelay = 250 * time.Millisecond
	retries    = 3
)

// A Client manages communication with the Rules Engine API using HTTP.
type Client struct {
	Logger          zerolog.Logger
	WatchRetryDelay time.Duration // Time between failed watch requests. Defaults to 1s.
	RetryDelay      time.Duration // Time between failed requests retries. Defaults to 1s.
	Retries         int           // Number of retries on retriable errors.
	baseURL         *url.URL
	httpClient      *http.Client

	Headers  map[string]string
	Rulesets *RulesetService
}

// New creates an HTTP client that uses a base url to communicate with the api server.
func New(baseURL string, opts ...Option) (*Client, error) {
	var c Client
	var err error

	c.Headers = make(map[string]string)

	c.baseURL, err = url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	if !strings.HasSuffix(c.baseURL.Path, "/") {
		c.baseURL.Path += "/"
	}

	for _, opt := range opts {
		opt(&c)
	}

	if c.httpClient == nil {
		c.httpClient = http.DefaultClient
	}

	c.Logger = zerolog.New(os.Stderr).With().Timestamp().Logger()
	c.WatchRetryDelay = watchDelay
	c.RetryDelay = retryDelay
	c.Retries = retries

	c.Rulesets = &RulesetService{
		client: &c,
	}

	return &c, nil
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

	for k, v := range c.Headers {
		req.Header.Set(k, v)
	}

	// If no User-Agent header is set, then we default it.
	if _, ok := req.Header["User-Agent"]; !ok {
		req.Header.Set("User-Agent", userAgent)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	req.Header.Set("Accept", "application/json")

	return req, nil
}

func isRetriableError(err error) bool {
	if err == nil {
		return false
	}

	if _, ok := err.(net.Error); ok {
		return true
	}

	if aerr, ok := err.(*reghttp.Error); ok {
		return aerr.Response.StatusCode == http.StatusInternalServerError ||
			aerr.Response.StatusCode == http.StatusRequestTimeout
	}

	return false
}

func (c *Client) try(ctx context.Context, req *http.Request, v interface{}) (*http.Response, error) {
	return c.tryN(ctx, req, v, c.Retries)
}

func (c *Client) tryN(ctx context.Context, req *http.Request, v interface{}, times int) (resp *http.Response, err error) {
	var (
		i       int
		reqBody *bytes.Reader
	)

	if req.Body != nil {
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		err = req.Body.Close()
		if err != nil {
			return nil, err
		}

		reqBody = bytes.NewReader(body)
	}

	for {
		if reqBody != nil {
			_, err = reqBody.Seek(0, io.SeekStart)
			if err != nil {
				return nil, err
			}

			req.Body = ioutil.NopCloser(reqBody)
		}

		resp, err = c.do(ctx, req, v)
		if err == nil || !isRetriableError(err) {
			break
		}

		i++
		if i >= times {
			break
		}

		c.Logger.Debug().Err(err).Msgf("Request failed %d times, retrying in %s...", i, c.RetryDelay)
		time.Sleep(c.RetryDelay)
	}

	return resp, err
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
		var apiErr reghttp.Error

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

// Header adds a key value pair to the headers sent on each request.
func Header(k, v string) Option {
	return func(c *Client) error {
		c.Headers[k] = v
		return nil
	}
}
