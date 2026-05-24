package asn

import "testing"

func TestParseSignalCSVIPKnowledge(t *testing.T) {
	rows, err := ParseSignalCSV("../../examples/testdata/asn/ip-knowledge-signals.csv")
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 3 {
		t.Fatalf("got %d rows", len(rows))
	}
	if rows[0].ASN != 15169 || rows[0].ASNType != TypeCloud {
		t.Fatalf("unexpected first row: %+v", rows[0])
	}
}

func TestParseSignalCSVASNSignalGraph(t *testing.T) {
	rows, err := ParseSignalCSV("../../examples/testdata/asn/asn-signal-graph.csv")
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("got %d rows", len(rows))
	}
	if rows[1].ASN != 64512 || !containsTag(rows[1].Tags, "vpn-adjacent") || !containsTag(rows[1].Tags, "tor-adjacent") {
		t.Fatalf("unexpected signal row: %+v", rows[1])
	}
}

func containsTag(tags []string, want string) bool {
	for _, tag := range tags {
		if tag == want {
			return true
		}
	}
	return false
}
