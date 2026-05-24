package output

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ipanalytics/ASNforge/internal/asn"
)

func TestCSVWriterEscaping(t *testing.T) {
	p := filepath.Join(t.TempDir(), "asn.csv")
	err := WriteASNCSV(p, []asn.Profile{{SchemaVersion: "s", BuildID: "b", ASN: 1, ASNName: "Name, Inc.", ASNType: asn.TypeUnknown}})
	if err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(p)
	if !strings.Contains(string(b), `"Name, Inc."`) {
		t.Fatalf("expected escaped csv, got %s", b)
	}
}

func TestJSONLWriter(t *testing.T) {
	p := filepath.Join(t.TempDir(), "out.jsonl")
	if err := WriteJSONL(p, []map[string]any{{"a": 1}, {"a": 2}}); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(p)
	if strings.Count(string(b), "\n") != 2 {
		t.Fatalf("expected two jsonl lines: %q", b)
	}
}
