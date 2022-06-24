package utils

import (
	"bytes"
	"database/sql"
	"encoding/hex"
	"eth2-exporter/price"
	"eth2-exporter/types"
	"fmt"
	"html"
	"html/template"
	"math"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/protolambda/ztyp/bitfields"

	eth1common "github.com/ethereum/go-ethereum/common"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func FormatMessageToHtml(message string) template.HTML {
	message = fmt.Sprint(strings.Replace(message, "Error: ", "", 1))
	return template.HTML(message)
}

// FormatSyncParticipationStatus will return a user-friendly format for an sync-participation-status number
func FormatSyncParticipationStatus(status uint64) template.HTML {
	if status == 0 {
		return `<span class="badge badge-pill bg-light text-dark" style="font-size: 12px; font-weight: 500;">Scheduled</span>`
	} else if status == 1 {
		return `<span class="badge badge-pill bg-success text-white" style="font-size: 12px; font-weight: 500;">Participated</span>`
	} else if status == 2 {
		return `<span class="badge badge-pill bg-danger text-white" style="font-size: 12px; font-weight: 500;">Missed</span>`
	} else if status == 3 {
		return `<span class="badge badge-pill bg-warning text-white" style="font-size: 12px; font-weight: 500;">No Block</span>`
	} else {
		return "Unknown"
	}
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
		return `<span class="badge badge-pill bg-warning text-white" style="font-size: 12px; font-weight: 500;">Orphaned</span>`
	} else {
		return "Unknown"
	}
}

// FormatAttestationStatusShort will return a user-friendly attestation for an attestation status number
func FormatAttestationStatusShort(status uint64) template.HTML {
	if status == 0 {
		return `<span title="Scheduled" data-toggle="tooltip" class="badge badge-pill bg-light text-dark" style="font-size: 12px; font-weight: 500;">Sche.</span>`
	} else if status == 1 {
		return `<span title="Attested" data-toggle="tooltip" class="badge badge-pill bg-success text-white" style="font-size: 12px; font-weight: 500;">Att.</span>`
	} else if status == 2 {
		return `<span title="Missed" data-toggle="tooltip" class="badge badge-pill bg-warning text-white" style="font-size: 12px; font-weight: 500;">Miss.</span>`
	} else if status == 3 {
		return `<span title="Orphaned" data-toggle="tooltip" class="badge badge-pill bg-warning text-white" style="font-size: 12px; font-weight: 500;">Orph.</span>`
	} else if status == 4 {
		return `<span title="Inactivity Leak" data-toggle="tooltip" class="badge badge-pill bg-danger text-white" style="font-size: 12px; font-weight: 500;">Leak</span>`
	} else if status == 5 {
		return `<span title="Inactive" data-toggle="tooltip" class="badge badge-pill bg-light text-dark" style="font-size: 12px; font-weight: 500;">Inac.</span>`
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
	exchangeRate := ExchangeRateForCurrency(currency)
	balance := FormatFloat((float64(balanceInt)/float64(1e9))*float64(exchangeRate), 2)

	return template.HTML(balance + " " + currency)
}

func FormatBalanceSql(balanceInt sql.NullInt64, currency string) template.HTML {
	if !balanceInt.Valid {
		return template.HTML("0 " + currency)
	}
	exchangeRate := ExchangeRateForCurrency(currency)
	balance := FormatFloat((float64(balanceInt.Int64)/float64(1e9))*float64(exchangeRate), 2)

	return template.HTML(balance + " " + currency)
}

func FormatBalanceGwei(balance *int64, currency string) template.HTML {
	if currency == "ETH" {
		balanceF := float64(*balance)
		if balance == nil {
			return template.HTML("<span> 0.00000 " + currency + "</span>")
		} else if *balance == 0 {
			return template.HTML("0")
		}

		if balanceF < 0 {
			return template.HTML(fmt.Sprintf("<span class=\"text-danger\">%.0f GWei</span>", balanceF))
		}
		return template.HTML(fmt.Sprintf("<span class=\"text-success\">+%.0f GWei</span>", balanceF))
	}
	return FormatBalanceChange(balance, currency)
}

// FormatBalanceChange will return a string for a balance change
func FormatBalanceChange(balance *int64, currency string) template.HTML {
	balanceF := float64(*balance) / float64(1e9)
	if currency == "ETH" {
		if balance == nil || *balance == 0 {
			return template.HTML("<span> 0 " + currency + "</span>")
		}

		if balanceF < 0 {
			return template.HTML(fmt.Sprintf("<span title=\"%.0f GWei\" data-toggle=\"tooltip\" class=\"text-danger\">%.5f ETH</span>", float64(*balance), balanceF))
		}
		return template.HTML(fmt.Sprintf("<span title=\"%.0f GWei\" data-toggle=\"tooltip\" class=\"text-success\">+%.5f ETH</span>", float64(*balance), balanceF))
	} else {
		if balance == nil {
			return template.HTML("<span> 0 " + currency + "</span>")
		}
		exchangeRate := ExchangeRateForCurrency(currency)
		balanceFormated := FormatFloat(balanceF*float64(exchangeRate), 2)

		if *balance > 0 {
			return template.HTML("<span class=\"text-success\">" + balanceFormated + " " + currency + "</span>")
		}
		if *balance < 0 {
			return template.HTML("<span class=\"text-danger\">" + balanceFormated + " " + currency + "</span>")
		}

		return template.HTML("pending")

	}
}

// FormatBalance will return a string for a balance
func FormatBalanceShort(balanceInt uint64, currency string) template.HTML {
	exchangeRate := ExchangeRateForCurrency(currency)
	balance := FormatFloat((float64(balanceInt)/float64(1e9))*float64(exchangeRate), 2)

	return template.HTML(balance)
}

func FormatAddCommas(n uint64) template.HTML {
	number := FormatFloat(float64(n), 2)

	return template.HTML(number)
}

// FormatBlockRoot will return the block-root formated as html
func FormatBlockRoot(blockRoot []byte) template.HTML {
	copyBtn := CopyButton(hex.EncodeToString(blockRoot))
	if len(blockRoot) < 32 {
		return "N/A"
	}
	return template.HTML(fmt.Sprintf("<a href=\"/block/%x\">%v</a>%v", blockRoot, FormatHash(blockRoot), copyBtn))
}

// FormatBlockSlot will return the block-slot formated as html
func FormatBlockSlot(blockSlot uint64) template.HTML {
	return template.HTML(fmt.Sprintf("<a href=\"/block/%d\">%s</a>", blockSlot, FormatAddCommas(blockSlot)))
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
func FormatBlockStatus(status uint64) template.HTML {
	// genesis <span class="badge text-dark" style="background: rgba(179, 159, 70, 0.8) none repeat scroll 0% 0%;">Genesis</span>
	if status == 0 {
		return `<span class="badge badge-pill bg-light text-dark" style="font-size: 12px; font-weight: 500;">Scheduled</span>`
	} else if status == 1 {
		return `<span class="badge badge-pill bg-success text-white" style="font-size: 12px; font-weight: 500;">Proposed</span>`
	} else if status == 2 {
		return `<span class="badge badge-pill bg-warning text-white" style="font-size: 12px; font-weight: 500;">Missed</span>`
	} else if status == 3 {
		return `<span class="badge badge-pill bg-secondary text-white" style="font-size: 12px; font-weight: 500;">Orphaned</span>`
	} else {
		return "Unknown"
	}
}

// FormatBlockStatusShort will return an html status for a block.
func FormatBlockStatusShort(status uint64) template.HTML {
	// genesis <span class="badge text-dark" style="background: rgba(179, 159, 70, 0.8) none repeat scroll 0% 0%;">Genesis</span>
	if status == 0 {
		return `<span title="Scheduled" data-toggle="tooltip" class="badge badge-pill bg-light text-dark" style="font-size: 12px; font-weight: 500;">Sche.</span>`
	} else if status == 1 {
		return `<span title="Proposed" data-toggle="tooltip" class="badge badge-pill bg-success text-white" style="font-size: 12px; font-weight: 500;">Prop.</span>`
	} else if status == 2 {
		return `<span title="Missed" data-toggle="tooltip" class="badge badge-pill bg-warning text-white" style="font-size: 12px; font-weight: 500;">Miss.</span>`
	} else if status == 3 {
		return `<span title="Orphaned" data-toggle="tooltip" class="badge badge-pill bg-secondary text-white" style="font-size: 12px; font-weight: 500;">Orph.</span>`
	} else {
		return "Unknown"
	}
}

// FormatCurrentBalance will return the current balance formated as string with 9 digits after the comma (1 gwei = 1e9 eth)
func FormatCurrentBalance(balanceInt uint64, currency string) template.HTML {
	if currency == "ETH" {
		exchangeRate := ExchangeRateForCurrency(currency)
		balance := float64(balanceInt) / float64(1e9)
		return template.HTML(fmt.Sprintf("%.5f %v", balance*exchangeRate, currency))
	} else {
		exchangeRate := ExchangeRateForCurrency(currency)
		balance := FormatFloat((float64(balanceInt)/float64(1e9))*float64(exchangeRate), 2)

		return template.HTML(fmt.Sprintf(`%s %v`, balance, currency))
	}
}

// FormatDepositAmount will return the deposit amount formated as string
func FormatDepositAmount(balanceInt uint64, currency string) template.HTML {
	exchangeRate := ExchangeRateForCurrency(currency)
	balance := float64(balanceInt) / float64(1e9)
	return template.HTML(fmt.Sprintf("%.0f %v", balance*exchangeRate, currency))
}

// FormatEffectiveBalance will return the effective balance formated as string with 1 digit after the comma
func FormatEffectiveBalance(balanceInt uint64, currency string) template.HTML {
	exchangeRate := ExchangeRateForCurrency(currency)
	balance := float64(balanceInt) / float64(1e9)
	return template.HTML(fmt.Sprintf("%.1f %v", balance*exchangeRate, currency))
}

// FormatEpoch will return the epoch formated as html
func FormatEpoch(epoch uint64) template.HTML {
	return template.HTML(fmt.Sprintf("<a href=\"/epoch/%d\">%s</a>", epoch, FormatAddCommas(epoch)))
}

// FormatEth1AddressString will return the eth1-address formated as html string
func FormatEth1AddressString(addr []byte) template.HTML {
	eth1Addr := eth1common.BytesToAddress(addr)
	return template.HTML(fmt.Sprintf("%s", eth1Addr.Hex()))
}

// FormatEth1AddressString will return the eth1-address formated as html string
func FormatEth1AddressStringLowerCase(addr []byte) template.HTML {
	return template.HTML(fmt.Sprintf("0x%x", addr))
}

// FormatEth1Address will return the eth1-address formated as html
func FormatEth1Address(addr []byte) template.HTML {
	copyBtn := CopyButton(hex.EncodeToString(addr))
	eth1Addr := eth1common.BytesToAddress(addr)

	if Config.Chain.Config.ConfigName == "prater" {
		return template.HTML(fmt.Sprintf("<a href=\"https://goerli.etherscan.io/address/0x%x\" class=\"text-monospace\">%s…</a>%s", addr, eth1Addr.Hex()[:8], copyBtn))
	}
	if Config.Chain.Config.ConfigName == "ropsten" {
		return template.HTML(fmt.Sprintf("<a href=\"https://ropsten.etherscan.io/address/0x%x\" class=\"text-monospace\">%s…</a>%s", addr, eth1Addr.Hex()[:8], copyBtn))
	}
	if Config.Chain.Config.ConfigName == "sepolia" {
		return template.HTML(fmt.Sprintf("<a href=\"https://sepolia.etherscan.io/address/0x%x\" class=\"text-monospace\">%s…</a>%s", addr, eth1Addr.Hex()[:8], copyBtn))
	}
	return template.HTML(fmt.Sprintf("<a href=\"https://etherchain.org/account/0x%x\" class=\"text-monospace\">%s…</a>%s", addr, eth1Addr.Hex()[:8], copyBtn))
}

// FormatEth1Block will return the eth1-block formated as html
func FormatEth1Block(block uint64) template.HTML {
	if Config.Chain.Config.ConfigName == "prater" {
		return template.HTML(fmt.Sprintf("<a href=\"https://goerli.etherscan.io/block/%[1]d\">%[1]d</a>", block))
	}
	if Config.Chain.Config.ConfigName == "ropsten" {
		return template.HTML(fmt.Sprintf("<a href=\"https://ropsten.etherscan.io/block/%[1]d\">%[1]d</a>", block))
	}
	if Config.Chain.Config.ConfigName == "sepolia" {
		return template.HTML(fmt.Sprintf("<a href=\"https://sepolia.etherscan.io/block/%[1]d\">%[1]d</a>", block))
	}
	return template.HTML(fmt.Sprintf("<a href=\"https://etherchain.org/block/%[1]d\">%[1]d</a>", block))
}

// FormatEth1BlockHash will return the eth1-block formated as html
func FormatEth1BlockHash(block []byte) template.HTML {
	copyBtn := CopyButton(hex.EncodeToString(block))
	if Config.Chain.Config.ConfigName == "prater" {
		return template.HTML(fmt.Sprintf("<a href=\"https://goerli.etherscan.io/block/%#[1]x\">%#[1]x</a>%s", block, copyBtn))
	}
	if Config.Chain.Config.ConfigName == "ropsten" {
		return template.HTML(fmt.Sprintf("<a href=\"https://ropsten.etherscan.io/block/%#[1]x\">%#[1]x</a>%s", block, copyBtn))
	}
	if Config.Chain.Config.ConfigName == "sepolia" {
		return template.HTML(fmt.Sprintf("<a href=\"https://sepolia.etherscan.io/block/%#[1]x\">%#[1]x</a>%s", block, copyBtn))
	}
	return template.HTML(fmt.Sprintf("<a href=\"https://etherchain.org/block/%#[1]x\">%[1]x</a>%s", block, copyBtn))
}

// FormatEth1TxHash will return the eth1-tx-hash formated as html
func FormatEth1TxHash(hash []byte) template.HTML {
	copyBtn := CopyButton(hex.EncodeToString(hash))

	if Config.Chain.Config.ConfigName == "prater" {
		return template.HTML(fmt.Sprintf(`<i class="fas fa-male mr-2"></i><a style="font-family: 'Roboto Mono'" href="https://goerli.etherscan.io/tx/0x%x">0x%v…</a>%v`, hash, hex.EncodeToString(hash)[:6], copyBtn))
	}
	if Config.Chain.Config.ConfigName == "ropsten" {
		return template.HTML(fmt.Sprintf(`<i class="fas fa-male mr-2"></i><a style="font-family: 'Roboto Mono'" href="https://ropsten.etherscan.io/tx/0x%x">0x%v…</a>%v`, hash, hex.EncodeToString(hash)[:6], copyBtn))
	}
	return template.HTML(fmt.Sprintf(`<i class="fas fa-male mr-2"></i><a style="font-family: 'Roboto Mono'" href="https://etherchain.org/tx/0x%x">0x%v…</a>%v`, hash, hex.EncodeToString(hash)[:6], copyBtn))
}

// FormatGlobalParticipationRate will return the global-participation-rate formated as html
func FormatGlobalParticipationRate(e uint64, r float64, currency string) template.HTML {
	p := message.NewPrinter(language.English)
	rr := fmt.Sprintf("%.2f%%", r*100)
	tpl := `
	<div style="position:relative;width:inherit;height:inherit;">
	  %.0[1]f <small class="text-muted ml-3">(%[2]v)</small>
	  <div class="progress" style="position:absolute;bottom:-6px;width:100%%;height:4px;">
		<div class="progress-bar" role="progressbar" style="width: %[2]v;" aria-valuenow="%[2]v" aria-valuemin="0" aria-valuemax="100"></div>
	  </div>
	</div>`
	return template.HTML(p.Sprintf(tpl, float64(e)/1e9*price.GetEthPrice(currency), rr))
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
	return template.HTML(fmt.Sprintf("<span aria-graffiti=\"%#x\"><a href=\"/blocks?q=%s\">%s</a></span>", graffiti, u, h))
}

// FormatHash will return a hash formated as html
// hash is required, trunc is optional.
// Only the first value in trunc_opt will be used.
func FormatHash(hash []byte, trunc_opt ...bool) template.HTML {
	trunc := true
	if len(trunc_opt) > 0 {
		trunc = trunc_opt[0]
	}

	// if len(hash) > 6 {
	// 	return template.HTML(fmt.Sprintf("<span class=\"text-monospace\">0x%x…%x</span>", hash[:3], hash[len(hash)-3:]))
	// }
	// return template.HTML(fmt.Sprintf("<span class=\"text-monospace\">0x%x</span>", hash))
	if len(hash) > 3 && trunc {
		return template.HTML(fmt.Sprintf("<span class=\"text-monospace\">%#x…</span>", hash[:3]))
	}
	return template.HTML(fmt.Sprintf("<span class=\"text-monospace\">%#x</span>", hash))
}

func CopyButton(clipboardText interface{}) string {
	return fmt.Sprintf(`<i class="fa fa-copy text-muted ml-2 p-1" role="button" data-toggle="tooltip" title="Copy to clipboard" data-clipboard-text=0x%v></i>`, clipboardText)
}

func CopyButtonWithTitle(clipboardText interface{}, title string) string {
	return fmt.Sprintf(`<i class="fa fa-copy text-muted ml-2 p-1" role="button" data-toggle="tooltip" title="%v" data-clipboard-text=0x%v></i>`, title, clipboardText)
}

func Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func FormatBitvector(b []byte) template.HTML {
	return formatBits(b, false)
}

func FormatBitlist(b []byte) template.HTML {
	return formatBits(b, true)
}

func formatBits(b []byte, hasDelimiterBit bool) template.HTML {
	var buf strings.Builder
	buf.WriteString("<div class=\"text-bitfield\">")
	perLine := 8
	for y := 0; y < len(b); y += perLine {
		start, end := y, y+perLine
		if end >= len(b) {
			end = len(b)
		}
		for x := start; x < end; x++ {
			v := b[x]
			var vStr string
			// if it has a delimiter, and this is the last byte, strip the delimiter bit, and omit the trailing zeroes
			if hasDelimiterBit && x == end-1 {
				i := bitfields.BitIndex(v)
				// mask out the delimit bit
				v &^= 1 << i
				// no padding. Reverse to put most significant bit on the right
				vStr = Reverse(fmt.Sprintf("%b", v))
			} else {
				// pad to full 8 bits, don't omit zeroes. Reverse to put most significant bit on the right
				vStr = Reverse(fmt.Sprintf("%08b", v))
			}
			buf.WriteString(fmt.Sprintf(" <span class=\"text-monospace\">%s</span>", vStr))
		}
		buf.WriteString("<br/>")
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
			buf.WriteString(fmt.Sprintf("\n"))
		} else if (i+1)%8 == 0 {
			buf.WriteString(fmt.Sprintf(" "))
		}
	}
	buf.WriteString("</pre>")
	return template.HTML(buf.String())
}

func FormatParticipation(v float64) template.HTML {
	return template.HTML(fmt.Sprintf("<span>%.2f %%</span>", v*100.0))
}

// FormatIncome will return a string for a balance
func FormatIncome(balanceInt int64, currency string) template.HTML {

	decimals := 2

	if currency == "ETH" {
		decimals = 5
	}

	exchangeRate := ExchangeRateForCurrency(currency)
	balance := (float64(balanceInt) / float64(1e9)) * float64(exchangeRate)
	balanceFormated := FormatFloat(balance, decimals)

	if balance > 0 {
		return template.HTML(fmt.Sprintf(`<span class="text-success"><b>+%s %v</b></span>`, balanceFormated, currency))
	} else if balance < 0 {
		return template.HTML(fmt.Sprintf(`<span class="text-danger"><b>%s %v</b></span>`, balanceFormated, currency))
	} else {
		return template.HTML(fmt.Sprintf(`<b>%s %v</b>`, balanceFormated, currency))
	}
}

func FormatIncomeSql(balanceInt sql.NullInt64, currency string) template.HTML {

	if !balanceInt.Valid {
		return template.HTML(fmt.Sprintf(`<b>0 %v</b>`, currency))
	}

	exchangeRate := ExchangeRateForCurrency(currency)
	balance := float64(balanceInt.Int64) / float64(1e9)

	if balance > 0 {
		return template.HTML(fmt.Sprintf(`<span class="text-success"><b>+%.4f %v</b></span>`, balance*exchangeRate, currency))
	} else if balance < 0 {
		return template.HTML(fmt.Sprintf(`<span class="text-danger"><b>%.4f %v</b></span>`, balance*exchangeRate, currency))
	} else {
		return template.HTML(fmt.Sprintf(`<b>%.4f %v</b>`, balance*exchangeRate, currency))
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
	return template.HTML(fmt.Sprintf("<span class=\"timestamp\" title=\"%v\" data-toggle=\"tooltip\" data-placement=\"top\" data-timestamp=\"%d\"></span>", time.Unix(ts, 0), ts))
}

// FormatTs will return a timestamp formated as html. This is supposed to be used together with client-side js
func FormatTsWithoutTooltip(ts int64) template.HTML {
	return template.HTML(fmt.Sprintf("<span class=\"timestamp\" data-timestamp=\"%d\"></span>", ts))
}

// FormatTimestamp will return a timestamp formated as html. This is supposed to be used together with client-side js
func FormatTimestampTs(ts time.Time) template.HTML {
	return template.HTML(fmt.Sprintf("<span class=\"timestamp\" title=\"%v\" data-timestamp=\"%d\"></span>", ts, ts.Unix()))
}

// FormatValidatorStatus will return the validator-status formated as html
// possible states
// pending, active_online, active_offline, exiting_online, exciting_offline, slashing_online, slashing_offline, exited, slashed
func FormatValidatorStatus(status string) template.HTML {
	if status == "deposited" || status == "deposited_valid" || status == "deposited_invalid" {
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

func FormatEth1AddressWithName(address []byte, name string) template.HTML {
	eth1Addr := eth1common.BytesToAddress(address)
	if name != "" {
		return template.HTML(fmt.Sprintf("<a href=\"https://etherchain.org/account/0x%x\" class=\"text-monospace\">%s</a>", eth1Addr, name))
	} else {
		return FormatEth1Address(address)
	}
}

// FormatValidatorInt64 will return html formatted text for a validator (for an int64 validator-id)
func FormatValidatorInt64(validator int64) template.HTML {
	return FormatValidator(uint64(validator))
}

// FormatValidatosrInt64 will return html formatted text for validators
func FormatValidatorsInt64(validators []int64) template.HTML {
	formatedValidators := make([]string, len(validators))
	for i, v := range validators {
		formatedValidators[i] = string(FormatValidatorInt64(v))
	}
	return template.HTML(strings.Join(formatedValidators, " "))
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

// FormatSlashedValidatorsInt64 will return html formatted text for slashed validators
func FormatSlashedValidatorsInt64(validators []int64) template.HTML {
	str := ""
	for i, v := range validators {
		if i == len(validators)+1 {
			str += fmt.Sprintf("<i class=\"fas fa-user-slash text-danger mr-2\"></i><a href=\"/validator/%v\">%v</a>", v, v)
		} else {
			str += fmt.Sprintf("<i class=\"fas fa-user-slash text-danger mr-2\"></i><a href=\"/validator/%v\">%v</a>, ", v, v)
		}
	}
	return template.HTML(str)
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
		return ""
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

func FormatPercentageColored(percentage float64) template.HTML {
	if math.IsInf(percentage, 0) || math.IsNaN(percentage) {
		percentage = 0
	} else {
		percentage = percentage * 100
	}
	if percentage == 100 {
		return template.HTML(fmt.Sprintf(`<span class="text-success">%.0f%%</span>`, percentage))
	} else if percentage >= 90 {
		return template.HTML(fmt.Sprintf(`<span class="text-success">%.0f%%</span>`, percentage))
	} else if percentage >= 80 {
		return template.HTML(fmt.Sprintf(`<span class="text-warning">%.0f%%</span>`, percentage))
	} else if percentage >= 60 {
		return template.HTML(fmt.Sprintf(`<span class="text-warning">%.0f%% </span>`, percentage))
	}
	return template.HTML(fmt.Sprintf(`<span class="text-danger">%.0f%%</span>`, percentage))
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

func KFormatterEthPrice(price uint64) string {
	if price > 999 {
		ethTruncPrice := fmt.Sprint(float64(int((float64(price)/float64(1000))*10))/float64(10)) + "k"
		return ethTruncPrice
	}
	return fmt.Sprint(price)
}

func FormatRPL(num string) string {
	floatNum, _ := strconv.ParseFloat(num, 64)
	return fmt.Sprintf("%.2f", floatNum/math.Pow10(18)) + " RPL"
}

func FormatFloat(num float64, precision int) string {
	p := message.NewPrinter(language.English)
	f := fmt.Sprintf("%%.%vf", precision)
	s := strings.TrimRight(strings.TrimRight(p.Sprintf(f, num), "0"), ".")
	r := []rune(p.Sprintf(s, num))
	return string(r)
}

func FormatNotificationChannel(ch types.NotificationChannel) string {
	label, ok := types.NotificationChannelLabels[ch]
	if !ok {
		return ""
	}
	return label
}
