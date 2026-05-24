# Third-Party Data

The code license is separate from generated dataset terms.

Generated artifacts may contain data derived from third-party registry and routing sources. Operators are responsible for redistribution rights for the source set they configure.

The v0.1 public-safe profile uses RIR delegated stats, the bgp.tools bulk table export, the bgp.tools ASN catalog export, and static ipanalytics signal feeds from IP-Knowledge-Layer and ASN-Signal-Graph. IP-Knowledge-Layer is published as CC0-1.0, and ASN-Signal-Graph documents code under Apache-2.0 with published datasets under CC0-1.0.

The default profile avoids CAIDA by default. Optional research profiles may have additional restrictions and should be documented before publication.

`config/research-caida.yaml` is intentionally separate from `public-safe`. CAIDA datasets are governed by CAIDA acceptable-use, citation, and redistribution terms. Operators should configure CAIDA file paths or URLs only after confirming they may use and redistribute the generated artifacts for their deployment.

The `release-caida` GitHub Actions workflow runs monthly and can also be triggered manually. It publishes prereleases. Review CAIDA terms before enabling it for public distribution.
