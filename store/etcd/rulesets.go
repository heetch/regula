package etcd

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"path"
	"regexp"
	"strconv"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/heetch/regula"
	"github.com/heetch/regula/rule"
	"github.com/heetch/regula/store"
	"github.com/pkg/errors"
	"github.com/segmentio/ksuid"
)

// RulesetService manages the rulesets using etcd.
type RulesetService struct {
	Client    *clientv3.Client
	Namespace string
}

// List returns all the rulesets entries under the given prefix.
func (s *RulesetService) List(ctx context.Context, prefix string) (*store.RulesetEntries, error) {
	resp, err := s.Client.KV.Get(ctx, s.rulesetPath(prefix, ""), clientv3.WithPrefix())
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch all entries")
	}

	// if a prefix is provided it must always return results
	// otherwise it doesn't exist.
	if resp.Count == 0 && prefix != "" {
		return nil, store.ErrNotFound
	}

	var entries store.RulesetEntries
	entries.Revision = strconv.FormatInt(resp.Header.Revision, 10)
	entries.Entries = make([]store.RulesetEntry, len(resp.Kvs))
	for i, pair := range resp.Kvs {
		err = json.Unmarshal(pair.Value, &entries.Entries[i])
		if err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal entry")
		}
	}

	return &entries, nil
}

// Latest returns the latest version of the ruleset entry which corresponds to the given path.
// It returns store.ErrNotFound if the path doesn't exist or if it's not a ruleset.
func (s *RulesetService) Latest(ctx context.Context, path string) (*store.RulesetEntry, error) {
	if path == "" {
		return nil, store.ErrNotFound
	}

	resp, err := s.Client.KV.Get(ctx, s.rulesetPath(path, "")+"/", clientv3.WithLastKey()...)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to fetch the entry: %s", path)
	}

	// Count will be 0 if the path doesn't exist or if it's not a ruleset.
	if resp.Count == 0 {
		return nil, store.ErrNotFound
	}

	var entry store.RulesetEntry
	err = json.Unmarshal(resp.Kvs[0].Value, &entry)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal entry")
	}

	return &entry, nil
}

// OneByVersion returns the ruleset entry which corresponds to the given path at the given version.
// It returns store.ErrNotFound if the path doesn't exist or if it's not a ruleset.
func (s *RulesetService) OneByVersion(ctx context.Context, path, version string) (*store.RulesetEntry, error) {
	if path == "" {
		return nil, store.ErrNotFound
	}

	resp, err := s.Client.KV.Get(ctx, s.rulesetPath(path, version))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to fetch the entry: %s", path)
	}

	// Count will be 0 if the path doesn't exist or if it's not a ruleset.
	if resp.Count == 0 {
		return nil, store.ErrNotFound
	}

	var entry store.RulesetEntry
	err = json.Unmarshal(resp.Kvs[0].Value, &entry)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal entry")
	}

	return &entry, nil
}

// Put adds a version of the given ruleset using an uuid.
func (s *RulesetService) Put(ctx context.Context, path string, ruleset *regula.Ruleset) (*store.RulesetEntry, error) {
	err := validateRulesetName(path)
	if err != nil {
		return nil, err
	}

	err = validateParamNames(ruleset)
	if err != nil {
		return nil, err
	}

	// generate checksum from the ruleset for comparison purpose
	h := md5.New()
	err = json.NewEncoder(h).Encode(ruleset)
	if err != nil {
		return nil, err
	}
	checksum := string(h.Sum(nil))

	k, err := ksuid.NewRandom()
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate version")
	}
	v := k.String()

	re := store.RulesetEntry{
		Path:    path,
		Version: v,
		Ruleset: ruleset,
	}

	raw, err := json.Marshal(&re)
	if err != nil {
		return nil, errors.Wrap(err, "failed to encode entry")
	}

	resp, err := s.Client.KV.Txn(ctx).
		If(
			// if last stored checksum equal the current one
			clientv3.Compare(clientv3.Value(s.checksumPath(path)), "=", checksum),
		).
		Then(
			// the latest ruleset is the same as this one
			// we fetch the latest ruleset entry
			clientv3.OpGet(s.rulesetPath(path, "")+"/", clientv3.WithLastKey()...),
		).
		Else(
			// if the checksum is different from the last one OR
			// if the checksum doesn't exist (first time we create a ruleset for this path)
			// we create a new version
			clientv3.OpPut(s.rulesetPath(path, v), string(raw)),
			// we create/update the checksum for the last ruleset version
			clientv3.OpPut(s.checksumPath(path), checksum),
		).
		Commit()
	if err != nil {
		return nil, errors.Wrap(err, "failed to put entry")
	}

	// the checksum is the same as the last ruleset saved, we return an error
	// and the entry
	if resp.Succeeded {
		if len(resp.Responses) == 0 || resp.Responses[0].GetResponseRange().Count == 0 {
			return nil, errors.New("succeeded txn response not received")
		}

		var entry store.RulesetEntry
		err = json.Unmarshal(resp.Responses[0].GetResponseRange().Kvs[0].Value, &entry)
		if err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal entry")
		}

		return &entry, store.ErrNotModified
	}

	return &re, nil
}

// regex used to validate ruleset name.
var rgxRuleset = regexp.MustCompile(`^[a-z]+(?:[a-z0-9-\/]?[a-z0-9])*$`)

func validateRulesetName(path string) error {
	if !rgxRuleset.MatchString(path) {
		return &store.ValidationError{
			Field:  "path",
			Value:  path,
			Reason: "invalid format",
		}
	}

	return nil
}

// operandsGetter is used to check if a type implements it,
// if so, we can retrieve the operands.
type operandsGetter interface {
	Operands() []rule.Expr
}

// regex used to validate parameters name.
var rgxParam = regexp.MustCompile(`^[a-z]+(?:[a-z0-9-]?[a-z0-9])*$`)

// list of reserved words that shouldn't be used as parameters.
var reservedWords = []string{
	"version",
	"list",
	"eval",
	"watch",
	"revision",
}

func validateParamNames(rs *regula.Ruleset) error {
	// fn will run recursively through the tree of Expr until it finds a rule.Param to validate it.
	var fn func(expr rule.Expr) error

	fn = func(expr rule.Expr) error {
		if r, ok := expr.(*rule.Rule); ok {
			err := fn(r.Expr)
			if err != nil {
				return err
			}
		}

		if o, ok := expr.(operandsGetter); ok {
			ops := o.Operands()
			for _, op := range ops {
				err := fn(op)
				if err != nil {
					return err
				}
			}
		}

		if p, ok := expr.(*rule.Param); ok {
			if !rgxParam.MatchString(p.Name) {
				return &store.ValidationError{
					Field:  "param",
					Value:  p.Name,
					Reason: "invalid format",
				}
			}

			for _, w := range reservedWords {
				if p.Name == w {
					return &store.ValidationError{
						Field:  "param",
						Value:  p.Name,
						Reason: "forbidden value",
					}
				}
			}
		}

		return nil
	}

	for _, r := range rs.Rules {
		err := fn(r.Expr)
		if err != nil {
			return err
		}
	}

	return nil
}

// Watch the given prefix for anything new.
func (s *RulesetService) Watch(ctx context.Context, prefix string, revision string) (*store.RulesetEvents, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	opts := []clientv3.OpOption{clientv3.WithPrefix()}
	if i, _ := strconv.ParseInt(revision, 10, 64); i > 0 {
		// watch from the next revision
		opts = append(opts, clientv3.WithRev(i+1))
	}

	wc := s.Client.Watch(ctx, s.rulesetPath(prefix, ""), opts...)
	for {
		select {
		case wresp := <-wc:
			if err := wresp.Err(); err != nil {
				return nil, errors.Wrapf(err, "failed to watch prefix: '%s'", prefix)
			}

			if len(wresp.Events) == 0 {
				continue
			}

			events := make([]store.RulesetEvent, len(wresp.Events))
			for i, ev := range wresp.Events {
				switch ev.Type {
				case mvccpb.PUT:
					events[i].Type = store.RulesetPutEvent
				}

				var e store.RulesetEntry
				err := json.Unmarshal(ev.Kv.Value, &e)
				if err != nil {
					return nil, errors.Wrap(err, "failed to unmarshal entry")
				}
				events[i].Path = e.Path
				events[i].Ruleset = e.Ruleset
				events[i].Version = e.Version
			}

			return &store.RulesetEvents{
				Events:   events,
				Revision: strconv.FormatInt(wresp.Header.Revision, 10),
			}, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

}

// Eval evaluates a ruleset given a path and a set of parameters. It implements the regula.Evaluator interface.
func (s *RulesetService) Eval(ctx context.Context, path string, params rule.Params) (*regula.EvalResult, error) {
	re, err := s.Latest(ctx, path)
	if err != nil {
		if err == store.ErrNotFound {
			return nil, regula.ErrRulesetNotFound
		}

		return nil, err
	}

	v, err := re.Ruleset.Eval(params)
	if err != nil {
		return nil, err
	}

	return &regula.EvalResult{
		Value:   v,
		Version: re.Version,
	}, nil
}

// EvalVersion evaluates a ruleset given a path and a set of parameters. It implements the regula.Evaluator interface.
func (s *RulesetService) EvalVersion(ctx context.Context, path, version string, params rule.Params) (*regula.EvalResult, error) {
	re, err := s.OneByVersion(ctx, path, version)
	if err != nil {
		if err == store.ErrNotFound {
			return nil, regula.ErrRulesetNotFound
		}

		return nil, err
	}

	v, err := re.Ruleset.Eval(params)
	if err != nil {
		return nil, err
	}

	return &regula.EvalResult{
		Value:   v,
		Version: re.Version,
	}, nil
}

func (s *RulesetService) rulesetPath(p, v string) string {
	return path.Join(s.Namespace, "rulesets", p, v)
}

func (s *RulesetService) checksumPath(p string) string {
	return path.Join(s.Namespace, "checksums", p)
}
