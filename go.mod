module eth2-exporter

go 1.13

require (
	github.com/bmizerany/assert v0.0.0-20160611221934-b7ed37b82869 // indirect
	github.com/ethereum/go-ethereum v1.9.14
	github.com/gogo/protobuf v1.3.1
	github.com/gorilla/mux v1.7.3
	github.com/hashicorp/golang-lru v0.5.4
	github.com/jackc/pgx/v4 v4.6.0
	github.com/jmoiron/sqlx v1.2.0
	github.com/juliangruber/go-intersect v1.0.1-0.20200323101606-4bd944a17692
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/kr/text v0.2.0 // indirect
	github.com/lib/pq v1.2.0
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/phyber/negroni-gzip v0.0.0-20180113114010-ef6356a5d029
	github.com/prysmaticlabs/ethereumapis v0.0.0-20200608211251-7dafd77461b5
	github.com/prysmaticlabs/go-bitfield v0.0.0-20200322041314-62c2aee71669
	github.com/prysmaticlabs/prysm v1.0.0-alpha.11
	github.com/sirupsen/logrus v1.6.0
	github.com/urfave/negroni v1.0.0
	github.com/zesik/proxyaddr v0.0.0-20161218060608-ec32c535184d
	google.golang.org/grpc v1.29.1
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	gopkg.in/yaml.v2 v2.2.8
)

replace github.com/ethereum/go-ethereum => github.com/prysmaticlabs/bazel-go-ethereum v0.0.0-20200530091827-df74fa9e9621
