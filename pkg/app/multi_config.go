package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// MultiConfig describes a single HTTP listener that serves multiple isolated
// goja-site instances selected by request Host.
type MultiConfig struct {
	Addr       string       `json:"addr" yaml:"addr"`
	DataDir    string       `json:"dataDir" yaml:"dataDir"`
	BaseDomain string       `json:"baseDomain" yaml:"baseDomain"`
	Dev        bool         `json:"dev" yaml:"dev"`
	Sites      []SiteConfig `json:"sites" yaml:"sites"`
}

// SiteConfig describes one hosted site. If Host or DBPath are omitted, they are
// derived from Name plus MultiConfig.BaseDomain/DataDir.
type SiteConfig struct {
	Name       string   `json:"name" yaml:"name"`
	Host       string   `json:"host" yaml:"host"`
	ScriptDirs []string `json:"scripts" yaml:"scripts"`
	DBPath     string   `json:"dbPath" yaml:"dbPath"`
}

func LoadMultiConfig(path string) (MultiConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return MultiConfig{}, fmt.Errorf("read multi-site config: %w", err)
	}
	var cfg MultiConfig
	switch strings.ToLower(filepath.Ext(path)) {
	case ".json":
		err = json.Unmarshal(data, &cfg)
	default:
		err = yaml.Unmarshal(data, &cfg)
	}
	if err != nil {
		return MultiConfig{}, fmt.Errorf("parse multi-site config %s: %w", path, err)
	}
	if err := cfg.Normalize(); err != nil {
		return MultiConfig{}, err
	}
	return cfg, nil
}

func (cfg *MultiConfig) Normalize() error {
	if cfg.Addr == "" {
		cfg.Addr = ":8080"
	}
	if cfg.DataDir == "" {
		cfg.DataDir = "./data/sites"
	}
	cfg.BaseDomain = strings.Trim(strings.ToLower(cfg.BaseDomain), ".")
	if len(cfg.Sites) == 0 {
		return fmt.Errorf("multi-site config requires at least one site")
	}
	seenNames := map[string]struct{}{}
	seenHosts := map[string]struct{}{}
	for i := range cfg.Sites {
		site := &cfg.Sites[i]
		site.Name = strings.TrimSpace(site.Name)
		if site.Name == "" {
			return fmt.Errorf("site %d: name is required", i)
		}
		if !validSiteName(site.Name) {
			return fmt.Errorf("site %q: name must contain only lowercase letters, numbers, and dashes", site.Name)
		}
		if _, ok := seenNames[site.Name]; ok {
			return fmt.Errorf("duplicate site name %q", site.Name)
		}
		seenNames[site.Name] = struct{}{}
		if len(site.ScriptDirs) == 0 {
			return fmt.Errorf("site %q: scripts is required", site.Name)
		}
		if site.Host == "" {
			if cfg.BaseDomain == "" {
				return fmt.Errorf("site %q: host is required when baseDomain is empty", site.Name)
			}
			site.Host = site.Name + "." + cfg.BaseDomain
		}
		site.Host = normalizeHost(site.Host)
		if _, ok := seenHosts[site.Host]; ok {
			return fmt.Errorf("duplicate site host %q", site.Host)
		}
		seenHosts[site.Host] = struct{}{}
		if site.DBPath == "" {
			site.DBPath = filepath.Join(cfg.DataDir, site.Name, "app.db")
		}
	}
	return nil
}

var siteNameRE = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*[a-z0-9]$|^[a-z0-9]$`)

func validSiteName(name string) bool { return siteNameRE.MatchString(name) }

func normalizeHost(host string) string {
	host = strings.TrimSpace(strings.ToLower(host))
	host = strings.TrimSuffix(host, ".")
	if i := strings.LastIndex(host, ":"); i >= 0 {
		// Strip a port from normal host:port values. Bracketed IPv6 host routing is
		// intentionally not supported for public multi-site names.
		if !strings.Contains(host[i+1:], "]") && strings.Count(host, ":") == 1 {
			host = host[:i]
		}
	}
	return host
}
