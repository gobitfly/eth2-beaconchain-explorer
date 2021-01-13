module eth2-exporter

go 1.15

require (
	cloud.google.com/go/firestore v1.4.0 // indirect
	firebase.google.com/go v3.13.0+incompatible
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751
	github.com/bmizerany/assert v0.0.0-20160611221934-b7ed37b82869 // indirect
	github.com/chromedp/cdproto v0.0.0-20200709115526-d1f6fc58448b
	github.com/chromedp/chromedp v0.5.3
	github.com/cloudflare/roughtime v0.0.0-20200528200038-bacff06d032d // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.0 // indirect
	github.com/davecgh/go-spew v1.1.1
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/ethereum/go-ethereum v1.9.14
	github.com/evanw/esbuild v0.8.23
	github.com/go-openapi/spec v0.19.12 // indirect
	github.com/gogo/protobuf v1.3.1
	github.com/google/go-querystring v1.0.0
	github.com/gorilla/context v1.1.1
	github.com/gorilla/csrf v1.7.0
	github.com/gorilla/mux v1.7.3
	github.com/gorilla/sessions v1.2.0
	github.com/hashicorp/golang-lru v0.5.4
	github.com/jackc/pgx/v4 v4.6.0
	github.com/jmoiron/sqlx v1.2.0
	github.com/juliangruber/go-intersect v1.0.1-0.20200323101606-4bd944a17692
	github.com/kataras/i18n v0.0.5
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/kr/text v0.2.0 // indirect
	github.com/lib/pq v1.2.0
	github.com/mailgun/mailgun-go/v4 v4.1.3
	github.com/mailru/easyjson v0.7.6 // indirect
	github.com/mssola/user_agent v0.5.2
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/phyber/negroni-gzip v0.0.0-20180113114010-ef6356a5d029
	github.com/pkg/errors v0.9.1
	github.com/protolambda/zrnt v0.12.4
	github.com/protolambda/ztyp v0.1.0
	github.com/prysmaticlabs/ethereumapis v0.0.0-20200909190233-a9190508298e
	github.com/prysmaticlabs/go-bitfield v0.0.0-20200618145306-2ae0807bef65
	github.com/prysmaticlabs/go-ssz v0.0.0-20200612203617-6d5c9aa213ae
	github.com/prysmaticlabs/prysm v1.0.0-alpha.25.0.20200917185001-3db678499074
	github.com/sirupsen/logrus v1.6.0
	github.com/stripe/stripe-go/v72 v72.24.0
	github.com/swaggo/http-swagger v0.0.0-20200308142732-58ac5e232fba
	github.com/swaggo/swag v1.6.9
	github.com/tdewolff/parse v2.3.4+incompatible // indirect
	github.com/urfave/cli/v2 v2.3.0 // indirect
	github.com/urfave/negroni v1.0.0
	github.com/zesik/proxyaddr v0.0.0-20161218060608-ec32c535184d
	golang.org/x/crypto v0.0.0-20200728195943-123391ffb6de
	golang.org/x/text v0.3.4
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e // indirect
	google.golang.org/api v0.36.0
	google.golang.org/grpc v1.33.2
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	gopkg.in/yaml.v2 v2.3.0
)

replace github.com/ethereum/go-ethereum => github.com/prysmaticlabs/bazel-go-ethereum v0.0.0-20200626171358-a933315235ec
