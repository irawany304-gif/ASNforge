package build

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ipanalytics/ASNforge/internal/download"
	"github.com/ipanalytics/ASNforge/internal/output"
	"github.com/ipanalytics/ASNforge/internal/version"
)

type Summary struct {
	ASNProfiles          int     `json:"asn_profiles"`
	Prefixes             int     `json:"prefixes"`
	RoutedPrefixes       int     `json:"routed_prefixes"`
	MOASPrefixes         int     `json:"moas_prefixes"`
	PrivateASNRecords    int     `json:"private_asn_records"`
	ReservedASNRecords   int     `json:"reserved_asn_records"`
	UnknownTypeASNs      int     `json:"unknown_type_asns"`
	MMDBInsertedPrefixes int     `json:"mmdb_inserted_prefixes"`
	BuildDurationSeconds float64 `json:"build_duration_seconds"`
}

type Quality struct {
	Verdict  string   `json:"verdict"`
	Warnings []string `json:"warnings"`
	Errors   []string `json:"errors"`
}

type Metadata struct {
	SchemaVersion string                `json:"schema_version"`
	BuildID       string                `json:"build_id"`
	GeneratedAt   string                `json:"generated_at"`
	GitCommit     string                `json:"git_commit,omitempty"`
	ToolVersion   string                `json:"tool_version"`
	ConfigProfile string                `json:"config_profile"`
	Sources       []download.SourceFile `json:"sources"`
	Artifacts     []output.Artifact     `json:"artifacts"`
	Summary       Summary               `json:"summary"`
	Quality       Quality               `json:"quality"`
}

func WriteMetadata(outDir string, md Metadata) error {
	if md.GitCommit == "" {
		md.GitCommit = gitCommit()
	}
	if md.ToolVersion == "" {
		md.ToolVersion = version.Version
	}
	b, err := json.MarshalIndent(md, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(outDir, "metadata.json"), append(b, '\n'), 0o644)
}

func gitCommit() string {
	cmd := exec.Command("git", "rev-parse", "--short=12", "HEAD")
	b, err := cmd.Output()
	if err != nil {
		return ""
	}
	return string(bytesTrimSpace(b))
}

func bytesTrimSpace(b []byte) []byte {
	for len(b) > 0 && (b[len(b)-1] == '\n' || b[len(b)-1] == '\r' || b[len(b)-1] == ' ' || b[len(b)-1] == '\t') {
		b = b[:len(b)-1]
	}
	return b
}
