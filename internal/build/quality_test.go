package build

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ipanalytics/ASNforge/internal/asn"
	"github.com/ipanalytics/ASNforge/internal/bgp"
)

func TestQualityReportGeneration(t *testing.T) {
	dir := t.TempDir()
	mmdb := filepath.Join(dir, "x.mmdb")
	if err := os.WriteFile(mmdb, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	profiles := []asn.Profile{{ASN: 1, ASNType: asn.TypeUnknown}}
	prefixes := []bgp.PrefixOrigin{{Prefix: "1.1.1.0/24"}}
	q, _ := Evaluate(profiles, prefixes, mmdb, 1)
	if q.Verdict != "WARN" {
		t.Fatalf("expected WARN for unknown ratio, got %+v", q)
	}
	if err := WriteQualityReport(filepath.Join(dir, "q.md"), "b", "g", nil, nil, profiles, prefixes, q, []string{"ok"}); err != nil {
		t.Fatal(err)
	}
}
