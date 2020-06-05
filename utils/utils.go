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
		"formatBalance":               FormatBalance,
		"formatCurrentBalance":        FormatCurrentBalance,
		"formatEffectiveBalance":      FormatEffectiveBalance,
		"formatBlockStatus":           FormatBlockStatus,
		"formatBlockSlot":             FormatBlockSlot,
		"formatSlotToTimestamp":       FormatSlotToTimestamp,
		"formatDepositAmount":         FormatDepositAmount,
		"formatEpoch":                 FormatEpoch,
		"formatEth1Block":             FormatEth1Block,
		"formatEth1Address":           FormatEth1Address,
		"formatEth1TxHash":            FormatEth1TxHash,
		"formatGraffiti":              FormatGraffiti,
		"formatHash":                  FormatHash,
		"formatIncome":                FormatIncome,
		"formatValidator":             FormatValidator,
		"formatValidatorInt64":        FormatValidatorInt64,
		"formatValidatorStatus":       FormatValidatorStatus,
		"formatPercentage":            FormatPercentage,
		"formatPublicKey":             FormatPublicKey,
		"formatSlashedValidator":      FormatSlashedValidator,
		"formatSlashedValidatorInt64": FormatSlashedValidatorInt64,
		"formatTimestamp":             FormatTimestamp,
		"epochOfSlot":                 EpochOfSlot,
		"mod":                         func(i, j int) bool { return i%j == 0 },
		"sub":                         func(i, j int) int { return i - j },
		"add":                         func(i, j int) int { return i + j },
	}
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
