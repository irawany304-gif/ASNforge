package caida

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
)

type RelationshipCounts struct {
	ASN       uint32
	Peers     int
	Customers int
	Providers int
}

func ParseRelationships(path string) (map[uint32]RelationshipCounts, error) {
	f, err := openMaybeCompressed(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	out := map[uint32]RelationshipCounts{}
	sc := bufio.NewScanner(f)
	line := 0
	for sc.Scan() {
		line++
		s := strings.TrimSpace(sc.Text())
		if s == "" || strings.HasPrefix(s, "#") {
			continue
		}
		parts := strings.Split(s, "|")
		if len(parts) < 3 {
			return nil, fmt.Errorf("%s:%d: expected asn1|asn2|relationship", path, line)
		}
		a, err := strconv.ParseUint(parts[0], 10, 32)
		if err != nil {
			return nil, fmt.Errorf("%s:%d: invalid ASN: %w", path, line, err)
		}
		b, err := strconv.ParseUint(parts[1], 10, 32)
		if err != nil {
			return nil, fmt.Errorf("%s:%d: invalid ASN: %w", path, line, err)
		}
		rel, err := strconv.Atoi(parts[2])
		if err != nil {
			return nil, fmt.Errorf("%s:%d: invalid relationship: %w", path, line, err)
		}
		left, right := out[uint32(a)], out[uint32(b)]
		left.ASN, right.ASN = uint32(a), uint32(b)
		switch rel {
		case -1:
			left.Customers++
			right.Providers++
		case 0:
			left.Peers++
			right.Peers++
		case 1:
			left.Providers++
			right.Customers++
		default:
			continue
		}
		out[left.ASN] = left
		out[right.ASN] = right
	}
	return out, sc.Err()
}
