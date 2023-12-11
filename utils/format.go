package utils

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"eth2-exporter/price"
	"eth2-exporter/types"
	"fmt"
	"html"
	"html/template"
	"math"
	"math/big"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/prysmaticlabs/go-bitfield"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	itypes "github.com/gobitfly/eth-rewards/types"
)

const CalculatingHint = `Calculating…`

func FormatMessageToHtml(message string) template.HTML {
	message = fmt.Sprint(strings.Replace(message, "Error: ", "", 1))
	return template.HTML(message)
}

// FormatSyncParticipationStatus will return a user-friendly format for an sync-participation-status number
func FormatSyncParticipationStatus(status, blockSlot uint64) template.HTML {
	if status == 0 {
		return `<span class="badge badge-pill bg-danger text-white" style="font-size: 12px; font-weight: 500;">Missed</span>`
	} else if status == 1 {
		return `<span class="badge badge-pill bg-success text-white" style="font-size: 12px; font-weight: 500;">Participated</span>`
	} else if status == 2 {
		return `<span class="badge badge-pill bg-light text-dark" style="font-size: 12px; font-weight: 500;">Scheduled</span>`
	} else if status == 3 {
		return template.HTML(fmt.Sprintf(`<span class="badge badge-pill bg-warning text-white" style="font-size: 12px; font-weight: 500;" data-toggle="tooltip" data-html="true" data-placement="top" title='Slot %v was missed, it does not contain a block'>Missed</span>`, FormatAddCommas(blockSlot)))
	} else {
		return "Unknown"
	}
}

// FormatSyncParticipationStatus will return a user-friendly format for an sync-participation-status number
func FormatSyncParticipations(participants uint64) template.HTML {
	return template.HTML(fmt.Sprintf(`<span>%v/%v</span>`, participants, Config.Chain.ClConfig.SyncCommitteeSize))
}

// FormatAttestationStatus will return a user-friendly attestation for an attestation status number
func FormatAttestationStatus(status uint64) template.HTML {
	if status == 0 {
		return `<span class="badge badge-pill bg-light text-dark" style="font-size: 12px; font-weight: 500;">Scheduled</span>`
	} else if status == 1 {
		return `<span class="badge badge-pill bg-success text-white" style="font-size: 12px; font-weight: 500;">Attested</span>`
	} else if status == 2 {
		return `<span class="badge badge-pill bg-warning text-white" style="font-size: 12px; font-weight: 500;">Missed</span>`
	} else if status == 3 {
		return `<span class="badge badge-pill bg-warning text-white" style="font-size: 12px; font-weight: 500;">Missed (Orphaned)</span>`
	} else {
		return "Unknown"
	}
}

// FormatAttestationStatusShort will return a user-friendly attestation for an attestation status number
func FormatAttestationStatusShort(status uint64) template.HTML {
	if status == 0 {
		return `<span title="Scheduled Attestation" data-toggle="tooltip" class="mx-1 badge badge-pill bg-light text-dark" style="font-size: 12px; font-weight: 500;">Sche.</span>`
	} else if status == 1 {
		return `<span title="Attested" data-toggle="tooltip" class="mx-1 badge badge-pill bg-success text-white" style="font-size: 12px; font-weight: 500;">Att.</span>`
	} else if status == 2 {
		return `<span title="Missed Attestation" data-toggle="tooltip" class="mx-1 badge badge-pill bg-warning text-white" style="font-size: 12px; font-weight: 500;">Miss.</span>`
	} else if status == 3 {
		return `<span title="Missed Attestation (Orphaned)" data-toggle="tooltip" class="mx-1 badge badge-pill bg-warning text-white" style="font-size: 12px; font-weight: 500;">Orph.</span>`
	} else if status == 4 {
		return `<span title="Inactivity Leak" data-toggle="tooltip" class="mx-1 badge badge-pill bg-danger text-white" style="font-size: 12px; font-weight: 500;">Leak</span>`
	} else if status == 5 {
		return `<span title="Inactive" data-toggle="tooltip" class="mx-1 badge badge-pill bg-light text-dark" style="font-size: 12px; font-weight: 500;">Inac.</span>`
	} else {
		return "Unknown"
	}
}

// FormatAttestorAssignmentKey will format attestor assignment keys
func FormatAttestorAssignmentKey(AttesterSlot, CommitteeIndex, MemberIndex uint64) string {
	return fmt.Sprintf("%v-%v-%v", AttesterSlot, CommitteeIndex, MemberIndex)
}

// FormatBalance will return a string for a balance
func FormatBalance(balanceInt uint64, currency string) template.HTML {
	exchangeRate := price.GetPrice(Config.Frontend.ClCurrency, currency)
	balance := FormatFloat((float64(balanceInt)/float64(Config.Frontend.ClCurrencyDivisor))*float64(exchangeRate), 2)

	return template.HTML(balance + " " + currency)
}

// FormatBalance will return a string for a balance
func FormatEligibleBalance(balanceInt uint64, currency string) template.HTML {
	if balanceInt == 0 {
		return `<span class="text-small text-muted">` + CalculatingHint + `</span>`
	}
	exchangeRate := price.GetPrice(Config.Frontend.ClCurrency, currency)
	balance := FormatFloat((float64(balanceInt)/float64(Config.Frontend.ClCurrencyDivisor))*float64(exchangeRate), 2)

	return template.HTML(balance)
}

func FormatBalanceSql(balanceInt sql.NullInt64, currency string) template.HTML {
	if !balanceInt.Valid {
		return template.HTML("0 " + currency)
	}
	exchangeRate := price.GetPrice(Config.Frontend.ClCurrency, currency)
	balance := FormatFloat((float64(balanceInt.Int64)/float64(Config.Frontend.ClCurrencyDivisor))*float64(exchangeRate), 5)

	return template.HTML(balance + " " + currency)
}

func FormatBalanceGwei(balance *int64, currency string) template.HTML {
	if currency == Config.Frontend.ClCurrency {
		if balance == nil {
			return template.HTML("<span> 0.00000 " + currency + "</span>")
		} else if *balance == 0 {
			return template.HTML("0")
		}

		balanceF := float64(*balance)
		if balanceF < 0 {
			return template.HTML(fmt.Sprintf("<span class=\"text-danger\">%.0f GWei</span>", balanceF))
		}
		return template.HTML(fmt.Sprintf("<span class=\"text-success\">+%.0f GWei</span>", balanceF))
	}
	return FormatBalanceChange(balance, currency)
}

func ClToMainCurrency(valIf interface{}) decimal.Decimal {
	val := IfToDec(valIf)
	res := val.DivRound(decimal.NewFromInt(Config.Frontend.ClCurrencyDivisor), 18)
	if Config.Frontend.ClCurrency == Config.Frontend.MainCurrency {
		return res
	}
	return res.Mul(decimal.NewFromFloat(price.GetPrice(Config.Frontend.ClCurrency, Config.Frontend.MainCurrency)))
}

func ElToMainCurrency(valIf interface{}) decimal.Decimal {
	val := IfToDec(valIf)
	res := val.DivRound(decimal.NewFromInt(Config.Frontend.ElCurrencyDivisor), 18)
	if Config.Frontend.ElCurrency == Config.Frontend.MainCurrency {
		return res
	}
	return res.Mul(decimal.NewFromFloat(price.GetPrice(Config.Frontend.ElCurrency, Config.Frontend.MainCurrency)))
}

func ClToCurrency(valIf interface{}, currency string) decimal.Decimal {
	val := IfToDec(valIf)
	res := val.DivRound(decimal.NewFromInt(Config.Frontend.ClCurrencyDivisor), 18)
	if currency == Config.Frontend.ClCurrency {
		return res
	}
	return res.Mul(decimal.NewFromFloat(price.GetPrice(Config.Frontend.ClCurrency, currency)))
}

func ElToCurrency(valIf interface{}, currency string) decimal.Decimal {
	val := IfToDec(valIf)
	res := val.DivRound(decimal.NewFromInt(Config.Frontend.ElCurrencyDivisor), 18)
	if currency == Config.Frontend.ElCurrency {
		return res
	}
	return res.Mul(decimal.NewFromFloat(price.GetPrice(Config.Frontend.ElCurrency, currency)))
}

func ClToCurrencyGwei(valIf interface{}, currency string) decimal.Decimal {
	val := IfToDec(valIf)
	if currency == Config.Frontend.ClCurrency {
		return val
	}
	return val.Mul(decimal.NewFromFloat(price.GetPrice(Config.Frontend.ClCurrency, currency)))
}

func FormatElCurrency(value interface{}, targetCurrency string, digitsAfterComma int, showCurrencySymbol, showPlusSign, colored, truncateAndAddTooltip bool) template.HTML {
	return formatCurrency(ElToCurrency(value, Config.Frontend.ElCurrency), Config.Frontend.ElCurrency, targetCurrency, digitsAfterComma, showCurrencySymbol, showPlusSign, colored, truncateAndAddTooltip)
}

func FormatClCurrency(value interface{}, targetCurrency string, digitsAfterComma int, showCurrencySymbol, showPlusSign, colored, truncateAndAddTooltip bool) template.HTML {
	return formatCurrency(ClToCurrency(value, Config.Frontend.ClCurrency), Config.Frontend.ClCurrency, targetCurrency, digitsAfterComma, showCurrencySymbol, showPlusSign, colored, truncateAndAddTooltip)
}

func FormatElCurrencyString(value interface{}, targetCurrency string, digitsAfterComma int, showCurrencySymbol, showPlusSign, truncateAndAddTooltip bool) string {
	return formatCurrencyString(ElToCurrency(value, Config.Frontend.ElCurrency), Config.Frontend.ElCurrency, targetCurrency, digitsAfterComma, showCurrencySymbol, showPlusSign, truncateAndAddTooltip)
}

func FormatClCurrencyString(value interface{}, targetCurrency string, digitsAfterComma int, showCurrencySymbol, showPlusSign, truncateAndAddTooltip bool) string {
	return formatCurrencyString(ClToCurrency(value, Config.Frontend.ClCurrency), Config.Frontend.ClCurrency, targetCurrency, digitsAfterComma, showCurrencySymbol, showPlusSign, truncateAndAddTooltip)
}

func formatCurrencyString(valIf interface{}, valueCurrency, targetCurrency string, digitsAfterComma int, showCurrencySymbol, showPlusSign, truncateAndAddTooltip bool) string {
	val := IfToDec(valIf)

	valPriced := val
	if valueCurrency != targetCurrency {
		valPriced = val.Mul(decimal.NewFromFloat(price.GetPrice(valueCurrency, targetCurrency)))
	}

	currencyStr := ""
	if showCurrencySymbol {
		currencyStr = " " + price.GetCurrencySymbol(targetCurrency)
	}

	amountStr := ""
	tooltipStartStr := ""
	tooltipEndStr := ""
	if truncateAndAddTooltip {
		amountStr = valPriced.Truncate(int32(digitsAfterComma)).String()

		// only add tooltip if the value is actually truncated
		valStr := valPriced.String()
		if valStr != amountStr {
			tooltipStartStr = fmt.Sprintf(`<span data-toggle="tooltip" data-placement="top" title="%s%s">`, valPriced, currencyStr)
			tooltipEndStr = `</span>`
		}

		// add trailing zeros to always have the same amount of digits after the comma
		dotIndex := strings.Index(valStr, ".")
		if dotIndex >= 0 {
			missingZeros := digitsAfterComma - (len(amountStr) - dotIndex - 1)
			if missingZeros > 0 {
				amountStr += strings.Repeat("0", missingZeros)
			}
		}
	} else {
		amountStr = valPriced.StringFixed(int32(digitsAfterComma))
	}

	plusSignStr := ""
	if showPlusSign && valPriced.Cmp(decimal.NewFromInt(0)) >= 0 {
		plusSignStr = "+"
	}

	return fmt.Sprintf(`%s%s%s%s%s`, tooltipStartStr, plusSignStr, amountStr, currencyStr, tooltipEndStr)
}

func formatCurrency(valIf interface{}, valueCurrency, targetCurrency string, digitsAfterComma int, showCurrencySymbol, showPlusSign, colored, truncateAndAddTooltip bool) template.HTML {
	result := formatCurrencyString(valIf, valueCurrency, targetCurrency, digitsAfterComma, showCurrencySymbol, showPlusSign, truncateAndAddTooltip)
	classes := ""

	if colored {
		val := IfToDec(valIf)
		if val.Cmp(decimal.NewFromInt(0)) >= 0 {
			classes = ` class="text-success"`
		} else {
			classes = ` class="text-danger"`
		}
	}

	return template.HTML(fmt.Sprintf(`<span%s>%s</span>`, classes, result))
}

// IfToDec trys to parse given parameter to decimal.Decimal, it only logs on error
func IfToDec(valIf interface{}) decimal.Decimal {
	var err error
	var val decimal.Decimal
	switch v := valIf.(type) {
	case *float64:
		val = decimal.NewFromFloat(*v)
	case *int64:
		val = decimal.NewFromInt(*v)
	case *uint64:
		val, err = decimal.NewFromString(fmt.Sprintf("%v", *v))
	case int, int64, float64, uint64, *big.Float:
		val, err = decimal.NewFromString(fmt.Sprintf("%v", valIf))
	case []uint8:
		val = decimal.NewFromBigInt(new(big.Int).SetBytes(v), 0)
	case *big.Int:
		val = decimal.NewFromBigInt(v, 0)
	case decimal.Decimal:
		val = v
	default:
		logger.WithFields(logrus.Fields{"type": reflect.TypeOf(valIf), "val": valIf}).Errorf("invalid value passed to IfToDec")
	}
	if err != nil {
		logger.WithFields(logrus.Fields{"type": reflect.TypeOf(valIf), "val": valIf, "error": err}).Errorf("invalid value passed to IfToDec")
	}
	return val
}

func FormatBalanceChangeFormatted(balance *int64, currencyName string, details *itypes.ValidatorEpochIncome) template.HTML {
	currencySymbol := "GWei"
	currencyFunc := ClToCurrencyGwei
	if currencyName != Config.Frontend.MainCurrency {
		currencySymbol = currencyName
		currencyFunc = ClToCurrency
	}

	if balance == nil || *balance == 0 {
		return template.HTML(fmt.Sprintf("<span class=\"float-right\">0 %s</span>", currencySymbol))
	}

	maxDigits := uint(6)

	income := ""
	if details != nil {
		income += fmt.Sprintf("Att. Source: %s %s<br/>", FormatAddCommasFormatted(currencyFunc(IfToDec(details.AttestationSourceReward).Sub(IfToDec(details.AttestationSourcePenalty)), currencyName).InexactFloat64(), maxDigits), currencySymbol)
		income += fmt.Sprintf("Att. Target: %s %s<br/>", FormatAddCommasFormatted(currencyFunc(IfToDec(details.AttestationTargetReward).Sub(IfToDec(details.AttestationTargetPenalty)), currencyName).InexactFloat64(), maxDigits), currencySymbol)
		income += fmt.Sprintf("Att. Head Vote: %s %s<br/>", FormatAddCommasFormatted(currencyFunc(details.AttestationHeadReward, currencyName).InexactFloat64(), maxDigits), currencySymbol)

		if details.FinalityDelayPenalty > 0 {
			income += fmt.Sprintf("Finality Delay Penalty: %s %s<br/>", FormatAddCommasFormatted(currencyFunc(details.FinalityDelayPenalty, currencyName).InexactFloat64()*-1, maxDigits), currencySymbol)
		}

		if details.ProposerSlashingInclusionReward > 0 {
			income += fmt.Sprintf("Proposer Slashing Inc. Reward: %s %s<br/>", FormatAddCommasFormatted(currencyFunc(details.ProposerSlashingInclusionReward, currencyName).InexactFloat64(), maxDigits), currencySymbol)
		}

		if details.ProposerAttestationInclusionReward > 0 {
			income += fmt.Sprintf("Proposer Att. Inc. Reward: %s %s<br/>", FormatAddCommasFormatted(currencyFunc(details.ProposerAttestationInclusionReward, currencyName).InexactFloat64(), maxDigits), currencySymbol)
		}

		if details.ProposerSyncInclusionReward > 0 {
			income += fmt.Sprintf("Proposer Sync Inc. Reward: %s %s<br/>", FormatAddCommasFormatted(currencyFunc(details.ProposerSyncInclusionReward, currencyName).InexactFloat64(), maxDigits), currencySymbol)
		}

		if details.SyncCommitteeReward > 0 {
			income += fmt.Sprintf("Sync Comm. Reward: %s %s<br/>", FormatAddCommasFormatted(currencyFunc(details.SyncCommitteeReward, currencyName).InexactFloat64(), maxDigits), currencySymbol)
		}

		if details.SyncCommitteePenalty > 0 {
			income += fmt.Sprintf("Sync Comm. Penalty: %s %s<br/>", FormatAddCommasFormatted(currencyFunc(details.SyncCommitteePenalty, currencyName).InexactFloat64()*-1, maxDigits), currencySymbol)
		}

		if details.SlashingReward > 0 {
			income += fmt.Sprintf("Slashing Reward: %s %s<br/>", FormatAddCommasFormatted(currencyFunc(details.SlashingReward, currencyName).InexactFloat64(), maxDigits), currencySymbol)
		}

		if details.SlashingPenalty > 0 {
			income += fmt.Sprintf("Slashing Penalty: %s %s<br/>", FormatAddCommasFormatted(currencyFunc(details.SlashingPenalty, currencyName).InexactFloat64()*-1, maxDigits), currencySymbol)
		}

		income += fmt.Sprintf("Total: %s %s", FormatAddCommasFormatted(currencyFunc(details.TotalClRewards(), currencyName).InexactFloat64(), maxDigits), currencySymbol)
	}

	if *balance < 0 {
		return template.HTML(fmt.Sprintf("<span title='%s' data-html=\"true\" data-toggle=\"tooltip\" class=\"text-danger float-right\">%s %s</span>", income, FormatAddCommasFormatted(currencyFunc(*balance, currencyName).InexactFloat64(), maxDigits), currencySymbol))
	}
	return template.HTML(fmt.Sprintf("<span title='%s' data-html=\"true\" data-toggle=\"tooltip\" class=\"text-success float-right\">+%s %s</span>", income, FormatAddCommasFormatted(currencyFunc(*balance, currencyName).InexactFloat64(), maxDigits), currencySymbol))
}

// FormatBalanceChange will return a string for a balance change
func FormatBalanceChange(balance *int64, currency string) template.HTML {
	if currency == Config.Frontend.ClCurrency {
		if balance == nil || *balance == 0 {
			return template.HTML("<span> 0 " + currency + "</span>")
		}
		if *balance < 0 {
			return template.HTML(fmt.Sprintf("<span class=\"text-danger float-right\">%s GWei</span>", FormatAddCommasFormatted(ClToCurrencyGwei(*balance, currency).InexactFloat64(), 0)))
		}
		return template.HTML(fmt.Sprintf("<span class=\"text-success float-right\">+%s GWei</span>", FormatAddCommasFormatted(ClToCurrencyGwei(*balance, currency).InexactFloat64(), 0)))
	}
	if balance == nil {
		return template.HTML("<span> 0 " + currency + "</span>")
	}
	balanceFormated := FormatFloat(ClToCurrency(*balance, currency).InexactFloat64(), 2)
	if *balance > 0 {
		return template.HTML("<span class=\"text-success\">" + balanceFormated + " " + currency + "</span>")
	}
	if *balance < 0 {
		return template.HTML("<span class=\"text-danger\">" + balanceFormated + " " + currency + "</span>")
	}
	return template.HTML("pending")
}

// FormatBalance will return a string for a balance
func FormatBalanceShort(balanceInt uint64, currency string) template.HTML {
	exchangeRate := price.GetPrice(Config.Frontend.ClCurrency, currency)
	balance := FormatFloat((float64(balanceInt)/float64(1e9))*float64(exchangeRate), 2)

	return template.HTML(balance)
}

func FormatFloatWithDigits(num float64, min, max int) template.HTML {
	return template.HTML(FormatFloatWithDigitsString(num, min, max))
}

// FormatFloatWithDigitsString formats num with max amount of digits after comma but stop after min number of non-zero-digits after comma. In other words it can be used to format a number with the least amount of characters keeping a threshold of significant digits.
//
// examples:
//
//	FormatFloatWithDigitsString(0.01234,2,2) = "0.01"
//	FormatFloatWithDigitsString(0.01234,2,3) = "0.012"
//	FormatFloatWithDigitsString(0.01234,2,4) = "0.012"
//	FormatFloatWithDigitsString(0.01234,3,4) = "0.0123"
func FormatFloatWithDigitsString(num float64, min, max int) string {
	if max > 18 {
		max = 18
	}
	if min > max {
		min = max
	}
	a := fmt.Sprintf(fmt.Sprintf("%%.%df", max), num)
	b := strings.Split(a, ".")
	if len(b) < 2 {
		return b[0]
	}
	idx := strings.IndexAny(b[1], "123456789")
	if idx == -1 {
		return b[0]
	}
	if idx+min > len(b[1]) {
		return b[0] + "." + b[1]
	}
	return b[0] + "." + b[1][:idx+min]
}

func FormatAddCommasFormatted(num float64, precision uint) template.HTML {
	p := message.NewPrinter(language.English)
	s := p.Sprintf(fmt.Sprintf("%%.%vf", precision), num)
	if precision > 0 {
		s = strings.TrimRight(strings.TrimRight(s, "0"), ".")
	}
	return template.HTML(strings.ReplaceAll(string([]rune(p.Sprintf(s, num))), ",", `<span class="thousands-separator"></span>`))
}

func FormatBigNumberAddCommasFormated(val hexutil.Big, precision uint) template.HTML {
	return FormatAddCommasFormatted(float64(val.ToInt().Int64()), 0)
}

func FormatAddCommas(n uint64) template.HTML {
	number := FormatFloat(float64(n), 2)

	number = strings.ReplaceAll(number, ",", `<span class="thousands-separator"></span>`)
	return template.HTML(number)
}

// FormatBlockRoot will return the block-root formated as html
func FormatBlockRoot(blockRoot []byte) template.HTML {
	copyBtn := CopyButton(hex.EncodeToString(blockRoot))
	if len(blockRoot) < 32 {
		return "N/A"
	}
	return template.HTML(fmt.Sprintf("<a href=\"/slot/%x\">%v</a>%v", blockRoot, FormatHash(blockRoot), copyBtn))
}

// FormatBlockSlot will return the block-slot formated as html
func FormatBlockSlot(blockSlot uint64) template.HTML {
	return template.HTML(fmt.Sprintf("<a href=\"/slot/%d\">%s</a>", blockSlot, FormatAddCommas(blockSlot)))
}

// FormatAttestationInclusionSlot will return the block-slot formated as html
func FormatAttestationInclusionSlot(blockSlot uint64) template.HTML {
	if blockSlot == 0 {
		return template.HTML("-")
	} else {
		return FormatBlockSlot(blockSlot)
	}
}

// FormatAttestationInclusionSlot will return the block-slot formated as html
func FormatInclusionDelay(inclusionSlot uint64, delay int64) template.HTML {
	if inclusionSlot == 0 {
		return template.HTML("-")
	} else if delay > 32 {
		return template.HTML(fmt.Sprintf("<span class=\"text-danger\">%[1]d</span>", delay))
	} else if delay > 3 {
		return template.HTML(fmt.Sprintf("<span class=\"text-warning\">%[1]d</span>", delay))
	} else {
		return template.HTML(fmt.Sprintf("<span class=\"text-success\">%[1]d</span>", delay))
	}
}

// FormatSlotToTimestamp will return the time elapsed since blockSlot
func FormatSlotToTimestamp(blockSlot uint64) template.HTML {
	time := SlotToTime(blockSlot)
	return FormatTimestamp(time.Unix())
}

// FormatBlockStatus will return an html status for a block.
func FormatBlockStatus(status, slot uint64) template.HTML {
	if slot == 0 {
		return `<span class="badge badge-pill text-dark" style="background: rgba(179, 159, 70, 0.8); font-size: 12px; font-weight: 500;">Genesis</span>`
	} else if status == 0 && SlotToTime(slot).Before(time.Now().Add(time.Minute*-1)) {
		return `<span class="badge badge-pill bg-light text-dark" style="font-size: 12px; font-weight: 500;">Missed</span>`
	} else if status == 0 {
		return `<span class="badge badge-pill bg-light text-dark" style="font-size: 12px; font-weight: 500;">Scheduled</span>`
	} else if status == 1 {
		return `<span class="badge badge-pill bg-success text-white" style="font-size: 12px; font-weight: 500;">Proposed</span>`
	} else if status == 2 {
		return `<span class="badge badge-pill bg-warning text-white" style="font-size: 12px; font-weight: 500;">Missed</span>`
	} else if status == 3 {
		return `<span class="badge badge-pill bg-secondary text-white" style="font-size: 12px; font-weight: 500;">Missed (Orphaned)</span>`
	} else {
		return "Unknown"
	}
}

// FormatBlockStatusShort will return an html status for a block.
func FormatBlockStatusShort(status, slot uint64) template.HTML {
	// genesis <span class="badge text-dark" style="background: rgba(179, 159, 70, 0.8) none repeat scroll 0% 0%;">Genesis</span>
	if status == 0 && SlotToTime(slot).Before(time.Now().Add(time.Minute*-1)) {
		return `<span title="Scheduled Block" data-toggle="tooltip" class="mx-1 badge badge-pill bg-light text-dark" style="font-size: 12px; font-weight: 500;">Miss.</span>`
	} else if status == 0 {
		return `<span title="Scheduled Block" data-toggle="tooltip" class="mx-1 badge badge-pill bg-light text-dark" style="font-size: 12px; font-weight: 500;">Sche.</span>`
	} else if status == 1 {
		return `<span title="Proposed Block" data-toggle="tooltip" class="mx-1 badge badge-pill bg-success text-white" style="font-size: 12px; font-weight: 500;">Prop.</span>`
	} else if status == 2 {
		return `<span title="Missed Block" data-toggle="tooltip" class="mx-1 badge badge-pill bg-warning text-white" style="font-size: 12px; font-weight: 500;">Miss.</span>`
	} else if status == 3 {
		return `<span title="Missed Block (Orphaned)" data-toggle="tooltip" class="mx-1 badge badge-pill bg-secondary text-white" style="font-size: 12px; font-weight: 500;">Orph.</span>`
	} else {
		return "Unknown"
	}
}

// FormatBlockStatusShort will return an html status for a block.
func FormatWithdrawalShort(slot uint64, amount uint64) template.HTML {
	return template.HTML(fmt.Sprintf("<span title=\"Withdrawal processed in epoch %v during slot %v for %v\" data-toggle=\"tooltip\" class=\"mx-1 badge badge-pill bg-success text-white\" style=\"font-size: 12px; font-weight: 500;\"><i class=\"fas fa-money-bill\"></i></span>", EpochOfSlot(slot), slot, FormatCurrentBalance(amount, Config.Frontend.ClCurrency)))
}

func FormatTransactionType(txnType uint8) string {
	switch txnType {
	case 0:
		return "0 (legacy)"
	case 1:
		return "1 (Access-list)"
	case 2:
		return "2 (EIP-1559)"
	case 3:
		return "3 (Blob, EIP-4844)"
	default:
		return fmt.Sprintf("%v (???)", txnType)
	}
}

// FormatCurrentBalance will return the current balance formated as string with 9 digits after the comma (1 gwei = 1e9 eth)
func FormatCurrentBalance(balanceInt uint64, currency string) template.HTML {
	return template.HTML(fmt.Sprintf(`%s %v`, exchangeAndTrim(Config.Frontend.ClCurrency, currency, float64(balanceInt), false), currency))
}

// FormatDepositAmount will return the deposit amount formated as string
func FormatDepositAmount(balanceInt uint64, currency string) template.HTML {
	exchangeRate := price.GetPrice(Config.Frontend.ClCurrency, currency)
	balance := float64(balanceInt) / float64(Config.Frontend.ClCurrencyDivisor)
	return template.HTML(fmt.Sprintf("%.0f %v", balance*exchangeRate, currency))
}

// FormatEffectiveBalance will return the effective balance formated as string with 1 digit after the comma
func FormatEffectiveBalance(balanceInt uint64, currency string) template.HTML {
	exchangeRate := price.GetPrice(Config.Frontend.ClCurrency, currency)
	balance := float64(balanceInt) / float64(Config.Frontend.ClCurrencyDivisor)
	return template.HTML(fmt.Sprintf("%.1f %v", balance*exchangeRate, currency))
}

// FormatEpoch will return the epoch formated as html
func FormatEpoch(epoch uint64) template.HTML {
	return template.HTML(fmt.Sprintf("<a href=\"/epoch/%d\">%s</a>", epoch, FormatAddCommas(epoch)))
}

// FormatEth1AddressStringLowerCase will return the eth1-address formated as html string in lower case
func FormatEth1AddressStringLowerCase(addr []byte) template.HTML {
	return template.HTML(fmt.Sprintf("0x%x", addr))
}

// FormatEth1Address will return the eth1-address formated as html
func FormatEth1Address(addr []byte) template.HTML {
	eth1Addr := FixAddressCasing(fmt.Sprintf("%x", addr))
	copyBtn := CopyButton(eth1Addr)
	return template.HTML(fmt.Sprintf("<a href=\"/address/%s\" class=\"text-monospace\">%s…</a>%s", eth1Addr, eth1Addr[:8], copyBtn))
}

// FormatEth1Block will return the eth1-block formated as html
func FormatEth1Block(block uint64) template.HTML {
	return template.HTML(fmt.Sprintf("<a href=\"/block/%[1]d\">%[1]d</a>", block))
}

// FormatEth1BlockHash will return the eth1-block formated as html
func FormatEth1BlockHash(block []byte) template.HTML {
	copyBtn := CopyButton(hex.EncodeToString(block))
	return template.HTML(fmt.Sprintf("<a href=\"/block/%#[1]x\">%#[1]x</a>%s", block, copyBtn))
}

// FormatEth1TxHash will return the eth1-tx-hash formated as html
func FormatEth1TxHash(hash []byte) template.HTML {
	copyBtn := CopyButton(hex.EncodeToString(hash))
	return template.HTML(fmt.Sprintf(`<i class="fas fa-male mr-2"></i><a style="font-family: 'Roboto Mono'" href="/tx/0x%x">0x%v…</a>%v`, hash, hex.EncodeToString(hash)[:6], copyBtn))
}

// FormatGlobalParticipationRate will return the global-participation-rate formated as html
func FormatGlobalParticipationRate(e uint64, r float64, currency string) template.HTML {
	if e == 0 {
		return `<span class="text-small text-muted">` + CalculatingHint + `</span>`
	}
	p := message.NewPrinter(language.English)
	rr := fmt.Sprintf("%v%%", math.Round(r*10000)/100)
	tpl := `
	<div style="position:relative;width:inherit;height:inherit;">
	  %.0[1]f <small class="text-muted ml-3">(%[2]v)</small>
	  <div class="progress" style="position:absolute;bottom:-6px;width:100%%;height:4px;">
		<div class="progress-bar" role="progressbar" style="width: %[2]v;" aria-valuenow="%[2]v" aria-valuemin="0" aria-valuemax="100"></div>
	  </div>
	</div>`
	return template.HTML(p.Sprintf(tpl, float64(e)/float64(Config.Frontend.ClCurrencyDivisor)*price.GetPrice(Config.Frontend.ClCurrency, currency), rr))
}

// When 'finalized' is false and 'count' is 0, a in-progress hint is returned (three dots if 'shortenCalcHint' is true)
// If 'count' is positive or 'finalized' is true, 'count' is returned as a string
func FormatCount(count uint64, finalized bool, shortenCalcHint bool) template.HTML {
	if finalized || count > 0 {
		return template.HTML(fmt.Sprintf("%v", count))
	}
	if shortenCalcHint {
		return template.HTML("…")
	}
	return template.HTML(CalculatingHint)
}

func FormatEtherValue(currency string, ethPrice decimal.Decimal, currentPrice template.HTML) template.HTML {
	p := message.NewPrinter(language.English)
	currencySymbol := price.GetCurrencySymbol(currency)
	return template.HTML(p.Sprintf(`<span>%[1]s %[2]s</span> <span class="text-muted">@ %[1]s%[3]s/%[4]s</span>`, currencySymbol, ethPrice.StringFixed(2), currentPrice, Config.Frontend.ElCurrency))
}

func FormatPricedValue(val interface{}, valueCurrency, targetCurrency string) template.HTML {
	p := message.NewPrinter(language.English)
	pp := IfToDec(price.GetPrice(valueCurrency, targetCurrency))
	v := IfToDec(val)
	targetBalance := v.Mul(pp)
	valueSymbol := price.GetCurrencySymbol(valueCurrency)
	targetSymbol := price.GetCurrencySymbol(targetCurrency)
	return template.HTML(p.Sprintf(`<span>%[1]s %[2]s</span> <span class="text-muted">@ %[3]s %[4]s/%[5]s`, targetSymbol, targetBalance.StringFixed(2), valueSymbol, pp.StringFixed(2), targetSymbol))
}

// FormatGraffiti will return the graffiti formated as html
func FormatGraffiti(graffiti []byte) template.HTML {
	s := strings.Map(fixUtf, string(bytes.Trim(graffiti, "\x00")))
	h := template.HTMLEscapeString(s)
	if len(s) <= 6 {
		return template.HTML(fmt.Sprintf(`<span aria-graffiti="%#x">%s</span>`, graffiti, h))
	}
	if len(h) >= 8 {
		return template.HTML(fmt.Sprintf(`<span aria-graffiti="%#x" data-toggle="tooltip" data-placement="top" title="%s">%s...</span>`, graffiti, h, h[:8]))
	}
	return template.HTML(fmt.Sprintf(`<span aria-graffiti="%#x" data-toggle="tooltip" data-placement="top" title="%s">%s...</span>`, graffiti, h, h[:]))
}

// FormatGraffitiAsLink will return the graffiti formated as html-link
func FormatGraffitiAsLink(graffiti []byte) template.HTML {
	s := strings.Map(fixUtf, string(bytes.Trim(graffiti, "\x00")))
	h := template.HTMLEscapeString(s)
	u := url.QueryEscape(s)
	return template.HTML(fmt.Sprintf("<span aria-graffiti=\"%#x\"><a href=\"/slots?q=%s\">%s</a></span>", graffiti, u, h))
}

/*
  - FormatHash will return a hash formated as html
    hash is required, trunc is optional.
    Only the first value in trunc_opt will be used.
    ATTENTION: IT TRUNCATES BY DEFAULT, PASS FALSE TO trunc_opt TO DISABLE
*/
func FormatHash(hash []byte, trunc_opt ...bool) template.HTML {
	return template.HTML(fmt.Sprintf("<span class=\"text-monospace\">%s</span>", FormatHashRaw(hash, trunc_opt...)))
}

/*
  - FormatHashRaw will return a hash formated
    hash is required, trunc is optional.
    Only the first value in trunc_opt will be used.
    ATTENTION: IT TRUNCATES BY DEFAULT, PASS FALSE TO trunc_opt TO DISABLE
*/
func FormatHashRaw(hash []byte, trunc_opt ...bool) string {
	s := fmt.Sprintf("%#x", hash)
	if len(s) == 42 { // if it's an address, we checksum it (0x + 40)
		s = common.BytesToAddress(hash).Hex()
	}
	if len(s) >= 10 && (len(trunc_opt) < 1 || trunc_opt[0]) {
		return fmt.Sprintf("%s…%s", s[:6], s[len(s)-4:])
	}
	return s
}

// WithdrawalCredentialsToAddress converts withdrawalCredentials to an address if possible
func WithdrawalCredentialsToAddress(credentials []byte) ([]byte, error) {
	if IsValidWithdrawalCredentials(fmt.Sprintf("%#x", credentials)) && bytes.Equal(credentials[:1], []byte{0x01}) {
		return credentials[12:], nil
	}
	return nil, fmt.Errorf("invalid withdrawal credentials")
}

// AddressToWithdrawalCredentials converts a valid address to withdrawalCredentials
func AddressToWithdrawalCredentials(address []byte) ([]byte, error) {
	if IsValidEth1Address(fmt.Sprintf("%#x", address)) {
		credentials := make([]byte, 12, 32)
		credentials[0] = 0x01
		credentials = append(credentials, address...)
		return credentials, nil
	}
	return nil, fmt.Errorf("invalid eth1 address")
}

func FormatHashWithCopy(hash []byte) template.HTML {
	if len(hash) == 0 {
		return "N/A"
	}

	copyBtn := CopyButton(hex.EncodeToString(hash))
	return template.HTML(fmt.Sprintf(`<span>%v</span> %v`, FormatHash(hash), copyBtn))
}

func formatWithdrawalHash(hash []byte) template.HTML {
	var colorClass string
	if hash[0] == 0x01 {
		colorClass = "text-success"
	} else {
		colorClass = "text-warning"
	}

	return template.HTML(fmt.Sprintf("<span class=\"text-monospace %s\">%#x</span><span class=\"text-monospace\">%x…%x</span>", colorClass, hash[:1], hash[1:2], hash[len(hash)-2:]))
}

func FormatWithdawalCredentials(hash []byte, addCopyButton bool) template.HTML {
	if len(hash) != 32 {
		return "INVALID CREDENTIALS"
	}

	var text template.HTML
	if hash[0] == 0x01 {
		text = template.HTML(fmt.Sprintf("<a href=\"/address/0x%x\">%s</a>", hash[12:], formatWithdrawalHash(hash)))
	} else {
		text = formatWithdrawalHash(hash)
	}

	if addCopyButton {
		text += template.HTML(fmt.Sprintf("<i class=\"fa fa-copy text-muted p-1\" role=\"button\" data-toggle=\"tooltip\" title=\"Copy to clipboard\" data-clipboard-text=\"%#x\"></i>", hash))
	}

	return text
}

func FormatAddressToWithdrawalCredentials(address []byte, addCopyButton bool) template.HTML {
	credentials, err := hex.DecodeString("010000000000000000000000")
	if err != nil {
		return "INVALID CREDENTIALS"
	}
	credentials = append(credentials, address...)

	return FormatWithdawalCredentials(credentials, addCopyButton)
}

func CopyButton(clipboardText interface{}) string {
	value := fmt.Sprintf("%v", clipboardText)
	if len(value) < 2 || value[0] != '0' || value[1] != 'x' {
		value = "0x" + value
	}
	return fmt.Sprintf(`<i class="fa fa-copy text-muted text-white ml-2 p-1" style="opacity: .8;" role="button" data-toggle="tooltip" title="Copy to clipboard" data-clipboard-text=%s></i>`, value)
}

func CopyButtonText(clipboardText interface{}) string {
	return fmt.Sprintf(`<i class="fa fa-copy text-muted ml-2 p-1" role="button" data-toggle="tooltip" title="Copy to clipboard" data-clipboard-text=%v></i>`, clipboardText)
}

func FormatBitlist(b []byte) template.HTML {
	p := bitfield.Bitlist(b)
	return formatBits(p.BytesNoTrim(), int(p.Len()))
}

func formatBits(b []byte, length int) template.HTML {
	var buf strings.Builder
	buf.WriteString("<div class=\"text-bitfield text-monospace\">")
	perLine := 8
	for y := 0; y < len(b); y += perLine {
		start, end := y*8, (y+perLine)*8
		if end >= length {
			end = length
		}
		for x := start; x < end; x++ {
			if x%8 == 0 {
				if x != 0 {
					buf.WriteString("</span> ")
				}
				buf.WriteString("<span>")
			}
			bit := BitAtVector(b, x)
			if bit {
				buf.WriteString("1")
			} else {
				buf.WriteString("0")
			}
		}
		buf.WriteString("</span><br/>")
	}
	buf.WriteString("</div>")
	return template.HTML(buf.String())
}

func formatBitvectorValidators(bits []byte, validators []uint64) template.HTML {
	invalidLen := false
	if len(bits)*8 != len(validators) {
		invalidLen = true
	}
	var buf strings.Builder
	buf.WriteString("<pre class=\"text-monospace\" style=\"font-size:1rem;\">")
	for i := 0; i < len(bits)*8; i++ {
		if invalidLen {
			if BitAtVector(bits, i) {
				buf.WriteString("1")
			} else {
				buf.WriteString("0")
			}
		} else {
			val := validators[i]
			if BitAtVector(bits, i) {
				buf.WriteString(fmt.Sprintf("<a title=\"Validator %[1]d\" href=\"/validator/%[1]d\">1</a>", val))
			} else {
				buf.WriteString(fmt.Sprintf("<a title=\"Validator %[1]d\" href=\"/validator/%[1]d\">0</a>", val))
			}
		}

		if (i+1)%64 == 0 {
			buf.WriteString("\n")
		} else if (i+1)%8 == 0 {
			buf.WriteString(" ")
		}
	}
	buf.WriteString("</pre>")
	return template.HTML(buf.String())
}

func FormatParticipation(v float64) template.HTML {
	return template.HTML(fmt.Sprintf("<span>%.2f %%</span>", v*100.0))
}

func FormatIncomeClEl(income types.ClEl, currency string) template.HTML {
	className := "text-success"
	if income.Total.Cmp(decimal.Zero) < 0 {
		className = "text-danger"
	}

	if income.Cl.Cmp(decimal.Zero) != 0 || income.El.Cmp(decimal.Zero) != 0 {
		return template.HTML(fmt.Sprintf(`
		<span class="%s" data-toggle="tooltip"
			data-html="true"
			title="
			CL: %s <br> 
			EL: %s">
			<b>%s</b>
		</span>`,
			className,
			FormatElCurrency(income.Cl, currency, 5, true, true, false, false),
			FormatElCurrency(income.El, currency, 5, true, true, false, false), // we use FormatElCurrency here because all values in income-struct are in el-currency
			FormatElCurrency(income.Total, currency, 5, true, true, false, false)))
	} else {
		return template.HTML(fmt.Sprintf(`<span><b>%s</b></span>`, FormatElCurrency(income.Total, currency, 5, true, true, false, false)))
	}
}

func FormatIncomeClElInt64(income types.ClElInt64, currency string) template.HTML {
	var incomeTrimmed string = exchangeAndTrim(Config.Frontend.ClCurrency, currency, income.Total, true)
	className := "text-success"
	if income.Total < 0 {
		className = "text-danger"
	}

	if income.Cl != 0 || income.El != 0 {
		return template.HTML(fmt.Sprintf(`
		<span class="%s" data-toggle="tooltip"
			data-html="true"
			title="
			CL: %s <br> 
			EL: %s">
			<b>%s %s</b>
		</span>`,
			className,
			FormatClCurrency(income.Cl, currency, 5, true, true, false, false),
			FormatClCurrency(income.El, currency, 5, true, true, false, false), // we use FormatClCurrency here because all values in income-struct are in Gwei
			incomeTrimmed,
			currency))
	} else {
		return template.HTML(fmt.Sprintf(`<span>%s %s</span>`, incomeTrimmed, currency))
	}
}

func FormatIncome(balance interface{}, currency string, includeCurrency bool) template.HTML {
	balanceFloat64 := IfToDec(balance).InexactFloat64()
	var income string = exchangeAndTrim(Config.Frontend.ElCurrency, currency, balanceFloat64, true)

	if includeCurrency {
		currency = " " + currency
	} else {
		currency = ""
	}

	if balanceFloat64 > 0 {
		return template.HTML(fmt.Sprintf(`<span class="text-success"><b>%s%s</b></span>`, income, currency))
	} else if balanceFloat64 < 0 {
		return template.HTML(fmt.Sprintf(`<span class="text-danger"><b>%s%s</b></span>`, income, currency))
	} else {
		return template.HTML(fmt.Sprintf(`<span>0%s</span>`, currency))
	}
}

func exchangeAndTrim(valueCurrency, exCurrency string, amount float64, addPositiveSign bool) string {
	decimals := 5
	preCommaDecimals := 2

	if valueCurrency != Config.Frontend.ClCurrency {
		decimals = 4
		preCommaDecimals = 4
	}

	exchangeRate := price.GetPrice(valueCurrency, exCurrency)
	exchangedAmount := float64(amount) * exchangeRate
	// lost precision here but we don't need it for frontend
	income, _ := trimAmount(big.NewInt(int64(exchangedAmount)), 9, preCommaDecimals, decimals, addPositiveSign)
	return income
}

func FormatIncomeSql(balanceInt sql.NullInt64, currency string) template.HTML {

	if !balanceInt.Valid {
		return template.HTML(fmt.Sprintf(`<b>0 %v</b>`, currency))
	}

	exchangeRate := price.GetPrice(Config.Frontend.ClCurrency, currency)
	balance := float64(balanceInt.Int64) / float64(1e9)

	if balance > 0 {
		return template.HTML(fmt.Sprintf(`<span class="text-success"><b>+%v %v</b></span>`, FormatFloat(balance*exchangeRate, 5), currency))
	} else if balance < 0 {
		return template.HTML(fmt.Sprintf(`<span class="text-danger"><b>%v %v</b></span>`, FormatFloat(balance*exchangeRate, 5), currency))
	} else {
		return template.HTML(fmt.Sprintf(`<b>%v %v</b>`, balance*exchangeRate, currency))
	}
}

func FormatSqlInt64(i sql.NullInt64) template.HTML {
	if !i.Valid {
		return "-"
	} else {
		return template.HTML(fmt.Sprintf(`%v`, i.Int64))
	}
}

// FormatPercentage will return a string for a percentage
func FormatPercentage(percentage float64) string {
	if math.IsInf(percentage, 0) || math.IsNaN(percentage) {
		return fmt.Sprintf("%.0f", float64(0))
	}
	return fmt.Sprintf("%.0f", percentage*float64(100))
}

// FormatPercentageWithPrecision will return a string for a percentage
func FormatPercentageWithPrecision(percentage float64, precision int) string {
	return fmt.Sprintf("%."+strconv.Itoa(precision)+"f", percentage*float64(100))
}

// FormatPercentageWithGPrecision will return a string for a percentage the maximum number of significant digits (trailing zeros are removed).
func FormatPercentageWithGPrecision(percentage float64, precision int) string {
	return fmt.Sprintf("%."+strconv.Itoa(precision)+"g", percentage*float64(100))
}

// FormatPublicKey will return html formatted text for a validator-public-key
func FormatPublicKey(validator []byte) template.HTML {
	copyBtn := CopyButton(hex.EncodeToString(validator))
	// return template.HTML(fmt.Sprintf("<i class=\"fas fa-male\"></i> <a href=\"/validator/0x%x\">%v</a>", validator, FormatHash(validator)))
	return template.HTML(fmt.Sprintf(`<i class="fas fa-male mr-2"></i><a style="font-family: 'Roboto Mono'" href="/validator/0x%x">0x%v…</a>%v`, validator, hex.EncodeToString(validator)[:6], copyBtn))
}

func FormatMachineName(machineName string) template.HTML {
	if machineName == "" {
		machineName = "Default"
	}
	return template.HTML(fmt.Sprintf("<i class=\"fas fa-hdd\"></i> %v", machineName))
}

// FormatTimestamp will return a timestamp formated as html. This is supposed to be used together with client-side js
func FormatTimestamp(ts int64) template.HTML {
	return template.HTML(fmt.Sprintf("<span class=\"timestamp\" data-toggle=\"tooltip\" data-placement=\"top\" data-timestamp=\"%d\"></span>", ts))
}

// FormatTsWithoutTooltip will return a timestamp formated as html. This is supposed to be used together with client-side js
func FormatTsWithoutTooltip(ts int64) template.HTML {
	return template.HTML(fmt.Sprintf("<span class=\"timestamp\" data-timestamp=\"%d\"></span>", ts))
}

// FormatValidatorStatus will return the validator-status formated as html
// possible states
// pending, active_online, active_offline, exiting_online, exciting_offline, slashing_online, slashing_offline, exited, slashed
func FormatValidatorStatus(status string) template.HTML {
	if status == "deposited" || status == "deposited_invalid" {
		return "<span><b>Deposited</b></span>"
	} else if status == "pending" {
		return "<span><b>Pending</b></span>"
	} else if status == "active_online" {
		return "<b>Active</b> <i class=\"fas fa-power-off fa-sm text-success\"></i>"
	} else if status == "active_offline" {
		return "<span data-toggle=\"tooltip\" title=\"No attestation in the last 2 epochs\"><b>Active</b> <i class=\"fas fa-power-off fa-sm text-danger\"></i></span>"
	} else if status == "exiting_online" {
		return "<b>Exiting</b> <i class=\"fas fa-power-off fa-sm text-success\"></i>"
	} else if status == "exiting_offline" {
		return "<span data-toggle=\"tooltip\" title=\"No attestation in the last 2 epochs\"><b>Exiting</b> <i class=\"fas fa-power-off fa-sm text-danger\"></i></span>"
	} else if status == "slashing_online" {
		return "<b>Slashing</b> <i class=\"fas fa-power-off fa-sm text-success\"></i>"
	} else if status == "slashing_offline" {
		return "<span data-toggle=\"tooltip\" title=\"No attestation in the last 2 epochs\"><b>Slashing</b> <i class=\"fas fa-power-off fa-sm text-danger\"></i></span>"
	} else if status == "exited" {
		return "<span><b>Exited</b></span>"
	} else if status == "slashed" {
		return "<span><b>Slashed</b></span>"
	}
	return "<b>Unknown</b>"
}

func formatPool(tag []string) string {
	if len(tag) > 1 {
		tagType := tag[0]
		tagName := strings.Split(tag[len(tag)-1], " ")
		if len(tagName) > 1 {
			_, err := strconv.ParseInt(tagName[len(tagName)-1], 10, 64)
			if err == nil {
				name := ""
				for _, s := range tagName[:len(tagName)-1] {
					if s == "-" {
						continue
					}
					name += s + " "
				}
				return fmt.Sprintf(`<a href='/pools' style="all: unset; cursor: pointer;" data-toggle="tooltip" title="This validator is part of a staking-pool"><span style="font-size: 18px;" class="bg-light text-dark badge-pill pr-2 pl-0 mr-1"><span class="bg-dark text-light rounded-left mr-1 px-1">%s</span> %s</span></a>`, tagType, name)
			}
		}
		return fmt.Sprintf(`<a href='/pools' style="all: unset; cursor: pointer;" data-toggle="tooltip" title="This validator is part of a staking-pool"><span style="font-size: 18px;" class="bg-light text-dark badge-pill pr-2 pl-0 mr-1"><span class="bg-dark text-light rounded-left mr-1 px-1">%s</span> %s</span></a>`, tagType, tag[len(tag)-1])
	}
	return ""
}

func formatSpecialTag(tag string) string {
	special_tag := strings.Split(tag, ":")
	if len(special_tag) > 1 {
		if special_tag[0] == "pool" {
			return formatPool(special_tag)
		}
	}
	return fmt.Sprintf(`<span style="font-size: 18px;" class="badge bg-dark text-light mr-1">%s</span>`, tag)
}

// FormatValidatorTag will return html formated text of a validator-tag.
// Depending on the tag it will describe the tag in a tooltip and link to more information regarding the tag.
func FormatValidatorTag(tag string) template.HTML {
	var result string
	switch tag {
	case "rocketpool":
		result = `<span style="background-color: rgba(240, 149, 45, .2); font-size: 18px;" class="badge-pill mr-1 font-weight-normal" data-toggle="tooltip" title="Rocket Pool Validator"><a style="color: var(--yellow);" href="/pools/rocketpool">Rocket Pool</a></span>`
	case "ssv":
		result = `<span style="background-color: rgba(238, 113, 18, .2); font-size: 18px;" class="badge-pill mr-1 font-weight-normal" data-toggle="tooltip" title="Secret Shared Validator"><a style="color: var(--orange);" href="https://github.com/bloxapp/ssv/">SSV</a></span>`
	default:
		result = formatSpecialTag(tag)
	}
	return template.HTML(result)
}

func FormatValidatorTags(tags []string) template.HTML {
	str := ""
	for _, tag := range tags {
		str += string(FormatValidatorTag(tag)) + " "
	}
	return template.HTML(str)
}

// FormatValidator will return html formatted text for a validator
func FormatValidator(validator uint64) template.HTML {
	return template.HTML(fmt.Sprintf("<i class=\"fas fa-male mr-2\"></i><a href=\"/validator/%v\">%v</a>", validator, validator))
}

func FormatValidatorWithName(validator interface{}, name string) template.HTML {
	var validatorRead string
	var validatorLink string
	switch v := validator.(type) {
	case []byte:
		if len(v) > 2 {
			validatorRead = fmt.Sprintf("0x%x…%x", v[:2], v[len(v)-2:])
		}
		validatorLink = fmt.Sprintf("%x", v)
	default:
		validatorRead = fmt.Sprintf("%v", v)
		validatorLink = fmt.Sprintf("%v", v)
	}

	if name != "" {
		return template.HTML(fmt.Sprintf("<i class=\"fas fa-male mr-2\"></i> <a href=\"/validator/%v\"><span class=\"text-truncate\">"+html.EscapeString(name)+"</span></a>", validatorLink))
	} else {
		return template.HTML(fmt.Sprintf("<i class=\"fas fa-male mr-2\"></i> <a href=\"/validator/%v\">%v</a>", validatorLink, validatorRead))
	}
}

// FormatValidatorInt64 will return html formatted text for a validator (for an int64 validator-id)
func FormatValidatorInt64(validator int64) template.HTML {
	return FormatValidator(uint64(validator))
}

// FormatSlashedValidatorInt64 will return html formatted text for a slashed validator
func FormatSlashedValidatorInt64(validator int64) template.HTML {
	return template.HTML(fmt.Sprintf("<i class=\"fas fa-user-slash text-danger mr-2\"></i><a href=\"/validator/%v\">%v</a>", validator, validator))
}

// FormatSlashedValidator will return html formatted text for a slashed validator
func FormatSlashedValidator(validator uint64) template.HTML {
	return template.HTML(fmt.Sprintf("<i class=\"fas fa-user-slash text-danger mr-2\"></i><a href=\"/validator/%v\">%v</a>", validator, validator))
}

// FormatSlashedValidator will return html formatted text for a slashed validator
func FormatSlashedValidatorWithName(validator uint64, name string) template.HTML {
	if name != "" {
		return template.HTML(fmt.Sprintf("<i class=\"fas fa-user-slash text-danger mr-2\"></i><a href=\"/validator/%v\">%v (<span class=\"text-truncate\">"+html.EscapeString(name)+"</span>)</a>", validator, validator))
	} else {
		return FormatSlashedValidator(validator)
	}
}

// FormatSlashedValidators will return html formatted text for slashed validators
func FormatSlashedValidators(validators []uint64) template.HTML {
	vals := make([]string, 0, len(validators))
	for _, v := range validators {
		vals = append(vals, fmt.Sprintf("<i class=\"fas fa-user-slash text-danger mr-2\"></i><a href=\"/validator/%v\">%v</a>", v, v))
	}
	return template.HTML(strings.Join(vals, ","))
}

// FormatSlashedValidators will return html formatted text for slashed validators
func FormatSlashedValidatorsWithName(validators []uint64, nameMap map[uint64]string) template.HTML {
	vals := make([]string, 0, len(validators))
	for _, v := range validators {
		name := nameMap[v]
		if name != "" {
			vals = append(vals, string(FormatSlashedValidatorWithName(v, name)))
		} else {
			vals = append(vals, string(FormatSlashedValidator(v)))
		}
	}
	return template.HTML(strings.Join(vals, ","))
}

// FormatYesNo will return yes or no formated as html
func FormatYesNo(yes bool) template.HTML {
	if yes {
		return `<span class="badge badge-pill bg-success text-white" style="font-size: 12px; font-weight: 500;">Yes</span>`
	}
	return `<span class="badge badge-pill bg-warning text-white" style="font-size: 12px; font-weight: 500;">No</span>`
}

func FormatValidatorName(name string) template.HTML {
	str := strings.Map(fixUtf, template.HTMLEscapeString(name))
	return template.HTML(fmt.Sprintf("<b><abbr title=\"This name has been set by the owner of this validator. Pool tags have been set by the beaconcha.in team.\">%s</abbr></b>", str))
}

func FormatAttestationInclusionEffectiveness(eff float64) template.HTML {
	tooltipText := "The attestation inclusion effectiveness should be 80% or higher to minimize reward penalties."
	if eff == 0 {
		return template.HTML(`<span class="text-danger" data-toggle="tooltip" title="Validator did not attest during the last 100 epochs"> N/A <i class="fas fa-frown"></i>`)
	} else if eff >= 100 {
		return template.HTML(fmt.Sprintf(`<span class="text-success" data-toggle="tooltip" title="%s"> %.0f%% - Perfect <i class="fas fa-grin-stars"></i>`, tooltipText, eff))
	} else if eff > 80 {
		return template.HTML(fmt.Sprintf(`<span class="text-success" data-toggle="tooltip" title="%s"> %.0f%% - Good <i class="fas fa-smile"></i></span>`, tooltipText, eff))
	} else if eff > 60 {
		return template.HTML(fmt.Sprintf(`<span class="text-warning" data-toggle="tooltip" title="%s"> %.0f%% - Fair <i class="fas fa-meh"></i></span>`, tooltipText, eff))
	} else {
		return template.HTML(fmt.Sprintf(`<span class="text-danger" data-toggle="tooltip" title="%s"> %.0f%% - Bad <i class="fas fa-frown"></i></span>`, tooltipText, eff))
	}
}

func FormatPercentageColoredEmoji(percentage float64) template.HTML {
	if math.IsInf(percentage, 0) || math.IsNaN(percentage) {
		percentage = 0
	} else {
		percentage = percentage * 100
	}
	if percentage == 100 {
		return template.HTML(fmt.Sprintf(`<span class="text-success">%.0f%% <i class="fas fa-grin-stars"></i></span>`, percentage))
	} else if percentage >= 90 {
		return template.HTML(fmt.Sprintf(`<span class="text-success">%.0f%% <i class="fas fa-smile"></i></span>`, percentage))
	} else if percentage >= 80 {
		return template.HTML(fmt.Sprintf(`<span class="text-warning">%.0f%% <i class="fas fa-smile"></i></span>`, percentage))
	} else if percentage >= 60 {
		return template.HTML(fmt.Sprintf(`<span class="text-warning">%.0f%% <i class="fas fa-meh"></i></span>`, percentage))
	}
	return template.HTML(fmt.Sprintf(`<span class="text-danger">%.0f%% <i class="fas fa-frown"></i></span>`, percentage))
}

func DerefString(str *string) string {
	if str != nil {
		return *str
	}
	return ""
}

// TrLang returns translated text based on language tag and text id
func TrLang(lang string, key string) template.HTML {
	I18n := getLocaliser()
	return template.HTML(I18n.Tr(lang, key))
}

func KFormatterEthPrice(price uint64) template.HTML {
	if price > 999 {
		ethTruncPrice := fmt.Sprint(float64(int((float64(price)/float64(1000))*10))/float64(10)) + "k"
		return template.HTML(ethTruncPrice)
	}
	return template.HTML(fmt.Sprint(price))
}

func FormatRPL(num string) string {
	floatNum, _ := strconv.ParseFloat(num, 64)
	return fmt.Sprintf("%.2f", floatNum/math.Pow10(18)) + " RPL"
}

func FormatETH(num string) string {
	floatNum, _ := strconv.ParseFloat(num, 64)
	return fmt.Sprintf("%.4f", floatNum/math.Pow10(18)) + " " + Config.Frontend.ClCurrency
}

func FormatFloat(num float64, precision int) string {
	p := message.NewPrinter(language.English)
	f := fmt.Sprintf("%%.%vf", precision)
	s := strings.TrimRight(strings.TrimRight(p.Sprintf(f, num), "0"), ".")
	r := []rune(p.Sprintf(s, num))
	return string(r)
}

func FormatNotificationChannel(ch types.NotificationChannel) template.HTML {
	label, ok := types.NotificationChannelLabels[ch]
	if !ok {
		return ""
	}
	return label
}

func FormatTokenBalance(balance *types.Eth1AddressBalance) template.HTML {
	p := message.NewPrinter(language.English)

	tokenDecimals := decimal.NewFromBigInt(new(big.Int).SetBytes(balance.Metadata.Decimals), 0)
	ethDiv := decimal.NewFromInt(Config.Frontend.ElCurrencyDivisor)
	tokenDiv := decimal.NewFromInt(10).Pow(tokenDecimals)

	tokenBalance := decimal.NewFromBigInt(new(big.Int).SetBytes(balance.Balance), 0).DivRound(tokenDiv, 18)

	tokenPriceEth := decimal.New(0, 0)
	if len(balance.Metadata.Price) > 0 {
		tokenPriceEth = decimal.NewFromBigInt(new(big.Int).SetBytes(balance.Metadata.Price), 0).DivRound(decimal.NewFromInt(Config.Frontend.ElCurrencyDivisor), 18)
	}

	ethPriceUsd := decimal.NewFromFloat(price.GetPrice(Config.Frontend.ElCurrency, "USD"))
	tokenPriceUsd := ethPriceUsd.Mul(tokenPriceEth).Mul(tokenDiv).DivRound(ethDiv, 18)
	tokenBalanceUsd := tokenBalance.Mul(tokenPriceUsd)

	symbolTitle := FormatTokenSymbolTitle(balance.Metadata.Symbol)
	symbol := FormatTokenSymbol(balance.Metadata.Symbol)
	logo := ""
	if len(balance.Metadata.Logo) != 0 {
		logo = fmt.Sprintf(`<img class="mr-1" style="height: 1.2rem;" src="data:image/png;base64, %s">`, base64.StdEncoding.EncodeToString(balance.Metadata.Logo))
	}

	return template.HTML(p.Sprintf(`
	<div class="token-balance-col token-name text-truncate d-flex align-items-center justify-content-between flex-wrap">
		<div class="token-icon p-1">
			<a href='/token/0x%x?a=0x%x'>
				<span>%s</span> <span title="%s">%s</span>
			</a> 
		</div>
		<div class="token-price-balance p-1">
			<span class="text-muted" style="font-size: 90%%;">$ %s</span>
		</div>
	</div> 
	<div class="token-balance-col token-balance d-flex align-items-center justify-content-between flex-wrap">
		<div class="token-holdings p-1">
			<span class="token-holdings">%s</span>
		</div>
		<div class="token-price p-1">
			<span class="text-muted" style="font-size: 90%%;">$ %s</span>
		</div>
	</div>`, balance.Token, balance.Address, logo, symbolTitle, symbol, tokenPriceUsd.StringFixed(2), FormatThousandsEnglish(tokenBalance.String()), tokenBalanceUsd.StringFixed(2)))
}

func FormatAddressEthBalance(balance *types.Eth1AddressBalance) template.HTML {
	e := new(big.Int).SetBytes(balance.Metadata.Decimals)
	d := new(big.Int).Exp(big.NewInt(10), e, nil)
	balWei := decimal.NewFromBigInt(new(big.Int).SetBytes(balance.Balance), 0)
	balEth := balWei.DivRound(decimal.NewFromBigInt(d, 0), int32(e.Int64()))

	p := message.NewPrinter(language.English)
	return template.HTML(p.Sprintf(`
		<div class="d-flex align-items-center">
			<svg style="width: 1rem; height: 1rem;">
				<use xlink:href="#ethereum-diamond-logo"/>
			</svg> 
			<span class="token-holdings">%v %v</span>
		</div>`, balEth, Config.Frontend.ElCurrency))
}

func FormatTokenValue(balance *types.Eth1AddressBalance, fullAmountTooltip bool) template.HTML {
	decimals := new(big.Int).SetBytes(balance.Metadata.Decimals)
	p := message.NewPrinter(language.English)
	mul := decimal.NewFromFloat(float64(10)).Pow(decimal.NewFromBigInt(decimals, 0))
	num := decimal.NewFromBigInt(new(big.Int).SetBytes(balance.Balance), 0)
	tokenValue := num.DivRound(mul, int32(decimals.Int64()))
	tokenValueFormatted := FormatThousandsEnglish(tokenValue.String())

	tooltip := ""
	if fullAmountTooltip {
		tooltip = fmt.Sprintf(` data-toggle="tooltip" data-placement="top" title="%s"`, tokenValueFormatted)
	}
	return template.HTML(p.Sprintf("<span%s>%s</span>", tooltip, tokenValueFormatted))
}

func FormatErc20Decimals(balance []byte, metadata *types.ERC20Metadata) decimal.Decimal {
	decimals := new(big.Int).SetBytes(metadata.Decimals)
	mul := decimal.NewFromFloat(float64(10)).Pow(decimal.NewFromBigInt(decimals, 0))
	num := decimal.NewFromBigInt(new(big.Int).SetBytes(balance), 0)

	return num.DivRound(mul, int32(decimals.Int64()))
}

func FormatTokenName(balance *types.Eth1AddressBalance) template.HTML {
	logo := ""
	if len(balance.Metadata.Logo) != 0 {
		logo = fmt.Sprintf(`<img style="height: 20px;" src="data:image/png;base64, %s">`, base64.StdEncoding.EncodeToString(balance.Metadata.Logo))
	}
	symbolTitle := FormatTokenSymbolTitle(balance.Metadata.Symbol)
	symbol := FormatTokenSymbol(balance.Metadata.Symbol)
	return template.HTML(fmt.Sprintf(`<a href='/token/0x%x?a=0x%x' title="%s">%s %s</a>`, balance.Token, balance.Address, symbolTitle, logo, symbol))
}

func ToBase64(input []byte) string {
	return base64.StdEncoding.EncodeToString(input)
}

// FormatBalance will return a string for a balance
func FormatEth1TxStatus(status uint64) template.HTML {
	if status == 1 {
		return template.HTML("<h5 class=\"m-0\"><span class=\"badge badge-success badge-pill align-middle text-white\"><i class=\"fas fa-check-circle\"></i> Success</span></h5>")
	} else {
		return template.HTML("<h5 class=\"m-0\"><span class=\"badge badge-danger badge-pill align-middle text-white\"><i class=\"fas fa-times-circle\"></i> Failed</span></h5>")
	}
}

// FormatEth1AddressFull will return the eth1-address formated as html
func FormatEth1AddressFull(addr common.Address) template.HTML {
	return FormatAddress(addr.Bytes(), nil, "", false, false, true)
}
