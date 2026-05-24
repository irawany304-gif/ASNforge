package rir

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func ParseDelegatedFile(path string) ([]ASNAllocation, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var out []ASNAllocation
	sc := bufio.NewScanner(f)
	line := 0
	for sc.Scan() {
		line++
		s := strings.TrimSpace(sc.Text())
		if s == "" || strings.HasPrefix(s, "#") || strings.HasPrefix(s, "2|") {
			continue
		}
		parts := strings.Split(s, "|")
		if len(parts) < 7 || parts[2] != "asn" {
			continue
		}
		start, err := strconv.ParseUint(parts[3], 10, 32)
		if err != nil {
			return nil, fmt.Errorf("%s:%d: invalid ASN start: %w", path, line, err)
		}
		count, err := strconv.ParseUint(parts[4], 10, 32)
		if err != nil {
			return nil, fmt.Errorf("%s:%d: invalid ASN count: %w", path, line, err)
		}
		out = append(out, ASNAllocation{
			Registry: parts[0], RegistrationCountry: parts[1], StartASN: uint32(start),
			Count: uint32(count), Date: parts[5], Status: parts[6],
		})
	}
	return out, sc.Err()
}
