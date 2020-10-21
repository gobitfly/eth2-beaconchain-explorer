package utils

import (
	"database/sql"
	"encoding/hex"
	"eth2-exporter/types"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"math"
	"math/big"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"gopkg.in/yaml.v2"

	"github.com/kelseyhightower/envconfig"
)

// Config is the globally accessible configuration
var Config *types.Config

// GetTemplateFuncs will get the template functions
func GetTemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"includeHTML":                             IncludeHTML,
		"formatHTML":                              FormatMessageToHtml,
		"formatBalance":                           FormatBalance,
		"formatCurrentBalance":                    FormatCurrentBalance,
		"formatEffectiveBalance":                  FormatEffectiveBalance,
		"formatBlockStatus":                       FormatBlockStatus,
		"formatBlockSlot":                         FormatBlockSlot,
		"formatSlotToTimestamp":                   FormatSlotToTimestamp,
		"formatDepositAmount":                     FormatDepositAmount,
		"formatEpoch":                             FormatEpoch,
		"formatEth1Block":                         FormatEth1Block,
		"formatEth1Address":                       FormatEth1Address,
		"formatEth1TxHash":                        FormatEth1TxHash,
		"formatGraffiti":                          FormatGraffiti,
		"formatHash":                              FormatHash,
		"formatIncome":                            FormatIncome,
		"formatValidator":                         FormatValidator,
		"formatValidatorWithName":                 FormatValidatorWithName,
		"formatValidatorInt64":                    FormatValidatorInt64,
		"formatValidatorStatus":                   FormatValidatorStatus,
		"formatPercentage":                        FormatPercentage,
		"formatPublicKey":                         FormatPublicKey,
		"formatSlashedValidator":                  FormatSlashedValidator,
		"formatSlashedValidatorInt64":             FormatSlashedValidatorInt64,
		"formatTimestamp":                         FormatTimestamp,
		"formatTimestampTs":                       FormatTimestampTs,
		"formatValidatorName":                     FormatValidatorName,
		"formatAttestationInclusionEffectiveness": FormatAttestationInclusionEffectiveness,
		"epochOfSlot":                             EpochOfSlot,
		"contains":                                strings.Contains,
		"mod":                                     func(i, j int) bool { return i%j == 0 },
		"sub":                                     func(i, j int) int { return i - j },
		"add":                                     func(i, j int) int { return i + j },
		"div":                                     func(i, j float64) float64 { return i / j },
		"gtf":                                     func(i, j float64) bool { return i > j },
		"round":                                   func(i float64, n int) float64 { return math.Round(i*math.Pow10(n)) / math.Pow10(n) },
		"percent":                                 func(i float64) float64 { return i * 100 },
		"formatThousands": func(i float64) string {
			p := message.NewPrinter(language.English)
			return p.Sprintf("%.0f\n", i)
		},
	}
}

var LayoutPaths []string = []string{"templates/layout/layout.html", "templates/layout/nav.html"}

// IncludeHTML adds html to the page
func IncludeHTML(path string) template.HTML {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		log.Println("includeHTML - error reading file: %v", err)
		return ""
	}
	return template.HTML(string(b))
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

// EpochOfSlot will return the corresponding epoch of a slot
func EpochOfSlot(slot uint64) uint64 {
	return slot / Config.Chain.SlotsPerEpoch
}

// SlotToTime will return a time.Time to slot
func SlotToTime(slot uint64) time.Time {
	return time.Unix(int64(Config.Chain.GenesisTimestamp+slot*Config.Chain.SecondsPerSlot), 0)
}

// TimeToSlot will return time to slot in seconds
func TimeToSlot(timestamp uint64) uint64 {
	if Config.Chain.GenesisTimestamp > timestamp {
		return 0
	}
	return (timestamp - Config.Chain.GenesisTimestamp) / Config.Chain.SecondsPerSlot
}

// EpochToTime will return a time.Time for an epoch
func EpochToTime(epoch uint64) time.Time {
	return time.Unix(int64(Config.Chain.GenesisTimestamp+epoch*Config.Chain.SecondsPerSlot*Config.Chain.SlotsPerEpoch), 0)
}

// TimeToEpoch will return an epoch for a given time
func TimeToEpoch(ts time.Time) int64 {
	if int64(Config.Chain.GenesisTimestamp) > ts.Unix() {
		return 0
	}
	return (ts.Unix() - int64(Config.Chain.GenesisTimestamp)) / int64(Config.Chain.SecondsPerSlot) / int64(Config.Chain.SlotsPerEpoch)
}

// WaitForCtrlC will block/wait until a control-c is pressed
func WaitForCtrlC() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}

// ReadConfig will process a configuration
func ReadConfig(cfg *types.Config, path string) error {
	err := readConfigFile(cfg, path)

	if err != nil {
		return err
	}

	return readConfigEnv(cfg)
}

func readConfigFile(cfg *types.Config, path string) error {
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

// MustParseHex will parse a string into hex
func MustParseHex(hexString string) []byte {
	data, err := hex.DecodeString(strings.Replace(hexString, "0x", "", -1))
	if err != nil {
		log.Fatal(err)
	}
	return data
}

func IsApiRequest(r *http.Request) bool {
	query, ok := r.URL.Query()["format"]
	return ok && len(query) > 0 && query[0] == "json"
}

var eth1AddressRE = regexp.MustCompile("^0?x?[0-9a-fA-F]{40}$")
var zeroHashRE = regexp.MustCompile("^0?x?0+$")

// IsValidEth1Address verifies whether a string represents a valid eth1-address.
func IsValidEth1Address(s string) bool {
	return !zeroHashRE.MatchString(s) && eth1AddressRE.MatchString(s)
}

// https://github.com/badoux/checkmail/blob/f9f80cb795fa/checkmail.go#L37
var emailRE = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

// IsValidEmail verifies wheter a string represents a valid email-address.
func IsValidEmail(s string) bool {
	return emailRE.MatchString(s)
}

// RoundDecimals rounds (nearest) a number to the specified number of digits after comma
func RoundDecimals(f float64, n int) float64 {
	d := math.Pow10(n)
	return math.Round(f*d) / d
}

const charset = "abcdefghijklmnopqrstuvwxyz0123456789"

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

// RandomString returns a random hex-string
func RandomString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func SqlRowsToJSON(rows *sql.Rows) ([]interface{}, error) {
	columnTypes, err := rows.ColumnTypes()

	if err != nil {
		return nil, err
	}

	count := len(columnTypes)
	finalRows := []interface{}{}

	for rows.Next() {

		scanArgs := make([]interface{}, count)

		for i, v := range columnTypes {
			//log.Printf("name: %v, type: %v", v.Name(), v.DatabaseTypeName())
			switch v.DatabaseTypeName() {
			case "VARCHAR", "TEXT", "UUID":
				scanArgs[i] = new(sql.NullString)
				break
			case "BOOL":
				scanArgs[i] = new(sql.NullBool)
				break
			case "INT4", "INT8":
				scanArgs[i] = new(sql.NullInt64)
				break
			case "FLOAT8":
				scanArgs[i] = new(sql.NullFloat64)
				break
			case "TIMESTAMP":
				scanArgs[i] = new(sql.NullTime)
				break
			default:
				scanArgs[i] = new(sql.NullString)
			}
		}

		err := rows.Scan(scanArgs...)

		if err != nil {
			return nil, err
		}

		masterData := map[string]interface{}{}

		for i, v := range columnTypes {

			//log.Println(v.Name(), v.DatabaseTypeName())
			if z, ok := (scanArgs[i]).(*sql.NullBool); ok {
				masterData[v.Name()] = z.Bool
				continue
			}

			if z, ok := (scanArgs[i]).(*sql.NullString); ok {
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
				continue
			}

			if z, ok := (scanArgs[i]).(*sql.NullInt64); ok {
				masterData[v.Name()] = z.Int64
				continue
			}

			if z, ok := (scanArgs[i]).(*sql.NullFloat64); ok {
				masterData[v.Name()] = z.Float64
				continue
			}

			if z, ok := (scanArgs[i]).(*sql.NullFloat64); ok {
				masterData[v.Name()] = z.Float64
				continue
			}

			if z, ok := (scanArgs[i]).(*sql.NullInt32); ok {
				masterData[v.Name()] = z.Int32
				continue
			}

			if z, ok := (scanArgs[i]).(*sql.NullTime); ok {
				masterData[v.Name()] = z.Time.Unix()
				continue
			}

			masterData[v.Name()] = scanArgs[i]
		}

		finalRows = append(finalRows, masterData)
	}

	return finalRows, nil
}
