package rir

import "testing"

func TestParseDelegatedFile(t *testing.T) {
	got, err := ParseDelegatedFile("../../examples/testdata/rir/delegated-arin-extended-latest")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 {
		t.Fatalf("got %d allocations", len(got))
	}
	if got[0].RegistrationCountry != "US" || got[0].StartASN != 15169 {
		t.Fatalf("unexpected first allocation: %+v", got[0])
	}
}
