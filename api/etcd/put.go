package etcd

import (
	"context"
	"crypto/md5"

	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/golang/protobuf/proto"
	"github.com/heetch/regula/api"
	pb "github.com/heetch/regula/api/etcd/proto"
	"github.com/heetch/regula/rule"
	"github.com/pkg/errors"
	"github.com/segmentio/ksuid"
)

// Put stores the given rules under the rules tree. If no signature is found for the given path it returns an error.
func (s *RulesetService) Put(ctx context.Context, path string, rules []*rule.Rule) error {
	txfn := func(stm concurrency.STM) error {
		p := rulesPutter{s, stm}
		return p.put(ctx, path, rules)
	}

	_, err := concurrency.NewSTM(s.Client, txfn, concurrency.WithAbortContext(ctx))
	if err != nil && err != api.ErrRulesetNotModified && !api.IsValidationError(err) {
		return errors.Wrap(err, "failed to put ruleset")
	}

	return err
}

// rulesPutter is responsible for validating and storing rules, updating checksums and other actions
// that are required in order to add a new ruleset version correctly.
type rulesPutter struct {
	s   *RulesetService
	stm concurrency.STM
}

func (p *rulesPutter) put(ctx context.Context, path string, rules []*rule.Rule) error {
	// validate the rules
	err := p.validateRules(p.stm, path, rules)
	if err != nil {
		return err
	}

	// encode rules
	data, err := proto.Marshal(rulesToProtobuf(rules))
	if err != nil {
		return err
	}

	// update checksum if rules have changed
	changed, err := p.updateChecksum(p.stm, path, data)
	if err != nil {
		return err
	}

	if !changed {
		return api.ErrRulesetNotModified
	}

	// create a new version of the rules
	return p.createNewVersion(p.stm, path, data)
}

// validateRules fetches the signature from the store and validates all the rules against it.
func (p *rulesPutter) validateRules(stm concurrency.STM, path string, rules []*rule.Rule) error {
	data := stm.Get(p.s.signaturesPath(path))
	if data == "" {
		return api.ErrSignatureNotFound
	}

	var pbsig pb.Signature
	err := proto.Unmarshal([]byte(data), &pbsig)
	if err != nil {
		return errors.Wrap(err, "failed to decode signature")
	}

	sig := signatureFromProtobuf(&pbsig)
	for _, r := range rules {
		if err := api.ValidateRule(sig, r); err != nil {
			return err
		}
	}

	return nil
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
func (p *rulesPutter) createNewVersion(stm concurrency.STM, path string, data []byte) error {
	// create a new ruleset version
	k, err := ksuid.NewRandom()
	if err != nil {
		return errors.Wrap(err, "failed to generate rules version")
	}

	stm.Put(p.s.rulesPath(path, k.String()), string(data))

	return nil
}
