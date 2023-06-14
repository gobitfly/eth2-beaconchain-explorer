module eth2-exporter

go 1.18

require (
	cloud.google.com/go/bigtable v1.16.0
	cloud.google.com/go/secretmanager v1.9.0
	firebase.google.com/go v3.13.0+incompatible
	github.com/Gurpartap/storekit-go v0.0.0-20201205024111-36b6cd5c6a21
	github.com/alexedwards/scs/redisstore v0.0.0-20230217120314-6b1bedc0f08c
	github.com/alexedwards/scs/v2 v2.5.0
	github.com/asaskevich/govalidator v0.0.0-20210307081110-f21760c49a8d
	github.com/awa/go-iap v1.3.7
	github.com/aybabtme/uniplot v0.0.0-20151203143629-039c559e5e7e
	github.com/davecgh/go-spew v1.1.1
	github.com/ethereum/go-ethereum v1.11.3
	github.com/evanw/esbuild v0.8.23
	github.com/go-redis/redis/v8 v8.11.5
	github.com/gobitfly/eth-rewards v0.1.2-0.20230403064929-411ddc40a5f7
	github.com/gobitfly/eth.store v0.0.0-20230306141701-814b59fb0cea
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/golang/protobuf v1.5.2
	github.com/gomodule/redigo v1.8.0
	github.com/gorilla/context v1.1.1
	github.com/gorilla/csrf v1.7.0
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/websocket v1.5.0
	github.com/hashicorp/golang-lru v0.5.5-0.20210104140557-80c98217689d
	github.com/jackc/pgx/v4 v4.18.1
	github.com/jmoiron/sqlx v1.2.0
	github.com/juliangruber/go-intersect v1.1.0
	github.com/jung-kurt/gofpdf v1.0.3-0.20190309125859-24315acbbda5
	github.com/kataras/i18n v0.0.5
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/lib/pq v1.10.7
	github.com/mailgun/mailgun-go/v4 v4.1.3
	github.com/mitchellh/mapstructure v1.5.0
	github.com/mr-tron/base58 v1.2.0
	github.com/mssola/user_agent v0.5.2
	github.com/mvdan/xurls v1.1.0
	github.com/phyber/negroni-gzip v0.0.0-20180113114010-ef6356a5d029
	github.com/pkg/errors v0.9.1
	github.com/pressly/goose/v3 v3.10.0
	github.com/prometheus/client_golang v1.14.0
	github.com/protolambda/zrnt v0.12.4
	github.com/prysmaticlabs/go-bitfield v0.0.0-20210809151128-385d8c5e3fb7
	github.com/prysmaticlabs/go-ssz v0.0.0-20210121151755-f6208871c388
	github.com/prysmaticlabs/prysm/v3 v3.2.0
	github.com/rocket-pool/rocketpool-go v1.10.1-0.20230228020137-d5a680907dff
	github.com/rocket-pool/smartnode v1.9.3
	github.com/shopspring/decimal v1.3.1
	github.com/sirupsen/logrus v1.9.0
	github.com/skip2/go-qrcode v0.0.0-20200617195104-da1b6568686e
	github.com/stripe/stripe-go/v72 v72.50.0
	github.com/swaggo/http-swagger v1.3.0
	github.com/swaggo/swag v1.8.3
	github.com/urfave/negroni v1.0.0
	github.com/wealdtech/go-eth2-types/v2 v2.8.1
	github.com/wealdtech/go-eth2-util v1.8.1
	github.com/zesik/proxyaddr v0.0.0-20161218060608-ec32c535184d
	golang.org/x/crypto v0.7.0
	golang.org/x/sync v0.1.0
	golang.org/x/text v0.8.0
	google.golang.org/api v0.102.0
	google.golang.org/protobuf v1.28.1
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/donovanhide/eventsource v0.0.0-20210830082556-c59027999da0
	golang.org/x/exp v0.0.0-20230206171751-46f607a40771
)

require (
	cloud.google.com/go v0.105.0 // indirect
	cloud.google.com/go/compute v1.12.1 // indirect
	cloud.google.com/go/compute/metadata v0.2.1 // indirect
	cloud.google.com/go/iam v0.7.0 // indirect
	cloud.google.com/go/longrunning v0.3.0 // indirect
	github.com/alessio/shellescape v1.4.1 // indirect
	github.com/census-instrumentation/opencensus-proto v0.2.1 // indirect
	github.com/cncf/udpa/go v0.0.0-20210930031921-04548b0d99d4 // indirect
	github.com/cncf/xds/go v0.0.0-20211011173535-cb28da3451f1 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/deckarep/golang-set/v2 v2.1.0 // indirect
	github.com/dgraph-io/ristretto v0.1.1 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/envoyproxy/go-control-plane v0.10.2-0.20220325020618-49ff273808a1 // indirect
	github.com/envoyproxy/protoc-gen-validate v0.1.0 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.2.0 // indirect
	github.com/imdario/mergo v0.3.13 // indirect
	github.com/ipfs/go-cid v0.3.2 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/multiformats/go-base32 v0.1.0 // indirect
	github.com/multiformats/go-base36 v0.2.0 // indirect
	github.com/multiformats/go-multibase v0.1.1 // indirect
	github.com/multiformats/go-multihash v0.2.1 // indirect
	github.com/multiformats/go-varint v0.0.7 // indirect
	github.com/pbnjay/memory v0.0.0-20210728143218-7b4eea64cf58 // indirect
	github.com/protolambda/zssz v0.1.5 // indirect
	github.com/prysmaticlabs/fastssz v0.0.0-20221107182844-78142813af44 // indirect
	github.com/prysmaticlabs/gohashtree v0.0.2-alpha // indirect
	github.com/rs/cors v1.8.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/spaolacci/murmur3 v1.1.0 // indirect
	github.com/urfave/cli v1.22.12 // indirect
	github.com/wealdtech/go-bytesutil v1.2.1 // indirect
	github.com/wealdtech/go-ens/v3 v3.5.5 // indirect
	github.com/wealdtech/go-merkletree v1.0.1-0.20190605192610-2bb163c2ea2a // indirect
	github.com/wealdtech/go-multicodec v1.4.0 // indirect
	github.com/yusufpapurcu/wmi v1.2.2 // indirect
	google.golang.org/genproto v0.0.0-20221118155620-16455021b5e6 // indirect
	google.golang.org/grpc v1.52.3 // indirect
	lukechampine.com/blake3 v1.1.7 // indirect
)

require (
	cloud.google.com/go/firestore v1.4.0 // indirect
	cloud.google.com/go/storage v1.27.0 // indirect
	github.com/BurntSushi/toml v1.2.1 // indirect
	github.com/KyleBanks/depth v1.2.1 // indirect
	github.com/attestantio/go-eth2-client v0.15.7
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/btcsuite/btcd/btcec/v2 v2.3.2 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/coocood/freecache v1.2.3
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.1.0 // indirect
	github.com/fatih/color v1.14.1 // indirect
	github.com/ferranbt/fastssz v0.1.3 // indirect
	github.com/go-chi/chi v4.0.2+incompatible // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.20.0 // indirect
	github.com/go-openapi/spec v0.20.6 // indirect
	github.com/go-openapi/swag v0.21.1 // indirect
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/goccy/go-yaml v1.10.0 // indirect
	github.com/golang/glog v1.0.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/google/uuid v1.3.0
	github.com/googleapis/gax-go/v2 v2.6.0 // indirect
	github.com/gorilla/securecookie v1.1.1 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.11.3 // indirect
	github.com/hashicorp/go-version v1.6.0
	github.com/herumi/bls-eth-go-binary v1.29.1 // indirect
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgconn v1.14.0 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.3.2 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/pgtype v1.14.0
	github.com/jackc/puddle v1.3.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/klauspost/compress v1.16.0
	github.com/klauspost/cpuid/v2 v2.2.4 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.17 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/minio/highwayhash v1.0.2 // indirect
	github.com/minio/sha256-simd v1.0.0 // indirect
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/prometheus/client_model v0.3.0 // indirect
	github.com/prometheus/common v0.39.0 // indirect
	github.com/prometheus/procfs v0.9.0 // indirect
	github.com/r3labs/sse/v2 v2.7.4 // indirect
	github.com/rs/zerolog v1.26.1 // indirect
	github.com/shirou/gopsutil v3.21.11+incompatible // indirect
	github.com/supranational/blst v0.3.10 // indirect
	github.com/swaggo/files v0.0.0-20220610200504-28940afbdbfe // indirect
	github.com/syndtr/goleveldb v1.0.1-0.20220721030215-126854af5e6d
	github.com/thomaso-mirodin/intmath v0.0.0-20160323211736-5dc6d854e46e // indirect
	github.com/tklauser/go-sysconf v0.3.11 // indirect
	github.com/tklauser/numcpus v0.6.0 // indirect
	go.opencensus.io v0.24.0 // indirect
	golang.org/x/net v0.8.0 // indirect
	golang.org/x/oauth2 v0.3.0 // indirect
	golang.org/x/sys v0.6.0 // indirect
	golang.org/x/tools v0.7.0 // indirect
	golang.org/x/xerrors v0.0.0-20220907171357-04be3eba64a2 // indirect
	gonum.org/v1/gonum v0.12.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	gopkg.in/cenkalti/backoff.v1 v1.1.0 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/natefinch/npipe.v2 v2.0.0-20160621034901-c1b8fa8bdcce // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	k8s.io/client-go v0.25.0 // indirect
)

replace github.com/json-iterator/go => github.com/prestonvanloon/go v1.1.7-0.20190722034630-4f2e55fcf87b

// See https://github.com/prysmaticlabs/grpc-gateway/issues/2
replace github.com/grpc-ecosystem/grpc-gateway/v2 => github.com/prysmaticlabs/grpc-gateway/v2 v2.3.1-0.20210702154020-550e1cd83ec1

replace github.com/prysmaticlabs/prysm/v3 => github.com/gobitfly/prysm/v3 v3.0.0-20230216184552-2f3f1e8190d5

replace github.com/wealdtech/go-merkletree v1.0.1-0.20190605192610-2bb163c2ea2a => github.com/rocket-pool/go-merkletree v1.0.1-0.20220406020931-c262d9b976dd

// replace github.com/ethereum/go-ethereum => github.com/gobitfly/go-ethereum v1.8.13-0.20230227100926-e78d720a0bf6
