# MMDB Output

MaxMind DB is prefix-keyed. ASNForge MMDB maps IP address to compact origin ASN profile data.

Do not expect ASN -> profile lookup from the MMDB. Use `asnforge-asn.jsonl` or `asnforge-asn.csv` for direct ASN lookup.

The MMDB intentionally excludes full origin ASN arrays, detailed MOAS state, collector observations, and long provenance objects by default. Those details live in `asnforge-prefixes.jsonl` and `.csv` to preserve MMDB data-section deduplication.

Ambiguous MOAS prefixes can be represented with `moas=true`; selected-origin behavior is controlled by `moas_policy`.

