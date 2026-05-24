package asn

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type SignalRecord struct {
	ASN        uint32
	ASNName    string
	ASNType    string
	Tags       []string
	Confidence int
	Source     string
}

func ParseSignalCSV(path string) ([]SignalRecord, error) {
	f, err := os.Open(path)
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
		index[strings.ToLower(strings.TrimSpace(h))] = i
	}
	switch {
	case hasColumns(index, "prefix", "layer", "provider", "asn", "tags"):
		return parseIPKnowledgeSignals(path, r, index)
	case hasColumns(index, "asn", "signal_count", "vpn_signals", "tor_signals", "public_feed_signals"):
		return parseASNSignalGraph(path, r, index)
	default:
		return nil, fmt.Errorf("%s: unsupported ASN signal CSV header", path)
	}
}

func parseIPKnowledgeSignals(path string, r *csv.Reader, index map[string]int) ([]SignalRecord, error) {
	var out []SignalRecord
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
		asnValue := field(rec, index, "asn")
		if asnValue == "" {
			continue
		}
		n, err := parseASNValue(asnValue)
		if err != nil {
			return nil, fmt.Errorf("%s:%d: invalid ASN %q: %w", path, line, asnValue, err)
		}
		layer := strings.ToLower(field(rec, index, "layer"))
		provider := field(rec, index, "provider")
		rawTags := strings.FieldsFunc(field(rec, index, "tags"), func(r rune) bool {
			return r == ';' || r == ',' || r == '|' || r == ' '
		})
		tags := normalizeSignalTags(layer, rawTags)
		asnType := signalTypeFromLayer(layer, tags)
		confidence := confidenceFromFloat(field(rec, index, "confidence"), 70)
		if confidence > 85 {
			confidence = 85
		}
		out = append(out, SignalRecord{
			ASN: n, ASNName: field(rec, index, "asn_name"), ASNType: asnType,
			Tags: tags, Confidence: confidence, Source: "ip-knowledge-layer",
		})
		if provider != "" && len(tags) > 0 {
			_ = provider
		}
	}
	return out, nil
}

func parseASNSignalGraph(path string, r *csv.Reader, index map[string]int) ([]SignalRecord, error) {
	var out []SignalRecord
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
		n, err := parseASNValue(field(rec, index, "asn"))
		if err != nil {
			return nil, fmt.Errorf("%s:%d: invalid ASN: %w", path, line, err)
		}
		tags := []string{}
		if intField(rec, index, "vpn_signals") > 0 {
			tags = append(tags, "vpn-adjacent", "privacy-service")
		}
		if intField(rec, index, "tor_signals") > 0 {
			tags = append(tags, "tor-adjacent", "privacy-service")
		}
		if intField(rec, index, "malware_signals")+intField(rec, index, "drop_list_signals")+intField(rec, index, "phishing_signals")+intField(rec, index, "public_feed_signals") > 0 {
			tags = append(tags, "suspicious", "security")
		}
		confidence := 55
		switch strings.ToLower(field(rec, index, "confidence")) {
		case "high":
			confidence = 80
		case "medium":
			confidence = 70
		case "low":
			confidence = 60
		}
		out = append(out, SignalRecord{
			ASN: n, ASNName: field(rec, index, "org"), ASNType: TypeUnknown,
			Tags: NormalizeTags(tags), Confidence: confidence, Source: "asn-signal-graph",
		})
	}
	return out, nil
}

func hasColumns(index map[string]int, cols ...string) bool {
	for _, col := range cols {
		if _, ok := index[col]; !ok {
			return false
		}
	}
	return true
}

func field(rec []string, index map[string]int, name string) string {
	i, ok := index[name]
	if !ok || i >= len(rec) {
		return ""
	}
	return strings.TrimSpace(rec[i])
}

func parseASNValue(raw string) (uint32, error) {
	raw = strings.TrimPrefix(strings.ToUpper(strings.TrimSpace(raw)), "AS")
	n, err := strconv.ParseUint(raw, 10, 32)
	return uint32(n), err
}

func intField(rec []string, index map[string]int, name string) int {
	v, _ := strconv.Atoi(field(rec, index, name))
	return v
}

func confidenceFromFloat(raw string, fallback int) int {
	if raw == "" {
		return fallback
	}
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return fallback
	}
	if v <= 1 {
		v *= 100
	}
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return int(v)
}

func normalizeSignalTags(layer string, raw []string) []string {
	tags := []string{}
	layer = strings.ToLower(layer)
	switch layer {
	case "hosting-cloud":
		tags = append(tags, "cloud", "hosting")
	case "crawler-bot":
		tags = append(tags, "crawler")
	case "anonymity":
		tags = append(tags, "privacy-service")
	case "asn-signal":
		tags = append(tags, "vpn-adjacent")
	}
	for _, tag := range raw {
		switch strings.ToLower(strings.TrimSpace(tag)) {
		case "cloud", "cdn", "hosting", "dns", "email", "security", "crawler", "search", "anycast":
			tags = append(tags, strings.ToLower(tag))
		case "ai-crawler", "ai-crawler-adjacent":
			tags = append(tags, "ai-crawler-adjacent")
		case "tor", "tor-exit", "tor-middle", "tor-guard":
			tags = append(tags, "tor-adjacent")
		case "vpn", "vpn-adjacent":
			tags = append(tags, "vpn-adjacent")
		case "proxy", "anonymity-network", "privacy-service":
			tags = append(tags, "privacy-service")
		case "datacenter":
			tags = append(tags, "hosting")
		}
	}
	return NormalizeTags(tags)
}

func signalTypeFromLayer(layer string, tags []string) string {
	layer = strings.ToLower(layer)
	if layer == "crawler-bot" {
		return TypeCrawler
	}
	for _, tag := range tags {
		switch tag {
		case "cdn":
			return TypeCDN
		case "cloud":
			return TypeCloud
		case "hosting":
			return TypeHosting
		case "security":
			return TypeSecurity
		}
	}
	return TypeUnknown
}
