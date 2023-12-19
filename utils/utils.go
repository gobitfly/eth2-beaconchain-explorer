package utils

import (
	"bufio"
	"bytes"
	"context"
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
	"io"
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
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"gopkg.in/yaml.v3"

	"github.com/asaskevich/govalidator"
	"github.com/carlmjohnson/requests"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/params"
	"github.com/kataras/i18n"
	"github.com/kelseyhightower/envconfig"
	"github.com/lib/pq"
	"github.com/mvdan/xurls"
	"github.com/prysmaticlabs/prysm/v3/beacon-chain/core/signing"
	prysm_params "github.com/prysmaticlabs/prysm/v3/config/params"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"github.com/skip2/go-qrcode"
	confusables "github.com/skygeario/go-confusable-homoglyphs"
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
		"formatNotificationChannel":               FormatNotificationChannel,
		"formatBalanceSql":                        FormatBalanceSql,
		"formatCurrentBalance":                    FormatCurrentBalance,
		"formatElCurrency":                        FormatElCurrency,
		"formatClCurrency":                        FormatClCurrency,
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
		"formatBitlist":                           FormatBitlist,
		"formatBitvectorValidators":               formatBitvectorValidators,
		"formatParticipation":                     FormatParticipation,
		"formatIncome":                            FormatIncome,
		"formatIncomeSql":                         FormatIncomeSql,
		"formatSqlInt64":                          FormatSqlInt64,
		"formatValidator":                         FormatValidator,
		"formatValidatorWithName":                 FormatValidatorWithName,
		"formatValidatorInt64":                    FormatValidatorInt64,
		"formatValidatorStatus":                   FormatValidatorStatus,
		"formatPercentage":                        FormatPercentage,
		"formatPercentageWithPrecision":           FormatPercentageWithPrecision,
		"formatPercentageWithGPrecision":          FormatPercentageWithGPrecision,
		"formatPercentageColoredEmoji":            FormatPercentageColoredEmoji,
		"formatPublicKey":                         FormatPublicKey,
		"formatSlashedValidator":                  FormatSlashedValidator,
		"formatSlashedValidatorInt64":             FormatSlashedValidatorInt64,
		"formatTimestamp":                         FormatTimestamp,
		"formatTsWithoutTooltip":                  FormatTsWithoutTooltip,
		"formatValidatorName":                     FormatValidatorName,
		"formatAttestationInclusionEffectiveness": FormatAttestationInclusionEffectiveness,
		"formatValidatorTags":                     FormatValidatorTags,
		"formatValidatorTag":                      FormatValidatorTag,
		"formatRPL":                               FormatRPL,
		"formatETH":                               FormatETH,
		"formatFloat":                             FormatFloat,
		"formatAmount":                            FormatAmount,
		"formatBytes":                             FormatBytes,
		"formatBlobVersionedHash":                 FormatBlobVersionedHash,
		"formatBigAmount":                         FormatBigAmount,
		"formatBytesAmount":                       FormatBytesAmount,
		"formatYesNo":                             FormatYesNo,
		"formatAmountFormatted":                   FormatAmountFormatted,
		"formatAddressAsLink":                     FormatAddressAsLink,
		"formatBuilder":                           FormatBuilder,
		"formatDifficulty":                        FormatDifficulty,
		"getCurrencyLabel":                        price.GetCurrencyLabel,
		"config":                                  func() *types.Config { return Config },
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
		"addBigInt":                               func(i, j *big.Int) *big.Int { return new(big.Int).Add(i, j) },
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
		"bigDecimalShift": func(num []byte, shift []byte) string {
			numDecimal := decimal.NewFromBigInt(new(big.Int).SetBytes(num), 0)
			denomDecimal := decimal.NewFromBigInt(new(big.Int).Exp(big.NewInt(10), new(big.Int).SetBytes(shift), nil), 0)
			res := numDecimal.DivRound(denomDecimal, 18)
			return res.String()
		},
		"trimTrailingZero": func(num string) string {
			if strings.Contains(num, ".") {
				return strings.TrimRight(strings.TrimRight(num, "0"), ".")
			}
			return num
		},
		// ETH1 related formatting
		"formatEth1TxStatus":    FormatEth1TxStatus,
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
	b, err := os.ReadFile(path)
	if err != nil {
		log.Printf("includeHTML - error reading file: %v", err)
		return ""
	}
	return template.HTML(string(b))
}

func GraffitiToString(graffiti []byte) string {
	s := strings.Map(fixUtf, string(bytes.Trim(graffiti, "\x00")))
	s = strings.Replace(s, "\u0000", "", -1) // remove 0x00 bytes as it is not supported in postgres

	if !utf8.ValidString(s) {
		return "INVALID_UTF8_STRING"
	}

	return s
}

// FormatGraffitiString formats (and escapes) the graffiti
func FormatGraffitiString(graffiti string) string {
	return strings.Map(fixUtf, template.HTMLEscapeString(graffiti))
}

func HasProblematicUtfCharacters(s string) bool {
	// Check for null character ('\x00')
	if utf8.ValidString(s) && utf8.Valid([]byte(s)) {
		// Check for control characters ('\x01' to '\x1F' and '\x7F')
		for _, r := range s {
			if r <= 0x1F || r == 0x7F {
				return true
			}
		}
	} else {
		return true // Invalid UTF-8 sequence
	}

	return false
}

func fixUtf(r rune) rune {
	if r == utf8.RuneError {
		return -1
	}
	return r
}

func SyncPeriodOfEpoch(epoch uint64) uint64 {
	if epoch < Config.Chain.ClConfig.AltairForkEpoch {
		return 0
	}
	return epoch / Config.Chain.ClConfig.EpochsPerSyncCommitteePeriod
}

// FirstEpochOfSyncPeriod returns the first epoch of a given sync period.
//
// Please note that it will return the calculated first epoch of the sync period even if it is pre ALTAIR.
//
// Furthermore, for the very first actual sync period, it may return an epoch pre ALTAIR even though that is inccorect.
//
// For more information: https://eth2book.info/capella/annotated-spec/#sync-committee-updates
func FirstEpochOfSyncPeriod(syncPeriod uint64) uint64 {
	return syncPeriod * Config.Chain.ClConfig.EpochsPerSyncCommitteePeriod
}

func TimeToSyncPeriod(t time.Time) uint64 {
	return SyncPeriodOfEpoch(uint64(TimeToEpoch(t)))
}

// EpochOfSlot returns the corresponding epoch of a slot
func EpochOfSlot(slot uint64) uint64 {
	return slot / Config.Chain.ClConfig.SlotsPerEpoch
}

// DayOfSlot returns the corresponding day of a slot
func DayOfSlot(slot uint64) uint64 {
	return Config.Chain.ClConfig.SecondsPerSlot * slot / (24 * 3600)
}

// WeekOfSlot returns the corresponding week of a slot
func WeekOfSlot(slot uint64) uint64 {
	return Config.Chain.ClConfig.SecondsPerSlot * slot / (7 * 24 * 3600)
}

// SlotToTime returns a time.Time to slot
func SlotToTime(slot uint64) time.Time {
	return time.Unix(int64(Config.Chain.GenesisTimestamp+slot*Config.Chain.ClConfig.SecondsPerSlot), 0)
}

// TimeToSlot returns time to slot in seconds
func TimeToSlot(timestamp uint64) uint64 {
	if Config.Chain.GenesisTimestamp > timestamp {
		return 0
	}
	return (timestamp - Config.Chain.GenesisTimestamp) / Config.Chain.ClConfig.SecondsPerSlot
}

func TimeToFirstSlotOfEpoch(timestamp uint64) uint64 {
	slot := TimeToSlot(timestamp)
	lastEpochOffset := slot % Config.Chain.ClConfig.SlotsPerEpoch
	slot = slot - lastEpochOffset
	return slot
}

// EpochToTime will return a time.Time for an epoch
func EpochToTime(epoch uint64) time.Time {
	return time.Unix(int64(Config.Chain.GenesisTimestamp+epoch*Config.Chain.ClConfig.SecondsPerSlot*Config.Chain.ClConfig.SlotsPerEpoch), 0)
}

// TimeToDay will return a days since genesis for an timestamp
func TimeToDay(timestamp uint64) uint64 {
	const hoursInADay = float64(Day / time.Hour)
	return uint64(time.Unix(int64(timestamp), 0).Sub(time.Unix(int64(Config.Chain.GenesisTimestamp), 0)).Hours() / hoursInADay)
}

func DayToTime(day int64) time.Time {
	return time.Unix(int64(Config.Chain.GenesisTimestamp), 0).Add(Day * time.Duration(day))
}

// TimeToEpoch will return an epoch for a given time
func TimeToEpoch(ts time.Time) int64 {
	if int64(Config.Chain.GenesisTimestamp) > ts.Unix() {
		return 0
	}
	return (ts.Unix() - int64(Config.Chain.GenesisTimestamp)) / int64(Config.Chain.ClConfig.SecondsPerSlot) / int64(Config.Chain.ClConfig.SlotsPerEpoch)
}

func WeiToEther(wei *big.Int) decimal.Decimal {
	return decimal.NewFromBigInt(wei, 0).DivRound(decimal.NewFromInt(params.Ether), 18)
}

func WeiBytesToEther(wei []byte) decimal.Decimal {
	return WeiToEther(new(big.Int).SetBytes(wei))
}

func GWeiToEther(gwei *big.Int) decimal.Decimal {
	return decimal.NewFromBigInt(gwei, 0).Div(decimal.NewFromInt(params.GWei))
}

func GWeiBytesToEther(gwei []byte) decimal.Decimal {
	return GWeiToEther(new(big.Int).SetBytes(gwei))
}

// WaitForCtrlC will block/wait until a control-c is pressed
func WaitForCtrlC() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}

// ReadConfig will process a configuration
func ReadConfig(cfg *types.Config, path string) error {

	configPathFromEnv := os.Getenv("BEACONCHAIN_CONFIG")

	if configPathFromEnv != "" { // allow the location of the config file to be passed via env args
		path = configPathFromEnv
	}
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

	if cfg.Frontend.SiteBrand == "" {
		cfg.Frontend.SiteBrand = "beaconcha.in"
	}

	if cfg.Chain.ClConfigPath == "" {
		// var prysmParamsConfig *prysmParams.BeaconChainConfig
		switch cfg.Chain.Name {
		case "mainnet":
			err = yaml.Unmarshal([]byte(config.MainnetChainYml), &cfg.Chain.ClConfig)
		case "prater":
			err = yaml.Unmarshal([]byte(config.PraterChainYml), &cfg.Chain.ClConfig)
		case "ropsten":
			err = yaml.Unmarshal([]byte(config.RopstenChainYml), &cfg.Chain.ClConfig)
		case "sepolia":
			err = yaml.Unmarshal([]byte(config.SepoliaChainYml), &cfg.Chain.ClConfig)
		case "gnosis":
			err = yaml.Unmarshal([]byte(config.GnosisChainYml), &cfg.Chain.ClConfig)
		case "holesky":
			err = yaml.Unmarshal([]byte(config.HoleskyChainYml), &cfg.Chain.ClConfig)
		default:
			return fmt.Errorf("tried to set known chain-config, but unknown chain-name: %v (path: %v)", cfg.Chain.Name, cfg.Chain.ClConfigPath)
		}
		if err != nil {
			return err
		}
		// err = prysmParams.SetActive(prysmParamsConfig)
		// if err != nil {
		// 	return fmt.Errorf("error setting chainConfig (%v) for prysmParams: %w", cfg.Chain.Name, err)
		// }
	} else if cfg.Chain.ClConfigPath == "node" {
		nodeEndpoint := fmt.Sprintf("http://%s:%s", cfg.Indexer.Node.Host, cfg.Indexer.Node.Port)

		jr := &types.ConfigJsonResponse{}

		err := requests.
			URL(nodeEndpoint + "/eth/v1/config/spec").
			ToJSON(jr).
			Fetch(context.Background())

		if err != nil {
			return err
		}

		chainCfg := types.ClChainConfig{
			PresetBase:                              jr.Data.PresetBase,
			ConfigName:                              jr.Data.ConfigName,
			TerminalTotalDifficulty:                 jr.Data.TerminalTotalDifficulty,
			TerminalBlockHash:                       jr.Data.TerminalBlockHash,
			TerminalBlockHashActivationEpoch:        mustParseUint(jr.Data.TerminalBlockHashActivationEpoch),
			MinGenesisActiveValidatorCount:          mustParseUint(jr.Data.MinGenesisActiveValidatorCount),
			MinGenesisTime:                          int64(mustParseUint(jr.Data.MinGenesisTime)),
			GenesisForkVersion:                      jr.Data.GenesisForkVersion,
			GenesisDelay:                            mustParseUint(jr.Data.GenesisDelay),
			AltairForkVersion:                       jr.Data.AltairForkVersion,
			AltairForkEpoch:                         mustParseUint(jr.Data.AltairForkEpoch),
			BellatrixForkVersion:                    jr.Data.BellatrixForkVersion,
			BellatrixForkEpoch:                      mustParseUint(jr.Data.BellatrixForkEpoch),
			CappellaForkVersion:                     jr.Data.CapellaForkVersion,
			CappellaForkEpoch:                       mustParseUint(jr.Data.CapellaForkEpoch),
			DenebForkVersion:                        jr.Data.DenebForkVersion,
			DenebForkEpoch:                          mustParseUint(jr.Data.DenebForkEpoch),
			SecondsPerSlot:                          mustParseUint(jr.Data.SecondsPerSlot),
			SecondsPerEth1Block:                     mustParseUint(jr.Data.SecondsPerEth1Block),
			MinValidatorWithdrawabilityDelay:        mustParseUint(jr.Data.MinValidatorWithdrawabilityDelay),
			ShardCommitteePeriod:                    mustParseUint(jr.Data.ShardCommitteePeriod),
			Eth1FollowDistance:                      mustParseUint(jr.Data.Eth1FollowDistance),
			InactivityScoreBias:                     mustParseUint(jr.Data.InactivityScoreBias),
			InactivityScoreRecoveryRate:             mustParseUint(jr.Data.InactivityScoreRecoveryRate),
			EjectionBalance:                         mustParseUint(jr.Data.EjectionBalance),
			MinPerEpochChurnLimit:                   mustParseUint(jr.Data.MinPerEpochChurnLimit),
			ChurnLimitQuotient:                      mustParseUint(jr.Data.ChurnLimitQuotient),
			ProposerScoreBoost:                      mustParseUint(jr.Data.ProposerScoreBoost),
			DepositChainID:                          mustParseUint(jr.Data.DepositChainID),
			DepositNetworkID:                        mustParseUint(jr.Data.DepositNetworkID),
			DepositContractAddress:                  jr.Data.DepositContractAddress,
			MaxCommitteesPerSlot:                    mustParseUint(jr.Data.MaxCommitteesPerSlot),
			TargetCommitteeSize:                     mustParseUint(jr.Data.TargetCommitteeSize),
			MaxValidatorsPerCommittee:               mustParseUint(jr.Data.TargetCommitteeSize),
			ShuffleRoundCount:                       mustParseUint(jr.Data.ShuffleRoundCount),
			HysteresisQuotient:                      mustParseUint(jr.Data.HysteresisQuotient),
			HysteresisDownwardMultiplier:            mustParseUint(jr.Data.HysteresisDownwardMultiplier),
			HysteresisUpwardMultiplier:              mustParseUint(jr.Data.HysteresisUpwardMultiplier),
			SafeSlotsToUpdateJustified:              mustParseUint(jr.Data.SafeSlotsToUpdateJustified),
			MinDepositAmount:                        mustParseUint(jr.Data.MinDepositAmount),
			MaxEffectiveBalance:                     mustParseUint(jr.Data.MaxEffectiveBalance),
			EffectiveBalanceIncrement:               mustParseUint(jr.Data.EffectiveBalanceIncrement),
			MinAttestationInclusionDelay:            mustParseUint(jr.Data.MinAttestationInclusionDelay),
			SlotsPerEpoch:                           mustParseUint(jr.Data.SlotsPerEpoch),
			MinSeedLookahead:                        mustParseUint(jr.Data.MinSeedLookahead),
			MaxSeedLookahead:                        mustParseUint(jr.Data.MaxSeedLookahead),
			EpochsPerEth1VotingPeriod:               mustParseUint(jr.Data.EpochsPerEth1VotingPeriod),
			SlotsPerHistoricalRoot:                  mustParseUint(jr.Data.SlotsPerHistoricalRoot),
			MinEpochsToInactivityPenalty:            mustParseUint(jr.Data.MinEpochsToInactivityPenalty),
			EpochsPerHistoricalVector:               mustParseUint(jr.Data.EpochsPerHistoricalVector),
			EpochsPerSlashingsVector:                mustParseUint(jr.Data.EpochsPerSlashingsVector),
			HistoricalRootsLimit:                    mustParseUint(jr.Data.HistoricalRootsLimit),
			ValidatorRegistryLimit:                  mustParseUint(jr.Data.ValidatorRegistryLimit),
			BaseRewardFactor:                        mustParseUint(jr.Data.BaseRewardFactor),
			WhistleblowerRewardQuotient:             mustParseUint(jr.Data.WhistleblowerRewardQuotient),
			ProposerRewardQuotient:                  mustParseUint(jr.Data.ProposerRewardQuotient),
			InactivityPenaltyQuotient:               mustParseUint(jr.Data.InactivityPenaltyQuotient),
			MinSlashingPenaltyQuotient:              mustParseUint(jr.Data.MinSlashingPenaltyQuotient),
			ProportionalSlashingMultiplier:          mustParseUint(jr.Data.ProportionalSlashingMultiplier),
			MaxProposerSlashings:                    mustParseUint(jr.Data.MaxProposerSlashings),
			MaxAttesterSlashings:                    mustParseUint(jr.Data.MaxAttesterSlashings),
			MaxAttestations:                         mustParseUint(jr.Data.MaxAttestations),
			MaxDeposits:                             mustParseUint(jr.Data.MaxDeposits),
			MaxVoluntaryExits:                       mustParseUint(jr.Data.MaxVoluntaryExits),
			InvactivityPenaltyQuotientAltair:        mustParseUint(jr.Data.InactivityPenaltyQuotientAltair),
			MinSlashingPenaltyQuotientAltair:        mustParseUint(jr.Data.MinSlashingPenaltyQuotientAltair),
			ProportionalSlashingMultiplierAltair:    mustParseUint(jr.Data.ProportionalSlashingMultiplierAltair),
			SyncCommitteeSize:                       mustParseUint(jr.Data.SyncCommitteeSize),
			EpochsPerSyncCommitteePeriod:            mustParseUint(jr.Data.EpochsPerSyncCommitteePeriod),
			MinSyncCommitteeParticipants:            mustParseUint(jr.Data.MinSyncCommitteeParticipants),
			InvactivityPenaltyQuotientBellatrix:     mustParseUint(jr.Data.InactivityPenaltyQuotientBellatrix),
			MinSlashingPenaltyQuotientBellatrix:     mustParseUint(jr.Data.MinSlashingPenaltyQuotientBellatrix),
			ProportionalSlashingMultiplierBellatrix: mustParseUint(jr.Data.ProportionalSlashingMultiplierBellatrix),
			MaxBytesPerTransaction:                  mustParseUint(jr.Data.MaxBytesPerTransaction),
			MaxTransactionsPerPayload:               mustParseUint(jr.Data.MaxTransactionsPerPayload),
			BytesPerLogsBloom:                       mustParseUint(jr.Data.BytesPerLogsBloom),
			MaxExtraDataBytes:                       mustParseUint(jr.Data.MaxExtraDataBytes),
			MaxWithdrawalsPerPayload:                mustParseUint(jr.Data.MaxWithdrawalsPerPayload),
			MaxValidatorsPerWithdrawalSweep:         mustParseUint(jr.Data.MaxValidatorsPerWithdrawalsSweep),
			MaxBlsToExecutionChange:                 mustParseUint(jr.Data.MaxBlsToExecutionChanges),
		}

		if jr.Data.AltairForkEpoch == "" {
			chainCfg.AltairForkEpoch = 18446744073709551615
		}
		if jr.Data.BellatrixForkEpoch == "" {
			chainCfg.BellatrixForkEpoch = 18446744073709551615
		}
		if jr.Data.CapellaForkEpoch == "" {
			chainCfg.CappellaForkEpoch = 18446744073709551615
		}
		if jr.Data.DenebForkEpoch == "" {
			chainCfg.DenebForkEpoch = 18446744073709551615
		}

		cfg.Chain.ClConfig = chainCfg

		type GenesisResponse struct {
			Data struct {
				GenesisTime           string `json:"genesis_time"`
				GenesisValidatorsRoot string `json:"genesis_validators_root"`
				GenesisForkVersion    string `json:"genesis_fork_version"`
			} `json:"data"`
		}

		gtr := &GenesisResponse{}

		err = requests.
			URL(nodeEndpoint + "/eth/v1/beacon/genesis").
			ToJSON(gtr).
			Fetch(context.Background())

		if err != nil {
			return err
		}

		cfg.Chain.GenesisTimestamp = mustParseUint(gtr.Data.GenesisTime)
		cfg.Chain.GenesisValidatorsRoot = gtr.Data.GenesisValidatorsRoot

		logger.Infof("loaded chain config from node with genesis time %s", gtr.Data.GenesisTime)

	} else {
		f, err := os.Open(cfg.Chain.ClConfigPath)
		if err != nil {
			return fmt.Errorf("error opening Chain Config file %v: %w", cfg.Chain.ClConfigPath, err)
		}
		var chainConfig *types.ClChainConfig
		decoder := yaml.NewDecoder(f)
		err = decoder.Decode(&chainConfig)
		if err != nil {
			return fmt.Errorf("error decoding Chain Config file %v: %v", cfg.Chain.ClConfigPath, err)
		}
		cfg.Chain.ClConfig = *chainConfig
	}

	type MinimalELConfig struct {
		ByzantiumBlock      *big.Int `yaml:"BYZANTIUM_FORK_BLOCK,omitempty"`      // Byzantium switch block (nil = no fork, 0 = already on byzantium)
		ConstantinopleBlock *big.Int `yaml:"CONSTANTINOPLE_FORK_BLOCK,omitempty"` // Constantinople switch block (nil = no fork, 0 = already activated)
	}
	if cfg.Chain.ElConfigPath == "" {
		minimalCfg := MinimalELConfig{}
		switch cfg.Chain.Name {
		case "mainnet":
			err = yaml.Unmarshal([]byte(config.MainnetChainYml), &minimalCfg)
		case "prater":
			err = yaml.Unmarshal([]byte(config.PraterChainYml), &minimalCfg)
		case "ropsten":
			err = yaml.Unmarshal([]byte(config.RopstenChainYml), &minimalCfg)
		case "sepolia":
			err = yaml.Unmarshal([]byte(config.SepoliaChainYml), &minimalCfg)
		case "gnosis":
			err = yaml.Unmarshal([]byte(config.GnosisChainYml), &minimalCfg)
		case "holesky":
			err = yaml.Unmarshal([]byte(config.HoleskyChainYml), &minimalCfg)
		default:
			return fmt.Errorf("tried to set known chain-config, but unknown chain-name: %v (path: %v)", cfg.Chain.Name, cfg.Chain.ElConfigPath)
		}
		if err != nil {
			return err
		}
		if minimalCfg.ByzantiumBlock == nil {
			minimalCfg.ByzantiumBlock = big.NewInt(0)
		}
		if minimalCfg.ConstantinopleBlock == nil {
			minimalCfg.ConstantinopleBlock = big.NewInt(0)
		}
		cfg.Chain.ElConfig = &params.ChainConfig{
			ChainID:             big.NewInt(int64(cfg.Chain.Id)),
			ByzantiumBlock:      minimalCfg.ByzantiumBlock,
			ConstantinopleBlock: minimalCfg.ConstantinopleBlock,
		}
	} else {
		f, err := os.Open(cfg.Chain.ElConfigPath)
		if err != nil {
			return fmt.Errorf("error opening EL Chain Config file %v: %w", cfg.Chain.ElConfigPath, err)
		}
		var chainConfig *params.ChainConfig
		decoder := json.NewDecoder(f)
		err = decoder.Decode(&chainConfig)
		if err != nil {
			return fmt.Errorf("error decoding EL Chain Config file %v: %v", cfg.Chain.ElConfigPath, err)
		}
		cfg.Chain.ElConfig = chainConfig
	}

	cfg.Chain.Name = cfg.Chain.ClConfig.ConfigName

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
		case "holesky":
			cfg.Chain.GenesisTimestamp = 1695902400
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
		case "holesky":
			cfg.Chain.GenesisValidatorsRoot = "0x9143aa7c615a7f7115e2b6aac319c03529df8242ae705fba9df39b79c59fa8b1"
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

	if cfg.Frontend.ClCurrency == "" {
		switch cfg.Chain.Name {
		case "gnosis":
			cfg.Frontend.MainCurrency = "GNO"
			cfg.Frontend.ClCurrency = "mGNO"
			cfg.Frontend.ClCurrencyDecimals = 18
			cfg.Frontend.ClCurrencyDivisor = 1e9
		default:
			cfg.Frontend.MainCurrency = "ETH"
			cfg.Frontend.ClCurrency = "ETH"
			cfg.Frontend.ClCurrencyDecimals = 18
			cfg.Frontend.ClCurrencyDivisor = 1e9
		}
	}

	if cfg.Frontend.ElCurrency == "" {
		switch cfg.Chain.Name {
		case "gnosis":
			cfg.Frontend.ElCurrency = "xDAI"
			cfg.Frontend.ElCurrencyDecimals = 18
			cfg.Frontend.ElCurrencyDivisor = 1e18
		default:
			cfg.Frontend.ElCurrency = "ETH"
			cfg.Frontend.ElCurrencyDecimals = 18
			cfg.Frontend.ElCurrencyDivisor = 1e18
		}
	}

	if cfg.Frontend.SiteTitle == "" {
		cfg.Frontend.SiteTitle = "Open Source Ethereum Explorer"
	}

	if cfg.Frontend.Keywords == "" {
		cfg.Frontend.Keywords = "open source ethereum block explorer, ethereum block explorer, beacon chain explorer, ethereum blockchain explorer"
	}

	if cfg.Frontend.Ratelimits.FreeDay == 0 {
		cfg.Frontend.Ratelimits.FreeDay = 30000
	}
	if cfg.Frontend.Ratelimits.FreeMonth == 0 {
		cfg.Frontend.Ratelimits.FreeMonth = 30000
	}
	if cfg.Frontend.Ratelimits.SapphierDay == 0 {
		cfg.Frontend.Ratelimits.SapphierDay = 100000
	}
	if cfg.Frontend.Ratelimits.SapphierMonth == 0 {
		cfg.Frontend.Ratelimits.SapphierMonth = 500000
	}
	if cfg.Frontend.Ratelimits.EmeraldDay == 0 {
		cfg.Frontend.Ratelimits.EmeraldDay = 200000
	}
	if cfg.Frontend.Ratelimits.EmeraldMonth == 0 {
		cfg.Frontend.Ratelimits.EmeraldMonth = 1000000
	}
	if cfg.Frontend.Ratelimits.DiamondDay == 0 {
		cfg.Frontend.Ratelimits.DiamondDay = 6000000
	}
	if cfg.Frontend.Ratelimits.DiamondMonth == 0 {
		cfg.Frontend.Ratelimits.DiamondMonth = 6000000
	}

	if cfg.Chain.Id != 0 {
		switch cfg.Chain.Name {
		case "mainnet", "ethereum":
			cfg.Chain.Id = 1
		case "prater", "goerli":
			cfg.Chain.Id = 5
		case "holesky":
			cfg.Chain.Id = 17000
		case "sepolia":
			cfg.Chain.Id = 11155111
		case "gnosis":
			cfg.Chain.Id = 100
		}
	}

	// we check for maching chain id just for safety
	if cfg.Chain.Id != 0 && cfg.Chain.Id != cfg.Chain.ClConfig.DepositChainID {
		logrus.Fatalf("cfg.Chain.Id != cfg.Chain.ClConfig.DepositChainID: %v != %v", cfg.Chain.Id, cfg.Chain.ClConfig.DepositChainID)
	}

	cfg.Chain.Id = cfg.Chain.ClConfig.DepositChainID

	logrus.WithFields(logrus.Fields{
		"genesisTimestamp":       cfg.Chain.GenesisTimestamp,
		"genesisValidatorsRoot":  cfg.Chain.GenesisValidatorsRoot,
		"configName":             cfg.Chain.ClConfig.ConfigName,
		"depositChainID":         cfg.Chain.ClConfig.DepositChainID,
		"depositNetworkID":       cfg.Chain.ClConfig.DepositNetworkID,
		"depositContractAddress": cfg.Chain.ClConfig.DepositContractAddress,
		"clCurrency":             cfg.Frontend.ClCurrency,
		"elCurrency":             cfg.Frontend.ElCurrency,
		"mainCurrency":           cfg.Frontend.MainCurrency,
	}).Infof("did init config")

	return nil
}

func mustParseUint(str string) uint64 {

	if str == "" {
		return 0
	}

	nbr, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		logrus.Fatalf("fatal error parsing uint %s: %v", str, err)
	}

	return nbr
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
var hashRE = regexp.MustCompile("^(0x)?[0-9a-fA-F]{96}$")

// IsValidEth1Address verifies whether a string represents a valid eth1-address.
func IsValidEth1Address(s string) bool {
	return !zeroHashRE.MatchString(s) && eth1AddressRE.MatchString(s)
}

// IsEth1Address verifies whether a string represents an eth1-address.
// In contrast to IsValidEth1Address, this also returns true for the 0x0 address
func IsEth1Address(s string) bool {
	return eth1AddressRE.MatchString(s)
}

// IsValidEth1Tx verifies whether a string represents a valid eth1-tx-hash.
func IsValidEth1Tx(s string) bool {
	return !zeroHashRE.MatchString(s) && eth1TxRE.MatchString(s)
}

// IsEth1Tx verifies whether a string represents an eth1-tx-hash.
// In contrast to IsValidEth1Tx, this also returns true for the 0x0 address
func IsEth1Tx(s string) bool {
	return eth1TxRE.MatchString(s)
}

// IsHash verifies whether a string represents an eth1-hash.
func IsHash(s string) bool {
	return hashRE.MatchString(s)
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
	body, err := io.ReadAll(req.Body) // Read the response from Google
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
	return strings.ToLower(Config.Chain.ClConfig.ConfigName)
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
// 		body, err := io.ReadAll(resp.Body)
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

func GetEtherscanAPIBaseUrl(provideDefault bool) string {
	const mainnetBaseUrl = "api.etherscan.io"
	const goerliBaseUrl = "api-goerli.etherscan.io"
	const sepoliaBaseUrl = "api-sepolia.etherscan.io"

	// check config first
	if len(Config.EtherscanAPIBaseURL) > 0 {
		return Config.EtherscanAPIBaseURL
	}

	// check chain id
	switch Config.Chain.ClConfig.DepositChainID {
	case 1: // mainnet
		return mainnetBaseUrl
	case 5: // goerli
		return goerliBaseUrl
	case 11155111: // sepolia
		return sepoliaBaseUrl
	}

	// use default
	if provideDefault {
		return mainnetBaseUrl
	}
	return ""
}

func getABIFromEtherscan(address []byte) (*types.ContractMetadata, error) {
	baseUrl := GetEtherscanAPIBaseUrl(false)
	if len(baseUrl) < 1 {
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

	body, err := io.ReadAll(resp.Body)
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

	return template.HTML(fmt.Sprintf(`<sub title="%s %s the ETH.STORE® indicator by %s%.2f%%" data-toggle="tooltip" class="%s">(%s%.2f%%)</sub>`, pool, ou, prefix, val, textClass, prefix, val))
}

func FormatPoolPerformance(val float64) template.HTML {
	return template.HTML(fmt.Sprintf(`<span data-toggle="tooltip" title=%f%%>%s%%</span>`, val, fmt.Sprintf("%.2f", val)))
}

func FormatTokenSymbolTitle(symbol string) string {
	if isMaliciousToken(symbol) {
		return fmt.Sprintf("The token symbol (%s) has been hidden because it contains a URL or a confusable character", symbol)
	}
	return ""
}

func FormatTokenSymbol(symbol string) string {
	if isMaliciousToken(symbol) {
		return "[hidden-symbol] ⚠️"
	}
	return symbol
}

func isMaliciousToken(symbol string) bool {
	containsUrls := len(xurls.Relaxed.FindAllString(symbol, -1)) > 0
	isConfusable := len(confusables.IsConfusable(symbol, false, []string{"LATIN", "COMMON"})) > 0
	isMixedScript := confusables.IsMixedScript(symbol, nil)
	return containsUrls || isConfusable || isMixedScript || strings.ToUpper(symbol) == "ETH"
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
	minTimeToWithdrawal := time.Now().Add(time.Second * time.Duration((distance/Config.Chain.ClConfig.MaxValidatorsPerWithdrawalSweep)*Config.Chain.ClConfig.SecondsPerSlot))
	timeToWithdrawal := time.Now().Add(time.Second * time.Duration((float64(distance)/float64(Config.Chain.ClConfig.MaxWithdrawalsPerPayload))*float64(Config.Chain.ClConfig.SecondsPerSlot)))

	if timeToWithdrawal.Before(minTimeToWithdrawal) {
		return minTimeToWithdrawal
	}

	return timeToWithdrawal
}

func EpochsPerDay() uint64 {
	return (uint64(Day.Seconds()) / Config.Chain.ClConfig.SlotsPerEpoch) / Config.Chain.ClConfig.SecondsPerSlot
}

func GetFirstAndLastEpochForDay(day uint64) (firstEpoch uint64, lastEpoch uint64) {
	firstEpoch = day * EpochsPerDay()
	lastEpoch = firstEpoch + EpochsPerDay() - 1
	return firstEpoch, lastEpoch
}

func GetLastBalanceInfoSlotForDay(day uint64) uint64 {
	return ((day+1)*EpochsPerDay() - 1) * Config.Chain.ClConfig.SlotsPerEpoch
}

// ForkVersionAtEpoch returns the forkversion active a specific epoch
func ForkVersionAtEpoch(epoch uint64) *types.ForkVersion {
	if epoch >= Config.Chain.ClConfig.CappellaForkEpoch {
		return &types.ForkVersion{
			Epoch:           Config.Chain.ClConfig.CappellaForkEpoch,
			CurrentVersion:  MustParseHex(Config.Chain.ClConfig.CappellaForkVersion),
			PreviousVersion: MustParseHex(Config.Chain.ClConfig.BellatrixForkVersion),
		}
	}
	if epoch >= Config.Chain.ClConfig.BellatrixForkEpoch {
		return &types.ForkVersion{
			Epoch:           Config.Chain.ClConfig.BellatrixForkEpoch,
			CurrentVersion:  MustParseHex(Config.Chain.ClConfig.BellatrixForkVersion),
			PreviousVersion: MustParseHex(Config.Chain.ClConfig.AltairForkVersion),
		}
	}
	if epoch >= Config.Chain.ClConfig.AltairForkEpoch {
		return &types.ForkVersion{
			Epoch:           Config.Chain.ClConfig.AltairForkEpoch,
			CurrentVersion:  MustParseHex(Config.Chain.ClConfig.AltairForkVersion),
			PreviousVersion: MustParseHex(Config.Chain.ClConfig.GenesisForkVersion),
		}
	}
	return &types.ForkVersion{
		Epoch:           0,
		CurrentVersion:  MustParseHex(Config.Chain.ClConfig.GenesisForkVersion),
		PreviousVersion: MustParseHex(Config.Chain.ClConfig.GenesisForkVersion),
	}
}

// LogFatal logs a fatal error with callstack info that skips callerSkip many levels with arbitrarily many additional infos.
// callerSkip equal to 0 gives you info directly where LogFatal is called.
func LogFatal(err error, errorMsg interface{}, callerSkip int, additionalInfos ...map[string]interface{}) {
	logErrorInfo(err, callerSkip, additionalInfos...).Fatal(errorMsg)
}

// LogError logs an error with callstack info that skips callerSkip many levels with arbitrarily many additional infos.
// callerSkip equal to 0 gives you info directly where LogError is called.
func LogError(err error, errorMsg interface{}, callerSkip int, additionalInfos ...map[string]interface{}) {
	logErrorInfo(err, callerSkip, additionalInfos...).Error(errorMsg)
}

func logErrorInfo(err error, callerSkip int, additionalInfos ...map[string]interface{}) *logrus.Entry {
	logFields := logrus.NewEntry(logrus.New())

	pc, fullFilePath, line, ok := runtime.Caller(callerSkip + 2)
	if ok {
		logFields = logFields.WithFields(logrus.Fields{
			"_file":     filepath.Base(fullFilePath),
			"_function": runtime.FuncForPC(pc).Name(),
			"_line":     line,
		})
	} else {
		logFields = logFields.WithField("runtime", "Callstack cannot be read")
	}

	errColl := []string{}
	for {
		errColl = append(errColl, fmt.Sprint(err))
		nextErr := errors.Unwrap(err)
		if nextErr != nil {
			err = nextErr
		} else {
			break
		}
	}

	errMarkSign := "~"
	for idx := 0; idx < (len(errColl) - 1); idx++ {
		errInfoText := fmt.Sprintf("%serrInfo_%v%s", errMarkSign, idx, errMarkSign)
		nextErrInfoText := fmt.Sprintf("%serrInfo_%v%s", errMarkSign, idx+1, errMarkSign)
		if idx == (len(errColl) - 2) {
			nextErrInfoText = fmt.Sprintf("%serror%s", errMarkSign, errMarkSign)
		}

		// Replace the last occurrence of the next error in the current error
		lastIdx := strings.LastIndex(errColl[idx], errColl[idx+1])
		if lastIdx != -1 {
			errColl[idx] = errColl[idx][:lastIdx] + nextErrInfoText + errColl[idx][lastIdx+len(errColl[idx+1]):]
		}

		errInfoText = strings.ReplaceAll(errInfoText, errMarkSign, "")
		logFields = logFields.WithField(errInfoText, errColl[idx])
	}

	if err != nil {
		logFields = logFields.WithField("errType", fmt.Sprintf("%T", err)).WithError(err)
	}

	for _, infoMap := range additionalInfos {
		for name, info := range infoMap {
			logFields = logFields.WithField(name, info)
		}
	}

	return logFields
}

func GetSigningDomain() ([]byte, error) {
	beaconConfig := prysm_params.BeaconConfig()
	genForkVersion, err := hex.DecodeString(strings.Replace(Config.Chain.ClConfig.GenesisForkVersion, "0x", "", -1))
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

// SlotsPerSyncCommittee returns the count of slots per sync committee period
// (might be wrong for the first sync period at atlair which might be shorter, see https://eth2book.info/capella/annotated-spec/#sync-committee-updates)
func SlotsPerSyncCommittee() uint64 {
	return Config.Chain.ClConfig.EpochsPerSyncCommitteePeriod * Config.Chain.ClConfig.SlotsPerEpoch
}

// GetRemainingScheduledSyncDuties returns the remaining count of scheduled slots given the stats of the current period, while also accounting for exported slots.
//
// Parameters:
//   - validatorCount: the count of validators associated with the stats.
//   - stats: the current sync committee stats of the validators
//   - lastExportedEpoch: the last epoch that was exported into the validator_stats table
//   - firstEpochOfPeriod: the first epoch of the current sync committee period
func GetRemainingScheduledSyncDuties(validatorCount int, stats types.SyncCommitteesStats, lastExportedEpoch, firstEpochOfPeriod uint64) uint64 {
	// check how many sync duties remain in the current sync committee based on firstEpochOfPeriod
	slotsPerSyncCommittee := SlotsPerSyncCommittee()
	if firstEpochOfPeriod <= Config.Chain.ClConfig.AltairForkEpoch {
		if firstEpochOfPeriod+SlotsPerSyncCommittee() < Config.Chain.ClConfig.AltairForkEpoch {
			// not a valid sync committee as altair comes after the complete sync committee period
			return 0
		}

		// the first sync period at altair might be shorter, see https://eth2book.info/capella/annotated-spec/#sync-committee-updates
		firstEpochOfNextSyncPeriod := FirstEpochOfSyncPeriod(SyncPeriodOfEpoch(Config.Chain.ClConfig.AltairForkEpoch) + 1)
		slotsPerSyncCommittee = (firstEpochOfNextSyncPeriod - Config.Chain.ClConfig.AltairForkEpoch) * Config.Chain.ClConfig.SlotsPerEpoch
	}
	dutiesPerSyncCommittee := slotsPerSyncCommittee * uint64(validatorCount)

	// check how many duties are already exported
	exportedEpochs := uint64(0)
	if lastExportedEpoch >= firstEpochOfPeriod {
		exportedEpochs = lastExportedEpoch - firstEpochOfPeriod + 1
	}
	exportedDuties := exportedEpochs * Config.Chain.ClConfig.SlotsPerEpoch * uint64(validatorCount)

	// calculate how many duties are remaining i.e. are scheduled
	totalStats := stats.MissedSlots + stats.ParticipatedSlots + stats.ScheduledSlots
	return (dutiesPerSyncCommittee - ((exportedDuties + totalStats) % dutiesPerSyncCommittee)) % dutiesPerSyncCommittee
}

// AddSyncStats adds the sync stats of a set of validators from a given syncDutiesHistory to the given stats, if stats is nil a new stats object is created.
// Parameters:
//   - validators: the validators to add the stats for
//   - syncDutiesHistory: the sync duties history of all queried validators
//   - stats: the stats object to add the stats to, if nil a new stats object is created
func AddSyncStats(validators []uint64, syncDutiesHistory map[uint64]map[uint64]*types.ValidatorSyncParticipation, stats *types.SyncCommitteesStats) types.SyncCommitteesStats {
	if stats == nil {
		stats = &types.SyncCommitteesStats{}
	}
	for _, validator := range validators {
		v := syncDutiesHistory[validator]
		for _, r := range v {
			slotTime := SlotToTime(r.Slot)
			if r.Status == 0 && time.Since(slotTime) <= time.Minute {
				r.Status = 2
			}
			switch r.Status {
			case 0:
				stats.MissedSlots++
			case 1:
				stats.ParticipatedSlots++
			case 2:
				stats.ScheduledSlots++
			}
		}
	}
	return *stats
}

// To remove all round brackets (including its content) from a string
func RemoveRoundBracketsIncludingContent(input string) string {
	openCount := 0
	result := ""
	for {
		if len(input) == 0 {
			break
		}
		openIndex := strings.Index(input, "(")
		closeIndex := strings.Index(input, ")")
		if openIndex == -1 && closeIndex == -1 {
			if openCount == 0 {
				result += input
			}
			break
		} else if openIndex != -1 && (openIndex < closeIndex || closeIndex == -1) {
			openCount++
			if openCount == 1 {
				result += input[:openIndex]
			}
			input = input[openIndex+1:]
		} else {
			if openCount > 0 {
				openCount--
			} else if openIndex == -1 && len(result) == 0 {
				result += input[:closeIndex]
			}
			input = input[closeIndex+1:]
		}
	}
	return result
}

func Int64Min(x, y int64) int64 {
	if x < y {
		return x
	}
	return y
}

func Int64Max(x, y int64) int64 {
	if x > y {
		return x
	}
	return y
}

// Prompt asks for a string value using the label. For comand line interactions.
func CmdPrompt(label string) string {
	var s string
	r := bufio.NewReader(os.Stdin)
	for {
		fmt.Fprint(os.Stderr, label+" ")
		s, _ = r.ReadString('\n')
		if s != "" {
			break
		}
	}
	return strings.TrimSpace(s)
}

// UniqueStrings returns an array of strings containing each value of s only once
func UniqueStrings(s []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, str := range s {
		if _, ok := seen[str]; !ok {
			seen[str] = true
			result = append(result, str)
		}
	}
	return result
}

func SortedUniqueUint64(arr []uint64) []uint64 {
	if len(arr) <= 1 {
		return arr
	}

	sort.Slice(arr, func(i, j int) bool {
		return arr[i] < arr[j]
	})

	result := make([]uint64, 1, len(arr))
	result[0] = arr[0]
	for i := 1; i < len(arr); i++ {
		if arr[i-1] != arr[i] {
			result = append(result, arr[i])
		}
	}

	return result
}

type HttpReqHttpError struct {
	StatusCode int
	Url        string
	Body       []byte
}

func (err *HttpReqHttpError) Error() string {
	return fmt.Sprintf("error response: url: %s, status: %d, body: %s", err.Url, err.StatusCode, err.Body)
}

func HttpReq(ctx context.Context, method, url string, params, result interface{}) error {
	var err error
	var req *http.Request
	if params != nil {
		paramsJSON, err := json.Marshal(params)
		if err != nil {
			return fmt.Errorf("error marshaling params for request: %w, url: %v", err, url)
		}
		req, err = http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(paramsJSON))
		if err != nil {
			return fmt.Errorf("error creating request with params: %w, url: %v", err, url)
		}
	} else {
		req, err = http.NewRequestWithContext(ctx, method, url, nil)
		if err != nil {
			return fmt.Errorf("error creating request: %w, url: %v", err, url)
		}
	}
	req.Header.Set("Content-Type", "application/json")
	httpClient := &http.Client{Timeout: time.Minute}
	res, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return &HttpReqHttpError{
			StatusCode: res.StatusCode,
			Url:        url,
			Body:       body,
		}
	}
	if result != nil {
		err = json.NewDecoder(res.Body).Decode(result)
		if err != nil {
			return fmt.Errorf("error unmarshaling response: %w, url: %v", err, url)
		}
	}
	return nil
}

func ReverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func GetCurrentFuncName() string {
	pc, _, _, _ := runtime.Caller(1)
	return runtime.FuncForPC(pc).Name()
}

func GetParentFuncName() string {
	pc, _, _, _ := runtime.Caller(2)
	return runtime.FuncForPC(pc).Name()
}

// Returns true if the given block number is 0 and if it is (according to its timestamp) included in slot 0
//
// This is only true for networks that launch with active PoS at block 0 which requires
//
//   - Belatrix happening at epoch 0 (pre condition for merged networks)
//   - Genesis for PoS to happen at the same timestamp as the first block
func IsPoSBlock0(number uint64, ts int64) bool {
	if number > 0 {
		return false
	}

	if Config.Chain.ClConfig.BellatrixForkEpoch > 0 {
		return false
	}

	return time.Unix(int64(Config.Chain.GenesisTimestamp-Config.Chain.ClConfig.GenesisDelay), 0).UTC().Equal(time.Unix(ts, 0))
}

func GetMaxAllowedDayRangeValidatorStats(validatorAmount int) int {
	if validatorAmount > 100000 {
		return 0 // exact day only
	} else if validatorAmount > 10000 {
		return 3
	} else if validatorAmount > 1000 {
		return 10
	} else {
		return math.MaxInt
	}
}
