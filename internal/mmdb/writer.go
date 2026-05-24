package mmdb

import (
	"net"
	"net/netip"
	"os"

	"github.com/ipanalytics/ASNforge/internal/asn"
	"github.com/ipanalytics/ASNforge/internal/bgp"
	"github.com/maxmind/mmdbwriter"
	"github.com/maxmind/mmdbwriter/mmdbtype"
)

type Record struct {
	SchemaVersion       string   `json:"schema_version" maxminddb:"schema_version"`
	BuildID             string   `json:"build_id" maxminddb:"build_id"`
	ASN                 uint32   `json:"asn" maxminddb:"asn"`
	ASNName             string   `json:"asn_name" maxminddb:"asn_name"`
	ASNOrg              string   `json:"asn_org" maxminddb:"asn_org"`
	ASNType             string   `json:"asn_type" maxminddb:"asn_type"`
	ASNTags             []string `json:"asn_tags" maxminddb:"asn_tags"`
	RegistrationCountry string   `json:"registration_country" maxminddb:"registration_country"`
	RIR                 string   `json:"rir" maxminddb:"rir"`
	MOAS                bool     `json:"moas" maxminddb:"moas"`
	ASNConfidence       int      `json:"asn_confidence" maxminddb:"asn_confidence"`
}

func Write(path string, prefixes []bgp.PrefixOrigin, profiles map[uint32]asn.Profile) (int, error) {
	db, err := mmdbwriter.New(mmdbwriter.Options{
		DatabaseType:            "ASNForge-ASN-Profile",
		Description:             map[string]string{"en": "ASNForge compact IP to ASN profile"},
		DisableIPv4Aliasing:     true,
		IPVersion:               6,
		IncludeReservedNetworks: true,
	})
	if err != nil {
		return 0, err
	}
	count := 0
	for _, p := range prefixes {
		pr, err := netip.ParsePrefix(p.Prefix)
		if err != nil {
			return count, err
		}
		profile := profiles[p.SelectedOriginASN]
		rec := recordMap(p, profile)
		_, ipnet, err := net.ParseCIDR(pr.String())
		if err != nil {
			return count, err
		}
		if err := db.Insert(ipnet, rec); err != nil {
			return count, err
		}
		count++
	}
	f, err := os.Create(path)
	if err != nil {
		return count, err
	}
	defer f.Close()
	_, err = db.WriteTo(f)
	return count, err
}

func recordMap(p bgp.PrefixOrigin, prof asn.Profile) mmdbtype.Map {
	tags := mmdbtype.Slice{}
	for _, t := range prof.ASNTags {
		tags = append(tags, mmdbtype.String(t))
	}
	return mmdbtype.Map{
		"schema_version":       mmdbtype.String(p.SchemaVersion),
		"build_id":             mmdbtype.String(p.BuildID),
		"asn":                  mmdbtype.Uint32(p.SelectedOriginASN),
		"asn_name":             mmdbtype.String(prof.ASNName),
		"asn_org":              mmdbtype.String(prof.ASNOrg),
		"asn_type":             mmdbtype.String(prof.ASNType),
		"asn_tags":             tags,
		"registration_country": mmdbtype.String(prof.RegistrationCountry),
		"rir":                  mmdbtype.String(prof.RIR),
		"moas":                 mmdbtype.Bool(p.MOAS),
		"asn_confidence":       mmdbtype.Uint16(uint16(prof.ASNConfidence)),
	}
}
