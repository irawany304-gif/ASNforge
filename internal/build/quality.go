package build

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ipanalytics/ASNforge/internal/asn"
	"github.com/ipanalytics/ASNforge/internal/bgp"
	"github.com/ipanalytics/ASNforge/internal/download"
	"github.com/ipanalytics/ASNforge/internal/output"
)

func Evaluate(profiles []asn.Profile, prefixes []bgp.PrefixOrigin, mmdbPath string, maxMMDBSizeMB int64) (Quality, Summary) {
	q := Quality{Verdict: "PASS"}
	s := Summary{ASNProfiles: len(profiles), Prefixes: len(prefixes), RoutedPrefixes: len(prefixes)}
	for _, p := range profiles {
		if p.PrivateASN {
			s.PrivateASNRecords++
		}
		if p.ReservedASN {
			s.ReservedASNRecords++
		}
		if p.ASNType == asn.TypeUnknown {
			s.UnknownTypeASNs++
		}
	}
	for _, p := range prefixes {
		if p.MOAS {
			s.MOASPrefixes++
		}
	}
	if len(prefixes) == 0 {
		q.Warnings = append(q.Warnings, "no prefixes generated")
	}
	if len(profiles) == 0 {
		q.Warnings = append(q.Warnings, "no ASN profiles generated")
	}
	if len(profiles) > 0 && float64(s.UnknownTypeASNs)/float64(len(profiles)) > 0.95 {
		q.Warnings = append(q.Warnings, "unknown_type_asns > 95% of profiles")
	}
	if len(prefixes) > 0 && float64(s.MOASPrefixes)/float64(len(prefixes)) > 0.10 {
		q.Warnings = append(q.Warnings, "MOAS prefixes > 10% of prefixes")
	}
	if st, err := os.Stat(mmdbPath); err != nil {
		q.Errors = append(q.Errors, "MMDB missing")
	} else if maxMMDBSizeMB > 0 && st.Size() > maxMMDBSizeMB*1024*1024 {
		q.Warnings = append(q.Warnings, "MMDB file size exceeds configured maximum")
	}
	if len(q.Errors) > 0 {
		q.Verdict = "FAIL"
	} else if len(q.Warnings) > 0 {
		q.Verdict = "WARN"
	}
	return q, s
}

func ApplyProfileQualityPolicy(profile string, q *Quality, s *Summary) {
	if profile == "public-safe" && s.Prefixes < 100000 {
		q.Errors = append(q.Errors, fmt.Sprintf("public-safe prefix count too low: got %d, want at least 100000", s.Prefixes))
	}
	if len(q.Errors) > 0 {
		q.Verdict = "FAIL"
	} else if len(q.Warnings) > 0 {
		q.Verdict = "WARN"
	} else {
		q.Verdict = "PASS"
	}
}

func WriteQualityReport(path, buildID, generatedAt string, sources []download.SourceFile, artifacts []output.Artifact, profiles []asn.Profile, prefixes []bgp.PrefixOrigin, q Quality, smoke []string) error {
	var b strings.Builder
	fmt.Fprintf(&b, "# ASNForge quality report\n\nBuild id: `%s`\n\nGenerated: `%s`\n\n", buildID, generatedAt)
	b.WriteString("## Sources\n\n| name | sha256 | size |\n| --- | --- | ---: |\n")
	for _, s := range sources {
		fmt.Fprintf(&b, "| %s | `%s` | %d |\n", s.Name, s.SHA256, s.SizeBytes)
	}
	b.WriteString("\n## Artifacts\n\n| name | sha256 | size | records |\n| --- | --- | ---: | ---: |\n")
	for _, a := range artifacts {
		fmt.Fprintf(&b, "| %s | `%s` | %d | %d |\n", a.Name, a.SHA256, a.SizeBytes, a.Records)
	}
	typeCounts, tagCounts := map[string]int{}, map[string]int{}
	privateReserved, unknown := 0, 0
	for _, p := range profiles {
		typeCounts[p.ASNType]++
		if p.ASNType == asn.TypeUnknown {
			unknown++
		}
		if p.PrivateASN || p.ReservedASN {
			privateReserved++
		}
		for _, t := range p.ASNTags {
			tagCounts[t]++
		}
	}
	moas := 0
	for _, p := range prefixes {
		if p.MOAS {
			moas++
		}
	}
	fmt.Fprintf(&b, "\n## Summary\n\nASN profile count: %d\n\nPrefix count: %d\n\nMMDB prefix insert count: %d\n\nMOAS prefix count: %d\n\nPrivate/reserved ASN count: %d\n\nUnknown type count: %d\n\n", len(profiles), len(prefixes), len(prefixes), moas, privateReserved, unknown)
	b.WriteString("Top ASN types: " + topCounts(typeCounts) + "\n\n")
	b.WriteString("Top tags: " + topCounts(tagCounts) + "\n\n")
	fmt.Fprintf(&b, "Quality verdict: **%s**\n\n", q.Verdict)
	b.WriteString("Warnings:\n")
	for _, w := range q.Warnings {
		fmt.Fprintf(&b, "- %s\n", w)
	}
	if len(q.Warnings) == 0 {
		b.WriteString("- none\n")
	}
	b.WriteString("\nSmoke test results:\n")
	for _, s := range smoke {
		fmt.Fprintf(&b, "- %s\n", s)
	}
	if len(smoke) == 0 {
		b.WriteString("- not configured\n")
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

func topCounts(m map[string]int) string {
	type kv struct {
		k string
		v int
	}
	var rows []kv
	for k, v := range m {
		rows = append(rows, kv{k, v})
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].v == rows[j].v {
			return rows[i].k < rows[j].k
		}
		return rows[i].v > rows[j].v
	})
	limit := 5
	if len(rows) < limit {
		limit = len(rows)
	}
	parts := make([]string, 0, limit)
	for i := 0; i < limit; i++ {
		parts = append(parts, fmt.Sprintf("%s=%d", rows[i].k, rows[i].v))
	}
	if len(parts) == 0 {
		return "none"
	}
	return strings.Join(parts, ", ")
}

func ValidateReleaseDir(outDir string, strict bool) error {
	required := []string{"asnforge.mmdb", "asnforge-asn.jsonl", "asnforge-asn.csv", "asnforge-prefixes.jsonl", "asnforge-prefixes.csv", "metadata.json", "checksums.txt", "quality-report.md", "asnforge-diff.json", "manifest.json"}
	for _, name := range required {
		if _, err := os.Stat(filepath.Join(outDir, name)); err != nil {
			return fmt.Errorf("missing output %s: %w", name, err)
		}
	}
	lines, err := os.ReadFile(filepath.Join(outDir, "checksums.txt"))
	if err != nil {
		return err
	}
	if !strings.Contains(string(lines), "asnforge.mmdb") {
		return fmt.Errorf("checksums missing asnforge.mmdb")
	}
	metadataBytes, err := os.ReadFile(filepath.Join(outDir, "metadata.json"))
	if err != nil {
		return err
	}
	var metadata struct {
		ConfigProfile string  `json:"config_profile"`
		Summary       Summary `json:"summary"`
		Quality       Quality `json:"quality"`
		Sources       []struct {
			URL       string `json:"url"`
			LocalPath string `json:"local_path"`
		} `json:"sources"`
	}
	if err := json.Unmarshal(metadataBytes, &metadata); err != nil {
		return fmt.Errorf("parse metadata.json: %w", err)
	}
	if metadata.ConfigProfile == "public-safe" {
		if metadata.Summary.Prefixes < 100000 || metadata.Summary.MMDBInsertedPrefixes < 100000 {
			return fmt.Errorf("public-safe release has too few prefixes: prefixes=%d mmdb_inserted_prefixes=%d", metadata.Summary.Prefixes, metadata.Summary.MMDBInsertedPrefixes)
		}
		for _, source := range metadata.Sources {
			if strings.Contains(filepath.ToSlash(source.URL), "examples/testdata/") || strings.Contains(filepath.ToSlash(source.LocalPath), "examples/testdata/") {
				return fmt.Errorf("public-safe release uses testdata source: %s", source.URL)
			}
		}
	}
	if strict && metadata.Quality.Verdict == "FAIL" {
		return fmt.Errorf("quality verdict %s", metadata.Quality.Verdict)
	}
	return nil
}
