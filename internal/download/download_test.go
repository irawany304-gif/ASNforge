package download

import "testing"

func TestURLCacheName(t *testing.T) {
	a := urlCacheName("https://example.com/a/asn-signals.csv", "asn-signals.csv")
	b := urlCacheName("https://example.com/b/asn-signals.csv", "asn-signals.csv")
	if a == b {
		t.Fatal("expected URL-specific cache names")
	}
}
