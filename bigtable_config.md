# beaconcha.in bigtable configuration
This document summarized the bigtable configuration options and table definitions required to run the beaconcha.in explorer. All settings can be applied either by using the GCP bigtable web interface or the `cbt` tool.

----
Table name: `beaconchain_validators_history`

```
cbt -project $PROJECT -instance $INSTANCE createtable beaconchain_validators_history
```

Column families:
* Name: `vb` | GC Policy: None
* Name: `ha` | GC Policy: None
* Name: `at` | GC Policy: None
* Name: `pr` | GC Policy: None
* Name: `sc` | GC Policy: None
* Name: `sp` | GC Policy: None
* Name: `id` | GC Policy: None
* Name: `stats` | GC Policy: None

```
cbt -project $PROJECT -instance $INSTANCE createfamily beaconchain_validators_history vb
cbt -project $PROJECT -instance $INSTANCE createfamily beaconchain_validators_history ha
cbt -project $PROJECT -instance $INSTANCE createfamily beaconchain_validators_history at
cbt -project $PROJECT -instance $INSTANCE createfamily beaconchain_validators_history pr
cbt -project $PROJECT -instance $INSTANCE createfamily beaconchain_validators_history sc
cbt -project $PROJECT -instance $INSTANCE createfamily beaconchain_validators_history sp
cbt -project $PROJECT -instance $INSTANCE createfamily beaconchain_validators_history id
cbt -project $PROJECT -instance $INSTANCE createfamily beaconchain_validators_history stats
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