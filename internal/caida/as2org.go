package caida

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type AS2OrgRecord struct {
	ASN     uint32
	ASNName string
	OrgID   string
	OrgName string
	Country string
	Source  string
}

func ParseAS2Org(path string) ([]AS2OrgRecord, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	type asEntry struct {
		asn    uint32
		name   string
		orgID  string
		source string
	}
	type orgEntry struct {
		name    string
		country string
	}
	asns := []asEntry{}
	orgs := map[string]orgEntry{}
	sc := bufio.NewScanner(f)
	line := 0
	for sc.Scan() {
		line++
		s := strings.TrimSpace(sc.Text())
		if s == "" || strings.HasPrefix(s, "#") {
			continue
		}
		parts := strings.Split(s, "|")
		switch len(parts) {
		case 6:
			n, err := strconv.ParseUint(parts[0], 10, 32)
			if err != nil {
				return nil, fmt.Errorf("%s:%d: invalid ASN: %w", path, line, err)
			}
			asns = append(asns, asEntry{asn: uint32(n), name: parts[2], orgID: parts[3], source: parts[5]})
		case 5:
			orgs[parts[0]] = orgEntry{name: parts[2], country: parts[3]}
		default:
			return nil, fmt.Errorf("%s:%d: unsupported AS2Org row with %d fields", path, line, len(parts))
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	out := make([]AS2OrgRecord, 0, len(asns))
	for _, row := range asns {
		org := orgs[row.orgID]
		out = append(out, AS2OrgRecord{
			ASN: row.asn, ASNName: row.name, OrgID: row.orgID, OrgName: org.name, Country: org.country, Source: row.source,
		})
	}
	return out, nil
}
