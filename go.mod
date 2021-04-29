module eth2-exporter

go 1.15

require (
	cloud.google.com/go v0.81.0
	cloud.google.com/go/firestore v1.4.0 // indirect
	firebase.google.com/go v3.13.0+incompatible
	github.com/Gurpartap/storekit-go v0.0.0-20201205024111-36b6cd5c6a21
	github.com/PuerkitoBio/goquery v1.6.1
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751
	github.com/awa/go-iap v1.3.7
	github.com/bmizerany/assert v0.0.0-20160611221934-b7ed37b82869 // indirect
	github.com/chromedp/cdproto v0.0.0-20200709115526-d1f6fc58448b
	github.com/chromedp/chromedp v0.5.3
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/ethereum/go-ethereum v1.10.2
	github.com/evanw/esbuild v0.8.23
	github.com/gin-gonic/gin v1.6.3 // indirect
	github.com/gofrs/uuid v3.3.0+incompatible // indirect
	github.com/gogo/protobuf v1.3.1
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/gorilla/context v1.1.1
	github.com/gorilla/csrf v1.7.0
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/sessions v1.2.0
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/hashicorp/golang-lru v0.5.4
	github.com/jackc/pgx/v4 v4.6.0
	github.com/jmoiron/sqlx v1.2.0
	github.com/juliangruber/go-intersect v1.0.1-0.20200323101606-4bd944a17692
	github.com/jung-kurt/gofpdf v1.0.3-0.20190309125859-24315acbbda5
	github.com/kataras/i18n v0.0.5
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/lib/pq v1.2.0
	github.com/mailgun/mailgun-go/v4 v4.1.3
	github.com/mailru/easyjson v0.7.6 // indirect
	github.com/mitchellh/mapstructure v1.3.2
	github.com/mssola/user_agent v0.5.2
	github.com/phyber/negroni-gzip v0.0.0-20180113114010-ef6356a5d029
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.9.0
	github.com/protolambda/zrnt v0.12.4
	github.com/protolambda/ztyp v0.1.0
	github.com/prysmaticlabs/eth2-types v0.0.0-20210219172114-1da477c09a06
	github.com/prysmaticlabs/ethereumapis v0.0.0-20210311175904-cf9f64632dd4
	github.com/prysmaticlabs/go-bitfield v0.0.0-20210202205921-7fcea7c45dc8
	github.com/prysmaticlabs/go-ssz v0.0.0-20210121151755-f6208871c388
	github.com/prysmaticlabs/prysm v1.3.8-geth
	github.com/rocket-pool/rocketpool-go v0.0.10
	github.com/sirupsen/logrus v1.6.0
	github.com/stripe/stripe-go/v72 v72.50.0
	github.com/swaggo/http-swagger v0.0.0-20200308142732-58ac5e232fba
	github.com/swaggo/swag v1.7.0
	github.com/urfave/cli/v2 v2.3.0 // indirect
	github.com/urfave/negroni v1.0.0
	github.com/zesik/proxyaddr v0.0.0-20161218060608-ec32c535184d
	golang.org/x/crypto v0.0.0-20210220033148-5ea612d1eb83
	golang.org/x/net v0.0.0-20210415231046-e915ea6b2b7d // indirect
	golang.org/x/oauth2 v0.0.0-20210413134643-5e61552d6c78 // indirect
	golang.org/x/sys v0.0.0-20210415045647-66c3f260301c // indirect
	golang.org/x/term v0.0.0-20210220032956-6a3ed077a48d // indirect
	golang.org/x/text v0.3.6
	google.golang.org/api v0.44.0
	google.golang.org/genproto v0.0.0-20210415145412-64678f1ae2d5
	google.golang.org/grpc v1.37.0
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v2 v2.4.0
)

replace github.com/prysmaticlabs/prysm => github.com/gobitfly/prysm v1.3.8-geth
