package etcd

import (
	"context"

	"github.com/coreos/etcd/clientv3"
	"github.com/gogo/protobuf/proto"
	"github.com/heetch/regula"
	"github.com/heetch/regula/store"
	"github.com/pkg/errors"
)

// Create stores a signature in the signature tree. A signature represents a ruleset.
func (s *RulesetService) Create(ctx context.Context, path string, signature *regula.Signature) error {
	if err := store.ValidatePath(path); err != nil {
		return err
	}

	if err := store.ValidateSignature(signature); err != nil {
		return err
	}

	value, err := proto.Marshal(signatureToProtobuf(signature))
	if err != nil {
		return errors.Wrap(err, "failed to encode signature")
	}

	sigPath := s.signaturesPath(path)

	resp, err := s.Client.Txn(ctx).
		// if a key exists, its version is > 0
		If(clientv3.Compare(clientv3.Version(sigPath), "==", 0)).
		// store the signature
		Then(clientv3.OpPut(sigPath, string(value))).
		Commit()
	if err != nil {
		return errors.Wrap(err, "failed to store signature")
	}

	// if the If condition failed, it means the signature already exists
	if !resp.Succeeded {
		return store.ErrAlreadyExists
	}

	return nil
}
