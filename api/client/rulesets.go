package client

import (
	"context"
	"net/http"
	ppath "path"
	"strconv"
	"time"

	"github.com/heetch/regula"
	"github.com/heetch/regula/api"
	"github.com/heetch/regula/rule"
	"github.com/pkg/errors"
)

// RulesetService handles communication with the ruleset related
// methods of the Regula API.
type RulesetService struct {
	client *Client
}

func (s *RulesetService) joinPath(path string) string {
	path = "./" + ppath.Join("rulesets/", path)
	if path == "./rulesets" {
		return path + "/"
	}

	return path
}

// List fetches all the rulesets starting with the given prefix.
func (s *RulesetService) List(ctx context.Context, prefix string, opt *ListOptions) (*api.Rulesets, error) {
	req, err := s.client.newRequest("GET", s.joinPath(prefix), nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("list", "")

	if opt != nil {
		if opt.Limit != 0 {
			q.Add("limit", strconv.Itoa(opt.Limit))
		}

		if opt.Continue != "" {
			q.Add("continue", opt.Continue)
		}
	}

	req.URL.RawQuery = q.Encode()

	var rl api.Rulesets

	_, err = s.client.try(ctx, req, &rl)
	return &rl, err
}

// Eval evaluates the given ruleset with the given params.
// It implements the regula.Evaluator interface and thus can be passed to the regula.Engine.
func (s *RulesetService) Eval(ctx context.Context, path string, params rule.Params) (*regula.EvalResult, error) {
	return s.EvalVersion(ctx, path, "", params)
}

// EvalVersion evaluates the given ruleset version with the given params.
// It implements the regula.Evaluator interface and thus can be passed to the regula.Engine.
func (s *RulesetService) EvalVersion(ctx context.Context, path, version string, params rule.Params) (*regula.EvalResult, error) {
	req, err := s.client.newRequest("GET", s.joinPath(path), nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("eval", "true")
	for _, k := range params.Keys() {
		v, err := params.EncodeValue(k)
		if err != nil {
			return nil, err
		}

		q.Add(k, v)
	}
	if version != "" {
		q.Add("version", version)
	}
	req.URL.RawQuery = q.Encode()

	var resp api.EvalResult

	_, err = s.client.try(ctx, req, &resp)
	if err != nil {
		return nil, err
	}

	return &regula.EvalResult{
		Value:   resp.Value,
		Version: resp.Version,
	}, nil
}

// Put creates a ruleset version on the given path.
func (s *RulesetService) Put(ctx context.Context, path string, rs *regula.Ruleset) (*api.Ruleset, error) {
	req, err := s.client.newRequest("PUT", s.joinPath(path), rs)
	if err != nil {
		return nil, err
	}

	var resp api.Ruleset

	_, err = s.client.try(ctx, req, &resp)
	return &resp, err
}

// WatchResponse contains a list of events occured on a group of rulesets.
// If an error occurs during the watching, the Err field will be populated.
type WatchResponse struct {
	Events *api.Events
	Err    error
}

// Watch watchs the given path for changes and sends the events in the returned channel.
// If revision is empty it will start to watch for changes occuring from the moment the request is performed,
// otherwise it will watch for any changes occured from the given revision.
// The given context must be used to stop the watcher.
func (s *RulesetService) Watch(ctx context.Context, prefix string, revision string) <-chan WatchResponse {
	ch := make(chan WatchResponse)

	go func() {
		defer close(ch)

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			req, err := s.client.newRequest("GET", s.joinPath(prefix), nil)
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
			_, err = s.client.do(ctx, req, &events)
			if err != nil {
				if e, ok := err.(*api.Error); ok {
					switch e.Response.StatusCode {
					case http.StatusNotFound:
						ch <- WatchResponse{Err: err}
						return
					case http.StatusInternalServerError:
						s.client.Logger.Debug().Err(err).Msg("watch request failed: internal server error")
					default:
						s.client.Logger.Error().Err(err).Int("status", e.Response.StatusCode).Msg("watch request returned unexpected status")
					}
				} else {
					switch err {
					case context.Canceled:
						fallthrough
					case context.DeadlineExceeded:
						s.client.Logger.Debug().Msg("watch context done")
						return
					default:
						s.client.Logger.Error().Err(err).Msg("watch request failed")
					}
				}

				// avoid too many requests on errors.
				time.Sleep(s.client.WatchRetryDelay)
				continue
			}

			if events.Timeout {
				s.client.Logger.Debug().Msg("watch request timed out")
				time.Sleep(s.client.WatchRetryDelay)
				continue
			}

			ch <- WatchResponse{Events: &events}
			revision = events.Revision
		}
	}()

	return ch
}
