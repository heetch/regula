package etcd

import (
	"context"
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/heetch/regula"
	"github.com/heetch/regula/api"
	pb "github.com/heetch/regula/api/etcd/proto"
	"github.com/stretchr/testify/require"
)

func TestCreate(t *testing.T) {
	s, cleanup := newEtcdRulesetService(t)
	defer cleanup()

	sig := regula.Signature{ReturnType: "bool", Params: map[string]string{"id": "string"}}

	// create a ruleset
	err := s.Create(context.Background(), "a", &sig)
	require.NoError(t, err)

	// verify creation
	resp, err := s.Client.Get(context.Background(), s.signaturesPath("a"))
	require.NoError(t, err)
	var pbsig pb.Signature
	err = proto.Unmarshal(resp.Kvs[0].Value, &pbsig)
	require.NoError(t, err)
	require.EqualValues(t, &sig, signatureFromProtobuf(&pbsig))

	// create with the same path
	err = s.Create(context.Background(), "a", &sig)
	require.Error(t, api.ErrAlreadyExists, err)
}
