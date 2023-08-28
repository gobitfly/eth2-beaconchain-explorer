# beaconcha.in bigtable configuration
This document summarized the bigtable configuration options and table definitions required to run the beaconcha.in explorer. All settings can be applied either by using the GCP bigtable web interface or the `cbt` tool.

----
Table name: `beaconchain_validator_balances`

Column families:
* Name: `vb` | GC Policy: None

----
Table name: `beaconchain_validator_attestations`

Column families:
* Name: `at` | GC Policy: None

----
Table name: `beaconchain_validator_proposals`

Column families:
* Name: `pr` | GC Policy: None

----
Table name: `beaconchain_validator_sync`

Column families:
* Name: `sc` | GC Policy: None

----
Table name: `beaconchain_validator_income`

Column families:
* Name: `id` | GC Policy: None
* Name: `stats` | GC Policy: None

----
Table name: `beaconchain_validators`

Column families:
* Name: `at` | GC Policy: Version based policy with a maximum of 1 versions

----
Table name: `blocks`

Column families:
* Name: `default` | GC Policy: Version based policy with a maximum of 1 versions

----
Table name: `cache`

Column families:
* Name: `10_min` | GC Policy: Version based policy with a maximum of 1 versions and a maximum age of 10 minutes
* Name: `1_day` | GC Policy: Version based policy with a maximum of 1 versions and a maximum age of 1 day
* Name: `1_hour` | GC Policy: Version based policy with a maximum of 1 versions and a maximum age of 1 hour

----
Table name: `data`

Column families:
* Name: `c` | GC Policy: Age based policy with a max age of 1 day
* Name: `f` | GC Policy: None

----
Table name: `machine_metrics`

Column families:
* Name: `mm` | GC Policy: Age based policy with a max age of 31 days

----
Table name: `metadata`

Column families:
* Name: `a` | GC Policy: None
* Name: `c` | GC Policy: None
* Name: `erc1155` | GC Policy: None
* Name: `erc20` | GC Policy: None
* Name: `erc721` | GC Policy: None
* Name: `series` | GC Policy: Version based policy with a maximum of 1 versions

----
Table name: `metadata_updates`

Column families:
* Name: `blocks` | GC Policy: Age based policy with a max age of 1 day
* Name: `f` | GC Policy: None