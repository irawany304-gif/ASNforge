package smoke

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/ipanalytics/ASNforge/internal/mmdb"
)

type Suite struct {
	IPLookups     []Case `json:"ip_lookups"`
	ASNLookups    []Case `json:"asn_lookups"`
	PrefixLookups []Case `json:"prefix_lookups"`
}

type Case struct {
	IP     string         `json:"ip,omitempty"`
	ASN    uint32         `json:"asn,omitempty"`
	Prefix string         `json:"prefix,omitempty"`
	Expect map[string]any `json:"expect"`
}

func Run(path, outDir string) ([]string, error) {
	if path == "" {
		return nil, nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s Suite
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, err
	}
	var results []string
	for _, c := range s.IPLookups {
		rec, ok, err := mmdb.Inspect(filepath.Join(outDir, "asnforge.mmdb"), c.IP)
		if err != nil {
			return results, err
		}
		if !ok {
			return results, fmt.Errorf("smoke ip %s not found", c.IP)
		}
		if want, ok := c.Expect["asn"]; ok && fmt.Sprint(rec.ASN) != fmt.Sprint(want) {
			return results, fmt.Errorf("smoke ip %s expected asn %v got %d", c.IP, want, rec.ASN)
		}
		results = append(results, "ip "+c.IP+" PASS")
	}
	for _, c := range s.ASNLookups {
		if err := containsLine(filepath.Join(outDir, "asnforge-asn.jsonl"), `"asn":`+strconv.FormatUint(uint64(c.ASN), 10)); err != nil {
			return results, err
		}
		results = append(results, fmt.Sprintf("asn %d PASS", c.ASN))
	}
	for _, c := range s.PrefixLookups {
		if err := containsLine(filepath.Join(outDir, "asnforge-prefixes.jsonl"), `"prefix":"`+c.Prefix+`"`); err != nil {
			return results, err
		}
		results = append(results, "prefix "+c.Prefix+" PASS")
	}
	return results, nil
}

func containsLine(path, needle string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if !bytesContains(b, []byte(needle)) {
		return fmt.Errorf("smoke expected %q in %s", needle, path)
	}
	return nil
}

func bytesContains(b, sub []byte) bool {
	for i := 0; i+len(sub) <= len(b); i++ {
		match := true
		for j := range sub {
			if b[i+j] != sub[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
