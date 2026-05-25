package release

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ipanalytics/ASNforge/internal/build"
)

const (
	statsBegin = "<!-- ASNFORGE:RELEASE-STATS BEGIN -->"
	statsEnd   = "<!-- ASNFORGE:RELEASE-STATS END -->"
)

func WriteReleaseNotes(outDir, path string) error {
	md, diff, err := load(outDir)
	if err != nil {
		return err
	}
	var b strings.Builder
	writeReleaseStats(&b, md, diff, true)
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

func UpdateREADME(outDir, readmePath string) error {
	md, diff, err := load(outDir)
	if err != nil {
		return err
	}
	current, err := os.ReadFile(readmePath)
	if err != nil {
		return err
	}
	var block strings.Builder
	block.WriteString(statsBegin + "\n")
	writeReleaseStats(&block, md, diff, false)
	block.WriteString(statsEnd + "\n")
	updated := upsertBlock(string(current), block.String())
	return os.WriteFile(readmePath, []byte(updated), 0o644)
}

func load(outDir string) (build.Metadata, map[string]any, error) {
	var md build.Metadata
	b, err := os.ReadFile(filepath.Join(outDir, "metadata.json"))
	if err != nil {
		return md, nil, err
	}
	if err := json.Unmarshal(b, &md); err != nil {
		return md, nil, err
	}
	diff := map[string]any{}
	if b, err := os.ReadFile(filepath.Join(outDir, "asnforge-diff.json")); err == nil {
		_ = json.Unmarshal(b, &diff)
	}
	return md, diff, nil
}

func writeReleaseStats(b *strings.Builder, md build.Metadata, diff map[string]any, includeTitle bool) {
	if includeTitle {
		fmt.Fprintf(b, "# ASNForge %s data release\n\n", md.ConfigProfile)
	}
	fmt.Fprintf(b, "## Latest Release Stats\n\n")
	fmt.Fprintf(b, "| Field | Value |\n| --- | ---: |\n")
	fmt.Fprintf(b, "| Build ID | `%s` |\n", md.BuildID)
	fmt.Fprintf(b, "| Profile | `%s` |\n", md.ConfigProfile)
	fmt.Fprintf(b, "| Generated | `%s` |\n", md.GeneratedAt)
	fmt.Fprintf(b, "| Quality | `%s` |\n", md.Quality.Verdict)
	fmt.Fprintf(b, "| ASN profiles | %s |\n", human(md.Summary.ASNProfiles))
	fmt.Fprintf(b, "| Named ASN profiles | %s |\n", human(md.Summary.NamedASNProfiles))
	fmt.Fprintf(b, "| Prefixes | %s |\n", human(md.Summary.Prefixes))
	fmt.Fprintf(b, "| MMDB inserted prefixes | %s |\n", human(md.Summary.MMDBInsertedPrefixes))
	fmt.Fprintf(b, "| MOAS prefixes | %s |\n", human(md.Summary.MOASPrefixes))
	fmt.Fprintf(b, "| Private ASN records | %s |\n", human(md.Summary.PrivateASNRecords))
	fmt.Fprintf(b, "| Reserved ASN records | %s |\n", human(md.Summary.ReservedASNRecords))
	fmt.Fprintf(b, "| Unknown type ASNs | %s |\n", human(md.Summary.UnknownTypeASNs))
	fmt.Fprintf(b, "| Build duration seconds | %.2f |\n\n", md.Summary.BuildDurationSeconds)

	writeSources(b, md)
	writeArtifacts(b, md)
	writeDiff(b, diff)
	writeQuality(b, md)
}

func writeSources(b *strings.Builder, md build.Metadata) {
	fmt.Fprintf(b, "## Sources\n\n")
	fmt.Fprintf(b, "| Name | URL | Size | SHA256 |\n| --- | --- | ---: | --- |\n")
	for _, s := range md.Sources {
		fmt.Fprintf(b, "| `%s` | %s | %s | `%s` |\n", s.Name, sourceURL(s.URL), human64(s.SizeBytes), shortHash(s.SHA256))
	}
	fmt.Fprintln(b)
}

func writeArtifacts(b *strings.Builder, md build.Metadata) {
	rows := append([]struct {
		name    string
		size    int64
		records int
	}{}, struct {
		name    string
		size    int64
		records int
	}{})
	rows = rows[:0]
	for _, a := range md.Artifacts {
		if strings.HasSuffix(a.Name, ".gz") || a.Name == "metadata.json" || a.Name == "checksums.txt" || a.Name == "quality-report.md" || a.Name == "manifest.json" || a.Name == "asnforge-diff.json" {
			rows = append(rows, struct {
				name    string
				size    int64
				records int
			}{a.Name, a.SizeBytes, a.Records})
		}
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].name < rows[j].name })
	fmt.Fprintf(b, "## Artifacts\n\n")
	fmt.Fprintf(b, "| Artifact | Size | Records |\n| --- | ---: | ---: |\n")
	for _, r := range rows {
		recordValue := "-"
		if r.records > 0 {
			recordValue = human(r.records)
		}
		fmt.Fprintf(b, "| `%s` | %s | %s |\n", r.name, human64(r.size), recordValue)
	}
	fmt.Fprintln(b)
}

func writeDiff(b *strings.Builder, diff map[string]any) {
	if len(diff) == 0 {
		return
	}
	keys := []string{"baseline", "new_asns", "removed_asns", "changed_asn_profiles", "new_prefixes", "removed_prefixes", "changed_prefix_origins", "new_moas_prefixes", "resolved_moas_prefixes"}
	fmt.Fprintf(b, "## Numeric Diff\n\n")
	fmt.Fprintf(b, "| Metric | Value |\n| --- | ---: |\n")
	for _, key := range keys {
		if v, ok := diff[key]; ok {
			fmt.Fprintf(b, "| `%s` | %v |\n", key, v)
		}
	}
	fmt.Fprintln(b)
}

func writeQuality(b *strings.Builder, md build.Metadata) {
	fmt.Fprintf(b, "## Quality\n\n")
	if len(md.Quality.Warnings) == 0 && len(md.Quality.Errors) == 0 {
		fmt.Fprintf(b, "No warnings or errors.\n")
		return
	}
	for _, warning := range md.Quality.Warnings {
		fmt.Fprintf(b, "- WARN: %s\n", warning)
	}
	for _, err := range md.Quality.Errors {
		fmt.Fprintf(b, "- ERROR: %s\n", err)
	}
}

func upsertBlock(readme, block string) string {
	start := strings.Index(readme, statsBegin)
	end := strings.Index(readme, statsEnd)
	if start >= 0 && end > start {
		end += len(statsEnd)
		return readme[:start] + block + readme[end:]
	}
	insertAt := strings.Index(readme, "\n## Overview")
	if insertAt < 0 {
		return strings.TrimRight(readme, "\n") + "\n\n" + block
	}
	return readme[:insertAt+1] + "\n" + block + readme[insertAt+1:]
}

func sourceURL(url string) string {
	if url == "" {
		return "-"
	}
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		return fmt.Sprintf("[%s](%s)", filepath.Base(url), url)
	}
	return "`" + url + "`"
}

func shortHash(hash string) string {
	if len(hash) <= 12 {
		return hash
	}
	return hash[:12]
}

func human(v int) string {
	return human64(int64(v))
}

func human64(v int64) string {
	s := fmt.Sprintf("%d", v)
	n := len(s)
	if n <= 3 {
		return s
	}
	var out []byte
	first := n % 3
	if first == 0 {
		first = 3
	}
	out = append(out, s[:first]...)
	for i := first; i < n; i += 3 {
		out = append(out, ',')
		out = append(out, s[i:i+3]...)
	}
	return string(out)
}
