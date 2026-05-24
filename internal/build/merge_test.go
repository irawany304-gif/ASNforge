package build

import (
	"testing"

	"github.com/ipanalytics/ASNforge/internal/asn"
)

func TestApplyCatalogAndNameHeuristics(t *testing.T) {
	profiles := map[uint32]asn.Profile{
		15169: {ASN: 15169, ASNType: asn.TypeUnknown, FieldSources: map[string][]string{}},
	}
	ApplyCatalog([]asn.CatalogRecord{{ASN: 15169, Name: "Google LLC", ASNType: asn.TypeUnknown, Confidence: 55}}, profiles, "s", "b", "g")
	ApplyNameHeuristics(profiles)
	got := profiles[15169]
	if got.ASNName != "Google LLC" || got.ASNType != asn.TypeCloud {
		t.Fatalf("unexpected profile: %+v", got)
	}
}

func TestApplySignals(t *testing.T) {
	profiles := map[uint32]asn.Profile{
		13335: {ASN: 13335, ASNType: asn.TypeUnknown, FieldSources: map[string][]string{}},
	}
	ApplySignals([]asn.SignalRecord{{ASN: 13335, ASNName: "Cloudflare, Inc.", ASNType: asn.TypeCDN, Tags: []string{"cdn", "security"}, Confidence: 85, Source: "test-signals"}}, profiles, "s", "b", "g")
	got := profiles[13335]
	if got.ASNType != asn.TypeCDN || got.ASNConfidence != 85 || len(got.ASNTags) != 2 {
		t.Fatalf("unexpected profile: %+v", got)
	}
}
