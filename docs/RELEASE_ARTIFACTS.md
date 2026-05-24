# Release Artifacts

`asnforge.mmdb`: compact IP -> ASN profile MaxMind DB.

`asnforge-asn.jsonl` / `.csv`: canonical ASN profile table for direct ASN lookup and joins.

When built with `config/research-caida.yaml`, ASN profile rows may include optional CAIDA fields such as `as_org_id`, `as_org_name`, `caida_rank`, customer cone metrics, degree metrics, and relationship counts. These fields are intentionally kept out of the compact MMDB.

CAIDA research releases are published by the monthly/manual `release-caida` workflow as prerelease assets.

`asnforge-prefixes.jsonl` / `.csv`: canonical prefix-origin snapshot with origin ASN arrays, selected-origin policy, MOAS, collectors, and prefix confidence.

`.gz` files: compressed release assets.

`metadata.json`: build id, schema version, source hashes, artifact hashes, summary, and quality verdict.

`checksums.txt`: SHA256 checksums.

`quality-report.md`: human-readable quality summary.

`asnforge-diff.json`: v0.1 baseline or previous-release diff shape.

`manifest.json`: machine-readable artifact list.
