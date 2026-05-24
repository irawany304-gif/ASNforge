package bgp

import "testing"

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
