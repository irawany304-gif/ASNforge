package build

import (
	"fmt"
	"os"
	"sort"
	"strconv"

	"github.com/ipanalytics/ASNforge/internal/asn"
	"github.com/ipanalytics/ASNforge/internal/rir"
	"gopkg.in/yaml.v3"
)

type Overrides struct {
	ASNs map[string]struct {
		ASNName       string              `yaml:"asn_name" json:"asn_name"`
		ASNOrg        string              `yaml:"asn_org" json:"asn_org"`
		ASNType       string              `yaml:"asn_type" json:"asn_type"`
		ASNTags       []string            `yaml:"asn_tags" json:"asn_tags"`
		ASNConfidence int                 `yaml:"asn_confidence" json:"asn_confidence"`
		FieldSources  map[string][]string `yaml:"field_sources" json:"field_sources"`
	} `yaml:"asns" json:"asns"`
}

func ProfilesFromAllocations(allocs []rir.ASNAllocation, schema, buildID, generatedAt string) map[uint32]asn.Profile {
	profiles := map[uint32]asn.Profile{}
	for _, a := range allocs {
		for i := uint32(0); i < a.Count; i++ {
			n := a.StartASN + i
			if _, ok := profiles[n]; ok {
				continue
			}
			profiles[n] = asn.Profile{
				SchemaVersion: schema, BuildID: buildID, ASN: n, ASNType: asn.TypeUnknown,
				ASNTags: nil, RegistrationCountry: a.RegistrationCountry, RIR: a.Registry,
				AllocationStatus: a.Status, AllocationDate: a.Date, ASNConfidence: 50,
				FieldSources: map[string][]string{
					"registration_country": {a.Registry + "-delegated"},
					"rir":                  {a.Registry + "-delegated"},
				},
				PrivateASN: asn.IsPrivate(n), ReservedASN: asn.IsReserved(n),
				SourceUpdatedAt: a.Date, GeneratedAt: generatedAt,
			}
		}
	}
	return profiles
}

func EnsureProfilesForOrigins(profiles map[uint32]asn.Profile, origins []uint32, schema, buildID, generatedAt string) {
	for _, n := range origins {
		if _, ok := profiles[n]; ok {
			continue
		}
		profiles[n] = asn.Profile{
			SchemaVersion: schema, BuildID: buildID, ASN: n, ASNType: asn.TypeUnknown, ASNConfidence: 30,
			FieldSources: map[string][]string{"asn": {"bgp-origin"}}, PrivateASN: asn.IsPrivate(n), ReservedASN: asn.IsReserved(n),
			GeneratedAt: generatedAt,
		}
	}
}

func ApplyOverrides(path string, profiles map[uint32]asn.Profile) error {
	if path == "" {
		return nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var ov Overrides
	if err := yaml.Unmarshal(b, &ov); err != nil {
		return err
	}
	for key, v := range ov.ASNs {
		n64, err := strconv.ParseUint(key, 10, 32)
		if err != nil {
			return fmt.Errorf("override ASN %q: %w", key, err)
		}
		if v.ASNType != "" && !asn.AllowedTypes[v.ASNType] {
			return fmt.Errorf("override ASN %s invalid asn_type %q", key, v.ASNType)
		}
		for _, tag := range v.ASNTags {
			if !asn.AllowedTags[tag] {
				return fmt.Errorf("override ASN %s invalid tag %q", key, tag)
			}
		}
		p := profiles[uint32(n64)]
		p.ASN = uint32(n64)
		if p.ASNType == "" {
			p.ASNType = asn.TypeUnknown
		}
		if p.FieldSources == nil {
			p.FieldSources = map[string][]string{}
		}
		if v.ASNName != "" {
			p.ASNName = v.ASNName
		}
		if v.ASNOrg != "" {
			p.ASNOrg = v.ASNOrg
		}
		if v.ASNType != "" {
			p.ASNType = v.ASNType
		}
		if v.ASNTags != nil {
			p.ASNTags = asn.NormalizeTags(append(v.ASNTags, "manual-override"))
		}
		if v.ASNConfidence > 0 {
			p.ASNConfidence = v.ASNConfidence
		}
		for f, srcs := range v.FieldSources {
			p.FieldSources[f] = srcs
		}
		p.PrivateASN = asn.IsPrivate(p.ASN)
		p.ReservedASN = asn.IsReserved(p.ASN)
		profiles[p.ASN] = p
	}
	return nil
}

func SortedProfiles(m map[uint32]asn.Profile, policy string) []asn.Profile {
	keys := make([]uint32, 0, len(m))
	for k, p := range m {
		if policy == "drop" && (p.PrivateASN || p.ReservedASN) {
			continue
		}
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	out := make([]asn.Profile, 0, len(keys))
	for _, k := range keys {
		p := m[k]
		p.ASNTags = asn.NormalizeTags(p.ASNTags)
		out = append(out, p)
	}
	return out
}
