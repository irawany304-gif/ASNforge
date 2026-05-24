package output

import (
	"encoding/csv"
	"os"
	"strconv"
	"strings"

	"github.com/ipanalytics/ASNforge/internal/asn"
	"github.com/ipanalytics/ASNforge/internal/bgp"
)

var ASNCSVHeader = []string{"schema_version", "build_id", "asn", "asn_name", "asn_org", "asn_type", "asn_tags", "registration_country", "rir", "allocation_status", "allocation_date", "asn_confidence", "private_asn", "reserved_asn", "as_org_id", "as_org_name", "caida_rank", "caida_customer_cone_asns", "caida_customer_cone_prefixes", "caida_customer_cone_addresses", "caida_degree_peers", "caida_degree_customers", "caida_degree_providers", "caida_peer_count", "caida_customer_count", "caida_provider_count", "source_updated_at", "generated_at"}
var PrefixCSVHeader = []string{"schema_version", "build_id", "prefix", "origin_asns", "selected_origin_asn", "moas", "origin_policy", "observation_count", "source_collectors", "prefix_confidence", "rpki_state", "private_asn", "reserved_asn", "generated_at"}

func WriteASNCSV(path string, rows []asn.Profile) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	defer w.Flush()
	if err := w.Write(ASNCSVHeader); err != nil {
		return err
	}
	for _, r := range rows {
		if err := w.Write([]string{
			r.SchemaVersion, r.BuildID, strconv.FormatUint(uint64(r.ASN), 10), r.ASNName, r.ASNOrg, r.ASNType,
			strings.Join(asn.NormalizeTags(r.ASNTags), ";"), r.RegistrationCountry, r.RIR, r.AllocationStatus, r.AllocationDate,
			strconv.Itoa(r.ASNConfidence), strconv.FormatBool(r.PrivateASN), strconv.FormatBool(r.ReservedASN),
			r.ASOrgID, r.ASOrgName, strconv.Itoa(r.CAIDARank), strconv.Itoa(r.CAIDAConeASNs), strconv.Itoa(r.CAIDAConePrefixes),
			strconv.FormatUint(r.CAIDAConeAddresses, 10), strconv.Itoa(r.CAIDADegreePeers), strconv.Itoa(r.CAIDADegreeCustomers),
			strconv.Itoa(r.CAIDADegreeProviders), strconv.Itoa(r.CAIDAPeerCount), strconv.Itoa(r.CAIDACustomerCount), strconv.Itoa(r.CAIDAProviderCount),
			r.SourceUpdatedAt, r.GeneratedAt,
		}); err != nil {
			return err
		}
	}
	return w.Error()
}

func WritePrefixCSV(path string, rows []bgp.PrefixOrigin) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	defer w.Flush()
	if err := w.Write(PrefixCSVHeader); err != nil {
		return err
	}
	for _, r := range rows {
		if err := w.Write([]string{
			r.SchemaVersion, r.BuildID, r.Prefix, joinU32(r.OriginASNs), strconv.FormatUint(uint64(r.SelectedOriginASN), 10),
			strconv.FormatBool(r.MOAS), r.OriginPolicy, strconv.Itoa(r.ObservationCount), strings.Join(r.SourceCollectors, ";"),
			strconv.Itoa(r.PrefixConfidence), r.RPKIState, strconv.FormatBool(r.PrivateASN), strconv.FormatBool(r.ReservedASN), r.GeneratedAt,
		}); err != nil {
			return err
		}
	}
	return w.Error()
}

func joinU32(v []uint32) string {
	s := make([]string, 0, len(v))
	for _, x := range v {
		s = append(s, strconv.FormatUint(uint64(x), 10))
	}
	return strings.Join(s, ";")
}
