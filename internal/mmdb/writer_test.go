package mmdb

import (
	"path/filepath"
	"testing"

	"github.com/ipanalytics/ASNforge/internal/asn"
	"github.com/ipanalytics/ASNforge/internal/bgp"
)

func TestWriteAndInspectMMDB(t *testing.T) {
	p := filepath.Join(t.TempDir(), "x.mmdb")
	profiles := map[uint32]asn.Profile{15169: {ASN: 15169, ASNName: "Google LLC", ASNType: asn.TypeCloud, ASNConfidence: 100, RegistrationCountry: "US", RIR: "arin"}}
	prefixes := []bgp.PrefixOrigin{{SchemaVersion: "s", BuildID: "b", Prefix: "8.8.8.0/24", SelectedOriginASN: 15169}}
	if _, err := Write(p, prefixes, profiles); err != nil {
		t.Fatal(err)
	}
	rec, ok, err := Inspect(p, "8.8.8.8")
	if err != nil {
		t.Fatal(err)
	}
	if !ok || rec.ASN != 15169 || rec.ASNName != "Google LLC" {
		t.Fatalf("unexpected record ok=%v rec=%+v", ok, rec)
	}
}
