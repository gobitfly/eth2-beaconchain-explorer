module github.com/gobitfly/eth2-beaconchain-explorer

go 1.23.1

toolchain go1.23.3

require (
	cloud.google.com/go/bigtable v1.35.0
	cloud.google.com/go/secretmanager v1.14.2
	firebase.google.com/go/v4 v4.14.1
	github.com/ClickHouse/clickhouse-go/v2 v2.30.0
	github.com/Gurpartap/storekit-go v0.0.0-20201205024111-36b6cd5c6a21
	github.com/alexedwards/scs/redisstore v0.0.0-20230217120314-6b1bedc0f08c
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2
	github.com/awa/go-iap v1.26.1
	github.com/aws/aws-sdk-go-v2 v1.23.0
	github.com/aws/aws-sdk-go-v2/credentials v1.16.2
	github.com/aws/aws-sdk-go-v2/service/s3 v1.43.0
	github.com/aybabtme/uniplot v0.0.0-20151203143629-039c559e5e7e
	github.com/carlmjohnson/requests v0.23.4
	github.com/davecgh/go-spew v1.1.1
	github.com/doug-martin/goqu/v9 v9.19.0
	github.com/ethereum/go-ethereum v1.14.6-0.20250124151602-75526bb8e01b
	github.com/evanw/esbuild v0.8.23
	github.com/go-redis/redis/v8 v8.11.5
	github.com/gobitfly/eth-rewards v0.1.2-0.20230403064929-411ddc40a5f7
	github.com/gobitfly/scs/v2 v2.0.0-20240516120302-8754831e6b9b
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/golang-jwt/jwt/v4 v4.5.1
	github.com/golang/protobuf v1.5.4
	github.com/gomodule/redigo v1.8.0
	github.com/google/go-cmp v0.6.0
	github.com/gorilla/context v1.1.1
	github.com/gorilla/csrf v1.7.0
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/websocket v1.5.3
	github.com/hashicorp/golang-lru v0.5.5-0.20210104140557-80c98217689d
	github.com/jackc/pgx-shopspring-decimal v0.0.0-20220624020537-1d36b5a1853e
	github.com/jackc/pgx/v4 v4.18.1
	github.com/jackc/pgx/v5 v5.4.3
	github.com/jmoiron/sqlx v1.2.0
	github.com/juliangruber/go-intersect v1.1.0
	github.com/kataras/i18n v0.0.5
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/klauspost/pgzip v1.2.6
	github.com/lib/pq v1.10.7
	github.com/mailgun/mailgun-go/v4 v4.1.3
	github.com/mitchellh/mapstructure v1.5.0
	github.com/montanaflynn/stats v0.7.1
	github.com/mssola/user_agent v0.5.2
	github.com/mvdan/xurls v1.1.0
	github.com/parquet-go/parquet-go v0.25.1
	github.com/phyber/negroni-gzip v0.0.0-20180113114010-ef6356a5d029
	github.com/pkg/errors v0.9.1
	github.com/pressly/goose/v3 v3.10.0
	github.com/prometheus/client_golang v1.20.0
	github.com/protolambda/zrnt v0.32.2
	github.com/prysmaticlabs/go-bitfield v0.0.0-20240328144219-a1caa50c3a1e
	github.com/prysmaticlabs/go-ssz v0.0.0-20210121151755-f6208871c388
	github.com/prysmaticlabs/prysm/v5 v5.2.0
	github.com/rocket-pool/rocketpool-go v1.8.3-0.20240618173422-783b8668f5b4
	github.com/rocket-pool/smartnode v1.13.6
	github.com/shopspring/decimal v1.4.0
	github.com/sirupsen/logrus v1.9.3
	github.com/skip2/go-qrcode v0.0.0-20200617195104-da1b6568686e
	github.com/skygeario/go-confusable-homoglyphs v0.0.0-20191212061114-e2b2a60df110
	github.com/stripe/stripe-go/v72 v72.50.0
	github.com/urfave/negroni v1.0.0
	github.com/wealdtech/go-ens/v3 v3.6.0
	github.com/wealdtech/go-eth2-types/v2 v2.8.1
	github.com/wealdtech/go-eth2-util v1.8.1
	github.com/zesik/proxyaddr v0.0.0-20161218060608-ec32c535184d
	go.uber.org/atomic v1.11.0
	golang.org/x/crypto v0.32.0
	golang.org/x/sync v0.10.0
	golang.org/x/text v0.21.0
	golang.org/x/time v0.9.0
	google.golang.org/api v0.217.0
	google.golang.org/grpc v1.69.4
	google.golang.org/protobuf v1.36.3
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/donovanhide/eventsource v0.0.0-20210830082556-c59027999da0
	golang.org/x/exp v0.0.0-20240808152545-0cdaa3abc0fa
)

require (
	cel.dev/expr v0.16.2 // indirect
	cloud.google.com/go v0.118.0 // indirect
	cloud.google.com/go/auth v0.14.0 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.7 // indirect
	cloud.google.com/go/compute/metadata v0.6.0 // indirect
	cloud.google.com/go/iam v1.3.1 // indirect
	cloud.google.com/go/longrunning v0.6.4 // indirect
	cloud.google.com/go/monitoring v1.22.1 // indirect
	github.com/ClickHouse/ch-go v0.61.5 // indirect
	github.com/MicahParks/keyfunc v1.9.0 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/VictoriaMetrics/fastcache v1.12.2 // indirect
	github.com/alecthomas/units v0.0.0-20211218093645-b94a6e3cc137 // indirect
	github.com/alessio/shellescape v1.4.1 // indirect
	github.com/alexedwards/scs/v2 v2.5.0 // indirect
	github.com/allegro/bigcache v1.2.1 // indirect
	github.com/andybalholm/brotli v1.1.1 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.5.1 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.2.3 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.5.3 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.2.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.10.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.2.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.10.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.16.3 // indirect
	github.com/aws/smithy-go v1.17.0 // indirect
	github.com/bits-and-blooms/bitset v1.17.0 // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/census-instrumentation/opencensus-proto v0.4.1 // indirect
	github.com/cncf/xds/go v0.0.0-20240905190251-b4127c9b8d78 // indirect
	github.com/consensys/bavard v0.1.22 // indirect
	github.com/consensys/gnark-crypto v0.14.0 // indirect
	github.com/crackcomm/go-gitignore v0.0.0-20170627025303-887ab5e44cc3 // indirect
	github.com/crate-crypto/go-ipa v0.0.0-20240724233137-53bbb0ceb27a // indirect
	github.com/crate-crypto/go-kzg-4844 v1.1.0 // indirect
	github.com/deckarep/golang-set/v2 v2.6.0 // indirect
	github.com/dgraph-io/ristretto v0.1.1 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/envoyproxy/go-control-plane v0.13.1 // indirect
	github.com/envoyproxy/protoc-gen-validate v1.1.0 // indirect
	github.com/ethereum/c-kzg-4844 v1.0.0 // indirect
	github.com/ethereum/go-verkle v0.2.2 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/glendc/go-external-ip v0.1.0 // indirect
	github.com/go-faster/city v1.0.1 // indirect
	github.com/go-faster/errors v0.7.1 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/gofrs/flock v0.8.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/btree v1.1.3 // indirect
	github.com/google/s2a-go v0.1.9 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.4 // indirect
	github.com/holiman/bloomfilter/v2 v2.0.3 // indirect
	github.com/holiman/uint256 v1.3.2 // indirect
	github.com/ipfs/bbloom v0.0.4 // indirect
	github.com/ipfs/boxo v0.8.0 // indirect
	github.com/ipfs/go-bitfield v1.1.0 // indirect
	github.com/ipfs/go-block-format v0.1.2 // indirect
	github.com/ipfs/go-cid v0.4.1 // indirect
	github.com/ipfs/go-datastore v0.6.0 // indirect
	github.com/ipfs/go-ipfs-util v0.0.2 // indirect
	github.com/ipfs/go-ipld-cbor v0.0.6 // indirect
	github.com/ipfs/go-ipld-format v0.4.0 // indirect
	github.com/ipfs/go-ipld-legacy v0.1.1 // indirect
	github.com/ipfs/go-log v1.0.5 // indirect
	github.com/ipfs/go-log/v2 v2.5.1 // indirect
	github.com/ipfs/go-metrics-interface v0.0.1 // indirect
	github.com/ipld/go-codec-dagpb v1.6.0 // indirect
	github.com/ipld/go-ipld-prime v0.20.0 // indirect
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgconn v1.14.0 // indirect
	github.com/jackc/pgproto3/v2 v2.3.2 // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/jbenet/goprocess v0.1.4 // indirect
	github.com/libp2p/go-buffer-pool v0.1.0 // indirect
	github.com/mattn/go-runewidth v0.0.15 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mmcloughlin/addchain v0.4.0 // indirect
	github.com/mr-tron/base58 v1.2.0 // indirect
	github.com/multiformats/go-base32 v0.1.0 // indirect
	github.com/multiformats/go-base36 v0.2.0 // indirect
	github.com/multiformats/go-multibase v0.2.0 // indirect
	github.com/multiformats/go-multihash v0.2.3 // indirect
	github.com/multiformats/go-varint v0.0.7 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/olekukonko/tablewriter v0.0.5 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/paulmach/orb v0.11.1 // indirect
	github.com/pbnjay/memory v0.0.0-20210728143218-7b4eea64cf58 // indirect
	github.com/pierrec/lz4/v4 v4.1.21 // indirect
	github.com/planetscale/vtprotobuf v0.6.1-0.20240319094008-0393e58bdf10 // indirect
	github.com/polydawn/refmt v0.89.0 // indirect
	github.com/protolambda/zssz v0.1.5 // indirect
	github.com/prysmaticlabs/fastssz v0.0.0-20241008181541-518c4ce73516 // indirect
	github.com/prysmaticlabs/gohashtree v0.0.4-beta.0.20240624100937-73632381301b // indirect
	github.com/rivo/uniseg v0.4.4 // indirect
	github.com/segmentio/asm v1.2.0 // indirect
	github.com/spaolacci/murmur3 v1.1.0 // indirect
	github.com/wealdtech/go-bytesutil v1.2.1 // indirect
	github.com/wealdtech/go-merkletree v1.0.1-0.20190605192610-2bb163c2ea2a // indirect
	github.com/wealdtech/go-multicodec v1.4.0 // indirect
	github.com/whyrusleeping/cbor-gen v0.0.0-20230126041949-52956bd4c9aa // indirect
	github.com/whyrusleeping/chunker v0.0.0-20181014151217-fe64bd25879f // indirect
	github.com/yusufpapurcu/wmi v1.2.3 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.54.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.54.0 // indirect
	go.opentelemetry.io/otel v1.31.0 // indirect
	go.opentelemetry.io/otel/metric v1.31.0 // indirect
	go.opentelemetry.io/otel/sdk v1.31.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.31.0 // indirect
	go.opentelemetry.io/otel/trace v1.31.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	google.golang.org/appengine/v2 v2.0.2 // indirect
	google.golang.org/genproto v0.0.0-20241216192217-9240e9c98484 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250115164207-1a7da9e5054f // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250115164207-1a7da9e5054f // indirect
	lukechampine.com/blake3 v1.3.0 // indirect
	rsc.io/binaryregexp v0.2.0 // indirect
	rsc.io/tmplfunc v0.0.3 // indirect
)

require (
	cloud.google.com/go/firestore v1.17.0 // indirect
	cloud.google.com/go/storage v1.43.0 // indirect
	github.com/BurntSushi/toml v1.3.2 // indirect
	github.com/attestantio/go-eth2-client v0.19.9
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/coocood/freecache v1.2.3
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.3.0 // indirect
	github.com/fatih/color v1.16.0 // indirect
	github.com/ferranbt/fastssz v0.1.3 // indirect
	github.com/go-chi/chi v4.0.2+incompatible // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/goccy/go-yaml v1.10.0 // indirect
	github.com/golang/glog v1.2.2 // indirect
	github.com/golang/snappy v0.0.5-0.20220116011046-fa5810519dcb // indirect
	github.com/google/uuid v1.6.0
	github.com/googleapis/gax-go/v2 v2.14.1 // indirect
	github.com/gorilla/securecookie v1.1.1 // indirect
	github.com/hashicorp/go-version v1.6.0
	github.com/herumi/bls-eth-go-binary v1.29.1 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/pgtype v1.14.0
	github.com/josharian/intern v1.0.0 // indirect
	github.com/klauspost/compress v1.17.9
	github.com/klauspost/cpuid/v2 v2.2.8 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/minio/highwayhash v1.0.2 // indirect
	github.com/minio/sha256-simd v1.0.1 // indirect
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.55.0 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/shirou/gopsutil v3.21.11+incompatible // indirect
	github.com/supranational/blst v0.3.14 // indirect
	github.com/syndtr/goleveldb v1.0.1-0.20220721030215-126854af5e6d
	github.com/thomaso-mirodin/intmath v0.0.0-20160323211736-5dc6d854e46e // indirect
	github.com/tklauser/go-sysconf v0.3.13 // indirect
	github.com/tklauser/numcpus v0.7.0 // indirect
	golang.org/x/net v0.34.0
	golang.org/x/oauth2 v0.25.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
	golang.org/x/xerrors v0.0.0-20231012003039-104605ab7028 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace github.com/json-iterator/go => github.com/prestonvanloon/go v1.1.7-0.20190722034630-4f2e55fcf87b

// See https://github.com/prysmaticlabs/grpc-gateway/issues/2
replace github.com/grpc-ecosystem/grpc-gateway/v2 => github.com/prysmaticlabs/grpc-gateway/v2 v2.3.1-0.20210702154020-550e1cd83ec1

// replace github.com/prysmaticlabs/prysm/v3 => github.com/gobitfly/prysm/v3 v3.0.0-20230216184552-2f3f1e8190d5

replace github.com/wealdtech/go-merkletree v1.0.1-0.20190605192610-2bb163c2ea2a => github.com/rocket-pool/go-merkletree v1.0.1-0.20220406020931-c262d9b976dd

replace github.com/rocket-pool/rocketpool-go v1.8.2 => github.com/gobitfly/rocketpool-go v0.0.0-20240105082836-5bb7c83a2d08

// replace github.com/ethereum/go-ethereum => github.com/gobitfly/go-ethereum v1.8.13-0.20230227100926-e78d720a0bf6
