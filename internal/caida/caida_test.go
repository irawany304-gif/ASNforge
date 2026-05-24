package caida

import "testing"

func TestParseAS2Org(t *testing.T) {
	rows, err := ParseAS2Org("../../examples/testdata/caida/as-org2info.txt")
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 || rows[0].OrgName != "Google LLC" {
		t.Fatalf("unexpected rows: %+v", rows)
	}
}

func TestParseRelationships(t *testing.T) {
	counts, err := ParseRelationships("../../examples/testdata/caida/as-rel.txt")
	if err != nil {
		t.Fatal(err)
	}
	if counts[15169].Customers != 1 || counts[13335].Peers != 1 {
		t.Fatalf("unexpected counts: %+v", counts)
	}
}

func TestParseASRankCSV(t *testing.T) {
	rows, err := ParseASRankCSV("../../examples/testdata/caida/asrank.csv")
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 || rows[0].Rank != 25 || rows[0].ConeASNs != 1000 {
		t.Fatalf("unexpected rows: %+v", rows)
	}
}
