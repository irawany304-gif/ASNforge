package build

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ipanalytics/ASNforge/internal/asn"
	"github.com/ipanalytics/ASNforge/internal/bgp"
	"github.com/ipanalytics/ASNforge/internal/config"
	"github.com/ipanalytics/ASNforge/internal/download"
	"github.com/ipanalytics/ASNforge/internal/mmdb"
	"github.com/ipanalytics/ASNforge/internal/output"
	"github.com/ipanalytics/ASNforge/internal/rir"
	"github.com/ipanalytics/ASNforge/internal/smoke"
)

func Run(ctx context.Context, opts config.Options) (Metadata, error) {
	start := time.Now()
	cfg, err := config.Load(opts.ConfigPath)
	if err != nil {
		return Metadata{}, err
	}
	if opts.BuildID == "" {
		opts.BuildID = config.BuildID()
	}
	if opts.SchemaVersion == "" {
		opts.SchemaVersion = cfg.SchemaVersion
	}
	if opts.PrivateASNPolicy == "" {
		opts.PrivateASNPolicy = cfg.PrivateASNPolicy
	}
	if opts.MOASPolicy == "" {
		opts.MOASPolicy = cfg.MOASPolicy
	}
	if opts.MMDBPath == "" {
		opts.MMDBPath = filepath.Join(opts.OutDir, "asnforge.mmdb")
	}
	if err := config.ValidatePolicies(opts.PrivateASNPolicy, opts.MOASPolicy); err != nil {
		return Metadata{}, err
	}
	if err := validateSourceProfile(cfg); err != nil {
		return Metadata{}, err
	}
	generatedAt := time.Now().UTC().Format(time.RFC3339)
	if err := os.MkdirAll(opts.OutDir, 0o755); err != nil {
		return Metadata{}, err
	}

	fmt.Fprintln(os.Stderr, "asnforge: collecting source files")
	sources, err := collectSources(ctx, cfg, opts)
	if err != nil {
		return Metadata{}, err
	}
	fmt.Fprintln(os.Stderr, "asnforge: parsing RIR delegated data")
	var allocs []rir.ASNAllocation
	for _, sf := range sources {
		if sf.Name == "bgp" || sf.Name == "asn_catalog" || filepath.Base(filepath.Dir(sf.LocalPath)) == "bgp" {
			continue
		}
		if looksLikeBGP(sf.LocalPath) || looksLikeASNCatalog(sf) {
			continue
		}
		got, err := rir.ParseDelegatedFile(sf.LocalPath)
		if err != nil {
			return Metadata{}, err
		}
		allocs = append(allocs, got...)
	}
	profileMap := ProfilesFromAllocations(allocs, opts.SchemaVersion, opts.BuildID, generatedAt)

	fmt.Fprintln(os.Stderr, "asnforge: parsing ASN catalog data")
	catalogRecords, err := parseASNCatalogs(sources)
	if err != nil {
		return Metadata{}, err
	}
	ApplyCatalog(catalogRecords, profileMap, opts.SchemaVersion, opts.BuildID, generatedAt)

	fmt.Fprintln(os.Stderr, "asnforge: parsing prefix-origin observations")
	var observations []bgp.PrefixOriginObservation
	for _, sf := range sources {
		if looksLikeBGP(sf.LocalPath) {
			got, err := bgp.ParsePreprocessedFile(sf.LocalPath)
			if err != nil {
				return Metadata{}, err
			}
			observations = append(observations, got...)
		}
	}
	prefixes := bgp.Aggregate(observations, opts.MOASPolicy, opts.SchemaVersion, opts.BuildID, generatedAt)
	origins := originSet(prefixes)
	EnsureProfilesForOrigins(profileMap, origins, opts.SchemaVersion, opts.BuildID, generatedAt)
	ApplyNameHeuristics(profileMap)
	if err := ApplyOverrides(cfg.Overrides.Path, profileMap); err != nil {
		return Metadata{}, err
	}
	profiles := SortedProfiles(profileMap, opts.PrivateASNPolicy)
	profileMap = mapFromProfiles(profiles)

	if opts.PrivateASNPolicy == "drop" {
		prefixes = filterPrivatePrefixes(prefixes)
	}

	fmt.Fprintln(os.Stderr, "asnforge: writing tabular artifacts")
	asnJSONL := filepath.Join(opts.OutDir, "asnforge-asn.jsonl")
	asnCSV := filepath.Join(opts.OutDir, "asnforge-asn.csv")
	prefixJSONL := filepath.Join(opts.OutDir, "asnforge-prefixes.jsonl")
	prefixCSV := filepath.Join(opts.OutDir, "asnforge-prefixes.csv")
	if err := output.WriteJSONL(asnJSONL, profiles); err != nil {
		return Metadata{}, err
	}
	if err := output.WriteASNCSV(asnCSV, profiles); err != nil {
		return Metadata{}, err
	}
	if err := output.WriteJSONL(prefixJSONL, prefixes); err != nil {
		return Metadata{}, err
	}
	if err := output.WritePrefixCSV(prefixCSV, prefixes); err != nil {
		return Metadata{}, err
	}

	fmt.Fprintln(os.Stderr, "asnforge: writing MMDB")
	inserted, err := mmdb.Write(opts.MMDBPath, prefixes, profileMap)
	if err != nil {
		return Metadata{}, err
	}
	if cfg.Compression {
		for _, p := range []string{opts.MMDBPath, asnJSONL, asnCSV, prefixJSONL, prefixCSV} {
			if err := output.GzipFile(p); err != nil {
				return Metadata{}, err
			}
		}
	}

	q, summary := Evaluate(profiles, prefixes, opts.MMDBPath, cfg.MaxMMDBSizeMB)
	summary.MMDBInsertedPrefixes = inserted
	summary.BuildDurationSeconds = time.Since(start).Seconds()
	ApplyProfileQualityPolicy(cfg.Profile, &q, &summary)

	artifacts, err := collectArtifacts(opts.OutDir, profiles, prefixes)
	if err != nil {
		return Metadata{}, err
	}
	smokeResults, smokeErr := smoke.Run(cfg.Smoke.Path, opts.OutDir)
	if smokeErr != nil {
		q.Errors = append(q.Errors, smokeErr.Error())
		q.Verdict = "FAIL"
	}
	if err := WriteQualityReport(filepath.Join(opts.OutDir, "quality-report.md"), opts.BuildID, generatedAt, sources, artifacts, profiles, prefixes, q, smokeResults); err != nil {
		return Metadata{}, err
	}
	if err := writeDiff(filepath.Join(opts.OutDir, "asnforge-diff.json")); err != nil {
		return Metadata{}, err
	}
	artifacts, err = collectArtifacts(opts.OutDir, profiles, prefixes)
	if err != nil {
		return Metadata{}, err
	}
	if err := writeManifest(filepath.Join(opts.OutDir, "manifest.json"), artifacts); err != nil {
		return Metadata{}, err
	}
	artifacts, err = collectArtifacts(opts.OutDir, profiles, prefixes)
	if err != nil {
		return Metadata{}, err
	}
	if err := output.WriteChecksums(opts.OutDir, artifacts); err != nil {
		return Metadata{}, err
	}
	md := Metadata{
		SchemaVersion: opts.SchemaVersion, BuildID: opts.BuildID, GeneratedAt: generatedAt,
		ConfigProfile: cfg.Profile, Sources: sources, Artifacts: artifacts, Summary: summary, Quality: q,
	}
	if err := WriteMetadata(opts.OutDir, md); err != nil {
		return Metadata{}, err
	}
	if opts.Strict && q.Verdict != "PASS" {
		return md, fmt.Errorf("quality verdict %s", q.Verdict)
	}
	return md, nil
}

func validateSourceProfile(cfg config.Config) error {
	if cfg.Profile != "public-safe" {
		return nil
	}
	if cfg.Sources.BGP.Enabled && len(cfg.Sources.BGP.URLs) == 0 && len(cfg.Sources.BGP.Paths) == 0 {
		return fmt.Errorf("public-safe profile requires at least one production BGP prefix-origin URL or path; use config/local-dev.yaml for deterministic fixture builds")
	}
	for _, path := range cfg.Sources.BGP.Paths {
		clean := filepath.ToSlash(path)
		if strings.Contains(clean, "examples/testdata/") {
			return fmt.Errorf("public-safe profile must not use deterministic testdata path %q", path)
		}
	}
	return nil
}

func collectSources(ctx context.Context, cfg config.Config, opts config.Options) ([]download.SourceFile, error) {
	var out []download.SourceFile
	keys := make([]string, 0, len(cfg.Sources.RIR.Paths))
	for k := range cfg.Sources.RIR.Paths {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		sf, err := download.LocalSourceFile(cfg.Sources.RIR.Paths[k])
		if err != nil {
			return nil, err
		}
		sf.Name = k
		out = append(out, sf)
	}
	if !opts.SkipDownload {
		keys = keys[:0]
		for k := range cfg.Sources.RIR.URLs {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			sf, err := download.Download(ctx, opts.CacheDir, "rir", cfg.Sources.RIR.URLs[k])
			if err != nil {
				return nil, err
			}
			sf.Name = k
			out = append(out, sf)
		}
		bgps, err := download.DownloadAll(ctx, opts.CacheDir, "bgp", cfg.Sources.BGP.URLs)
		if err != nil {
			return nil, err
		}
		out = append(out, bgps...)
		catalogs, err := download.DownloadAll(ctx, opts.CacheDir, "asn_catalog", cfg.Sources.ASNCatalog.URLs)
		if err != nil {
			return nil, err
		}
		for i := range catalogs {
			catalogs[i].Name = "asn_catalog"
		}
		out = append(out, catalogs...)
	}
	for _, p := range cfg.Sources.BGP.Paths {
		sf, err := download.LocalSourceFile(p)
		if err != nil {
			return nil, err
		}
		sf.Name = "bgp"
		out = append(out, sf)
	}
	for _, p := range cfg.Sources.ASNCatalog.Paths {
		sf, err := download.LocalSourceFile(p)
		if err != nil {
			return nil, err
		}
		sf.Name = "asn_catalog"
		out = append(out, sf)
	}
	return out, nil
}

func looksLikeBGP(path string) bool {
	base := filepath.Base(path)
	return base == "prefix-origin.csv" || base == "prefix-origin.tsv" || filepath.Ext(base) == ".jsonl" || filepath.Base(filepath.Dir(path)) == "bgp"
}

func looksLikeASNCatalog(sf download.SourceFile) bool {
	if sf.Name == "asn_catalog" {
		return true
	}
	base := filepath.Base(sf.LocalPath)
	return base == "asns.csv" || filepath.Base(filepath.Dir(sf.LocalPath)) == "asn_catalog"
}

func parseASNCatalogs(sources []download.SourceFile) ([]asn.CatalogRecord, error) {
	var out []asn.CatalogRecord
	for _, sf := range sources {
		if !looksLikeASNCatalog(sf) {
			continue
		}
		rows, err := asn.ParseBGPToolsASNsCSV(sf.LocalPath)
		if err != nil {
			return nil, err
		}
		out = append(out, rows...)
	}
	return out, nil
}

func originSet(prefixes []bgp.PrefixOrigin) []uint32 {
	m := map[uint32]bool{}
	for _, p := range prefixes {
		for _, n := range p.OriginASNs {
			m[n] = true
		}
	}
	out := make([]uint32, 0, len(m))
	for n := range m {
		out = append(out, n)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func mapFromProfiles(profiles []asn.Profile) map[uint32]asn.Profile {
	m := map[uint32]asn.Profile{}
	for _, p := range profiles {
		m[p.ASN] = p
	}
	return m
}

func filterPrivatePrefixes(in []bgp.PrefixOrigin) []bgp.PrefixOrigin {
	out := in[:0]
	for _, p := range in {
		if !p.PrivateASN && !p.ReservedASN {
			out = append(out, p)
		}
	}
	return out
}

func collectArtifacts(outDir string, profiles []asn.Profile, prefixes []bgp.PrefixOrigin) ([]output.Artifact, error) {
	type spec struct {
		name, ct, desc string
		records        int
	}
	specs := []spec{
		{"asnforge.mmdb", "application/vnd.maxmind.maxmind-db", "Compact IP prefix to ASN profile database", len(prefixes)},
		{"asnforge.mmdb.gz", "application/gzip", "Compressed MMDB", len(prefixes)},
		{"asnforge-asn.jsonl", "application/x-ndjson", "Canonical ASN profile table", len(profiles)},
		{"asnforge-asn.jsonl.gz", "application/gzip", "Compressed ASN JSONL", len(profiles)},
		{"asnforge-asn.csv", "text/csv", "Canonical ASN profile CSV", len(profiles)},
		{"asnforge-asn.csv.gz", "application/gzip", "Compressed ASN CSV", len(profiles)},
		{"asnforge-prefixes.jsonl", "application/x-ndjson", "Canonical prefix-origin table", len(prefixes)},
		{"asnforge-prefixes.jsonl.gz", "application/gzip", "Compressed prefix JSONL", len(prefixes)},
		{"asnforge-prefixes.csv", "text/csv", "Canonical prefix-origin CSV", len(prefixes)},
		{"asnforge-prefixes.csv.gz", "application/gzip", "Compressed prefix CSV", len(prefixes)},
		{"quality-report.md", "text/markdown", "Quality report", 0},
		{"asnforge-diff.json", "application/json", "Release diff", 0},
		{"manifest.json", "application/json", "Artifact manifest", 0},
	}
	var out []output.Artifact
	for _, s := range specs {
		p := filepath.Join(outDir, s.name)
		if _, err := os.Stat(p); err != nil {
			continue
		}
		a, err := output.ArtifactInfo(s.name, p, s.records, s.ct, s.desc)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, nil
}

func writeManifest(path string, artifacts []output.Artifact) error {
	b, err := json.MarshalIndent(artifacts, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), 0o644)
}

func writeDiff(path string) error {
	v := map[string]any{
		"baseline": true, "new_asns": 0, "removed_asns": 0, "changed_asn_profiles": 0,
		"new_prefixes": 0, "removed_prefixes": 0, "changed_prefix_origins": 0,
		"new_moas_prefixes": 0, "resolved_moas_prefixes": 0,
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), 0o644)
}
