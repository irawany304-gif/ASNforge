# Data Sources

ASNForge v0.1 uses public-safe source classes by default.

RIR delegated stats provide registry allocation information, including ASN ranges, allocation status, allocation date, RIR, and `registration_country`. This is registry data, not geolocation.

BGP prefix-origin snapshots provide observed routing state. v0.1 implements normalized CSV/TSV input with `prefix,origin_asn,collector,observed_at` and keeps a clean `PrefixOriginSource` interface for later native MRT or `bgpdump` integration. The checked-in public-safe config includes a deterministic normalized fixture as a runnable fallback; operators should configure real preprocessed public feeds for production data releases.

Manual overrides provide curated corrections for ASN name, organization, type, tags, confidence, and field sources. Overrides take precedence over inferred fields.

Future optional source profiles may include PeeringDB, RPKI VRPs, and CAIDA AS Rank / AS relationships / AS2Org. CAIDA is not included in the default v0.1 public-safe release.
