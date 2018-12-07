package etcd

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/heetch/regula"
	"github.com/heetch/regula/rule"
	"github.com/heetch/regula/store"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
)

// RulesetService manages the rulesets using etcd.
type RulesetService struct {
	Client    *clientv3.Client
	Logger    zerolog.Logger
	Namespace string
}

func computeLimit(l int) int {
	if l <= 0 || l > 100 {
		return 50 // TODO: make this one configurable
	}
	return l
}

// List returns the rulesets entries under the given prefix.  If pathsOnly is set to true, only the rulesets paths will be returned.
// If the prefix is empty it returns entries from the beginning following the lexical ordering.
// If the given limit is lower or equal to 0 or greater than 100, it returns 50 entries.
func (s *RulesetService) List(ctx context.Context, prefix string, limit int, continueToken string, pathsOnly bool) (*store.RulesetEntries, error) {
	options := make([]clientv3.OpOption, 0, 3)

	var key string

	limit = computeLimit(limit)

	if continueToken != "" {
		lastPath, err := base64.URLEncoding.DecodeString(continueToken)
		if err != nil {
			return nil, store.ErrInvalidContinueToken
		}

		key = string(lastPath)

		var rangeEnd string
		if pathsOnly {
			rangeEnd = clientv3.GetPrefixRangeEnd(s.latestRulesetPath(prefix))
		} else {
			rangeEnd = clientv3.GetPrefixRangeEnd(s.rulesetsPath(prefix, ""))
		}
		options = append(options, clientv3.WithRange(rangeEnd))
	} else {
		key = prefix
		options = append(options, clientv3.WithPrefix())
	}

	options = append(options, clientv3.WithLimit(int64(limit)))

	if pathsOnly {
		return s.listPaths(ctx, key, prefix, limit, options)
	}
	return s.listRulesets(ctx, key, prefix, limit, options)
}

func (s *RulesetService) listPaths(ctx context.Context, key, prefix string, limit int, opts []clientv3.OpOption) (*store.RulesetEntries, error) {
	opts = append(opts, clientv3.WithKeysOnly())
	resp, err := s.Client.KV.Get(ctx, s.latestRulesetPath(key), opts...)
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
	for _, pair := range resp.Kvs {
		p := strings.TrimPrefix(string(pair.Key), s.latestRulesetPath("")+"/")
		entries.Entries = append(entries.Entries, store.RulesetEntry{Path: p})
	}

	if len(entries.Entries) < limit || !resp.More {
		return &entries, nil
	}

	lastEntry := entries.Entries[len(entries.Entries)-1]

	// we want to start immediately after the last key
	entries.Continue = base64.URLEncoding.EncodeToString([]byte(lastEntry.Path + "\x00"))

	return &entries, nil
}

func (s *RulesetService) listRulesets(ctx context.Context, key, prefix string, limit int, opts []clientv3.OpOption) (*store.RulesetEntries, error) {
	resp, err := s.Client.KV.Get(ctx, s.rulesetsPath(key, ""), opts...)
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
			s.Logger.Debug().Err(err).Bytes("entry", pair.Value).Msg("list: unmarshalling failed")
			return nil, errors.Wrap(err, "failed to unmarshal entry")
		}
	}

	if len(entries.Entries) < limit || !resp.More {
		return &entries, nil
	}

	lastEntry := entries.Entries[len(entries.Entries)-1]

	// we want to start immediately after the last key
	entries.Continue = base64.URLEncoding.EncodeToString([]byte(path.Join(lastEntry.Path, lastEntry.Version+"\x00")))

	return &entries, nil
}

// Latest returns the latest version of the ruleset entry which corresponds to the given path.
// It returns store.ErrNotFound if the path doesn't exist or if it's not a ruleset.
func (s *RulesetService) Latest(ctx context.Context, path string) (*store.RulesetEntry, error) {
	if path == "" {
		return nil, store.ErrNotFound
	}

	resp, err := s.Client.KV.Get(ctx, s.rulesetsPath(path, "")+"/", clientv3.WithLastKey()...)
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
		s.Logger.Debug().Err(err).Bytes("entry", resp.Kvs[0].Value).Msg("latest: unmarshalling failed")
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

	resp, err := s.Client.KV.Get(ctx, s.rulesetsPath(path, version))
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
		s.Logger.Debug().Err(err).Bytes("entry", resp.Kvs[0].Value).Msg("one-by-version: unmarshalling failed")
		return nil, errors.Wrap(err, "failed to unmarshal entry")
	}

	return &entry, nil
}

// Put adds a version of the given ruleset using an uuid.
func (s *RulesetService) Put(ctx context.Context, path string, ruleset *regula.Ruleset) (*store.RulesetEntry, error) {
	sig, err := validateRuleset(path, ruleset)
	if err != nil {
		return nil, err
	}

	var entry store.RulesetEntry

	txfn := func(stm concurrency.STM) error {
		// generate a checksum from the ruleset for comparison purpose
		h := md5.New()
		err = json.NewEncoder(h).Encode(ruleset)
		if err != nil {
			return errors.Wrap(err, "failed to generate checksum")
		}
		checksum := string(h.Sum(nil))

		// if nothing changed return latest ruleset
		if stm.Get(s.checksumsPath(path)) == checksum {
			v := stm.Get(stm.Get(s.latestRulesetPath(path)))

			err = json.Unmarshal([]byte(v), &entry)
			if err != nil {
				s.Logger.Debug().Err(err).Str("entry", v).Msg("put: entry unmarshalling failed")
				return errors.Wrap(err, "failed to unmarshal entry")
			}

			s.Logger.Debug().Str("path", path).Msg("cannot save the rulesets: it already exists")

			return store.ErrNotModified
		}

		// make sure signature didn't change
		rawSig := stm.Get(s.signaturesPath(path))
		if rawSig != "" {
			var curSig signature
			err := json.Unmarshal([]byte(rawSig), &curSig)
			if err != nil {
				s.Logger.Debug().Err(err).Str("signature", rawSig).Msg("put: signature unmarshalling failed")
				return errors.Wrap(err, "failed to decode ruleset signature")
			}

			err = curSig.matchWith(sig)
			if err != nil {
				return err
			}
		}

		// if no signature found, create one
		if rawSig == "" {
			v, err := json.Marshal(&sig)
			if err != nil {
				return errors.Wrap(err, "failed to encode updated signature")
			}

			stm.Put(s.signaturesPath(path), string(v))
		}

		// update checksum
		stm.Put(s.checksumsPath(path), checksum)

		// create a new ruleset version
		k, err := ksuid.NewRandom()
		if err != nil {
			return errors.Wrap(err, "failed to generate ruleset version")
		}
		version := k.String()

		re := store.RulesetEntry{
			Path:    path,
			Version: version,
			Ruleset: ruleset,
		}

		raw, err := json.Marshal(&re)
		if err != nil {
			return errors.Wrap(err, "failed to encode entry")
		}

		stm.Put(s.rulesetsPath(path, version), string(raw))

		// update the pointer to the latest ruleset
		stm.Put(s.latestRulesetPath(path), s.rulesetsPath(path, version))

		entry = re
		return nil
	}

	_, err = concurrency.NewSTM(s.Client, txfn, concurrency.WithAbortContext(ctx))
	if err != nil && err != store.ErrNotModified && !store.IsValidationError(err) {
		return nil, errors.Wrap(err, "failed to put ruleset")
	}

	return &entry, err
}

type signature struct {
	ReturnType string
	ParamTypes map[string]string
}

func newSignature(rs *regula.Ruleset) *signature {
	pt := make(map[string]string)
	for _, p := range rs.Params() {
		pt[p.Name] = p.Type
	}

	return &signature{
		ParamTypes: pt,
		ReturnType: rs.Type,
	}
}

func (s *signature) matchWith(other *signature) error {
	if s.ReturnType != other.ReturnType {
		return &store.ValidationError{
			Field:  "return type",
			Value:  other.ReturnType,
			Reason: fmt.Sprintf("signature mismatch: return type must be of type %s", s.ReturnType),
		}
	}

	for name, tp := range other.ParamTypes {
		stp, ok := s.ParamTypes[name]
		if !ok {
			return &store.ValidationError{
				Field:  "param",
				Value:  name,
				Reason: "signature mismatch: unknown parameter",
			}
		}

		if tp != stp {
			return &store.ValidationError{
				Field:  "param type",
				Value:  tp,
				Reason: fmt.Sprintf("signature mismatch: param must be of type %s", stp),
			}
		}
	}

	return nil
}

func validateRuleset(path string, rs *regula.Ruleset) (*signature, error) {
	err := validateRulesetName(path)
	if err != nil {
		return nil, err
	}

	sig := newSignature(rs)

	for _, r := range rs.Rules {
		params := r.Params()
		err = validateParamNames(params)
		if err != nil {
			return nil, err
		}
	}

	return sig, nil
}

// regex used to validate ruleset names.
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

func validateParamNames(params []rule.Param) error {
	for i := range params {
		if !rgxParam.MatchString(params[i].Name) {
			return &store.ValidationError{
				Field:  "param",
				Value:  params[i].Name,
				Reason: "invalid format",
			}
		}

		for _, w := range reservedWords {
			if params[i].Name == w {
				return &store.ValidationError{
					Field:  "param",
					Value:  params[i].Name,
					Reason: "forbidden value",
				}
			}
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

	wc := s.Client.Watch(ctx, s.rulesetsPath(prefix, ""), opts...)
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
				default:
					s.Logger.Debug().Str("type", string(ev.Type)).Msg("watch: ignoring event type")
					continue
				}

				var e store.RulesetEntry
				err := json.Unmarshal(ev.Kv.Value, &e)
				if err != nil {
					s.Logger.Debug().Bytes("entry", ev.Kv.Value).Msg("watch: unmarshalling failed")
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

func (s *RulesetService) rulesetsPath(p, v string) string {
	return path.Join(s.Namespace, "rulesets", "entries", p, v)
}

func (s *RulesetService) checksumsPath(p string) string {
	return path.Join(s.Namespace, "rulesets", "checksums", p)
}

func (s *RulesetService) signaturesPath(p string) string {
	return path.Join(s.Namespace, "rulesets", "signatures", p)
}

func (s *RulesetService) latestRulesetPath(p string) string {
	return path.Join(s.Namespace, "rulesets", "latest", p)
}
