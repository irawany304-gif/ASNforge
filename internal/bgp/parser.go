package bgp

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/ipanalytics/ASNforge/internal/download"
)

type PrefixOriginSource interface {
	Name() string
	Download(ctx context.Context, cacheDir string) ([]download.SourceFile, error)
	Parse(ctx context.Context, files []download.SourceFile) ([]PrefixOriginObservation, error)
}

func ParsePreprocessedFile(path string) ([]PrefixOriginObservation, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r := csv.NewReader(f)
	r.FieldsPerRecord = -1
	if strings.HasSuffix(path, ".tsv") {
		r.Comma = '\t'
	}
	var out []PrefixOriginObservation
	line := 0
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("%s:%d: %w", path, line+1, err)
		}
		line++
		if len(rec) == 0 || strings.HasPrefix(strings.TrimSpace(rec[0]), "#") {
			continue
		}
		if line == 1 && strings.EqualFold(rec[0], "prefix") {
			continue
		}
		if len(rec) < 3 {
			return nil, fmt.Errorf("%s:%d: expected prefix,origin_asn,collector", path, line)
		}
		asn, err := strconv.ParseUint(strings.TrimSpace(rec[1]), 10, 32)
		if err != nil {
			return nil, fmt.Errorf("%s:%d: invalid origin ASN: %w", path, line, err)
		}
		observed := ""
		if len(rec) > 3 {
			observed = strings.TrimSpace(rec[3])
		}
		out = append(out, PrefixOriginObservation{
			Prefix: strings.TrimSpace(rec[0]), OriginASN: uint32(asn),
			Collector: strings.TrimSpace(rec[2]), ObservedAt: observed,
		})
	}
	return out, nil
}
