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

// rulesPath returns the path where the rulesets are stored in etcd.
// Key: <namespace>/rulesets/rulesets/<path>/<version>
// Value: rules
func (s *RulesetService) rulesetsPath(p, v string) string {
	// If the version parameter is not empty, we concatenate to the path separated by the versionSeparator value.
	if v != "" {
		p += versionSeparator + v
	}
	return path.Join(s.Namespace, "rulesets", "rulesets", p)
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

// latestVersionPath returns the path where the latest version string of each ruleset is stored in etcd.
// Key: <namespace>/rulesets/latest/<path>
// Value: version string
func (s *RulesetService) latestVersionPath(p string) string {
	return path.Join(s.Namespace, "rulesets", "latest", p)
}

// versionsPath returns the path where the versions of each rulesets are stored in etcd.
// Key: <namespace>/rulesets/versions/<path>
// Value: [version string, ]
func (s *RulesetService) versionsPath(p string) string {
	return path.Join(s.Namespace, "rulesets", "versions", p)
}

func pathVersionFromKey(key string) (string, string) {
	chunks := strings.Split(key, versionSeparator)
	return chunks[0], chunks[1]
}
