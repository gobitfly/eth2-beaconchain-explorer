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
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/prysmaticlabs/go-bitfield"
	"github.com/shopspring/decimal"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	itypes "github.com/gobitfly/eth-rewards/types"
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
		return `<span class="badge badge-pill bg-warning text-white" style="font-size: 12px; font-weight: 500;">Missed (Orphaned)</span>`
	} else {
		return "Unknown"
	}
}

// FormatAttestationStatusShort will return a user-friendly attestation for an attestation status number
func FormatAttestationStatusShort(status uint64) template.HTML {
	if status == 0 {
		return `<span title="Scheduled" data-toggle="tooltip" class="mx-1 badge badge-pill bg-light text-dark" style="font-size: 12px; font-weight: 500;">Sche.</span>`
	} else if status == 1 {
		return `<span title="Attested" data-toggle="tooltip" class="mx-1 badge badge-pill bg-success text-white" style="font-size: 12px; font-weight: 500;">Att.</span>`
	} else if status == 2 {
		return `<span title="Missed" data-toggle="tooltip" class="mx-1 badge badge-pill bg-warning text-white" style="font-size: 12px; font-weight: 500;">Miss.</span>`
	} else if status == 3 {
		return `<span title="Missed (Orphaned)" data-toggle="tooltip" class="mx-1 badge badge-pill bg-warning text-white" style="font-size: 12px; font-weight: 500;">Orph.</span>`
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
	exchangeRate := ExchangeRateForCurrency(currency)
	balance := FormatFloat((float64(balanceInt)/float64(1e9))*float64(exchangeRate), 2)

	return template.HTML(balance + " " + currency)
}

// FormatBalance will return a string for a balance
func FormatEligibleBalance(balanceInt uint64, currency string) template.HTML {
	if balanceInt == 0 {
		return `<span class="text-small text-muted">Calculating...</span>`
	}
	exchangeRate := ExchangeRateForCurrency(currency)
	balance := FormatFloat((float64(balanceInt)/float64(1e9))*float64(exchangeRate), 2)

	return template.HTML(balance)
}

func FormatBalanceSql(balanceInt sql.NullInt64, currency string) template.HTML {
	if !balanceInt.Valid {
		return template.HTML("0 " + currency)
	}
	exchangeRate := ExchangeRateForCurrency(currency)
	balance := FormatFloat((float64(balanceInt.Int64)/float64(1e9))*float64(exchangeRate), 5)

	return template.HTML(balance + " " + currency)
}

func FormatBalanceGwei(balance *int64, currency string) template.HTML {
	if currency == "ETH" {
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

func FormatBalanceChangeFormated(balance *int64, currencyName string, details *itypes.ValidatorEpochIncome) template.HTML {

	income := ""
	if details != nil {

		income += fmt.Sprintf("Att. Source: %s GWei<br/>", FormatAddCommasFormated(float64(int64(details.AttestationSourceReward)-int64(details.AttestationSourcePenalty)), 0))
		income += fmt.Sprintf("Att. Target: %s GWei<br/>", FormatAddCommasFormated(float64(int64(details.AttestationTargetReward)-int64(details.AttestationTargetPenalty)), 0))
		income += fmt.Sprintf("Att. Head Vote: %s GWei<br/>", FormatAddCommasFormated(float64(details.AttestationHeadReward), 0))

		if details.FinalityDelayPenalty > 0 {
			income += fmt.Sprintf("Finality Delay Penalty: %s GWei<br/>", FormatAddCommasFormated(float64(details.FinalityDelayPenalty)*-1, 0))
		}

		if details.ProposerSlashingInclusionReward > 0 {
			income += fmt.Sprintf("Proposer Slashing Inc. Reward: %s GWei<br/>", FormatAddCommasFormated(float64(details.ProposerSlashingInclusionReward), 0))
		}

		if details.ProposerAttestationInclusionReward > 0 {
			income += fmt.Sprintf("Proposer Att. Inc. Reward: %s GWei<br/>", FormatAddCommasFormated(float64(details.ProposerAttestationInclusionReward), 0))
		}

		if details.ProposerSyncInclusionReward > 0 {
			income += fmt.Sprintf("Proposer Sync Inc. Reward: %s GWei<br/>", FormatAddCommasFormated(float64(details.ProposerSyncInclusionReward), 0))
		}

		if details.SyncCommitteeReward > 0 {
			income += fmt.Sprintf("Sync Comm. Reward: %s GWei<br/>", FormatAddCommasFormated(float64(details.SyncCommitteeReward), 0))
		}

		if details.SyncCommitteePenalty > 0 {
			income += fmt.Sprintf("Sync Comm. Penalty: %s GWei<br/>", FormatAddCommasFormated(float64(details.SyncCommitteePenalty)*-1, 0))
		}

		if details.SlashingReward > 0 {
			income += fmt.Sprintf("Slashing Reward: %s GWei<br/>", FormatAddCommasFormated(float64(details.SlashingReward), 0))
		}

		if details.SlashingPenalty > 0 {
			income += fmt.Sprintf("Slashing Penalty: %s GWei<br/>", FormatAddCommasFormated(float64(details.SlashingPenalty)*-1, 0))
		}

		income += fmt.Sprintf("Total: %s GWei", FormatAddCommasFormated(float64(details.TotalClRewards()), 0))
	}

	if currencyName == "ETH" {
		if balance == nil || *balance == 0 {
			return template.HTML("<span>0 GWei</span>")
		}
		if *balance < 0 {
			return template.HTML(fmt.Sprintf("<span title='%s' data-html=\"true\" data-toggle=\"tooltip\" class=\"text-danger\">%s GWei</span>", income, FormatAddCommasFormated(float64(*balance), 0)))
		}
		return template.HTML(fmt.Sprintf("<span title='%s' data-html=\"true\" data-toggle=\"tooltip\" class=\"text-success\">+%s GWei</span>", income, FormatAddCommasFormated(float64(*balance), 0)))
	} else {
		if balance == nil {
			return template.HTML("<span>0 " + currencyName + "</span>")
		}
		if *balance == 0 {
			return template.HTML("pending")
		}

		balanceF := float64(*balance) / float64(1e9)
		exchangeRate := ExchangeRateForCurrency(currencyName)
		value := balanceF * float64(exchangeRate)

		if *balance < 0 {
			return template.HTML(fmt.Sprintf("<span class=\"text-danger float-right\" data-toggle=\"tooltip\" data-placement=\"top\" title=\"%f\">%s %s</span>", value, FormatAddCommasFormated(value, 2), currencyName))
		}
		return template.HTML(fmt.Sprintf("<span class=\"text-success float-right\" data-toggle=\"tooltip\" data-placement=\"top\" title=\"%f\">+%s %s</span>", value, FormatAddCommasFormated(value, 2), currencyName))
	}
}

// FormatBalanceChange will return a string for a balance change
func FormatBalanceChange(balance *int64, currency string) template.HTML {
	balanceF := float64(*balance) / float64(1e9)
	if currency == "ETH" {
		if balance == nil || *balance == 0 {
			return template.HTML("<span> 0 " + currency + "</span>")
		}

		if balanceF < 0 {
			return template.HTML(fmt.Sprintf("<span class=\"text-danger float-right\">%s GWei</span>", FormatAddCommasFormated(float64(*balance), 0)))
		}
		return template.HTML(fmt.Sprintf("<span class=\"text-success float-right\">+%s GWei</span>", FormatAddCommasFormated(float64(*balance), 0)))
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

func FormatAddCommasFormated(num float64, precision uint) template.HTML {
	p := message.NewPrinter(language.English)
	s := p.Sprintf(fmt.Sprintf("%%.%vf", precision), num)
	if precision > 0 {
		s = strings.TrimRight(strings.TrimRight(s, "0"), ".")
	}
	return template.HTML(strings.ReplaceAll(string([]rune(p.Sprintf(s, num))), ",", `<span class="thousands-separator"></span>`))
}

func FormatBigNumberAddCommasFormated(val hexutil.Big, precision uint) template.HTML {
	return FormatAddCommasFormated(float64(val.ToInt().Int64()), 0)
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
func FormatBlockStatus(status uint64) template.HTML {
	// genesis <span class="badge text-dark" style="background: rgba(179, 159, 70, 0.8) none repeat scroll 0% 0%;">Genesis</span>
	if status == 0 {
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
func FormatBlockStatusShort(status uint64) template.HTML {
	// genesis <span class="badge text-dark" style="background: rgba(179, 159, 70, 0.8) none repeat scroll 0% 0%;">Genesis</span>
	if status == 0 {
		return `<span title="Scheduled" data-toggle="tooltip" class="mx-1 badge badge-pill bg-light text-dark" style="font-size: 12px; font-weight: 500;">Sche.</span>`
	} else if status == 1 {
		return `<span title="Proposed" data-toggle="tooltip" class="mx-1 badge badge-pill bg-success text-white" style="font-size: 12px; font-weight: 500;">Prop.</span>`
	} else if status == 2 {
		return `<span title="Missed" data-toggle="tooltip" class="mx-1 badge badge-pill bg-warning text-white" style="font-size: 12px; font-weight: 500;">Miss.</span>`
	} else if status == 3 {
		return `<span title="Missed (Orphaned)" data-toggle="tooltip" class="mx-1 badge badge-pill bg-secondary text-white" style="font-size: 12px; font-weight: 500;">Orph.</span>`
	} else {
		return "Unknown"
	}
}

func FormatBlockStatusStyle(status uint64) template.HTML {
	if status == 0 {
		return `<div title="This slot is still scheduled" data-toggle="tooltip" class="style-badge style-badge-single-char style-bg-neutral-1"><div class="style-status-tag-text" >S<span class="d-none d-sm-inline">cheduled</span></div></div>`
	} else if status == 1 {
		return `<div title="This block has been proposed" data-toggle="tooltip" class="style-badge style-badge-single-char style-bg-good"><div class="style-status-tag-text text-white" >P<span class="d-none d-sm-inline">roposed</span></div></div>`
	} else if status == 2 {
		return `<div title="This block proposal has been missed" data-toggle="tooltip" class="style-badge style-badge-single-char style-bg-bad"><div class="style-status-tag-text" >M<span class="d-none d-sm-inline">issed</span></div></div>`
	} else if status == 3 {
		return `<div title="This block has been orphaned" data-toggle="tooltip" class="style-badge style-badge-single-char style-bg-neutral-2"><div class="style-status-tag-text text-white" >O<span class="d-none d-sm-inline">rphaned</span></div></div>`
	} else {
		return `<div title="This shouldn't be possible" data-toggle="tooltip" class="style-badge style-badge-single-char style-bg-neutral-2"><div class="style-status-tag-text text-white" >?</div></div>`
	}
}

func FormatEpochStatus(finalized bool, participationRate float64) template.HTML {
	if finalized {
		return `<div title="This epoch is finalized" data-toggle="tooltip" class="style-badge style-bg-good"><div class="style-status-tag-text text-white" ><span class="d-sm-none">FIN</span><span class="d-none d-sm-inline">Finalized</span></div></div>`
	} else if participationRate > 0.66 && participationRate < 1 {
		// since the latest epoch in the db always has a participation rate of 1, check for < 1 instead of <= 1
		return `<div title="This epoch is not finalized but safe, making a revert unlikely" data-toggle="tooltip" class="style-badge style-badge-long style-bg-neutral-2"><div class="style-status-tag-text text-white" ><span class="d-sm-none">NFS</span><span class="d-none d-sm-inline">Not finalized (Safe)</span></div></div>`
	} else {
		return `<div title="This epoch is not finalized" data-toggle="tooltip" class="style-badge style-bg-neutral-1"><div class="style-status-tag-text" ><span class="d-sm-none">NF</span><span class="d-none d-sm-inline">Not finalized</span></div></div>`
	}
}

// FormatBlockStatusShort will return an html status for a block.
func FormatWithdrawalShort(slot uint64, amount uint64) template.HTML {
	return template.HTML(fmt.Sprintf("<span title=\"Withdrawal processed in epoch %v during slot %v for %v\" data-toggle=\"tooltip\" class=\"mx-1 badge badge-pill bg-success text-white\" style=\"font-size: 12px; font-weight: 500;\"><i class=\"fas fa-money-bill\"></i></span>", EpochOfSlot(slot), slot, FormatCurrentBalance(amount, "ETH")))
}

func FormatTransactionType(txnType uint8) string {
	switch txnType {
	case 0:
		return "0 (legacy)"
	case 1:
		return "1 (Access-list)"
	case 2:
		return "2 (EIP-1559)"
	default:
		return fmt.Sprintf("%v (???)", txnType)
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
	return template.HTML(fmt.Sprintf("%.0f %v", balance*exchangeRate, currency))
}

// FormatEpoch will return the epoch formated as html
func FormatEpoch(epoch uint64) template.HTML {
	return template.HTML(fmt.Sprintf("<a href=\"/epoch/%d\">%s</a>", epoch, FormatAddCommas(epoch)))
}

// FormatEth1AddressString will return the eth1-address formated as html string
func FormatEth1AddressString(addr []byte) template.HTML {
	eth1Addr := common.BytesToAddress(addr)
	return template.HTML(eth1Addr.Hex())
}

// FormatEth1AddressString will return the eth1-address formated as html string
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

func FormatGlobalParticipationRate(voted uint64, participationRate float64, currency string) template.HTML {
	if voted == 0 {
		return `<span>Calculating...</span>`
	}
	p := message.NewPrinter(language.English)
	rr := fmt.Sprintf("%.2f%%", participationRate*100)
	tpl := `
	<div style="position:relative;width:inherit;height:inherit;">
		<span>%.0[1]f</span><span class="style-paragraph-3 ml-3">(%[2]v)
	  <div class="progress" style="width:100%%;height:4px;">
		<div class="progress-bar" role="progressbar" style="width: %[2]v;" aria-valuenow="%[2]v" aria-valuemin="0" aria-valuemax="100"></div>
	  </div>
	</div>`
	return template.HTML(p.Sprintf(tpl, float64(voted)/1e9*price.GetEthPrice(currency), rr))
}

func FormatEtherValue(symbol string, ethPrice *big.Float, currentPrice template.HTML) template.HTML {
	p := message.NewPrinter(language.English)
	ep, _ := ethPrice.Float64()
	return template.HTML(p.Sprintf(`<span>%s %.2f</span> <span class="text-muted">@ %s/ETH</span>`, symbol, ep, currentPrice))
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

// FormatHash will return a hash formated as html
// hash is required, trunc is optional.
// Only the first value in trunc_opt will be used.
func FormatHash(hash []byte, trunc_opt ...bool) template.HTML {
	trunc := true
	if len(trunc_opt) > 0 {
		trunc = trunc_opt[0]
	}

	// return template.HTML(fmt.Sprintf("<span class=\"text-monospace\">0x%x</span>", hash))
	if len(hash) > 3 && trunc {
		return template.HTML(fmt.Sprintf("<span class=\"style-hash\">%#x…%x</span>", hash[:2], hash[len(hash)-2:]))
	}
	return template.HTML(fmt.Sprintf("<span class=\"style-hash\">%#x</span>", hash))
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

func FormatName(name string, trunc_opt ...bool) template.HTML {
	trunc := true
	if len(trunc_opt) > 0 {
		trunc = trunc_opt[0]
	}

	// return template.HTML(fmt.Sprintf("<span class=\"text-monospace\">0x%x</span>", hash))
	if len(name) > 8 && trunc {
		return template.HTML(fmt.Sprintf("<span class=\"text-monospace\">%s…</span>", name[:8]))
	}
	return template.HTML(fmt.Sprintf("<span class=\"text-monospace\">%s</span>", name))
}

func AddCopyButton(element template.HTML, copyContent string) template.HTML {
	return template.HTML(fmt.Sprintf(`<span title="%s" data-toggle="tooltip">%v<span>`, copyContent, element)) + " " + template.HTML(CopyButton(copyContent))
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

func CopyButtonWithTitle(clipboardText interface{}, title string) string {
	value := fmt.Sprintf("%v", clipboardText)
	if len(value) < 2 || value[0] != '0' || value[1] != 'x' {
		value = "0x" + value
	}
	return fmt.Sprintf(`<i class="fa fa-copy text-muted ml-2 p-1" role="button" data-toggle="tooltip" title="%v" data-clipboard-text=%s></i>`, title, value)
}

func Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func FormatBitvector(b []byte) template.HTML {
	return formatBits(b, len(b)*8)
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

func FormatIncomeClElInt64(income types.ClElInt64, currency string) template.HTML {
	var incomeTrimmed string = exchangeAndTrim(currency, income.Total)
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
		</span>`, className, FormatExchangedAmount(income.Cl, currency), FormatExchangedAmount(income.El, currency), incomeTrimmed, currency))
	} else {
		return template.HTML(fmt.Sprintf(`<span>%s%s</span>`, incomeTrimmed, currency))
	}
}

// FormatIncome will return a string for a balance
func FormatIncome(balanceInt int64, currency string) template.HTML {
	return formatIncome(balanceInt, currency, true)
}

func FormatIncomeNoCurrency(balanceInt int64, currency string) template.HTML {
	return formatIncome(balanceInt, currency, false)
}

func formatIncome(balanceInt int64, currency string, includeCurrency bool) template.HTML {
	var income string = exchangeAndTrim(currency, balanceInt)

	if includeCurrency {
		currency = " " + currency
	} else {
		currency = ""
	}

	if balanceInt > 0 {
		return template.HTML(fmt.Sprintf(`<span class="text-success"><b>%s%s</b></span>`, income, currency))
	} else if balanceInt < 0 {
		return template.HTML(fmt.Sprintf(`<span class="text-danger"><b>%s%s</b></span>`, income, currency))
	} else {
		return template.HTML(fmt.Sprintf(`<span>0%s</span>`, currency))
	}
}

func FormatExchangedAmount(balanceInt int64, currency string) template.HTML {
	income := exchangeAndTrim(currency, balanceInt)
	return template.HTML(fmt.Sprintf(`<span>%s %s</span>`, income, currency))
}

func exchangeAndTrim(currency string, amount int64) string {
	decimals := 5
	preCommaDecimals := 1

	if currency != "ETH" {
		decimals = 2
		preCommaDecimals = 4
	}

	exchangeRate := ExchangeRateForCurrency(currency)
	exchangedAmount := float64(amount) * exchangeRate
	// lost precision here but we don't need it for frontend
	income, _ := trimAmount(big.NewInt(int64(exchangedAmount)), 9, preCommaDecimals, decimals, true)
	return income
}

func FormatIncomeSql(balanceInt sql.NullInt64, currency string) template.HTML {

	if !balanceInt.Valid {
		return template.HTML(fmt.Sprintf(`<b>0 %v</b>`, currency))
	}

	exchangeRate := ExchangeRateForCurrency(currency)
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
		return "Deposited"
	} else if status == "pending" {
		return "Pending"
	} else if status == "active_online" {
		return `<span class="text-success">Active</span>`
	} else if status == "active_offline" {
		return `<span class="text-warning" data-toggle="tooltip" title="No attestation in the last 2 epochs">Active</span>`
	} else if status == "exiting_online" {
		return `Exiting`
	} else if status == "exiting_offline" {
		return `<span class="text-danger" data-toggle="tooltip" title="No attestation in the last 2 epochs">Exiting</span>`
	} else if status == "slashing_online" {
		return `<span class="text-danger">Slashing</span>`
	} else if status == "slashing_offline" {
		return `<span class="text-danger" data-toggle="tooltip" title="No attestation in the last 2 epochs">Slashing</span>`
	} else if status == "exited" {
		return "Exited"
	} else if status == "slashed" {
		return `<span class="text-danger">Slashed</span>`
	}
	return "Unknown"
}

// FormatValidatorTag will return html formated text of a validator-tag.
// Depending on the tag it will describe the tag in a tooltip and link to more information regarding the tag.
func formatValidatorTag(tag string) string {
	var result string
	switch {
	case tag == "rocketpool", tag == "pool:rocketpool":
		result = `<a href="/pools/rocketpool" title="Rocket Pool Validator" data-toggle="tooltip" class="mr-2 badge badge-pill style-bg-neutral-1 style-status-tag-text text-dark">Rocket Pool</a>`
	case strings.HasPrefix(tag, "pool:"):
		result = fmt.Sprintf(`<a href="/pools" title="This validator is part of a staking-pool." data-toggle="tooltip" class="mr-2 badge badge-pill style-bg-neutral-1 style-status-tag-text text-dark">Pool: %s</a>`, strings.TrimPrefix(tag, "pool:"))
	case strings.HasPrefix(tag, "name:"):
		result = fmt.Sprintf(`<span title="This name has been set by the owner of this validator." data-toggle="tooltip" class="mr-2 badge badge-pill style-bg-neutral-1 style-status-tag-text text-dark">%s</span>`, strings.TrimPrefix(tag, "name:"))
	case tag == "ssv":
		result = `<a href="https://github.com/bloxapp/ssv/" title="Secret Shared Validator" data-toggle="tooltip" class=""mr-2 badge badge-pill style-bg-neutral-1 style-status-tag-text text-dark">SSV</a>`
	default:
		result = fmt.Sprintf(`<span class="mr-2 badge badge-pill style-bg-neutral-1 style-status-tag-text text-dark">%s</span>`, tag)
	}

	return result
}

func FormatValidatorTags(tags []string) template.HTML {
	str := ""
	// some validators have duplicate tags, we dedup them
	dedupMap := make(map[string]string)
	for _, tag := range tags {
		// some tags have a number at the end, we want to remove that
		re := regexp.MustCompile(`(.+) - \d+`)
		match := re.FindStringSubmatch(tag)
		if len(match) > 1 {
			tag = match[1]
		}
		trimmedTag := tag
		trimmedTag = strings.TrimPrefix(trimmedTag, "pool:")
		trimmedTag = strings.ReplaceAll(trimmedTag, " ", "")
		trimmedTag = strings.ToLower(trimmedTag)
		dedupedTag, ok := dedupMap[trimmedTag]
		if !ok || tag < dedupedTag {
			dedupMap[trimmedTag] = tag
		}
	}
	for _, tag := range dedupMap {
		str += formatValidatorTag(tag) + " "
	}
	return template.HTML(str)
}

// FormatValidator will return html formatted text for a validator
func FormatValidator(validator uint64) template.HTML {
	return template.HTML(fmt.Sprintf("<a href=\"/validator/%v\">%v</a>", validator, validator))
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
		return template.HTML(fmt.Sprintf("<a href=\"/validator/%v\"><span class=\"text-truncate\">"+html.EscapeString(name)+"</span></a>", validatorLink))
	} else {
		return template.HTML(fmt.Sprintf("<a href=\"/validator/%v\">%v</a>", validatorLink, validatorRead))
	}
}

func FormatEth1AddressWithName(address []byte, name string) template.HTML {
	eth1Addr := common.BytesToAddress(address)
	if name != "" {
		return template.HTML(fmt.Sprintf("<a href=\"/address/0x%x\" class=\"text-monospace\">%s</a>", eth1Addr, name))
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
	return fmt.Sprintf("%.4f", floatNum/math.Pow10(18)) + " ETH"
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

func FormatBlockReward(blockNumber int64) template.HTML {
	var reward *big.Int

	if blockNumber < 4370000 {
		reward = big.NewInt(5e+18)
	} else if blockNumber < 7280000 {
		reward = big.NewInt(3e+18)
	} else {
		reward = big.NewInt(2e+18)
	}

	return FormatAmount(reward, "Ether", 5)
}

func FormatTokenBalance(balance *types.Eth1AddressBalance) template.HTML {
	mul := decimal.NewFromFloat(float64(10)).Pow(decimal.NewFromBigInt(new(big.Int).SetBytes(balance.Metadata.Decimals), 0))
	num := decimal.NewFromBigInt(new(big.Int).SetBytes(balance.Balance), 0)
	p := message.NewPrinter(language.English)

	priceS := string(balance.Metadata.Price)
	price := decimal.New(0, 0)
	if priceS != "" {
		var err error
		price, err = decimal.NewFromString(priceS)
		if err != nil {
			logger.WithError(err).Errorf("error getting price from string - FormatTokenBalance price: %v", priceS)
		}
	}
	// numPrice := num.Div(mul).Mul(price)

	logo := ""
	if len(balance.Metadata.Logo) != 0 {
		logo = fmt.Sprintf(`<img class="mr-1" style="height: 1.2rem;" src="data:image/png;base64, %s">`, base64.StdEncoding.EncodeToString(balance.Metadata.Logo))
	}
	symbolTitle := FormatTokenSymbolTitle(balance.Metadata.Symbol)
	symbol := FormatTokenSymbol(balance.Metadata.Symbol)
	pflt, _ := price.Float64()
	flt, _ := num.Div(mul).Round(5).Float64()
	bflt, _ := price.Mul(num.Div(mul)).Float64()
	return template.HTML(p.Sprintf(`
	<div class="token-balance-col token-name text-truncate d-flex align-items-center justify-content-between flex-wrap">
		<div class="token-icon p-1">
			<a href='/token/0x%x?a=0x%x'>
				<span>%s</span> <span title="%s">%s</span>
			</a> 
		</div>
		<div class="token-price-balance p-1">
			<span class="text-muted" style="font-size: 90%%;">$%.2f</span>
		</div>
	</div> 
	<div class="token-balance-col token-balance d-flex align-items-center justify-content-between flex-wrap">
		<div class="token-holdings p-1">
			<span class="token-holdings">%s</span>
		</div>
		<div class="token-price p-1">
			<span class="text-muted" style="font-size: 90%%;">@ $%.2f</span>
		</div>
	</div>`, balance.Token, balance.Address, logo, symbolTitle, symbol, bflt, FormatThousandsEnglish(strconv.FormatFloat(flt, 'f', -1, 64)), pflt))
}

func FormatAddressEthBalance(balance *types.Eth1AddressBalance) template.HTML {
	e := new(big.Int).SetBytes(balance.Metadata.Decimals)
	d := new(big.Int).Exp(big.NewInt(10), e, nil)
	balWei := new(big.Float).SetInt(new(big.Int).SetBytes(balance.Balance))
	balEth := new(big.Float).Quo(balWei, new(big.Float).SetInt(d))
	p := message.NewPrinter(language.English)
	return template.HTML(p.Sprintf(fmt.Sprintf(`
		<div class="d-flex align-items-center">
			<svg style="width: 1rem; height: 1rem;">
				<use xlink:href="#ethereum-diamond-logo"/>
			</svg> 
			<span class="token-holdings">%%.%df Ether</span>
		</div>`, e.Int64()), balEth))
}

func FormatTokenValue(balance *types.Eth1AddressBalance) template.HTML {
	decimals := new(big.Int).SetBytes(balance.Metadata.Decimals)
	p := message.NewPrinter(language.English)
	mul := decimal.NewFromFloat(float64(10)).Pow(decimal.NewFromBigInt(decimals, 0))
	num := decimal.NewFromBigInt(new(big.Int).SetBytes(balance.Balance), 0)
	f, _ := num.DivRound(mul, int32(decimals.Int64())).Float64()

	return template.HTML(p.Sprintf("%s", FormatThousandsEnglish(strconv.FormatFloat(f, 'f', -1, 64))))
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

// FormatTimestamp will return a timestamp formated as html. This is supposed to be used together with client-side js
func FormatTimestampUInt64(ts uint64) template.HTML {
	return template.HTML(fmt.Sprintf("<span class=\"timestamp\" title=\"%v\" data-toggle=\"tooltip\" data-placement=\"top\" data-timestamp=\"%d\"></span>", time.Unix(int64(ts), 0), ts))
}

// FormatEth1AddressFull will return the eth1-address formated as html
func FormatEth1AddressFull(addr common.Address) template.HTML {
	return FormatAddress(addr.Bytes(), nil, "", false, false, true)
}

func FormatHeaderHash(address []byte) template.HTML {
	if l := len(address) * 2; l < 8 {
		return template.HTML(fmt.Sprintf("0x%x", address))
	}
	return template.HTML(fmt.Sprintf(`
	<h2 class="overflow-auto text-nowrap style-header-account mb-0">
		0x<span style="color: var(--primary)">%x</span><span data-truncate-middle="%x"></span><span style="color: var(--primary)">%x</span>
	</h2>`, address[:2], address[2:len(address)-2], address[len(address)-2:]))
}

func FormatValidatorHistoryEvent(event types.ValidatorHistoryEvent) template.HTML {
	colorMap := map[uint64]string{
		0: "var(--text-addition)",
		1: "var(--green)",
		2: "var(--red)",
		3: "#f7a53c",
	}
	str := `
		<div class="style-history-event">
			<svg viewBox="0 0 14 12" xmlns="http://www.w3.org/2000/svg" fill="%s">
				<path d="M6.22302 0V3H9.33467L6.22302 0Z" />
				<path d="M7.6746 10.4625C7.60167 10.4789 7.52874 10.4906 7.45581 10.4953C7.43393 10.4977 7.41205 10.5 7.39017 10.5H5.83435C5.68606 10.5 5.55235 10.4203 5.48672 10.2937L5.27279 9.87891C5.23146 9.79922 5.14881 9.75 5.05886 9.75C4.96892 9.75 4.88383 9.79922 4.84494 9.87891L4.63101 10.2937C4.56051 10.432 4.40736 10.5141 4.24935 10.5C4.09134 10.4859 3.9552 10.3805 3.91144 10.2352L3.50061 8.92969L3.26237 9.69844C3.11408 10.1742 2.65949 10.5 2.14412 10.5H1.94478C1.73086 10.5 1.55583 10.3313 1.55583 10.125C1.55583 9.91875 1.73086 9.75 1.94478 9.75H2.14412C2.31672 9.75 2.46744 9.64219 2.51606 9.48281L2.87828 8.32266C2.96093 8.05781 3.21375 7.875 3.50061 7.875C3.78746 7.875 4.04029 8.05781 4.12294 8.32266L4.40493 9.22734C4.58482 9.08203 4.81334 9 5.05643 9C5.44296 9 5.79545 9.21094 5.96805 9.54375L6.07501 9.75H6.29137C6.21601 9.54375 6.20142 9.31875 6.25734 9.09844L6.62198 7.68984C6.69005 7.425 6.83105 7.18594 7.03039 6.99375L9.33495 4.77187V3.75H6.2233C5.79302 3.75 5.44539 3.41484 5.44539 3V0H1.55583C0.697691 0 0 0.672656 0 1.5V10.5C0 11.3273 0.697691 12 1.55583 12H7.77913C8.63726 12 9.33495 11.3273 9.33495 10.5V10.0477C9.26932 10.0734 9.20368 10.0945 9.13561 10.1109L7.6746 10.4625ZM0.997916 5.67937C0.848411 5.47359 0.77378 5.23102 0.77378 4.95188V4.78031C0.77378 4.75758 0.782045 4.73812 0.798576 4.72195C0.815107 4.70578 0.835284 4.69805 0.859107 4.69805H1.49262C1.5162 4.69805 1.53638 4.70602 1.55315 4.72195C1.56968 4.73812 1.57795 4.75734 1.57795 4.78031V4.91063C1.57795 5.12109 1.67519 5.29617 1.86991 5.43562C2.06439 5.57531 2.33009 5.64492 2.66727 5.64492C2.95656 5.64492 3.1751 5.58562 3.32217 5.46656C3.46925 5.34773 3.54291 5.19445 3.54291 5.00672C3.54291 4.87875 3.50255 4.77 3.42184 4.6807C3.34114 4.59141 3.22615 4.51031 3.07665 4.43695C2.92714 4.36383 2.703 4.27219 2.40399 4.1625C2.06706 4.04367 1.79406 3.93164 1.58548 3.82617C1.37666 3.72094 1.20212 3.57797 1.06234 3.39726C0.922313 3.21656 0.852301 2.99133 0.852301 2.72133C0.852301 2.31867 1.00424 2.00062 1.30787 1.76742C1.6115 1.53398 2.01723 1.4175 2.52506 1.4175C2.88095 1.4175 3.19528 1.47703 3.46828 1.59586C3.74103 1.71492 3.9535 1.88062 4.10544 2.09344C4.25737 2.30625 4.33322 2.55211 4.33322 2.83125V2.94797C4.33322 2.97094 4.32495 2.99039 4.30842 3.00633C4.29165 3.0225 4.27171 3.03023 4.24789 3.03023H3.60733C3.5835 3.03023 3.56333 3.02227 3.5468 3.00633C3.53002 2.99039 3.522 2.97094 3.522 2.94797V2.8725C3.522 2.65758 3.4306 2.4757 3.24803 2.32688C3.06522 2.17828 2.81264 2.10375 2.48981 2.10375C2.22872 2.10375 2.02598 2.15648 1.88109 2.26148C1.7362 2.36672 1.664 2.51531 1.664 2.7075C1.664 2.84484 1.70193 2.95687 1.77802 3.04383C1.85386 3.13078 1.96788 3.20976 2.11981 3.28055C2.27175 3.35156 2.50658 3.44179 2.82455 3.55172C3.16149 3.67523 3.43084 3.7875 3.63237 3.88805C3.83389 3.98883 4.00722 4.12945 4.15211 4.31016C4.29675 4.49086 4.3692 4.71633 4.3692 4.98609C4.3692 5.39789 4.21143 5.72508 3.89588 5.96742C3.58034 6.21 3.14957 6.33117 2.60382 6.33117C2.23358 6.33117 1.91099 6.27398 1.6358 6.15961C1.36062 6.04523 1.14815 5.88515 0.998646 5.67914L0.997916 5.67937Z" />
				<path d="M13.7155 3.61172L13.3654 3.27422C12.9862 2.90859 12.3712 2.90859 11.9895 3.27422L11.2748 3.96328L13.0008 5.62734L13.7155 4.93828C14.0947 4.57266 14.0947 3.97969 13.7155 3.61172Z" />
				<path d="M7.58242 7.52363C7.48275 7.61973 7.41225 7.73926 7.37822 7.87285L7.01357 9.28145C6.97954 9.41035 7.01843 9.54394 7.11567 9.63769C7.21291 9.73144 7.35147 9.76895 7.48518 9.73613L8.9462 9.38457C9.08233 9.35176 9.20874 9.28379 9.30841 9.1877L12.4492 6.15723L10.7232 4.49316L7.58242 7.52363Z" />
			</svg>
			<svg viewBox="0 0 14 12" xmlns="http://www.w3.org/2000/svg" fill="%s">
				<path d="M7.6746 10.4625C7.60167 10.4789 7.52874 10.4906 7.45581 10.4953C7.43393 10.4977 7.41205 10.5 7.39017 10.5H5.83435C5.68606 10.5 5.55235 10.4203 5.48672 10.2938L5.27279 9.87891C5.23146 9.79922 5.14881 9.75 5.05886 9.75C4.96892 9.75 4.88383 9.79922 4.84494 9.87891L4.63101 10.2938C4.56051 10.432 4.40736 10.5141 4.24935 10.5C4.09134 10.4859 3.9552 10.3805 3.91144 10.2352L3.50061 8.92969L3.26237 9.69844C3.11408 10.1742 2.65949 10.5 2.14412 10.5H1.94478C1.73086 10.5 1.55583 10.3313 1.55583 10.125C1.55583 9.91875 1.73086 9.75 1.94478 9.75H2.14412C2.31672 9.75 2.46744 9.64219 2.51606 9.48281L2.87828 8.32266C2.96093 8.05781 3.21375 7.875 3.50061 7.875C3.78746 7.875 4.04029 8.05781 4.12294 8.32266L4.40493 9.22734C4.58482 9.08203 4.81334 9 5.05643 9C5.44296 9 5.79545 9.21094 5.96805 9.54375L6.07501 9.75H6.29137C6.21601 9.54375 6.20142 9.31875 6.25734 9.09844L6.62198 7.68984C6.69005 7.425 6.83105 7.18594 7.03039 6.99375L9.33495 4.77187V3.75H6.2233C5.79302 3.75 5.44539 3.41484 5.44539 3V0H1.55583C0.697691 0 0 0.672656 0 1.5V10.5C0 11.3273 0.697691 12 1.55583 12H7.77913C8.63726 12 9.33495 11.3273 9.33495 10.5V10.0477C9.26932 10.0734 9.20368 10.0945 9.13561 10.1109L7.6746 10.4625ZM0.80198 2.16562C0.778156 2.16562 0.757979 2.15766 0.741448 2.14172C0.724674 2.12578 0.716652 2.10633 0.716652 2.08336V1.55484C0.716652 1.53211 0.724918 1.51266 0.741448 1.49648C0.757979 1.48031 0.778156 1.47258 0.80198 1.47258H4.2688C4.29238 1.47258 4.31255 1.48055 4.32933 1.49648C4.34586 1.51266 4.35412 1.53187 4.35412 1.55484V2.08336C4.35412 2.10633 4.34586 2.12578 4.32933 2.14172C4.31255 2.15789 4.29262 2.16562 4.2688 2.16562H2.95899C2.93516 2.16562 2.92349 2.17711 2.92349 2.19984V6.19406C2.92349 6.21703 2.91523 6.23648 2.8987 6.25242C2.88192 6.26859 2.86199 6.27633 2.83817 6.27633H2.18326C2.15944 6.27633 2.13926 6.26836 2.12273 6.25242C2.10596 6.23648 2.09793 6.21703 2.09793 6.19406V2.19984C2.09793 2.17711 2.08602 2.16562 2.06244 2.16562H0.80198Z" />
				<path d="M6.22302 0V3H9.33467L6.22302 0Z" />
				<path d="M13.7155 3.61172L13.3654 3.27422C12.9862 2.90859 12.3712 2.90859 11.9895 3.27422L11.2748 3.96328L13.0008 5.62734L13.7155 4.93828C14.0947 4.57266 14.0947 3.97969 13.7155 3.61172Z" />
				<path d="M7.58242 7.52363C7.48275 7.61973 7.41225 7.73926 7.37822 7.87285L7.01357 9.28145C6.97954 9.41035 7.01843 9.54395 7.11567 9.6377C7.21291 9.73145 7.35147 9.76894 7.48518 9.73613L8.9462 9.38457C9.08233 9.35176 9.20874 9.28379 9.30841 9.1877L12.4492 6.15723L10.7232 4.49316L7.58242 7.52363Z" />
			</svg>
			<svg viewBox="0 0 14 12" xmlns="http://www.w3.org/2000/svg" fill="%s">
				<path d="M6.22302 0V3H9.33467L6.22302 0Z" />
				<path d="M7.6746 10.4625C7.60167 10.4789 7.52874 10.4906 7.45581 10.4953C7.43393 10.4977 7.41205 10.5 7.39017 10.5H5.83435C5.68606 10.5 5.55235 10.4203 5.48672 10.2938L5.27279 9.87891C5.23146 9.79922 5.14881 9.75 5.05886 9.75C4.96892 9.75 4.88383 9.79922 4.84494 9.87891L4.63101 10.2938C4.56051 10.432 4.40736 10.5141 4.24935 10.5C4.09134 10.4859 3.9552 10.3805 3.91144 10.2352L3.50061 8.92969L3.26237 9.69844C3.11408 10.1742 2.65949 10.5 2.14412 10.5H1.94478C1.73086 10.5 1.55583 10.3313 1.55583 10.125C1.55583 9.91875 1.73086 9.75 1.94478 9.75H2.14412C2.31672 9.75 2.46744 9.64219 2.51606 9.48281L2.87828 8.32266C2.96093 8.05781 3.21375 7.875 3.50061 7.875C3.78746 7.875 4.04028 8.05781 4.12294 8.32266L4.40493 9.22734C4.58482 9.08203 4.81334 9 5.05643 9C5.44296 9 5.79545 9.21094 5.96805 9.54375L6.07501 9.75H6.29137C6.21601 9.54375 6.20142 9.31875 6.25734 9.09844L6.62198 7.68984C6.69005 7.425 6.83105 7.18594 7.03039 6.99375L9.33495 4.77188V3.75H6.2233C5.79302 3.75 5.44539 3.41484 5.44539 3V0H1.55583C0.697691 0 0 0.672656 0 1.5V10.5C0 11.3273 0.697691 12 1.55583 12H7.77913C8.63726 12 9.33495 11.3273 9.33495 10.5V10.0477C9.26932 10.0734 9.20368 10.0945 9.13561 10.1109L7.6746 10.4625ZM1.07838 6.21445C1.05456 6.21445 1.03438 6.20648 1.01785 6.19055C1.00108 6.17461 0.993054 6.15516 0.993054 6.13219V1.49273C0.993054 1.47 1.00132 1.45055 1.01785 1.43438C1.03438 1.41844 1.05456 1.41047 1.07838 1.41047H1.73329C1.75687 1.41047 1.77704 1.41844 1.79382 1.43438C1.81035 1.45055 1.81861 1.46977 1.81861 1.49273V3.41437C1.81861 3.43734 1.83028 3.44859 1.85411 3.44859H3.66932C3.6929 3.44859 3.70481 3.43734 3.70481 3.41437V1.49273C3.70481 1.47 3.71308 1.45055 3.72961 1.43438C3.74614 1.4182 3.76631 1.41047 3.79014 1.41047H4.44504C4.46862 1.41047 4.4888 1.41844 4.50557 1.43438C4.5221 1.45055 4.53037 1.46977 4.53037 1.49273V6.13195C4.53037 6.15492 4.5221 6.17438 4.50557 6.19031C4.4888 6.20648 4.46887 6.21422 4.44504 6.21422H3.79014C3.76631 6.21422 3.74614 6.20625 3.72961 6.19031C3.71283 6.17438 3.70481 6.15492 3.70481 6.13195V4.16906C3.70481 4.14633 3.6929 4.13484 3.66932 4.13484H1.85435C1.83053 4.13484 1.81886 4.14633 1.81886 4.16906V6.13195C1.81886 6.15492 1.81059 6.17438 1.79406 6.19031C1.77729 6.20648 1.75735 6.21422 1.73353 6.21422H1.07862L1.07838 6.21445Z" />
				<path d="M7.58242 7.52363C7.48275 7.61973 7.41225 7.73926 7.37822 7.87285L7.01357 9.28145C6.97954 9.41035 7.01843 9.54395 7.11567 9.6377C7.21291 9.73145 7.35148 9.76895 7.48518 9.73613L8.9462 9.38457C9.08233 9.35176 9.20874 9.28379 9.30841 9.1877L12.4492 6.15723L10.7232 4.49316L7.58242 7.52363Z" />
				<path d="M13.7155 3.61172L13.3654 3.27422C12.9862 2.90859 12.3712 2.90859 11.9895 3.27422L11.2748 3.96328L13.0008 5.62734L13.7155 4.93828C14.0947 4.57266 14.0947 3.97969 13.7155 3.61172Z" />
			</svg>
			<i class="fas fa-cube" style="color: %s"></i>
			<span data-toggle="tooltip" title="%s"><i class="fas fa-sync" style="color: %s"></i></span>
			<i class="fas fa-user-slash" style="color: %s"></i>
			<span data-toggle="tooltip" title="%s"><i class="fas fa-money-bill" style="color: %s"></i></span>
		</div>
	`
	var syncTooltip string
	if event.SyncParticipationStatus > 0 {
		syncTooltip = fmt.Sprintf("%d/%d", event.SyncParticipationCount, Config.Chain.Config.SlotsPerEpoch)
	}

	var withdrawalTooltip string
	if event.WithdrawalStatus > 0 {
		withdrawalTooltip = fmt.Sprintf("Withdrawal processed for %v", FormatCurrentBalance(event.WithdrawalAmount, "ETH"))
	}

	return template.HTML(fmt.Sprintf(str, colorMap[event.AttestationSourceStatus], colorMap[event.AttestationTargetStatus], colorMap[event.AttestationHeadStatus], colorMap[event.BlockProposalStatus], syncTooltip, colorMap[event.SyncParticipationStatus], colorMap[event.SlashingStatus], withdrawalTooltip, colorMap[event.WithdrawalStatus]))
}
