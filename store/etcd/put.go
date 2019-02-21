package etcd

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/gogo/protobuf/proto"
	"github.com/heetch/regula"
	"github.com/heetch/regula/rule"
	"github.com/heetch/regula/store"
	pb "github.com/heetch/regula/store/etcd/proto"
	"github.com/pkg/errors"
	"github.com/segmentio/ksuid"
)

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

			var pbrse pb.RulesetEntry
			err = proto.Unmarshal([]byte(v), &pbrse)
			if err != nil {
				s.Logger.Debug().Err(err).Str("entry", v).Msg("put: entry unmarshalling failed")
				return errors.Wrap(err, "failed to unmarshal entry")
			}

			s.Logger.Debug().Str("path", path).Msg("ruleset didn't change, returning without creating a new version")

			entry.Path = pbrse.Path
			entry.Version = pbrse.Version
			entry.Ruleset = rulesetFromProtobuf(pbrse.Ruleset)

			return store.ErrNotModified
		}

		// make sure signature didn't change
		rawSig := stm.Get(s.signaturesPath(path))
		if rawSig != "" {
			var pbsig pb.Signature

			err := proto.Unmarshal([]byte(rawSig), &pbsig)
			if err != nil {
				s.Logger.Debug().Err(err).Str("signature", rawSig).Msg("put: signature unmarshalling failed")
				return errors.Wrap(err, "failed to decode ruleset signature")
			}

			err = compareSignature(signatureFromProtobuf(&pbsig), sig)
			if err != nil {
				return err
			}
		}

		// if no signature found, create one
		if rawSig == "" {
			b, err := proto.Marshal(signatureToProtobuf(sig))
			if err != nil {
				return errors.Wrap(err, "failed to encode updated signature")
			}

			stm.Put(s.signaturesPath(path), string(b))
		}

		// update checksum
		stm.Put(s.checksumsPath(path), checksum)

		// create a new ruleset version
		k, err := ksuid.NewRandom()
		if err != nil {
			return errors.Wrap(err, "failed to generate ruleset version")
		}
		version := k.String()

		err = s.putVersions(stm, path, version)
		if err != nil {
			return err
		}

		re := store.RulesetEntry{
			Path:    path,
			Version: version,
			Ruleset: ruleset,
		}

		err = s.putEntry(stm, &re)
		if err != nil {
			return err
		}

		// update the pointer to the latest ruleset
		stm.Put(s.latestRulesetPath(path), s.entriesPath(path, version))

		entry = re
		return nil
	}

	_, err = concurrency.NewSTM(s.Client, txfn, concurrency.WithAbortContext(ctx))
	if err != nil && err != store.ErrNotModified && !store.IsValidationError(err) {
		return nil, errors.Wrap(err, "failed to put ruleset")
	}

	return &entry, err
}

func compareSignature(base, other *regula.Signature) error {
	if base.ReturnType != other.ReturnType {
		return &store.ValidationError{
			Field:  "return type",
			Value:  other.ReturnType,
			Reason: fmt.Sprintf("signature mismatch: return type must be of type %s", base.ReturnType),
		}
	}

	for name, tp := range other.ParamTypes {
		stp, ok := base.ParamTypes[name]
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

func validateRuleset(path string, rs *regula.Ruleset) (*regula.Signature, error) {
	err := validateRulesetName(path)
	if err != nil {
		return nil, err
	}

	sig := regula.NewSignature(rs)

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

// putVersions stores the new version or appends it to the existing ones under the key <namespace>/rulesets/versions/<path>.
func (s *RulesetService) putVersions(stm concurrency.STM, path, version string) error {
	var v pb.Versions

	res := stm.Get(s.versionsPath(path))
	if res != "" {
		err := proto.Unmarshal([]byte(res), &v)
		if err != nil {
			s.Logger.Debug().Err(err).Str("path", path).Msg("put: versions unmarshalling failed")
			return errors.Wrap(err, "failed to unmarshal versions")
		}
	}

	v.Versions = append(v.Versions, version)
	bvs, err := proto.Marshal(&v)
	if err != nil {
		return errors.Wrap(err, "failed to encode versions")
	}
	stm.Put(s.versionsPath(path), string(bvs))

	return nil
}

func (s *RulesetService) putEntry(stm concurrency.STM, rse *store.RulesetEntry) error {
	pbrse := pb.RulesetEntry{
		Path:    rse.Path,
		Version: rse.Version,
		Ruleset: rulesetToProtobuf(rse.Ruleset),
	}

	b, err := proto.Marshal(&pbrse)
	if err != nil {
		return errors.Wrap(err, "failed to encode entry")
	}
	stm.Put(s.entriesPath(rse.Path, rse.Version), string(b))

	return nil
}
