# Classification

ASNForge uses one closed primary `asn_type`: `isp`, `hosting`, `cloud`, `cdn`, `transit`, `enterprise`, `education`, `government`, `ix`, `security`, `crawler`, or `unknown`.

Tags use a controlled vocabulary listed in `schemas/asn-profile.schema.json`.

Manual overrides take precedence. Without PeeringDB, CAIDA, or provider catalogs, v0.1 is conservative and defaults most ASNs to `unknown` unless a manual override or strong name heuristic applies.

ASN type is a scored classification, not an authoritative fact.

Confidence is source agreement and completeness, not badness, trustworthiness, or enforcement risk.

