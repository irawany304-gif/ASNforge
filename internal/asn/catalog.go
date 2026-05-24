package asn

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type CatalogRecord struct {
	ASN        uint32
	Name       string
	SourceType string
	ASNType    string
	Tags       []string
	Confidence int
}

func ParseBGPToolsASNsCSV(path string) ([]CatalogRecord, error) {
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
	required := []string{"asn", "name", "class"}
	for _, key := range required {
		if _, ok := index[key]; !ok {
			return nil, fmt.Errorf("%s: missing required column %q", path, key)
		}
	}
	var out []CatalogRecord
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
		rawASN := strings.TrimSpace(rec[index["asn"]])
		rawASN = strings.TrimPrefix(strings.ToUpper(rawASN), "AS")
		n, err := strconv.ParseUint(rawASN, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("%s:%d: invalid ASN %q: %w", path, line, rec[index["asn"]], err)
		}
		class := strings.TrimSpace(rec[index["class"]])
		asnType, tags, confidence := ClassifyBGPToolsClass(class)
		name := strings.TrimSpace(rec[index["name"]])
		out = append(out, CatalogRecord{
			ASN: uint32(n), Name: name, SourceType: class, ASNType: asnType,
			Tags: tags, Confidence: confidence,
		})
	}
	return out, nil
}

func ClassifyBGPToolsClass(class string) (string, []string, int) {
	c := strings.ToLower(strings.TrimSpace(class))
	switch c {
	case "content", "cdn":
		return TypeCDN, []string{"cdn"}, 80
	case "hosting":
		return TypeHosting, []string{"hosting"}, 80
	case "enterprise":
		return TypeEnterprise, []string{"enterprise"}, 75
	case "education":
		return TypeEducation, []string{"education"}, 75
	case "government":
		return TypeGovernment, []string{"government"}, 75
	case "transit":
		return TypeTransit, []string{"transit", "backbone"}, 80
	case "eyeball", "isp":
		return TypeISP, []string{"broadband"}, 75
	case "ix", "internet exchange":
		return TypeIX, []string{"ix"}, 75
	case "non-profit", "nonprofit":
		return TypeEnterprise, []string{"enterprise"}, 60
	default:
		return TypeUnknown, nil, 55
	}
}
