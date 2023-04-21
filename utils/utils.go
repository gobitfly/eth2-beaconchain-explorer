package utils

import (
	"bytes"
	securerand "crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"eth2-exporter/config"
	"eth2-exporter/price"
	"eth2-exporter/types"
	"fmt"
	"html/template"
	"image/color"
	"io/ioutil"
	"log"
	"math"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"gopkg.in/yaml.v3"

	"github.com/asaskevich/govalidator"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/params"
	"github.com/kataras/i18n"
	"github.com/kelseyhightower/envconfig"
	"github.com/lib/pq"
	"github.com/mvdan/xurls"
	"github.com/prysmaticlabs/prysm/v3/beacon-chain/core/signing"
	prysm_params "github.com/prysmaticlabs/prysm/v3/config/params"
	"github.com/sirupsen/logrus"
	"github.com/skip2/go-qrcode"
)

// Config is the globally accessible configuration
var Config *types.Config

var ErrRateLimit = errors.New("## RATE LIMIT ##")

var localiser *i18n.I18n

// making sure language files are loaded only once
func getLocaliser() *i18n.I18n {
	if localiser == nil {
		localiser, err := i18n.New(i18n.Glob("locales/*/*"), "en-US", "ru-RU")
		if err != nil {
			log.Println(err)
		}
		return localiser
	}
	return localiser
}

var HashLikeRegex = regexp.MustCompile(`^[0-9a-fA-F]{0,96}$`)

// GetTemplateFuncs will get the template functions
func GetTemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"includeHTML":                             IncludeHTML,
		"includeSvg":                              IncludeSvg,
		"formatHTML":                              FormatMessageToHtml,
		"formatBalance":                           FormatBalance,
		"formatBalanceChange":                     FormatBalanceChange,
		"formatNotificationChannel":               FormatNotificationChannel,
		"formatBalanceSql":                        FormatBalanceSql,
		"formatCurrentBalance":                    FormatCurrentBalance,
		"formatEffectiveBalance":                  FormatEffectiveBalance,
		"formatBlockStatus":                       FormatBlockStatus,
		"formatBlockSlot":                         FormatBlockSlot,
		"formatSlotToTimestamp":                   FormatSlotToTimestamp,
		"formatDepositAmount":                     FormatDepositAmount,
		"formatEpoch":                             FormatEpoch,
		"fixAddressCasing":                        FixAddressCasing,
		"formatAddressLong":                       FormatAddressLong,
		"formatHashLong":                          FormatHashLong,
		"formatEth1Block":                         FormatEth1Block,
		"formatEth1BlockHash":                     FormatEth1BlockHash,
		"formatEth1Address":                       FormatEth1Address,
		"formatEth1AddressStringLowerCase":        FormatEth1AddressStringLowerCase,
		"formatEth1TxHash":                        FormatEth1TxHash,
		"formatGraffiti":                          FormatGraffiti,
		"formatHash":                              FormatHash,
		"formatWithdawalCredentials":              FormatWithdawalCredentials,
		"formatAddressToWithdrawalCredentials":    FormatAddressToWithdrawalCredentials,
		"formatBitvector":                         FormatBitvector,
		"formatBitlist":                           FormatBitlist,
		"formatBitvectorValidators":               formatBitvectorValidators,
		"formatParticipation":                     FormatParticipation,
		"formatIncome":                            FormatIncome,
		"formatIncomeNoCurrency":                  FormatIncomeNoCurrency,
		"formatIncomeSql":                         FormatIncomeSql,
		"formatSqlInt64":                          FormatSqlInt64,
		"formatValidator":                         FormatValidator,
		"formatValidatorWithName":                 FormatValidatorWithName,
		"formatValidatorInt64":                    FormatValidatorInt64,
		"formatValidatorStatus":                   FormatValidatorStatus,
		"formatPercentage":                        FormatPercentage,
		"formatPercentageWithPrecision":           FormatPercentageWithPrecision,
		"formatPercentageWithGPrecision":          FormatPercentageWithGPrecision,
		"formatPercentageColored":                 FormatPercentageColored,
		"formatPercentageColoredEmoji":            FormatPercentageColoredEmoji,
		"formatPublicKey":                         FormatPublicKey,
		"formatSlashedValidator":                  FormatSlashedValidator,
		"formatSlashedValidatorInt64":             FormatSlashedValidatorInt64,
		"formatTimestamp":                         FormatTimestamp,
		"formatTsWithoutTooltip":                  FormatTsWithoutTooltip,
		"formatTimestampTs":                       FormatTimestampTs,
		"formatTime":                              FormatTime,
		"formatValidatorName":                     FormatValidatorName,
		"formatAttestationInclusionEffectiveness": FormatAttestationInclusionEffectiveness,
		"formatValidatorTags":                     FormatValidatorTags,
		"formatValidatorTag":                      FormatValidatorTag,
		"formatRPL":                               FormatRPL,
		"formatETH":                               FormatETH,
		"formatFloat":                             FormatFloat,
		"formatAmount":                            FormatAmount,
		"formatExchangedAmount":                   FormatExchangedAmount,
		"formatBigAmount":                         FormatBigAmount,
		"formatBytesAmount":                       FormatBytesAmount,
		"formatYesNo":                             FormatYesNo,
		"formatAmountFormatted":                   FormatAmountFormatted,
		"formatAddressAsLink":                     FormatAddressAsLink,
		"formatBuilder":                           FormatBuilder,
		"formatDifficulty":                        FormatDifficulty,
		"getCurrencyLabel":                        price.GetCurrencyLabel,
		"epochOfSlot":                             EpochOfSlot,
		"dayToTime":                               DayToTime,
		"contains":                                strings.Contains,
		"roundDecimals":                           RoundDecimals,
		"bigIntCmp":                               func(i *big.Int, j int) int { return i.Cmp(big.NewInt(int64(j))) },
		"mod":                                     func(i, j int) bool { return i%j == 0 },
		"sub":                                     func(i, j int) int { return i - j },
		"subUI64":                                 func(i, j uint64) uint64 { return i - j },
		"add":                                     func(i, j int) int { return i + j },
		"addI64":                                  func(i, j int64) int64 { return i + j },
		"addUI64":                                 func(i, j uint64) uint64 { return i + j },
		"addFloat64":                              func(i, j float64) float64 { return i + j },
		"mul":                                     func(i, j float64) float64 { return i * j },
		"div":                                     func(i, j float64) float64 { return i / j },
		"divInt":                                  func(i, j int) float64 { return float64(i) / float64(j) },
		"nef":                                     func(i, j float64) bool { return i != j },
		"gtf":                                     func(i, j float64) bool { return i > j },
		"ltf":                                     func(i, j float64) bool { return i < j },
		"round": func(i float64, n int) float64 {
			return math.Round(i*math.Pow10(n)) / math.Pow10(n)
		},
		"percent": func(i float64) float64 { return i * 100 },
		"formatThousands": func(i float64) string {
			p := message.NewPrinter(language.English)
			return p.Sprintf("%.0f\n", i)
		},
		"formatThousandsFancy": func(i float64) string {
			p := message.NewPrinter(language.English)
			return p.Sprintf("%v\n", i)
		},
		"formatThousandsInt": func(i int) string {
			p := message.NewPrinter(language.English)
			return p.Sprintf("%d", i)
		},
		"formatStringThousands": FormatThousandsEnglish,
		"derefString":           DerefString,
		"trLang":                TrLang,
		"firstCharToUpper":      func(s string) string { return cases.Title(language.English).String(s) },
		"eqsp": func(a, b *string) bool {
			if a != nil && b != nil {
				return *a == *b
			}
			return false
		},
		"stringsJoin":     strings.Join,
		"formatAddCommas": FormatAddCommas,
		"encodeToString":  hex.EncodeToString,

		"formatTokenBalance":      FormatTokenBalance,
		"formatAddressEthBalance": FormatAddressEthBalance,
		"toBase64":                ToBase64,
		"bytesToNumberString": func(input []byte) string {
			return new(big.Int).SetBytes(input).String()
		},
		"bigQuo": func(num []byte, denom []byte) string {
			numFloat := new(big.Float).SetInt(new(big.Int).SetBytes(num))
			denomFloat := new(big.Float).SetInt(new(big.Int).SetBytes(denom))
			res := new(big.Float).Quo(numFloat, denomFloat)
			return res.Text('f', int(res.MinPrec()))
		},
		"bigDecimalShift": func(num []byte, shift []byte) string {
			numFloat := new(big.Float).SetInt(new(big.Int).SetBytes(num))
			denom := new(big.Int).Exp(big.NewInt(10), new(big.Int).SetBytes(shift), nil)
			// shift := new(big.Float).SetInt(new(big.Int).SetBytes(shift))
			res := new(big.Float).Quo(numFloat, new(big.Float).SetInt(denom))
			return res.Text('f', int(res.MinPrec()))
		},
		"trimTrailingZero": func(num string) string {
			if strings.Contains(num, ".") {
				return strings.TrimRight(strings.TrimRight(num, "0"), ".")
			}
			return num
		},
		// ETH1 related formatting
		"formatEth1TxStatus":    FormatEth1TxStatus,
		"formatTimestampUInt64": FormatTimestampUInt64,
		"formatEth1AddressFull": FormatEth1AddressFull,
		"byteToString": func(num []byte) string {
			return string(num)
		},
		"bigToInt": func(val *hexutil.Big) *big.Int {
			if val != nil {
				return val.ToInt()
			}
			return nil
		},
		"formatBigNumberAddCommasFormated": FormatBigNumberAddCommasFormated,
		"formatEthstoreComparison":         FormatEthstoreComparison,
		"formatPoolPerformance":            FormatPoolPerformance,
		"formatTokenSymbolTitle":           FormatTokenSymbolTitle,
		"formatTokenSymbol":                FormatTokenSymbol,
		"formatTokenSymbolHTML":            FormatTokenSymbolHTML,
		"dict": func(values ...interface{}) (map[string]interface{}, error) {
			if len(values)%2 != 0 {
				return nil, errors.New("invalid dict call")
			}
			dict := make(map[string]interface{}, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, errors.New("dict keys must be strings")
				}
				dict[key] = values[i+1]
			}
			return dict, nil
		},
	}
}

// IncludeHTML adds html to the page
func IncludeHTML(path string) template.HTML {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		log.Printf("includeHTML - error reading file: %v", err)
		return ""
	}
	return template.HTML(string(b))
}

func GraffitiToSring(graffiti []byte) string {
	s := strings.Map(fixUtf, string(bytes.Trim(graffiti, "\x00")))
	s = strings.Replace(s, "\u0000", "", -1) // rempove 0x00 bytes as it is not supported in postgres

	if !utf8.ValidString(s) {
		return "INVALID_UTF8_STRING"
	}

	return s
}

// FormatGraffitiString formats (and escapes) the graffiti
func FormatGraffitiString(graffiti string) string {
	return strings.Map(fixUtf, template.HTMLEscapeString(graffiti))
}

func fixUtf(r rune) rune {
	if r == utf8.RuneError {
		return -1
	}
	return r
}

func SyncPeriodOfEpoch(epoch uint64) uint64 {
	if epoch < Config.Chain.Config.AltairForkEpoch {
		return 0
	}
	return epoch / Config.Chain.Config.EpochsPerSyncCommitteePeriod
}

func FirstEpochOfSyncPeriod(syncPeriod uint64) uint64 {
	return syncPeriod * Config.Chain.Config.EpochsPerSyncCommitteePeriod
}

func TimeToSyncPeriod(t time.Time) uint64 {
	return SyncPeriodOfEpoch(uint64(TimeToEpoch(t)))
}

// EpochOfSlot returns the corresponding epoch of a slot
func EpochOfSlot(slot uint64) uint64 {
	return slot / Config.Chain.Config.SlotsPerEpoch
}

// DayOfSlot returns the corresponding day of a slot
func DayOfSlot(slot uint64) uint64 {
	return Config.Chain.Config.SecondsPerSlot * slot / (24 * 3600)
}

// WeekOfSlot returns the corresponding week of a slot
func WeekOfSlot(slot uint64) uint64 {
	return Config.Chain.Config.SecondsPerSlot * slot / (7 * 24 * 3600)
}

// SlotToTime returns a time.Time to slot
func SlotToTime(slot uint64) time.Time {
	return time.Unix(int64(Config.Chain.GenesisTimestamp+slot*Config.Chain.Config.SecondsPerSlot), 0)
}

// TimeToSlot returns time to slot in seconds
func TimeToSlot(timestamp uint64) uint64 {
	if Config.Chain.GenesisTimestamp > timestamp {
		return 0
	}
	return (timestamp - Config.Chain.GenesisTimestamp) / Config.Chain.Config.SecondsPerSlot
}

// EpochToTime will return a time.Time for an epoch
func EpochToTime(epoch uint64) time.Time {
	return time.Unix(int64(Config.Chain.GenesisTimestamp+epoch*Config.Chain.Config.SecondsPerSlot*Config.Chain.Config.SlotsPerEpoch), 0)
}

// TimeToDay will return a days since genesis for an timestamp
func TimeToDay(timestamp uint64) uint64 {
	return uint64(time.Unix(int64(timestamp), 0).Sub(time.Unix(int64(Config.Chain.GenesisTimestamp), 0)).Hours() / 24)
	// return time.Unix(int64(Config.Chain.GenesisTimestamp), 0).Add(time.Hour * time.Duration(24*int(day)))
}

func DayToTime(day int64) time.Time {
	return time.Unix(int64(Config.Chain.GenesisTimestamp), 0).Add(time.Hour * time.Duration(24*int(day)))
}

// TimeToEpoch will return an epoch for a given time
func TimeToEpoch(ts time.Time) int64 {
	if int64(Config.Chain.GenesisTimestamp) > ts.Unix() {
		return 0
	}
	return (ts.Unix() - int64(Config.Chain.GenesisTimestamp)) / int64(Config.Chain.Config.SecondsPerSlot) / int64(Config.Chain.Config.SlotsPerEpoch)
}

func WeiToEther(wei *big.Int) *big.Float {
	return new(big.Float).Quo(new(big.Float).SetInt(wei), big.NewFloat(params.Ether))
}

// WaitForCtrlC will block/wait until a control-c is pressed
func WaitForCtrlC() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}

// ReadConfig will process a configuration
func ReadConfig(cfg *types.Config, path string) error {

	if strings.HasPrefix(path, "projects/") {
		x, err := AccessSecretVersion(path)
		if err != nil {
			return fmt.Errorf("error getting config from secret store: %v", err)
		}
		err = yaml.Unmarshal([]byte(*x), cfg)
		if err != nil {
			return fmt.Errorf("error decoding config file %v: %v", path, err)
		}

		logger.Infof("seeded config file from secret store")
	} else {

		err := readConfigFile(cfg, path)
		if err != nil {
			return err
		}
	}

	readConfigEnv(cfg)
	err := readConfigSecrets(cfg)
	if err != nil {
		return err
	}

	if cfg.Chain.ConfigPath == "" {
		// var prysmParamsConfig *prysmParams.BeaconChainConfig
		switch cfg.Chain.Name {
		case "mainnet":
			err = yaml.Unmarshal([]byte(config.MainnetChainYml), &cfg.Chain.Config)
		case "prater":
			err = yaml.Unmarshal([]byte(config.PraterChainYml), &cfg.Chain.Config)
		case "ropsten":
			err = yaml.Unmarshal([]byte(config.RopstenChainYml), &cfg.Chain.Config)
		case "sepolia":
			err = yaml.Unmarshal([]byte(config.SepoliaChainYml), &cfg.Chain.Config)
		case "gnosis":
			err = yaml.Unmarshal([]byte(config.GnosisChainYml), &cfg.Chain.Config)
		default:
			return fmt.Errorf("tried to set known chain-config, but unknown chain-name")
		}
		if err != nil {
			return err
		}
		// err = prysmParams.SetActive(prysmParamsConfig)
		// if err != nil {
		// 	return fmt.Errorf("error setting chainConfig (%v) for prysmParams: %w", cfg.Chain.Name, err)
		// }
	} else {
		f, err := os.Open(cfg.Chain.ConfigPath)
		if err != nil {
			return fmt.Errorf("error opening Chain Config file %v: %w", cfg.Chain.ConfigPath, err)
		}
		var chainConfig *types.ChainConfig
		decoder := yaml.NewDecoder(f)
		err = decoder.Decode(&chainConfig)
		if err != nil {
			return fmt.Errorf("error decoding Chain Config file %v: %v", cfg.Chain.ConfigPath, err)
		}
		cfg.Chain.Config = *chainConfig
		// err = prysmParams.LoadChainConfigFile(cfg.Chain.ConfigPath, nil)
		// if err != nil {
		// 	return fmt.Errorf("error loading chainConfig (%v) for prysmParams: %w", cfg.Chain.ConfigPath, err)
		// }
	}
	cfg.Chain.Name = cfg.Chain.Config.ConfigName

	if cfg.Chain.GenesisTimestamp == 0 {
		switch cfg.Chain.Name {
		case "mainnet":
			cfg.Chain.GenesisTimestamp = 1606824023
		case "prater":
			cfg.Chain.GenesisTimestamp = 1616508000
		case "sepolia":
			cfg.Chain.GenesisTimestamp = 1655733600
		case "zhejiang":
			cfg.Chain.GenesisTimestamp = 1675263600
		case "gnosis":
			cfg.Chain.GenesisTimestamp = 1638993340
		default:
			return fmt.Errorf("tried to set known genesis-timestamp, but unknown chain-name")
		}
	}

	if cfg.Chain.GenesisValidatorsRoot == "" {
		switch cfg.Chain.Name {
		case "mainnet":
			cfg.Chain.GenesisValidatorsRoot = "0x4b363db94e286120d76eb905340fdd4e54bfe9f06bf33ff6cf5ad27f511bfe95"
		case "prater":
			cfg.Chain.GenesisValidatorsRoot = "0x043db0d9a83813551ee2f33450d23797757d430911a9320530ad8a0eabc43efb"
		case "sepolia":
			cfg.Chain.GenesisValidatorsRoot = "0xd8ea171f3c94aea21ebc42a1ed61052acf3f9209c00e4efbaaddac09ed9b8078"
		case "zhejiang":
			cfg.Chain.GenesisValidatorsRoot = "0x53a92d8f2bb1d85f62d16a156e6ebcd1bcaba652d0900b2c2f387826f3481f6f"
		case "gnosis":
			cfg.Chain.GenesisValidatorsRoot = "0xf5dcb5564e829aab27264b9becd5dfaa017085611224cb3036f573368dbb9d47"
		default:
			return fmt.Errorf("tried to set known genesis-validators-root, but unknown chain-name")
		}
	}

	if cfg.Chain.DomainBLSToExecutionChange == "" {
		cfg.Chain.DomainBLSToExecutionChange = "0x0A000000"
	}
	if cfg.Chain.DomainVoluntaryExit == "" {
		cfg.Chain.DomainVoluntaryExit = "0x04000000"
	}

	logrus.WithFields(logrus.Fields{
		"genesisTimestamp":       cfg.Chain.GenesisTimestamp,
		"genesisValidatorsRoot":  cfg.Chain.GenesisValidatorsRoot,
		"configName":             cfg.Chain.Config.ConfigName,
		"depositChainID":         cfg.Chain.Config.DepositChainID,
		"depositNetworkID":       cfg.Chain.Config.DepositNetworkID,
		"depositContractAddress": cfg.Chain.Config.DepositContractAddress,
	}).Infof("did init config")

	return nil
}

func readConfigFile(cfg *types.Config, path string) error {
	if path == "" {
		return yaml.Unmarshal([]byte(config.DefaultConfigYml), cfg)
	}

	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("error opening config file %v: %v", path, err)
	}

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(cfg)
	if err != nil {
		return fmt.Errorf("error decoding config file %v: %v", path, err)
	}

	return nil
}

func readConfigEnv(cfg *types.Config) error {
	return envconfig.Process("", cfg)
}

func readConfigSecrets(cfg *types.Config) error {
	return ProcessSecrets(cfg)
}

// MustParseHex will parse a string into hex
func MustParseHex(hexString string) []byte {
	data, err := hex.DecodeString(strings.Replace(hexString, "0x", "", -1))
	if err != nil {
		log.Fatal(err)
	}
	return data
}

func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Headers", "*, Authorization")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "*")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func IsApiRequest(r *http.Request) bool {
	query, ok := r.URL.Query()["format"]
	return ok && len(query) > 0 && query[0] == "json"
}

var eth1AddressRE = regexp.MustCompile("^(0x)?[0-9a-fA-F]{40}$")
var withdrawalCredentialsRE = regexp.MustCompile("^(0x)?00[0-9a-fA-F]{62}$")
var withdrawalCredentialsAddressRE = regexp.MustCompile("^(0x)?010000000000000000000000[0-9a-fA-F]{40}$")
var eth1TxRE = regexp.MustCompile("^(0x)?[0-9a-fA-F]{64}$")
var zeroHashRE = regexp.MustCompile("^(0x)?0+$")

// IsValidEth1Address verifies whether a string represents a valid eth1-address.
func IsValidEth1Address(s string) bool {
	return !zeroHashRE.MatchString(s) && eth1AddressRE.MatchString(s)
}

// IsEth1Address verifies whether a string represents an eth1-address. In contrast to IsValidEth1Address, this also returns true for the 0x0 address
func IsEth1Address(s string) bool {
	return eth1AddressRE.MatchString(s)
}

// IsValidEth1Tx verifies whether a string represents a valid eth1-tx-hash.
func IsValidEth1Tx(s string) bool {
	return !zeroHashRE.MatchString(s) && eth1TxRE.MatchString(s)
}

// IsValidWithdrawalCredentials verifies whether a string represents valid withdrawal credentials.
func IsValidWithdrawalCredentials(s string) bool {
	return withdrawalCredentialsRE.MatchString(s) || withdrawalCredentialsAddressRE.MatchString(s)
}

// https://github.com/badoux/checkmail/blob/f9f80cb795fa/checkmail.go#L37
var emailRE = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

// IsValidEmail verifies whether a string represents a valid email-address.
func IsValidEmail(s string) bool {
	return emailRE.MatchString(s)
}

// IsValidUrl verifies whether a string represents a valid Url.
func IsValidUrl(s string) bool {
	u, err := url.ParseRequestURI(s)
	if err != nil {
		return false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	if len(u.Host) == 0 {
		return false
	}
	return govalidator.IsURL(s)
}

// RoundDecimals rounds (nearest) a number to the specified number of digits after comma
func RoundDecimals(f float64, n int) float64 {
	d := math.Pow10(n)
	return math.Round(f*d) / d
}

// HashAndEncode digests the input with sha256 and returns it as hex string
func HashAndEncode(input string) string {
	codeHashedBytes := sha256.Sum256([]byte(input))
	return hex.EncodeToString(codeHashedBytes[:])
}

const charset = "abcdefghijklmnopqrstuvwxyz0123456789"

// RandomString returns a random hex-string
func RandomString(length int) string {
	b, _ := GenerateRandomBytesSecure(length)
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b)
}

func GenerateRandomBytesSecure(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := securerand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func SqlRowsToJSON(rows *sql.Rows) ([]interface{}, error) {
	columnTypes, err := rows.ColumnTypes()

	if err != nil {
		return nil, fmt.Errorf("error getting column types: %w", err)
	}

	count := len(columnTypes)
	finalRows := []interface{}{}

	for rows.Next() {

		scanArgs := make([]interface{}, count)

		for i, v := range columnTypes {
			switch v.DatabaseTypeName() {
			case "VARCHAR", "TEXT", "UUID":
				scanArgs[i] = new(sql.NullString)
			case "BOOL":
				scanArgs[i] = new(sql.NullBool)
			case "INT4", "INT8":
				scanArgs[i] = new(sql.NullInt64)
			case "FLOAT8":
				scanArgs[i] = new(sql.NullFloat64)
			case "TIMESTAMP":
				scanArgs[i] = new(sql.NullTime)
			case "_INT4", "_INT8":
				scanArgs[i] = new(pq.Int64Array)
			default:
				scanArgs[i] = new(sql.NullString)
			}
		}

		err := rows.Scan(scanArgs...)

		if err != nil {
			return nil, fmt.Errorf("error scanning rows: %w", err)
		}

		masterData := map[string]interface{}{}

		for i, v := range columnTypes {

			//log.Println(v.Name(), v.DatabaseTypeName())
			if z, ok := (scanArgs[i]).(*sql.NullBool); ok {
				if z.Valid {
					masterData[v.Name()] = z.Bool
				} else {
					masterData[v.Name()] = nil
				}
				continue
			}

			if z, ok := (scanArgs[i]).(*sql.NullString); ok {
				if z.Valid {
					if v.DatabaseTypeName() == "BYTEA" {
						if len(z.String) > 0 {
							masterData[v.Name()] = "0x" + hex.EncodeToString([]byte(z.String))
						} else {
							masterData[v.Name()] = nil
						}
					} else if v.DatabaseTypeName() == "NUMERIC" {
						nbr, _ := new(big.Int).SetString(z.String, 10)
						masterData[v.Name()] = nbr
					} else {
						masterData[v.Name()] = z.String
					}
				} else {
					masterData[v.Name()] = nil
				}
				continue
			}

			if z, ok := (scanArgs[i]).(*sql.NullInt64); ok {
				if z.Valid {
					masterData[v.Name()] = z.Int64
				} else {
					masterData[v.Name()] = nil
				}
				continue
			}

			if z, ok := (scanArgs[i]).(*sql.NullInt32); ok {
				if z.Valid {
					masterData[v.Name()] = z.Int32
				} else {
					masterData[v.Name()] = nil
				}
				continue
			}

			if z, ok := (scanArgs[i]).(*sql.NullFloat64); ok {
				if z.Valid {
					masterData[v.Name()] = z.Float64
				} else {
					masterData[v.Name()] = nil
				}
				continue
			}

			if z, ok := (scanArgs[i]).(*sql.NullTime); ok {
				if z.Valid {
					masterData[v.Name()] = z.Time.Unix()
				} else {
					masterData[v.Name()] = nil
				}
				continue
			}

			masterData[v.Name()] = scanArgs[i]
		}

		finalRows = append(finalRows, masterData)
	}

	return finalRows, nil
}

// GenerateAPIKey generates an API key for a user
func GenerateRandomAPIKey() (string, error) {
	const apiLength = 28
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

	max := big.NewInt(int64(len(letters)))
	key := make([]byte, apiLength)
	for i := 0; i < apiLength; i++ {
		num, err := securerand.Int(securerand.Reader, max)
		if err != nil {
			return "", err
		}
		key[i] = letters[num.Int64()]
	}

	apiKeyBase64 := base64.RawURLEncoding.EncodeToString(key)
	return apiKeyBase64, nil
}

func ExchangeRateForCurrency(currency string) float64 {
	return price.GetEthPrice(currency)
}

// Glob walks through a directory and returns files with a given extension
func Glob(dir string, ext string) ([]string, error) {
	files := []string{}
	err := filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if filepath.Ext(path) == ext {
			files = append(files, path)
		}
		return nil
	})

	return files, err
}

// ValidateReCAPTCHA validates a ReCaptcha server side
func ValidateReCAPTCHA(recaptchaResponse string) (bool, error) {
	// Check this URL verification details from Google
	// https://developers.google.com/recaptcha/docs/verify
	req, err := http.PostForm("https://www.google.com/recaptcha/api/siteverify", url.Values{
		"secret":   {Config.Frontend.RecaptchaSecretKey},
		"response": {recaptchaResponse},
	})
	if err != nil { // Handle error from HTTP POST to Google reCAPTCHA verify server
		return false, err
	}
	defer req.Body.Close()
	body, err := ioutil.ReadAll(req.Body) // Read the response from Google
	if err != nil {
		return false, err
	}

	var googleResponse types.GoogleRecaptchaResponse
	err = json.Unmarshal(body, &googleResponse) // Parse the JSON response from Google
	if err != nil {
		return false, err
	}
	if len(googleResponse.ErrorCodes) > 0 {
		err = fmt.Errorf("error validating ReCaptcha %v", googleResponse.ErrorCodes)
	} else {
		err = nil
	}

	if googleResponse.Score > 0.5 {
		return true, err
	}

	return false, fmt.Errorf("score too low threshold not reached, Score: %v - Required >0.5; %v", googleResponse.Score, err)
}

func BitAtVector(b []byte, i int) bool {
	bb := b[i/8]
	return (bb & (1 << uint(i%8))) > 0
}

func BitAtVectorReversed(b []byte, i int) bool {
	bb := b[i/8]
	return (bb & (1 << uint(7-(i%8)))) > 0
}

func GetNetwork() string {
	return strings.ToLower(Config.Chain.Config.ConfigName)
}

func ElementExists(arr []string, el string) bool {
	for _, e := range arr {
		if e == el {
			return true
		}
	}
	return false
}

func TryFetchContractMetadata(address []byte) (*types.ContractMetadata, error) {
	return getABIFromEtherscan(address)
}

// func getABIFromSourcify(address []byte) (*types.ContractMetadata, error) {
// 	httpClient := http.Client{
// 		Timeout: time.Second * 5,
// 	}

// 	resp, err := httpClient.Get(fmt.Sprintf("https://sourcify.dev/server/repository/contracts/full_match/%d/0x%x/metadata.json", 1, address))
// 	if err != nil {
// 		return nil, err
// 	}

// 	if resp.StatusCode == 200 {
// 		body, err := ioutil.ReadAll(resp.Body)
// 		if err != nil {
// 			return nil, err
// 		}

// 		data := &types.SourcifyContractMetadata{}
// 		err = json.Unmarshal(body, data)
// 		if err != nil {
// 			return nil, err
// 		}

// 		abiString, err := json.Marshal(data.Output.Abi)
// 		if err != nil {
// 			return nil, err
// 		}

// 		contractAbi, err := abi.JSON(bytes.NewReader(abiString))
// 		if err != nil {
// 			return nil, err
// 		}

// 		meta := &types.ContractMetadata{}
// 		meta.ABIJson = abiString
// 		meta.ABI = &contractAbi
// 		meta.Name = ""

// 		return meta, nil
// 	} else {
// 		return nil, fmt.Errorf("sourcify contract code not found")
// 	}
// }

func getABIFromEtherscan(address []byte) (*types.ContractMetadata, error) {
	baseUrl := ""
	switch dcid := Config.Chain.Config.DepositChainID; dcid {
	case 1: // mainnet
		baseUrl = "api.etherscan.io"
	case 5: // goerli
		baseUrl = "api-goerli.etherscan.io"
	case 11155111: // sepolia
		baseUrl = "api-sepolia.etherscan.io"
	default: // unsupported
		return nil, nil
	}

	httpClient := http.Client{Timeout: time.Second * 5}
	resp, err := httpClient.Get(fmt.Sprintf("https://%s/api?module=contract&action=getsourcecode&address=0x%x&apikey=%s", baseUrl, address, Config.EtherscanAPIKey))
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("StatusCode: '%d', Status: '%s'", resp.StatusCode, resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	headerData := &struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}{}
	err = json.Unmarshal(body, headerData)
	if err != nil {
		return nil, err
	}
	if headerData.Status == "0" {
		if headerData.Message == "NOTOK" {
			return nil, ErrRateLimit
		}
		return nil, fmt.Errorf("%s", headerData.Message)
	}

	data := &types.EtherscanContractMetadata{}
	err = json.Unmarshal(body, data)
	if err != nil {
		return nil, err
	}
	if data.Result[0].Abi == "Contract source code not verified" {
		return nil, nil
	}

	contractAbi, err := abi.JSON(strings.NewReader(data.Result[0].Abi))
	if err != nil {
		return nil, err
	}
	meta := &types.ContractMetadata{}
	meta.ABIJson = []byte(data.Result[0].Abi)
	meta.ABI = &contractAbi
	meta.Name = data.Result[0].ContractName
	return meta, nil
}

func FormatThousandsEnglish(number string) string {
	runes := []rune(number)
	cnt := 0
	for _, rune := range runes {
		if rune == '.' {
			break
		}
		cnt += 1
	}
	amt := cnt / 3
	rem := cnt % 3

	if rem == 0 {
		amt -= 1
	}

	res := make([]rune, 0, amt+rem)
	if amt <= 0 {
		return number
	}
	for i := 0; i < len(runes); i++ {
		if i != 0 && i == rem {
			res = append(res, ',')
			amt -= 1
		}

		if amt > 0 && i > rem && ((i-rem)%3) == 0 {
			res = append(res, ',')
			amt -= 1
		}

		res = append(res, runes[i])
	}

	return string(res)
}

// Generates a QR code for an address
// returns two transparent base64 encoded img strings for dark and light theme
// the first has a black QR code the second a white QR code
func GenerateQRCodeForAddress(address []byte) (string, string, error) {
	q, err := qrcode.New(FixAddressCasing(fmt.Sprintf("%x", address)), qrcode.Medium)
	if err != nil {
		return "", "", err
	}

	q.BackgroundColor = color.Transparent
	q.ForegroundColor = color.Black

	png, err := q.PNG(320)
	if err != nil {
		return "", "", err
	}

	q.ForegroundColor = color.White

	pngInverse, err := q.PNG(320)
	if err != nil {
		return "", "", err
	}

	return base64.StdEncoding.EncodeToString(png), base64.StdEncoding.EncodeToString(pngInverse), nil
}

// sliceContains reports whether the provided string is present in the given slice of strings.
func SliceContains(list []string, target string) bool {
	for _, s := range list {
		if s == target {
			return true
		}
	}
	return false
}

func FormatEthstoreComparison(pool string, val float64) template.HTML {
	prefix := ""
	textClass := "text-danger"
	ou := "underperforms"
	if val > 0 {
		prefix = "+"
		textClass = "text-success"
		ou = "outperforms"
	}

	return template.HTML(fmt.Sprintf(`<sub title="%s %s the ETH.STORE indicator by %s%.2f%%" data-toggle="tooltip" class="%s">(%s%.2f%%)</sub>`, pool, ou, prefix, val, textClass, prefix, val))
}

func FormatPoolPerformance(val float64) template.HTML {
	return template.HTML(fmt.Sprintf(`<span data-toggle="tooltip" title=%f%%>%s%%</span>`, val, fmt.Sprintf("%.2f", val)))
}

func FormatTokenSymbolTitle(symbol string) string {
	urls := xurls.Relaxed.FindAllString(symbol, -1)

	if len(urls) > 0 {
		return "The token symbol has been hidden as it contains a URL which might be a scam"
	}
	return ""
}

func FormatTokenSymbol(symbol string) string {
	urls := xurls.Relaxed.FindAllString(symbol, -1)

	if len(urls) > 0 {
		return "[hidden-symbol]"
	}
	return symbol
}

func FormatTokenSymbolHTML(tmpl template.HTML) template.HTML {
	tmplString := (string(tmpl))
	symbolTitle := FormatTokenSymbolTitle(tmplString)

	tmplString = FormatTokenSymbol(tmplString)
	tmpl = template.HTML(strings.ReplaceAll(tmplString, `title=""`, fmt.Sprintf(`title="%s"`, symbolTitle)))

	return tmpl
}

func ReverseSlice[S ~[]E, E any](s S) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

func AddBigInts(a, b []byte) []byte {
	return new(big.Int).Add(new(big.Int).SetBytes(a), new(big.Int).SetBytes(b)).Bytes()
}

// GetTimeToNextWithdrawal calculates the time it takes for the validators next withdrawal to be processed.
func GetTimeToNextWithdrawal(distance uint64) time.Time {
	minTimeToWithdrawal := time.Now().Add(time.Second * time.Duration((distance/Config.Chain.Config.MaxValidatorsPerWithdrawalSweep)*Config.Chain.Config.SecondsPerSlot))
	timeToWithdrawal := time.Now().Add(time.Second * time.Duration((float64(distance)/float64(Config.Chain.Config.MaxWithdrawalsPerPayload))*float64(Config.Chain.Config.SecondsPerSlot)))

	if timeToWithdrawal.Before(minTimeToWithdrawal) {
		return minTimeToWithdrawal
	}

	return timeToWithdrawal
}

func EpochsPerDay() uint64 {
	day := time.Hour * 24
	return (uint64(day.Seconds()) / Config.Chain.Config.SlotsPerEpoch) / Config.Chain.Config.SecondsPerSlot
}

// ForkVersionAtEpoch returns the forkversion active a specific epoch
func ForkVersionAtEpoch(epoch uint64) *types.ForkVersion {
	if epoch >= Config.Chain.Config.CappellaForkEpoch {
		return &types.ForkVersion{
			Epoch:           Config.Chain.Config.CappellaForkEpoch,
			CurrentVersion:  MustParseHex(Config.Chain.Config.CappellaForkVersion),
			PreviousVersion: MustParseHex(Config.Chain.Config.BellatrixForkVersion),
		}
	}
	if epoch >= Config.Chain.Config.BellatrixForkEpoch {
		return &types.ForkVersion{
			Epoch:           Config.Chain.Config.BellatrixForkEpoch,
			CurrentVersion:  MustParseHex(Config.Chain.Config.BellatrixForkVersion),
			PreviousVersion: MustParseHex(Config.Chain.Config.AltairForkVersion),
		}
	}
	if epoch >= Config.Chain.Config.AltairForkEpoch {
		return &types.ForkVersion{
			Epoch:           Config.Chain.Config.AltairForkEpoch,
			CurrentVersion:  MustParseHex(Config.Chain.Config.AltairForkVersion),
			PreviousVersion: MustParseHex(Config.Chain.Config.GenesisForkVersion),
		}
	}
	return &types.ForkVersion{
		Epoch:           0,
		CurrentVersion:  MustParseHex(Config.Chain.Config.GenesisForkVersion),
		PreviousVersion: MustParseHex(Config.Chain.Config.GenesisForkVersion),
	}
}

// LogFatal logs a fatal error with callstack info that skips callerSkip many levels with arbitrarily many additional infos.
// callerSkip equal to 0 gives you info directly where LogFatal is called.
func LogFatal(err error, errorMsg interface{}, callerSkip int, additionalInfos ...string) {
	logErrorInfo(err, callerSkip, additionalInfos...).Fatal(errorMsg)
}

// LogError logs an error with callstack info that skips callerSkip many levels with arbitrarily many additional infos.
// callerSkip equal to 0 gives you info directly where LogError is called.
func LogError(err error, errorMsg interface{}, callerSkip int, additionalInfos ...string) {
	logErrorInfo(err, callerSkip, additionalInfos...).Error(errorMsg)
}

func logErrorInfo(err error, callerSkip int, additionalInfos ...string) *logrus.Entry {
	logFields := logrus.NewEntry(logrus.New())

	pc, fullFilePath, line, ok := runtime.Caller(callerSkip + 2)
	if ok {
		logFields = logFields.WithFields(logrus.Fields{
			"cs_file":     filepath.Base(fullFilePath),
			"cs_function": runtime.FuncForPC(pc).Name(),
			"cs_line":     line,
		})
	} else {
		logFields = logFields.WithField("runtime", "Callstack cannot be read")
	}

	if err != nil {
		logFields = logFields.WithField("error type", fmt.Sprintf("%T", err)).WithError(err)
	}

	for idx, info := range additionalInfos {
		logFields = logFields.WithField(fmt.Sprintf("info_%v", idx), info)
	}

	return logFields
}

func GetSigningDomain() ([]byte, error) {
	beaconConfig := prysm_params.BeaconConfig()
	genForkVersion, err := hex.DecodeString(strings.Replace(Config.Chain.Config.GenesisForkVersion, "0x", "", -1))
	if err != nil {
		return nil, err
	}

	domain, err := signing.ComputeDomain(
		beaconConfig.DomainDeposit,
		genForkVersion,
		beaconConfig.ZeroHash[:],
	)

	if err != nil {
		return nil, err
	}

	return domain, err
}
