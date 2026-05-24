# Classification

ASNForge uses one closed primary `asn_type`: `isp`, `hosting`, `cloud`, `cdn`, `transit`, `enterprise`, `education`, `government`, `ix`, `security`, `crawler`, or `unknown`.

Tags use a controlled vocabulary listed in `schemas/asn-profile.schema.json`.

Manual overrides take precedence. The public-safe profile uses bgp.tools ASN classes and conservative name heuristics before defaulting to `unknown`. Without PeeringDB, CAIDA, or provider catalogs, v0.1 remains conservative and does not attempt fine-grained business classification.

ASN type is a scored classification, not an authoritative fact.

Confidence is source agreement and completeness, not badness, trustworthiness, or enforcement risk.
