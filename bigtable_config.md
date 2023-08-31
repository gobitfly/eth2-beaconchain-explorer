# beaconcha.in bigtable configuration
This document summarized the bigtable configuration options and table definitions required to run the beaconcha.in explorer. All settings can be applied either by using the GCP bigtable web interface or the `cbt` tool.

----
Table name: `beaconchain_validator_balances`

```
cbt -project $PROJECT -instance $INSTANCE createtable beaconchain_validator_balances
```

Column families:
* Name: `vb` | GC Policy: None

```
cbt -project $PROJECT -instance $INSTANCE createfamily beaconchain_validator_balances vb
cbt -project $PROJECT -instance $INSTANCE createfamily beaconchain_validator_balances ha
```

----
Table name: `beaconchain_validator_attestations`

```
cbt -project $PROJECT -instance $INSTANCE createtable beaconchain_validator_attestations
```

Column families:
* Name: `at` | GC Policy: None

```
cbt -project $PROJECT -instance $INSTANCE createfamily beaconchain_validator_attestations at
```
----
Table name: `beaconchain_validator_proposals`

```
cbt -project $PROJECT -instance $INSTANCE createtable beaconchain_validator_proposals
```

Column families:
* Name: `pr` | GC Policy: None

```
cbt -project $PROJECT -instance $INSTANCE createfamily beaconchain_validator_proposals pr
```
----
Table name: `beaconchain_validator_sync`

```
cbt -project $PROJECT -instance $INSTANCE createtable beaconchain_validator_sync
```

Column families:
* Name: `sc` | GC Policy: None

```
cbt -project $PROJECT -instance $INSTANCE createfamily beaconchain_validator_sync sc
cbt -project $PROJECT -instance $INSTANCE createfamily beaconchain_validator_sync sp
```
----
Table name: `beaconchain_validator_income`

```
cbt -project $PROJECT -instance $INSTANCE createtable beaconchain_validator_income
```

Column families:
* Name: `id` | GC Policy: None
* Name: `stats` | GC Policy: None

```
cbt -project $PROJECT -instance $INSTANCE createfamily beaconchain_validator_income id
cbt -project $PROJECT -instance $INSTANCE createfamily beaconchain_validator_income stats
```
----
Table name: `beaconchain_validators`

```
cbt -project $PROJECT -instance $INSTANCE createtable beaconchain_validators
```

Column families:
* Name: `at` | GC Policy: Version based policy with a maximum of 1 versions

```
cbt -project $PROJECT -instance $INSTANCE createfamily beaconchain_validators at

cbt -project $PROJECT -instance $INSTANCE setgcpolicy beaconchain_validators at maxversions=1
```
----
Table name: `blocks`

```
cbt -project $PROJECT -instance $INSTANCE createtable blocks
```

Column families:
* Name: `default` | GC Policy: Version based policy with a maximum of 1 versions

```
cbt -project $PROJECT -instance $INSTANCE createfamily blocks default

cbt -project $PROJECT -instance $INSTANCE setgcpolicy blocks default maxversions=1
```
----
Table name: `cache`

```
cbt -project $PROJECT -instance $INSTANCE createtable cache
```

Column families:
* Name: `10_min` | GC Policy: Version based policy with a maximum of 1 versions and a maximum age of 10 minutes
* Name: `1_day` | GC Policy: Version based policy with a maximum of 1 versions and a maximum age of 1 day
* Name: `1_hour` | GC Policy: Version based policy with a maximum of 1 versions and a maximum age of 1 hour

```
cbt -project $PROJECT -instance $INSTANCE createfamily cache 10_min
cbt -project $PROJECT -instance $INSTANCE createfamily cache 1_day
cbt -project $PROJECT -instance $INSTANCE createfamily cache 1_hour

cbt -project $PROJECT -instance $INSTANCE setgcpolicy cache 10_min maxage=10m and maxversions=1
cbt -project $PROJECT -instance $INSTANCE setgcpolicy cache 1_day maxage=1d and maxversions=1
cbt -project $PROJECT -instance $INSTANCE setgcpolicy cache 1_hour maxage=1h and maxversions=1
```
----
Table name: `data`

```
cbt -project $PROJECT -instance $INSTANCE createtable data
```

Column families:
* Name: `c` | GC Policy: Age based policy with a max age of 1 day
* Name: `f` | GC Policy: None

```
cbt -project $PROJECT -instance $INSTANCE createfamily data c
cbt -project $PROJECT -instance $INSTANCE createfamily data f

cbt -project $PROJECT -instance $INSTANCE setgcpolicy data c maxage=1d
```
----
Table name: `machine_metrics`

```
cbt -project $PROJECT -instance $INSTANCE createtable machine_metrics
```

Column families:
* Name: `mm` | GC Policy: Age based policy with a max age of 31 days

```
cbt -project $PROJECT -instance $INSTANCE createfamily machine_metrics mm

cbt -project $PROJECT -instance $INSTANCE setgcpolicy machine_metrics mm maxage=31d
```
----
Table name: `metadata`

```
cbt -project $PROJECT -instance $INSTANCE createtable metadata
```

Column families:
* Name: `a` | GC Policy: None
* Name: `c` | GC Policy: None
* Name: `erc1155` | GC Policy: None
* Name: `erc20` | GC Policy: None
* Name: `erc721` | GC Policy: None
* Name: `series` | GC Policy: Version based policy with a maximum of 1 versions

```
cbt -project $PROJECT -instance $INSTANCE createfamily metadata a
cbt -project $PROJECT -instance $INSTANCE createfamily metadata c
cbt -project $PROJECT -instance $INSTANCE createfamily metadata erc1155
cbt -project $PROJECT -instance $INSTANCE createfamily metadata erc20
cbt -project $PROJECT -instance $INSTANCE createfamily metadata erc721
cbt -project $PROJECT -instance $INSTANCE createfamily metadata series

cbt -project $PROJECT -instance $INSTANCE setgcpolicy metadata series maxversions=1
```
----
Table name: `metadata_updates`

```
cbt -project $PROJECT -instance $INSTANCE createtable metadata_updates
```

Column families:
* Name: `blocks` | GC Policy: Age based policy with a max age of 1 day
* Name: `f` | GC Policy: None

```
cbt -project $PROJECT -instance $INSTANCE createfamily metadata_updates blocks
cbt -project $PROJECT -instance $INSTANCE createfamily metadata_updates f

cbt -project $PROJECT -instance $INSTANCE setgcpolicy metadata_updates blocks maxage=1d
```