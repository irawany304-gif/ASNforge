package config

import (
	"flag"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Profile          string `yaml:"profile"`
	SchemaVersion    string `yaml:"schema_version"`
	PrivateASNPolicy string `yaml:"private_asn_policy"`
	MOASPolicy       string `yaml:"moas_policy"`
	Compression      bool   `yaml:"compression"`
	MaxMMDBSizeMB    int64  `yaml:"max_mmdb_size_mb"`
	SourceTimeout    string `yaml:"source_timeout"`
	Sources          struct {
		RIR struct {
			Enabled bool              `yaml:"enabled"`
			URLs    map[string]string `yaml:"urls"`
			Paths   map[string]string `yaml:"paths"`
		} `yaml:"rir"`
		BGP struct {
			Enabled    bool     `yaml:"enabled"`
			Mode       string   `yaml:"mode"`
			Collectors []string `yaml:"collectors"`
			URLs       []string `yaml:"urls"`
			Paths      []string `yaml:"paths"`
		} `yaml:"bgp"`
	} `yaml:"sources"`
	Overrides struct {
		Path string `yaml:"path"`
	} `yaml:"overrides"`
	Smoke struct {
		Path string `yaml:"path"`
	} `yaml:"smoke"`
}

type Options struct {
	ConfigPath       string
	OutDir           string
	CacheDir         string
	BuildID          string
	SchemaVersion    string
	PrivateASNPolicy string
	MOASPolicy       string
	MMDBPath         string
	SkipDownload     bool
	Strict           bool
	Format           string
	ASNTable         string
}

func Load(path string) (Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	var c Config
	if err := yaml.Unmarshal(b, &c); err != nil {
		return Config{}, err
	}
	if c.SchemaVersion == "" {
		c.SchemaVersion = "asnforge.v0.1"
	}
	if c.PrivateASNPolicy == "" {
		c.PrivateASNPolicy = "flag"
	}
	if c.MOASPolicy == "" {
		c.MOASPolicy = "mark_ambiguous"
	}
	if c.MaxMMDBSizeMB == 0 {
		c.MaxMMDBSizeMB = 256
	}
	if c.SourceTimeout == "" {
		c.SourceTimeout = "60s"
	}
	return c, nil
}

func AddCommonFlags(fs *flag.FlagSet, o *Options) {
	fs.StringVar(&o.ConfigPath, "config", "config/public-safe.yaml", "Path to build config")
	fs.StringVar(&o.OutDir, "out", "release/current", "Output release directory")
	fs.StringVar(&o.CacheDir, "cache", "data/cache", "Source cache directory")
	fs.StringVar(&o.BuildID, "build-id", "", "Optional explicit build id")
	fs.StringVar(&o.SchemaVersion, "schema-version", "asnforge.v0.1", "Schema version")
	fs.StringVar(&o.PrivateASNPolicy, "private-asn-policy", "", "flag, drop, keep")
	fs.StringVar(&o.MOASPolicy, "moas-policy", "", "mark_ambiguous, most_observed, lowest_asn")
	fs.StringVar(&o.MMDBPath, "mmdb", "", "Output or input MMDB path")
	fs.BoolVar(&o.SkipDownload, "skip-download", false, "Use cached source files")
	fs.BoolVar(&o.Strict, "strict", false, "Fail on quality warnings")
	fs.StringVar(&o.Format, "format", "text", "text or json")
	fs.StringVar(&o.ASNTable, "asn-table", "", "ASN table path")
}

func BuildID() string {
	return time.Now().UTC().Format("20060102-150405Z")
}

func ValidatePolicies(privatePolicy, moasPolicy string) error {
	switch privatePolicy {
	case "", "flag", "drop", "keep":
	default:
		return fmt.Errorf("invalid private ASN policy %q", privatePolicy)
	}
	switch moasPolicy {
	case "", "mark_ambiguous", "most_observed", "lowest_asn":
	default:
		return fmt.Errorf("invalid MOAS policy %q", moasPolicy)
	}
	return nil
}
