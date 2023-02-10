module eth2-exporter

go 1.18

require (
	cloud.google.com/go/bigtable v1.16.0
	cloud.google.com/go/secretmanager v1.5.0
	firebase.google.com/go v3.13.0+incompatible
	github.com/Gurpartap/storekit-go v0.0.0-20201205024111-36b6cd5c6a21
	github.com/asaskevich/govalidator v0.0.0-20210307081110-f21760c49a8d
	github.com/awa/go-iap v1.3.7
	github.com/aybabtme/uniplot v0.0.0-20151203143629-039c559e5e7e
	github.com/davecgh/go-spew v1.1.1
	github.com/ethereum/go-ethereum v1.10.23
	github.com/evanw/esbuild v0.8.23
	github.com/go-redis/redis/v8 v8.11.5
	github.com/gobitfly/eth-rewards v0.1.2-0.20230108135551-2d8c485e4c35
	github.com/gobitfly/eth.store v0.0.0-20221012115129-3b4606992669
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/golang/protobuf v1.5.2
	github.com/gorilla/context v1.1.1
	github.com/gorilla/csrf v1.7.0
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/sessions v1.2.0
	github.com/gorilla/websocket v1.5.0
	github.com/hashicorp/golang-lru v0.5.5-0.20210104140557-80c98217689d
	github.com/jackc/pgx/v4 v4.6.0
	github.com/jmoiron/sqlx v1.2.0
	github.com/juliangruber/go-intersect v1.1.0
	github.com/jung-kurt/gofpdf v1.0.3-0.20190309125859-24315acbbda5
	github.com/kataras/i18n v0.0.5
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/lib/pq v1.10.6
	github.com/mailgun/mailgun-go/v4 v4.1.3
	github.com/mitchellh/mapstructure v1.4.3
	github.com/mr-tron/base58 v1.2.0
	github.com/mssola/user_agent v0.5.2
	github.com/phyber/negroni-gzip v0.0.0-20180113114010-ef6356a5d029
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.12.2
	github.com/protolambda/zrnt v0.12.4
	github.com/protolambda/ztyp v0.1.0
	github.com/prysmaticlabs/go-bitfield v0.0.0-20210809151128-385d8c5e3fb7
	github.com/prysmaticlabs/prysm/v3 v3.1.1
	github.com/rocket-pool/rocketpool-go v1.3.0
	github.com/shopspring/decimal v1.3.1
	github.com/sirupsen/logrus v1.9.0
	github.com/skip2/go-qrcode v0.0.0-20200617195104-da1b6568686e
	github.com/stripe/stripe-go/v72 v72.50.0
	github.com/swaggo/http-swagger v1.3.0
	github.com/swaggo/swag v1.8.3
	github.com/urfave/negroni v1.0.0
	github.com/zesik/proxyaddr v0.0.0-20161218060608-ec32c535184d
	golang.org/x/crypto v0.0.0-20220622213112-05595931fe9d
	golang.org/x/sync v0.0.0-20220722155255-886fb9371eb4
	golang.org/x/text v0.5.0
	google.golang.org/api v0.85.0
	google.golang.org/genproto v0.0.0-20220801145646-83ce21fca29f
	google.golang.org/protobuf v1.28.1
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/donovanhide/eventsource v0.0.0-20210830082556-c59027999da0
	golang.org/x/exp v0.0.0-20220518171630-0b5c67f07fdf
)

require (
	cloud.google.com/go v0.102.1 // indirect
	cloud.google.com/go/compute v1.7.0 // indirect
	cloud.google.com/go/iam v0.3.0 // indirect
	contrib.go.opencensus.io/exporter/jaeger v0.2.1 // indirect
	github.com/aristanetworks/goarista v0.0.0-20200805130819-fd197cf57d96 // indirect
	github.com/benbjohnson/clock v1.3.0 // indirect
	github.com/btcsuite/btcd v0.23.1 // indirect
	github.com/btcsuite/btcd/chaincfg/chainhash v1.0.1 // indirect
	github.com/census-instrumentation/opencensus-proto v0.2.1 // indirect
	github.com/cheekybits/genny v1.0.0 // indirect
	github.com/cncf/udpa/go v0.0.0-20210930031921-04548b0d99d4 // indirect
	github.com/cncf/xds/go v0.0.0-20211011173535-cb28da3451f1 // indirect
	github.com/containerd/cgroups v1.0.4 // indirect
	github.com/coreos/go-systemd/v22 v22.3.2 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/davidlazar/go-crypto v0.0.0-20200604182044-b73af7476f6c // indirect
	github.com/dgraph-io/ristretto v0.1.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/elastic/gosigar v0.14.2 // indirect
	github.com/envoyproxy/go-control-plane v0.10.2-0.20220325020618-49ff273808a1 // indirect
	github.com/envoyproxy/protoc-gen-validate v0.1.0 // indirect
	github.com/flynn/noise v1.0.0 // indirect
	github.com/francoispqt/gojay v1.2.13 // indirect
	github.com/fsnotify/fsnotify v1.5.4 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-task/slim-sprig v0.0.0-20210107165309-348f09dbbbc0 // indirect
	github.com/godbus/dbus/v5 v5.1.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-jwt/jwt/v4 v4.3.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/gopacket v1.1.19 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.1.0 // indirect
	github.com/googleapis/go-type-adapters v1.0.0 // indirect
	github.com/gopherjs/gopherjs v0.0.0-20220410123724-9e86199038b0 // indirect
	github.com/holiman/uint256 v1.2.0 // indirect
	github.com/huin/goupnp v1.0.3 // indirect
	github.com/ipfs/go-cid v0.2.0 // indirect
	github.com/ipfs/go-ipfs-util v0.0.2 // indirect
	github.com/ipfs/go-log v1.0.5 // indirect
	github.com/ipfs/go-log/v2 v2.5.1 // indirect
	github.com/jackpal/go-nat-pmp v1.0.2 // indirect
	github.com/jbenet/go-temp-err-catcher v0.1.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/k0kubun/go-ansi v0.0.0-20180517002512-3bf9e2903213 // indirect
	github.com/kevinms/leakybucket-go v0.0.0-20200115003610-082473db97ca // indirect
	github.com/koron/go-ssdp v0.0.3 // indirect
	github.com/kr/pretty v0.3.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/libp2p/go-buffer-pool v0.1.0 // indirect
	github.com/libp2p/go-cidranger v1.1.0 // indirect
	github.com/libp2p/go-eventbus v0.2.1 // indirect
	github.com/libp2p/go-flow-metrics v0.0.3 // indirect
	github.com/libp2p/go-libp2p v0.20.3 // indirect
	github.com/libp2p/go-libp2p-asn-util v0.2.0 // indirect
	github.com/libp2p/go-libp2p-core v0.17.0 // indirect
	github.com/libp2p/go-libp2p-peerstore v0.7.0 // indirect
	github.com/libp2p/go-libp2p-pubsub v0.7.1-0.20220701163738-60cf38003244 // indirect
	github.com/libp2p/go-libp2p-resource-manager v0.3.0 // indirect
	github.com/libp2p/go-mplex v0.7.0 // indirect
	github.com/libp2p/go-msgio v0.2.0 // indirect
	github.com/libp2p/go-nat v0.1.0 // indirect
	github.com/libp2p/go-netroute v0.2.0 // indirect
	github.com/libp2p/go-openssl v0.0.7 // indirect
	github.com/libp2p/go-reuseport v0.2.0 // indirect
	github.com/libp2p/go-yamux/v3 v3.1.2 // indirect
	github.com/logrusorgru/aurora v2.0.3+incompatible // indirect
	github.com/lucas-clemente/quic-go v0.27.2 // indirect
	github.com/marten-seemann/qtls-go1-16 v0.1.5 // indirect
	github.com/marten-seemann/qtls-go1-17 v0.1.2 // indirect
	github.com/marten-seemann/qtls-go1-18 v0.1.2 // indirect
	github.com/marten-seemann/tcp v0.0.0-20210406111302-dfbc87cc63fd // indirect
	github.com/miekg/dns v1.1.50 // indirect
	github.com/mikioh/tcpinfo v0.0.0-20190314235526-30a79bb1804b // indirect
	github.com/mikioh/tcpopt v0.0.0-20190314235656-172688c1accc // indirect
	github.com/mitchellh/colorstring v0.0.0-20190213212951-d06e56a500db // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/multiformats/go-base32 v0.0.4 // indirect
	github.com/multiformats/go-base36 v0.1.0 // indirect
	github.com/multiformats/go-multiaddr v0.6.0 // indirect
	github.com/multiformats/go-multiaddr-dns v0.3.1 // indirect
	github.com/multiformats/go-multiaddr-fmt v0.1.0 // indirect
	github.com/multiformats/go-multibase v0.1.1 // indirect
	github.com/multiformats/go-multicodec v0.5.0 // indirect
	github.com/multiformats/go-multihash v0.2.0 // indirect
	github.com/multiformats/go-multistream v0.3.3 // indirect
	github.com/multiformats/go-varint v0.0.6 // indirect
	github.com/nxadm/tail v1.4.8 // indirect
	github.com/onsi/ginkgo v1.16.5 // indirect
	github.com/opencontainers/runtime-spec v1.0.2 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/pbnjay/memory v0.0.0-20210728143218-7b4eea64cf58 // indirect
	github.com/prometheus/prom2json v1.3.0 // indirect
	github.com/prysmaticlabs/fastssz v0.0.0-20220628121656-93dfe28febab // indirect
	github.com/prysmaticlabs/gohashtree v0.0.2-alpha // indirect
	github.com/prysmaticlabs/prombbolt v0.0.0-20210126082820-9b7adba6db7c // indirect
	github.com/r3labs/sse v0.0.0-20210224172625-26fe804710bc // indirect
	github.com/raulk/clock v1.1.0 // indirect
	github.com/raulk/go-watchdog v1.2.0 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/rogpeppe/go-internal v1.8.0 // indirect
	github.com/rs/cors v1.8.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/schollz/progressbar/v3 v3.3.4 // indirect
	github.com/smartystreets/assertions v1.13.0 // indirect
	github.com/smartystreets/goconvey v1.7.2 // indirect
	github.com/spacemonkeygo/spacelog v0.0.0-20180420211403-2296661a0572 // indirect
	github.com/spaolacci/murmur3 v1.1.0 // indirect
	github.com/stretchr/testify v1.8.0 // indirect
	github.com/trailofbits/go-mutexasserts v0.0.0-20200708152505-19999e7d3cef // indirect
	github.com/uber/jaeger-client-go v2.25.0+incompatible // indirect
	github.com/urfave/cli/v2 v2.10.2 // indirect
	github.com/wealdtech/go-bytesutil v1.1.1 // indirect
	github.com/whyrusleeping/multiaddr-filter v0.0.0-20160516205228-e903e4adabd7 // indirect
	github.com/whyrusleeping/timecache v0.0.0-20160911033111-cfcb2f1abfee // indirect
	github.com/xrash/smetrics v0.0.0-20201216005158-039620a65673 // indirect
	go.etcd.io/bbolt v1.3.5 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.8.0 // indirect
	go.uber.org/zap v1.21.0 // indirect
	golang.org/x/mod v0.6.0-dev.0.20220419223038-86c51ed26bb4 // indirect
	golang.org/x/term v0.3.0 // indirect
	google.golang.org/grpc v1.48.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
	k8s.io/apimachinery v0.25.0 // indirect
	k8s.io/klog/v2 v2.70.1 // indirect
	k8s.io/utils v0.0.0-20220728103510-ee6ede2d64ed // indirect
	lukechampine.com/blake3 v1.1.7 // indirect
	sigs.k8s.io/json v0.0.0-20220713155537-f223a00ba0e2 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.3 // indirect
	sigs.k8s.io/yaml v1.2.0 // indirect
)

require (
	cloud.google.com/go/firestore v1.4.0 // indirect
	cloud.google.com/go/storage v1.22.1 // indirect
	github.com/BurntSushi/toml v1.1.0 // indirect
	github.com/KyleBanks/depth v1.2.1 // indirect
	github.com/StackExchange/wmi v0.0.0-20210224194228-fe8f1750fd46 // indirect
	github.com/attestantio/go-eth2-client v0.11.4 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/btcsuite/btcd/btcec/v2 v2.2.0 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/coocood/freecache v1.2.3
	github.com/deckarep/golang-set v1.8.0 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.0.1 // indirect
	github.com/fatih/color v1.13.0 // indirect
	github.com/ferranbt/fastssz v0.0.0-20220103083642-bc5fefefa28b // indirect
	github.com/go-chi/chi v4.0.2+incompatible // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.20.0 // indirect
	github.com/go-openapi/spec v0.20.6 // indirect
	github.com/go-openapi/swag v0.21.1 // indirect
	github.com/go-sql-driver/mysql v1.6.0 // indirect
	github.com/go-stack/stack v1.8.0 // indirect
	github.com/goccy/go-yaml v1.9.5 // indirect
	github.com/golang/glog v1.0.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/go-cmp v0.5.8 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/googleapis/gax-go/v2 v2.4.0 // indirect
	github.com/gorilla/securecookie v1.1.1 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.11.3 // indirect
	github.com/hashicorp/go-version v1.6.0
	github.com/herumi/bls-eth-go-binary v0.0.0-20211108015406-b5186ba08dc7 // indirect
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgconn v1.5.0 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.0.1 // indirect
	github.com/jackc/pgservicefile v0.0.0-20200307190119-3430c5407db8 // indirect
	github.com/jackc/pgtype v1.3.0
	github.com/jackc/puddle v1.1.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/klauspost/compress v1.15.9
	github.com/klauspost/cpuid/v2 v2.0.14 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/mattn/go-runewidth v0.0.13 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/minio/highwayhash v1.0.2 // indirect
	github.com/minio/sha256-simd v1.0.0 // indirect
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.35.0 // indirect
	github.com/prometheus/procfs v0.7.3 // indirect
	github.com/r3labs/sse/v2 v2.7.4 // indirect
	github.com/rjeczalik/notify v0.9.1 // indirect
	github.com/rs/zerolog v1.26.1 // indirect
	github.com/shirou/gopsutil v3.21.4-0.20210419000835-c7a38de76ee5+incompatible // indirect
	github.com/supranational/blst v0.3.10 // indirect
	github.com/swaggo/files v0.0.0-20220610200504-28940afbdbfe // indirect
	github.com/syndtr/goleveldb v1.0.1-0.20220614013038-64ee5596c38a
	github.com/thomaso-mirodin/intmath v0.0.0-20160323211736-5dc6d854e46e // indirect
	github.com/tklauser/go-sysconf v0.3.9 // indirect
	github.com/tklauser/numcpus v0.3.0 // indirect
	go.opencensus.io v0.23.0 // indirect
	golang.org/x/net v0.4.0 // indirect
	golang.org/x/oauth2 v0.0.0-20220608161450-d0670ef3b1eb // indirect
	golang.org/x/sys v0.3.0 // indirect
	golang.org/x/time v0.0.0-20220210224613-90d013bbcef8 // indirect
	golang.org/x/tools v0.1.12 // indirect
	golang.org/x/xerrors v0.0.0-20220609144429-65e65417b02f // indirect
	gonum.org/v1/gonum v0.9.3 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	gopkg.in/cenkalti/backoff.v1 v1.1.0 // indirect
	gopkg.in/ini.v1 v1.62.0 // indirect
	gopkg.in/natefinch/npipe.v2 v2.0.0-20160621034901-c1b8fa8bdcce // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	k8s.io/client-go v0.25.0 // indirect
)

// replace github.com/ethereum/go-ethereum => github.com/prysmaticlabs/bazel-go-ethereum v0.0.0-20210707101027-e8523651bf6f

replace github.com/json-iterator/go => github.com/prestonvanloon/go v1.1.7-0.20190722034630-4f2e55fcf87b

// See https://github.com/prysmaticlabs/grpc-gateway/issues/2
replace github.com/grpc-ecosystem/grpc-gateway/v2 => github.com/prysmaticlabs/grpc-gateway/v2 v2.3.1-0.20210702154020-550e1cd83ec1
