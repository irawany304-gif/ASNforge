package asn

import "testing"

func TestPrivateReservedASN(t *testing.T) {
	if !IsReserved(0) || !IsReserved(23456) || !IsReserved(65535) || !IsReserved(4294967295) {
		t.Fatal("expected reserved ASNs")
	}
	if !IsPrivate(64512) || !IsPrivate(4200000000) {
		t.Fatal("expected private ASNs")
	}
	if IsPrivate(15169) || IsReserved(15169) {
		t.Fatal("public ASN misclassified")
	}
}
