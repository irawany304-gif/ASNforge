package asn

import "sort"

const (
	TypeISP        = "isp"
	TypeHosting    = "hosting"
	TypeCloud      = "cloud"
	TypeCDN        = "cdn"
	TypeTransit    = "transit"
	TypeEnterprise = "enterprise"
	TypeEducation  = "education"
	TypeGovernment = "government"
	TypeIX         = "ix"
	TypeSecurity   = "security"
	TypeCrawler    = "crawler"
	TypeUnknown    = "unknown"
)

var AllowedTypes = map[string]bool{
	TypeISP: true, TypeHosting: true, TypeCloud: true, TypeCDN: true, TypeTransit: true,
	TypeEnterprise: true, TypeEducation: true, TypeGovernment: true, TypeIX: true,
	TypeSecurity: true, TypeCrawler: true, TypeUnknown: true,
}

var AllowedTags = map[string]bool{
	"cloud": true, "cdn": true, "hosting": true, "residential": true, "broadband": true,
	"mobile": true, "transit": true, "backbone": true, "enterprise": true, "education": true,
	"government": true, "ix": true, "security": true, "vpn-adjacent": true,
	"privacy-service": true, "tor-adjacent": true, "crawler": true, "search": true,
	"ai-crawler-adjacent": true, "anycast": true, "dns": true, "email": true,
	"suspicious": true, "manual-override": true,
}

type Profile struct {
	SchemaVersion        string              `json:"schema_version"`
	BuildID              string              `json:"build_id"`
	ASN                  uint32              `json:"asn"`
	ASNName              string              `json:"asn_name"`
	ASNOrg               string              `json:"asn_org"`
	ASNType              string              `json:"asn_type"`
	ASNTags              []string            `json:"asn_tags"`
	RegistrationCountry  string              `json:"registration_country"`
	RIR                  string              `json:"rir"`
	AllocationStatus     string              `json:"allocation_status"`
	AllocationDate       string              `json:"allocation_date"`
	ASNConfidence        int                 `json:"asn_confidence"`
	FieldSources         map[string][]string `json:"field_sources,omitempty"`
	PrivateASN           bool                `json:"private_asn"`
	ReservedASN          bool                `json:"reserved_asn"`
	ASOrgID              string              `json:"as_org_id,omitempty"`
	ASOrgName            string              `json:"as_org_name,omitempty"`
	CAIDARank            int                 `json:"caida_rank,omitempty"`
	CAIDAConeASNs        int                 `json:"caida_customer_cone_asns,omitempty"`
	CAIDAConePrefixes    int                 `json:"caida_customer_cone_prefixes,omitempty"`
	CAIDAConeAddresses   uint64              `json:"caida_customer_cone_addresses,omitempty"`
	CAIDADegreePeers     int                 `json:"caida_degree_peers,omitempty"`
	CAIDADegreeCustomers int                 `json:"caida_degree_customers,omitempty"`
	CAIDADegreeProviders int                 `json:"caida_degree_providers,omitempty"`
	CAIDAPeerCount       int                 `json:"caida_peer_count,omitempty"`
	CAIDACustomerCount   int                 `json:"caida_customer_count,omitempty"`
	CAIDAProviderCount   int                 `json:"caida_provider_count,omitempty"`
	SourceUpdatedAt      string              `json:"source_updated_at"`
	GeneratedAt          string              `json:"generated_at"`
}

func NormalizeTags(tags []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(tags))
	for _, tag := range tags {
		if tag == "" || seen[tag] {
			continue
		}
		seen[tag] = true
		out = append(out, tag)
	}
	sort.Strings(out)
	return out
}
