package asn

import "testing"

func TestParseBGPToolsASNsCSV(t *testing.T) {
	rows, err := ParseBGPToolsASNsCSV("../../examples/testdata/asn/asns.csv")
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 4 {
		t.Fatalf("got %d rows", len(rows))
	}
	if rows[0].ASN != 15169 || rows[0].Name != "Google LLC" || rows[0].ASNType != TypeCDN {
		t.Fatalf("unexpected first row: %+v", rows[0])
	}
}

func TestClassifyBGPToolsClass(t *testing.T) {
	got, tags, conf := ClassifyBGPToolsClass("Eyeball")
	if got != TypeISP || len(tags) != 1 || tags[0] != "broadband" || conf == 0 {
		t.Fatalf("unexpected classification: %s %v %d", got, tags, conf)
	}
}
