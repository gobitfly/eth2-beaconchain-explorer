package utils

import (
	"encoding/hex"
	"eth2-exporter/types"
	"fmt"
	"gopkg.in/yaml.v2"
	"html/template"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

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
	}
}

// FormatBlockStatus will return an html status for a block
func FormatBlockStatus(status uint64) template.HTML {
	if status == 0 {
		return "<span class=\"badge badge-light\">Scheduled</span>"
	} else if status == 1 {
		return "<span class=\"badge badge-success\">Proposed</span>"
	} else if status == 2 {
		return "<span class=\"badge badge-warning\">Missed</span>"
	} else if status == 3 {
		return "<span class=\"badge badge-secondary\">Orphaned</span>"
	} else {
		return "Unknown"
	}
}

// FormatAttestationStatus will return a user-friendly attestation for an attestation status number
func FormatAttestationStatus(status uint64) string {
	if status == 0 {
		return "Scheduled"
	} else if status == 1 {
		return "Attested"
	} else if status == 2 {
		return "Missed"
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

// FormatBalance will return a string for a balance
func FormatBalance(balance uint64) string {
	return fmt.Sprintf("%.2f ETH", float64(balance)/float64(1000000000))
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
		return fmt.Errorf("Error opening config file %v: %v", path, err)
	}

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(cfg)
	if err != nil {
		return fmt.Errorf("Error decoding config file %v: %v", path, err)
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
