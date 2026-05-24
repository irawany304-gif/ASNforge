# Data Sources

ASNForge v0.1 uses public-safe source classes by default.

RIR delegated stats provide registry allocation information, including ASN ranges, allocation status, allocation date, RIR, and `registration_country`. This is registry data, not geolocation.

BGP prefix-origin snapshots provide observed routing state. v0.1 implements normalized CSV/TSV input with `prefix,origin_asn,collector,observed_at` and the bgp.tools bulk table export at `https://bgp.tools/table.jsonl`. The bgp.tools export provides JSON Lines records with `CIDR`, `ASN`, and `Hits`; ASNForge maps these to prefix-origin observations with collector `bgp.tools`.

ASN catalog enrichment uses `https://bgp.tools/asns.csv`, which provides ASN, name, and coarse class fields. ASNForge uses it for `asn_name`, `asn_org`, conservative `asn_type`, tags, and confidence. Manual overrides still take precedence.

The deterministic normalized fixture is scoped to `config/local-dev.yaml`. The public-safe profile uses the bgp.tools bulk table and ASN catalog exports and sets an identifying HTTP User-Agent.

Manual overrides provide curated corrections for ASN name, organization, type, tags, confidence, and field sources. Overrides take precedence over inferred fields.

Future optional source profiles may include PeeringDB, RPKI VRPs, and CAIDA AS Rank / AS relationships / AS2Org. CAIDA is not included in the default v0.1 public-safe release.
