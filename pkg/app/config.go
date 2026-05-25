package app

import (
	"fmt"

	"github.com/go-go-golems/goja-site/pkg/observability"
)

type DBPolicy string

const (
	DBPolicySimple  DBPolicy = "simple"
	DBPolicyGuarded DBPolicy = "guarded"
)

// Config describes one goja-site server process.
type Config struct {
	Addr          string
	DBPath        string
	ScriptDirs    []string
	Dev           bool
	DBPolicy      DBPolicy
	ReadOnly      bool
	AllowWrites   bool
	SiteName      string
	Observability *observability.Observability
}

func normalizeDBPolicy(policy DBPolicy) (DBPolicy, error) {
	switch policy {
	case "":
		return DBPolicyGuarded, nil
	case DBPolicySimple, DBPolicyGuarded:
		return policy, nil
	default:
		return "", fmt.Errorf("unsupported database policy %q (expected %q or %q)", policy, DBPolicySimple, DBPolicyGuarded)
	}
}

func normalizeDBPolicyConfig(cfg *Config) error {
	policy, err := normalizeDBPolicy(cfg.DBPolicy)
	if err != nil {
		return err
	}
	cfg.DBPolicy = policy
	if cfg.DBPolicy == DBPolicySimple && !cfg.AllowWrites {
		cfg.ReadOnly = true
	}
	return nil
}
