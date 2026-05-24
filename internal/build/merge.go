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

func ApplyCatalog(records []asn.CatalogRecord, profiles map[uint32]asn.Profile, schema, buildID, generatedAt string) {
	for _, rec := range records {
		p := profiles[rec.ASN]
		if p.ASN == 0 {
			p.SchemaVersion = schema
			p.BuildID = buildID
			p.ASN = rec.ASN
			p.ASNType = asn.TypeUnknown
			p.ASNConfidence = 30
			p.PrivateASN = asn.IsPrivate(rec.ASN)
			p.ReservedASN = asn.IsReserved(rec.ASN)
			p.GeneratedAt = generatedAt
		}
		if p.FieldSources == nil {
			p.FieldSources = map[string][]string{}
		}
		if rec.Name != "" && p.ASNName == "" {
			p.ASNName = rec.Name
			p.ASNOrg = rec.Name
			p.FieldSources["asn_name"] = appendSource(p.FieldSources["asn_name"], "bgp.tools-asns")
			p.FieldSources["asn_org"] = appendSource(p.FieldSources["asn_org"], "bgp.tools-asns")
		}
		if rec.ASNType != "" && rec.ASNType != asn.TypeUnknown && p.ASNType == asn.TypeUnknown {
			p.ASNType = rec.ASNType
			p.FieldSources["asn_type"] = appendSource(p.FieldSources["asn_type"], "bgp.tools-asns")
		}
		if len(rec.Tags) > 0 {
			p.ASNTags = asn.NormalizeTags(append(p.ASNTags, rec.Tags...))
			p.FieldSources["asn_tags"] = appendSource(p.FieldSources["asn_tags"], "bgp.tools-asns")
		}
		if rec.Confidence > p.ASNConfidence {
			p.ASNConfidence = rec.Confidence
		}
		profiles[rec.ASN] = p
	}
}

func ApplySignals(records []asn.SignalRecord, profiles map[uint32]asn.Profile, schema, buildID, generatedAt string) {
	for _, rec := range records {
		p := profiles[rec.ASN]
		if p.ASN == 0 {
			p.SchemaVersion = schema
			p.BuildID = buildID
			p.ASN = rec.ASN
			p.ASNType = asn.TypeUnknown
			p.ASNConfidence = 30
			p.PrivateASN = asn.IsPrivate(rec.ASN)
			p.ReservedASN = asn.IsReserved(rec.ASN)
			p.GeneratedAt = generatedAt
		}
		if p.FieldSources == nil {
			p.FieldSources = map[string][]string{}
		}
		if p.ASNName == "" && rec.ASNName != "" {
			p.ASNName = rec.ASNName
			p.ASNOrg = rec.ASNName
			p.FieldSources["asn_name"] = appendSource(p.FieldSources["asn_name"], rec.Source)
			p.FieldSources["asn_org"] = appendSource(p.FieldSources["asn_org"], rec.Source)
		}
		if p.ASNType == asn.TypeUnknown && rec.ASNType != "" && rec.ASNType != asn.TypeUnknown {
			p.ASNType = rec.ASNType
			p.FieldSources["asn_type"] = appendSource(p.FieldSources["asn_type"], rec.Source)
		}
		if len(rec.Tags) > 0 {
			p.ASNTags = asn.NormalizeTags(append(p.ASNTags, rec.Tags...))
			p.FieldSources["asn_tags"] = appendSource(p.FieldSources["asn_tags"], rec.Source)
		}
		if rec.Confidence > p.ASNConfidence {
			p.ASNConfidence = rec.Confidence
		}
		profiles[rec.ASN] = p
	}
}

func ApplyCAIDA(data caidaData, profiles map[uint32]asn.Profile, schema, buildID, generatedAt string) {
	for _, rec := range data.as2org {
		p := ensureProfile(profiles, rec.ASN, schema, buildID, generatedAt)
		if p.FieldSources == nil {
			p.FieldSources = map[string][]string{}
		}
		if p.ASNName == "" && rec.ASNName != "" {
			p.ASNName = rec.ASNName
			p.FieldSources["asn_name"] = appendSource(p.FieldSources["asn_name"], "caida-as2org")
		}
		if p.ASNOrg == "" && rec.OrgName != "" {
			p.ASNOrg = rec.OrgName
			p.FieldSources["asn_org"] = appendSource(p.FieldSources["asn_org"], "caida-as2org")
		}
		p.ASOrgID = rec.OrgID
		p.ASOrgName = rec.OrgName
		p.FieldSources["as_org_id"] = appendSource(p.FieldSources["as_org_id"], "caida-as2org")
		p.FieldSources["as_org_name"] = appendSource(p.FieldSources["as_org_name"], "caida-as2org")
		profiles[rec.ASN] = p
	}
	for _, rec := range data.asrank {
		p := ensureProfile(profiles, rec.ASN, schema, buildID, generatedAt)
		if p.FieldSources == nil {
			p.FieldSources = map[string][]string{}
		}
		p.CAIDARank = rec.Rank
		p.CAIDAConeASNs = rec.ConeASNs
		p.CAIDAConePrefixes = rec.ConePrefixes
		p.CAIDAConeAddresses = rec.ConeAddresses
		p.CAIDADegreePeers = rec.DegreePeers
		p.CAIDADegreeCustomers = rec.DegreeCustomers
		p.CAIDADegreeProviders = rec.DegreeProviders
		p.FieldSources["caida_rank"] = appendSource(p.FieldSources["caida_rank"], "caida-asrank")
		profiles[rec.ASN] = p
	}
	for asnID, rec := range data.relationships {
		p := ensureProfile(profiles, asnID, schema, buildID, generatedAt)
		if p.FieldSources == nil {
			p.FieldSources = map[string][]string{}
		}
		p.CAIDAPeerCount = rec.Peers
		p.CAIDACustomerCount = rec.Customers
		p.CAIDAProviderCount = rec.Providers
		p.FieldSources["caida_relationships"] = appendSource(p.FieldSources["caida_relationships"], "caida-as-relationships")
		if p.ASNType == asn.TypeUnknown && rec.Customers >= 100 {
			p.ASNType = asn.TypeTransit
			p.ASNTags = asn.NormalizeTags(append(p.ASNTags, "transit", "backbone"))
			p.FieldSources["asn_type"] = appendSource(p.FieldSources["asn_type"], "caida-as-relationships")
			p.FieldSources["asn_tags"] = appendSource(p.FieldSources["asn_tags"], "caida-as-relationships")
		}
		profiles[asnID] = p
	}
}

func ensureProfile(profiles map[uint32]asn.Profile, id uint32, schema, buildID, generatedAt string) asn.Profile {
	p := profiles[id]
	if p.ASN != 0 {
		return p
	}
	return asn.Profile{
		SchemaVersion: schema, BuildID: buildID, ASN: id, ASNType: asn.TypeUnknown,
		ASNConfidence: 30, FieldSources: map[string][]string{},
		PrivateASN: asn.IsPrivate(id), ReservedASN: asn.IsReserved(id), GeneratedAt: generatedAt,
	}
}

func appendSource(existing []string, source string) []string {
	for _, v := range existing {
		if v == source {
			return existing
		}
	}
	return append(existing, source)
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

func ApplyNameHeuristics(profiles map[uint32]asn.Profile) {
	for id, p := range profiles {
		if p.ASNType != asn.TypeUnknown {
			continue
		}
		classifiedType, tags, confidence := asn.ClassifyName(p.ASNName, p.ASNOrg)
		if classifiedType == asn.TypeUnknown {
			continue
		}
		if p.FieldSources == nil {
			p.FieldSources = map[string][]string{}
		}
		p.ASNType = classifiedType
		p.ASNTags = asn.NormalizeTags(append(p.ASNTags, tags...))
		p.FieldSources["asn_type"] = appendSource(p.FieldSources["asn_type"], "name-heuristic")
		p.FieldSources["asn_tags"] = appendSource(p.FieldSources["asn_tags"], "name-heuristic")
		if confidence > p.ASNConfidence {
			p.ASNConfidence = confidence
		}
		profiles[id] = p
	}
}
