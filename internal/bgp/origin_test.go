package bgp

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMOASSelectionPolicy(t *testing.T) {
	obs := []PrefixOriginObservation{
		{Prefix: "10.0.0.0/24", OriginASN: 64512, Collector: "a"},
		{Prefix: "10.0.0.0/24", OriginASN: 64513, Collector: "b"},
		{Prefix: "10.0.0.0/24", OriginASN: 64513, Collector: "c"},
	}
	most := Aggregate(obs, "most_observed", "s", "b", "g")
	if most[0].SelectedOriginASN != 64513 || !most[0].MOAS {
		t.Fatalf("unexpected most_observed: %+v", most[0])
	}
	amb := Aggregate(obs, "mark_ambiguous", "s", "b", "g")
	if amb[0].SelectedOriginASN != 0 {
		t.Fatalf("expected ambiguous selected origin 0")
	}
}

func TestPrefixSorting(t *testing.T) {
	rows := []PrefixOrigin{{Prefix: "2001:db8::/32"}, {Prefix: "10.0.1.0/24"}, {Prefix: "10.0.0.0/24"}}
	SortPrefixes(rows)
	if rows[0].Prefix != "10.0.0.0/24" || rows[2].Prefix != "2001:db8::/32" {
		t.Fatalf("unexpected order: %+v", rows)
	}
}

func TestParseBGPToolsJSONL(t *testing.T) {
	path := filepath.Join(t.TempDir(), "table.jsonl")
	data := "{\"CIDR\":\"8.8.8.0/24\",\"ASN\":15169,\"Hits\":713}\n{\"CIDR\":\"2001:4860:4860::/48\",\"ASN\":15169,\"Hits\":100}\n"
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := ParsePreprocessedFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Prefix != "8.8.8.0/24" || got[0].ObservationCount != 713 || got[0].Collector != "bgp.tools" {
		t.Fatalf("unexpected parsed rows: %+v", got)
	}
}
