package etcd

import (
	"context"
	"crypto/md5"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/golang/protobuf/proto"
	"github.com/heetch/regula"
	"github.com/heetch/regula/api"
	pb "github.com/heetch/regula/api/etcd/proto"
	"github.com/heetch/regula/rule"
	"github.com/pkg/errors"
	"github.com/segmentio/ksuid"
)

// Put stores the given rules under the rules tree. If no signature is found for the given path it returns an error.
func (s *RulesetService) Put(ctx context.Context, path string, rules []*rule.Rule) (*regula.Ruleset, error) {
	var ruleset *regula.Ruleset
	var err error

	txfn := func(stm concurrency.STM) error {
		p := rulesPutter{s, stm}
		ruleset, err = p.put(ctx, path, rules)
		return err
	}

	_, err = concurrency.NewSTM(s.Client, txfn, concurrency.WithAbortContext(ctx))
	if err != nil && err != api.ErrRulesetNotModified && !api.IsValidationError(err) {
		return nil, errors.Wrap(err, "failed to put ruleset")
	}

	return ruleset, err
}

// rulesPutter is responsible for validating and storing rules, updating checksums and other actions
// that are required in order to add a new ruleset version correctly.
type rulesPutter struct {
	s   *RulesetService
	stm concurrency.STM
}

func (p *rulesPutter) put(ctx context.Context, path string, rules []*rule.Rule) (*regula.Ruleset, error) {
	var err error

	ruleset := regula.Ruleset{
		Path:  path,
		Rules: rules,
	}

	// validate the rules
	ruleset.Signature, err = p.validateRules(p.stm, path, rules)
	if err != nil {
		return nil, err
	}

	// encode rules
	data, err := proto.Marshal(rulesToProtobuf(rules))
	if err != nil {
		return nil, err
	}

	// update checksum if rules have changed
	changed, err := p.updateChecksum(p.stm, path, data)
	if err != nil {
		return nil, err
	}

	if !changed {
		// fetch latest version string
		_, ruleset.Version = p.s.pathVersionFromKey((p.stm.Get(p.s.latestVersionPath(path))))

		return &ruleset, api.ErrRulesetNotModified
	}

	// create a new version of the ruleset
	ruleset.Version, err = p.createNewVersion(p.stm, path, data)
	if err != nil {
		return nil, err
	}

	// update the pointer to the latest ruleset version
	p.stm.Put(p.s.latestVersionPath(path), p.s.rulesPath(path, ruleset.Version))

	return &ruleset, p.updateVersionRegistry(p.stm, path, ruleset.Version)
}

// validateRules fetches the signature from the store and validates all the rules against it.
// if the rules are valid, it returns the signature.
func (p *rulesPutter) validateRules(stm concurrency.STM, path string, rules []*rule.Rule) (*regula.Signature, error) {
	data := stm.Get(p.s.signaturesPath(path))
	if data == "" {
		return nil, api.ErrSignatureNotFound
	}

	var pbsig pb.Signature
	err := proto.Unmarshal([]byte(data), &pbsig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode signature")
	}

	sig := signatureFromProtobuf(&pbsig)
	for _, r := range rules {
		if err := api.ValidateRule(sig, r); err != nil {
			return nil, err
		}
	}

	return sig, nil
}

// updateChecksum generates a checksum from the given data and stores it if it has changed.
// It returns a boolean that is true if the checksum has changed.
func (p *rulesPutter) updateChecksum(stm concurrency.STM, path string, data []byte) (bool, error) {
	// generate a checksum from the rules for comparison purpose
	h := md5.New()
	_, err := h.Write(data)
	if err != nil {
		return false, errors.Wrap(err, "failed to generate checksum")
	}

	checksum := string(h.Sum(nil))

	if stm.Get(p.s.checksumsPath(path)) == checksum {
		return false, nil
	}

	// update checksum
	stm.Put(p.s.checksumsPath(path), checksum)

	return true, nil
}

// createNewVersion adds a new entry under <namespace>/rulesets/rules/<path>/<version>.
func (p *rulesPutter) createNewVersion(stm concurrency.STM, path string, data []byte) (string, error) {
	// create a new ruleset version
	k, err := ksuid.NewRandom()
	if err != nil {
		return "", errors.Wrap(err, "failed to generate rules version")
	}
	version := k.String()

	stm.Put(p.s.rulesPath(path, version), string(data))

	return version, nil
}

// updateVersionRegistry stores the new version or appends it to the existing ones under the key <namespace>/rulesets/versions/<path>.
func (p *rulesPutter) updateVersionRegistry(stm concurrency.STM, path, version string) error {
	var v pb.Versions

	res := stm.Get(p.s.versionsPath(path))
	if res != "" {
		err := proto.Unmarshal([]byte(res), &v)
		if err != nil {
			p.s.Logger.Debug().Err(err).Str("path", path).Msg("put: versions unmarshalling failed")
			return errors.Wrap(err, "failed to unmarshal versions")
		}
	}

	v.Versions = append(v.Versions, version)
	bvs, err := proto.Marshal(&v)
	if err != nil {
		return errors.Wrap(err, "failed to encode versions")
	}
	stm.Put(p.s.versionsPath(path), string(bvs))

	return nil
}
