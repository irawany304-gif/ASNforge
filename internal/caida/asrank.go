package caida

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type ASRankRecord struct {
	ASN             uint32
	Rank            int
	ConeASNs        int
	ConePrefixes    int
	ConeAddresses   uint64
	DegreePeers     int
	DegreeCustomers int
	DegreeProviders int
}

func ParseASRankCSV(path string) ([]ASRankRecord, error) {
	f, err := openMaybeCompressed(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r := csv.NewReader(f)
	r.FieldsPerRecord = -1
	header, err := r.Read()
	if err != nil {
		return nil, err
	}
	index := map[string]int{}
	for i, h := range header {
		index[normalizeHeader(h)] = i
	}
	asnKey := firstKey(index, "asn", "id")
	if asnKey == "" {
		return nil, fmt.Errorf("%s: missing asn/id column", path)
	}
	var out []ASRankRecord
	line := 1
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("%s:%d: %w", path, line+1, err)
		}
		line++
		n, err := strconv.ParseUint(strings.TrimPrefix(strings.ToUpper(value(rec, index, asnKey)), "AS"), 10, 32)
		if err != nil {
			return nil, fmt.Errorf("%s:%d: invalid ASN: %w", path, line, err)
		}
		out = append(out, ASRankRecord{
			ASN:             uint32(n),
			Rank:            intValue(rec, index, "rank"),
			ConeASNs:        intValueAny(rec, index, "customer_cone_asns", "cone_asns", "cone_asn"),
			ConePrefixes:    intValueAny(rec, index, "customer_cone_prefixes", "cone_prefixes"),
			ConeAddresses:   uintValueAny(rec, index, "customer_cone_addresses", "cone_addresses"),
			DegreePeers:     intValueAny(rec, index, "degree_peers", "peers"),
			DegreeCustomers: intValueAny(rec, index, "degree_customers", "customers"),
			DegreeProviders: intValueAny(rec, index, "degree_providers", "providers", "transits"),
		})
	}
	return out, nil
}

func normalizeHeader(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, ".", "_")
	s = strings.ReplaceAll(s, "-", "_")
	return s
}

func firstKey(index map[string]int, keys ...string) string {
	for _, key := range keys {
		if _, ok := index[key]; ok {
			return key
		}
	}
	return ""
}

func value(rec []string, index map[string]int, key string) string {
	i, ok := index[key]
	if !ok || i >= len(rec) {
		return ""
	}
	return strings.TrimSpace(rec[i])
}

func intValue(rec []string, index map[string]int, key string) int {
	v, _ := strconv.Atoi(value(rec, index, key))
	return v
}

func intValueAny(rec []string, index map[string]int, keys ...string) int {
	for _, key := range keys {
		if v := intValue(rec, index, key); v != 0 {
			return v
		}
	}
	return 0
}

func uintValueAny(rec []string, index map[string]int, keys ...string) uint64 {
	for _, key := range keys {
		raw := value(rec, index, key)
		if raw == "" {
			continue
		}
		v, _ := strconv.ParseUint(raw, 10, 64)
		if v != 0 {
			return v
		}
	}
	return 0
}
