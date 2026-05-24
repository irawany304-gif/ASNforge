package bgp

import (
	"sort"

	"github.com/ipanalytics/ASNforge/internal/asn"
)

type PrefixOriginObservation struct {
	Prefix           string
	OriginASN        uint32
	Collector        string
	ObservedAt       string
	ObservationCount int
}

type PrefixOrigin struct {
	SchemaVersion     string   `json:"schema_version"`
	BuildID           string   `json:"build_id"`
	Prefix            string   `json:"prefix"`
	OriginASNs        []uint32 `json:"origin_asns"`
	SelectedOriginASN uint32   `json:"selected_origin_asn"`
	MOAS              bool     `json:"moas"`
	OriginPolicy      string   `json:"origin_policy"`
	ObservationCount  int      `json:"observation_count"`
	SourceCollectors  []string `json:"source_collectors"`
	PrefixConfidence  int      `json:"prefix_confidence"`
	RPKIState         string   `json:"rpki_state"`
	PrivateASN        bool     `json:"private_asn"`
	ReservedASN       bool     `json:"reserved_asn"`
	GeneratedAt       string   `json:"generated_at"`
}

func Aggregate(obs []PrefixOriginObservation, policy, schema, buildID, generatedAt string) []PrefixOrigin {
	type agg struct {
		counts     map[uint32]int
		collectors map[string]bool
		total      int
	}
	byPrefix := map[string]*agg{}
	for _, o := range obs {
		count := o.ObservationCount
		if count <= 0 {
			count = 1
		}
		a := byPrefix[o.Prefix]
		if a == nil {
			a = &agg{counts: map[uint32]int{}, collectors: map[string]bool{}}
			byPrefix[o.Prefix] = a
		}
		a.counts[o.OriginASN] += count
		a.collectors[o.Collector] = true
		a.total += count
	}
	out := make([]PrefixOrigin, 0, len(byPrefix))
	for prefix, a := range byPrefix {
		origins := make([]uint32, 0, len(a.counts))
		private, reserved := false, false
		for v := range a.counts {
			origins = append(origins, v)
			private = private || asn.IsPrivate(v)
			reserved = reserved || asn.IsReserved(v)
		}
		sort.Slice(origins, func(i, j int) bool { return origins[i] < origins[j] })
		collectors := make([]string, 0, len(a.collectors))
		for c := range a.collectors {
			collectors = append(collectors, c)
		}
		sort.Strings(collectors)
		moas := len(origins) > 1
		selected := selectOrigin(origins, a.counts, policy)
		conf := PrefixConfidence(a.total, len(collectors), moas, private || reserved)
		out = append(out, PrefixOrigin{
			SchemaVersion: schema, BuildID: buildID, Prefix: prefix, OriginASNs: origins,
			SelectedOriginASN: selected, MOAS: moas, OriginPolicy: policy, ObservationCount: a.total,
			SourceCollectors: collectors, PrefixConfidence: conf, RPKIState: "unknown",
			PrivateASN: private, ReservedASN: reserved, GeneratedAt: generatedAt,
		})
	}
	SortPrefixes(out)
	return out
}

func selectOrigin(origins []uint32, counts map[uint32]int, policy string) uint32 {
	if len(origins) == 0 {
		return 0
	}
	if len(origins) == 1 {
		return origins[0]
	}
	switch policy {
	case "most_observed":
		best := origins[0]
		for _, v := range origins[1:] {
			if counts[v] > counts[best] || (counts[v] == counts[best] && v < best) {
				best = v
			}
		}
		return best
	case "lowest_asn":
		return origins[0]
	default:
		return 0
	}
}

func PrefixConfidence(observationCount, collectorCount int, moas, privateOrReserved bool) int {
	v := 70
	if observationCount >= 2 {
		v += 10
	}
	if collectorCount > 1 {
		v += 10
	}
	if moas {
		v -= 20
	}
	if privateOrReserved {
		v -= 30
	}
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
}
