package etcd

import (
	"path"
	"strings"

	"github.com/coreos/etcd/clientv3"
	"github.com/rs/zerolog"
)

// versionSeparator separates the path from the version in the entries path in etcd.
// The purpose is to have the same ordering as the others namespace (latest, versions, ...).
const versionSeparator = "!"

// RulesetService manages the rulesets using etcd.
type RulesetService struct {
	Client    *clientv3.Client
	Logger    zerolog.Logger
	Namespace string
}

// rulesPath returns the path where the rules of rulesets are stored in etcd.
// Key: <namespace>/rulesets/rules/<path>/<version>
// Value: rules
func (s *RulesetService) rulesPath(p, v string) string {
	// If the version parameter is not empty, we concatenate to the path separated by the versionSeparator value.
	if v != "" {
		p += versionSeparator + v
	}
	return path.Join(s.Namespace, "rulesets", "rules", p)
}

// checksumsPath returns the path where the checksums are stored in etcd.
// Key: <namespace>/rulesets/checksums/<path>
// Value: checksum
func (s *RulesetService) checksumsPath(p string) string {
	return path.Join(s.Namespace, "rulesets", "checksums", p)
}

// signaturesPath returns the path where the signatures are stored in etcd.
// Key: <namespace>/rulesets/signatures/<path>
// Value: signature
func (s *RulesetService) signaturesPath(p string) string {
	return path.Join(s.Namespace, "rulesets", "signatures", p)
}

func (s *RulesetService) pathVersionFromKey(key string) (string, string) {
	key = strings.TrimPrefix(key, path.Join(s.Namespace, "rulesets", "rules")+"/")
	chunks := strings.Split(key, versionSeparator)
	return chunks[0], chunks[1]
}

func (s *RulesetService) pathFromKey(collection, key string) string {
	return strings.TrimPrefix(key, path.Join(s.Namespace, "rulesets", collection)+"/")
}
