package utils

import (
	"encoding/hex"
	"eth2-exporter/types"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/kelseyhightower/envconfig"
)

// PageSize is the number of records used when fetching RPC data
const PageSize = 500

// Config is the globally accessible configuration
var Config *types.Config

// GetTemplateFuncs will get the template functions
func GetTemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"formatBlockStatus": FormatBlockStatus,
		"formatValidator":   FormatValidator,
		"formatBalance":     FormatBalance,
		"formatPercentage":  FormatPercentage,
		"formatIncome":      FormatIncome,
		"mod":               func(i, j int) bool { return i%j == 0 },
		"sub":               func(i, j int) int { return i - j },
		"add":               func(i, j int) int { return i + j },
	}
}

// FormatBlockStatus will return an html status for a block
func FormatBlockStatus(status uint64) template.HTML {
	if status == 0 {
		return "<span class=\"badge bg-light text-dark\">Scheduled</span>"
	} else if status == 1 {
		return "<span class=\"badge bg-success text-white\">Proposed</span>"
	} else if status == 2 {
		return "<span class=\"badge bg-warning text-dark\">Missed</span>"
	} else if status == 3 {
		return "<span class=\"badge bg-secondary text-white\">Orphaned</span>"
	} else {
		return "Unknown"
	}
}

// FormatAttestationStatus will return a user-friendly attestation for an attestation status number
func FormatAttestationStatus(status uint64) string {
	if status == 0 {
		return "<span class=\"badge bg-light text-dark\">Scheduled</span>"
	} else if status == 1 {
		return "<span class=\"badge bg-success text-white\">Attested</span>"
	} else if status == 2 {
		return "<span class=\"badge bg-warning text-dark\">Missed</span>"
	} else {
		return "Unknown"
	}
}

func FormatValidatorStatus(status string) string {
	if status == "pending" {
		return "<span class=\"badge validator-pending text-dark\">pending</span>"
	} else if status == "active:online" {
		return "<span class=\"badge validator-active text-dark\">active <span class=\"badge badge-light bg-success\">on</span></span>"
	} else if status == "active:offline" {
		return "<span class=\"badge validator-active text-dark\">active <span class=\"badge badge-light bg-danger\">off</span></span>"
	} else if status == "exiting:online" {
		return "<span class=\"badge validator-exiting text-dark\">exiting <span class=\"badge badge-light bg-success\">on</span></span>"
	} else if status == "exiting:offline" {
		return "<span class=\"badge validator-exiting text-dark\">exiting <span class=\"badge badge-light bg-danger\">off</span></span>"
	} else if status == "slashing:online" {
		return "<span class=\"badge validator-slashing text-dark\">slashing <span class=\"badge badge-light bg-success\">on</span></span>"
	} else if status == "slashing:offline" {
		return "<span class=\"badge validator-slashing text-dark\">slashing <span class=\"badge badge-light bg-danger\">off</span></span>"
	} else if status == "exited" {
		return "<span class=\"badge validator-exited text-dark\">exited</span>"
	}
	return "Unknown"
}

// FormatValidator will return html formatted text for a validator
func FormatValidator(validator uint64) template.HTML {
	return template.HTML(fmt.Sprintf("<i class=\"fas fa-male\"></i> <a href=\"/validator/%v\">%v</a>", validator, validator))
}

// SlotToTime will return a time.Time to slot
func SlotToTime(slot uint64) time.Time {
	return time.Unix(int64(Config.Chain.GenesisTimestamp+slot*Config.Chain.SecondsPerSlot), 0)
}

// TimeToSlot will return time to slot in seconds
func TimeToSlot(timestamp uint64) uint64 {
	return (timestamp - Config.Chain.GenesisTimestamp) / Config.Chain.SecondsPerSlot
}

// EpochToTime will return a time.Time for an epoch
func EpochToTime(epoch uint64) time.Time {
	return time.Unix(int64(Config.Chain.GenesisTimestamp+epoch*Config.Chain.SecondsPerSlot*Config.Chain.SlotsPerEpoch), 0)
}

// TimeToEpoch will return an epoch for a given time
func TimeToEpoch(ts time.Time) int64 {
	return (ts.Unix() - int64(Config.Chain.GenesisTimestamp)) / int64(Config.Chain.SecondsPerSlot) / int64(Config.Chain.SlotsPerEpoch)
}

// FormatBalance will return a string for a balance
func FormatBalance(balance uint64) string {
	return fmt.Sprintf("%.2f ETH", float64(balance)/float64(1000000000))
}

// FormatIncome will return a string for a balance
func FormatIncome(income int64) template.HTML {
	if income > 0 {
		return template.HTML(fmt.Sprintf(`<span class="text-success"><b>+%.4f ETH</b></span>`, float64(income)/float64(1000000000)))
	} else if income < 0 {
		return template.HTML(fmt.Sprintf(`<span class="text-danger"><b>%.4f ETH</b></span>`, float64(income)/float64(1000000000)))
	} else {
		return template.HTML(fmt.Sprintf(`<b>%.4f ETH</b>`, float64(income)/float64(1000000000)))
	}
}

// FormatPercentage will return a string for a percentage
func FormatPercentage(percentage float64) string {
	return fmt.Sprintf("%.0f", percentage*float64(100))
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

// FormatPublicKey will format a public key
func FormatPublicKey(publicKey []byte) string {
	return fmt.Sprintf("%x", publicKey)
}

// FormatAttestorAssignmentKey will format attestor assignment keys
func FormatAttestorAssignmentKey(AttesterSlot, CommitteeIndex, MemberIndex uint64) string {
	return fmt.Sprintf("%v-%v-%v", AttesterSlot, CommitteeIndex, MemberIndex)
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
