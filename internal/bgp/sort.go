package bgp

import (
	"net/netip"
	"sort"
)

func SortPrefixes(p []PrefixOrigin) {
	sort.Slice(p, func(i, j int) bool {
		a, erra := netip.ParsePrefix(p[i].Prefix)
		b, errb := netip.ParsePrefix(p[j].Prefix)
		if erra != nil || errb != nil {
			return p[i].Prefix < p[j].Prefix
		}
		aa, bb := a.Addr(), b.Addr()
		if aa.Is4() != bb.Is4() {
			return aa.Is4()
		}
		if c := aa.Compare(bb); c != 0 {
			return c < 0
		}
		return a.Bits() < b.Bits()
	})
}
