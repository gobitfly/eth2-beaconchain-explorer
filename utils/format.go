package utils

import (
	"fmt"
	"html"
	"html/template"
	"strings"
	"time"

	eth1common "github.com/ethereum/go-ethereum/common"
)

func FormatMessageToHtml(message string) template.HTML {
	message = fmt.Sprint(strings.Replace(message, "Error: ", "", 1))
	return template.HTML(message)
}

// FormatAttestationStatus will return a user-friendly attestation for an attestation status number
func FormatAttestationStatus(status uint64) template.HTML {
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

// FormatAttestorAssignmentKey will format attestor assignment keys
func FormatAttestorAssignmentKey(AttesterSlot, CommitteeIndex, MemberIndex uint64) string {
	return fmt.Sprintf("%v-%v-%v", AttesterSlot, CommitteeIndex, MemberIndex)
}

// FormatBalance will return a string for a balance
func FormatBalance(balance uint64) template.HTML {
	return template.HTML(fmt.Sprintf("%.2f ETH", float64(balance)/float64(1e9)))
}

// FormatBlockRoot will return the block-root formated as html
func FormatBlockRoot(blockRoot []byte) template.HTML {
	if len(blockRoot) < 32 {
		return "N/A"
	}
	return template.HTML(fmt.Sprintf("<a href=\"/block/%x\">%v</a>", blockRoot, FormatHash(blockRoot)))
}

// FormatBlockSlot will return the block-slot formated as html
func FormatBlockSlot(blockSlot uint64) template.HTML {
	return template.HTML(fmt.Sprintf("<a href=\"/block/%[1]d\">%[1]d</a>", blockSlot))
}

// FormatSlotToTimestamp will return the momentjs time elapsed since blockSlot
func FormatSlotToTimestamp(blockSlot uint64) template.HTML {
	time := SlotToTime(blockSlot)
	return FormatTimestamp(time.Unix())
}

// FormatBlockStatus will return an html status for a block.
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

// FormatCurrentBalance will return the current balance formated as string with 9 digits after the comma (1 gwei = 1e9 eth)
func FormatCurrentBalance(balance uint64) template.HTML {
	return template.HTML(fmt.Sprintf("%.9f ETH", float64(balance)/float64(1e9)))
}

// FormatDepositAmount will return the deposit amount formated as string
func FormatDepositAmount(amount uint64) template.HTML {
	return template.HTML(fmt.Sprintf("%.0f ETH", float64(amount)/float64(1e9)))
}

// FormatEffectiveBalance will return the effective balance formated as string with 1 digit after the comma
func FormatEffectiveBalance(balance uint64) template.HTML {
	return template.HTML(fmt.Sprintf("%.1f ETH", float64(balance)/float64(1e9)))
}

// FormatEpoch will return the epoch formated as html
func FormatEpoch(epoch uint64) template.HTML {
	return template.HTML(fmt.Sprintf("<a href=\"/epoch/%[1]d\">%[1]d</a>", epoch))
}

// FormatEth1Address will return the eth1-address formated as html
func FormatEth1Address(addr []byte) template.HTML {
	eth1Addr := eth1common.BytesToAddress(addr)
	return template.HTML(fmt.Sprintf("<a href=\"https://goerli.etherscan.io/address/0x%x\" class=\"text-monospace\">%s…</a>", addr, eth1Addr.Hex()[:8]))
}

// FormatEth1Block will return the eth1-block formated as html
func FormatEth1Block(block uint64) template.HTML {
	return template.HTML(fmt.Sprintf("<a href=\"https://goerli.etherscan.io/block/%[1]d\">%[1]d</a>", block))
}

// FormatEth1TxHash will return the eth1-tx-hash formated as html
func FormatEth1TxHash(hash []byte) template.HTML {
	return template.HTML(fmt.Sprintf("<a href=\"https://goerli.etherscan.io/tx/0x%x\">%v</a>", hash, FormatHash(hash)))
}

// FormatGlobalParticipationRate will return the global-participation-rate formated as html
func FormatGlobalParticipationRate(e uint64, r float64) template.HTML {
	rr := fmt.Sprintf("%.0f%%", r*100)
	tpl := `<div>%.2[1]f <small class="text-muted ml-3">(%[2]v)</small></div><div class="progress" style="height:5px;"><div class="progress-bar" role="progressbar" style="width: %[2]v;" aria-valuenow="%[2]v" aria-valuemin="0" aria-valuemax="100"></div></div>`
	return template.HTML(fmt.Sprintf(tpl, float64(e)/1e9, rr))
}

// FormatGraffiti will return the graffiti formated as html
func FormatGraffiti(graffiti []byte) template.HTML {
	str := strings.Map(fixUtf, template.HTMLEscapeString(string(graffiti)))
	return template.HTML(fmt.Sprintf("<span aria-graffiti=\"%#x\">%s</span>", graffiti, str))
}

// FormatHash will return a hash formated as html
func FormatHash(hash []byte) template.HTML {
	// if len(hash) > 6 {
	// 	return template.HTML(fmt.Sprintf("<span class=\"text-monospace\">0x%x…%x</span>", hash[:3], hash[len(hash)-3:]))
	// }
	// return template.HTML(fmt.Sprintf("<span class=\"text-monospace\">0x%x</span>", hash))
	if len(hash) > 3 {
		return template.HTML(fmt.Sprintf("<span class=\"text-monospace\">%#x…</span>", hash[:3]))
	}
	return template.HTML(fmt.Sprintf("<span class=\"text-monospace\">%#x</span>", hash))
}

// FormatIncome will return a string for a balance
func FormatIncome(income int64) template.HTML {
	if income > 0 {
		return template.HTML(fmt.Sprintf(`<span class="text-success"><b>+%.4f ETH</b></span>`, float64(income)/float64(1e9)))
	} else if income < 0 {
		return template.HTML(fmt.Sprintf(`<span class="text-danger"><b>%.4f ETH</b></span>`, float64(income)/float64(1e9)))
	} else {
		return template.HTML(fmt.Sprintf(`<b>%.4f ETH</b>`, float64(income)/float64(1e9)))
	}
}

// FormatPercentage will return a string for a percentage
func FormatPercentage(percentage float64) string {
	return fmt.Sprintf("%.0f", percentage*float64(100))
}

// FormatPublicKey will return html formatted text for a validator-public-key
func FormatPublicKey(validator []byte) template.HTML {
	return template.HTML(fmt.Sprintf("<i class=\"fas fa-male\"></i> <a href=\"/validator/0x%x\">%v</a>", validator, FormatHash(validator)))
}

// FormatTimestamp will return a timestamp formated as html. This is supposed to be used together with client-side js
func FormatTimestamp(ts int64) template.HTML {
	return template.HTML(fmt.Sprintf("<span class=\"timestamp\" title=\"%v\" data-toggle=\"tooltip\" data-placement=\"top\" data-timestamp=\"%d\"></span>", time.Unix(ts, 0), ts))
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
		return "<b>Deposited</b>"
	} else if status == "pending" {
		return "<b>Pending</b>"
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
		return "<b>Exited</b>"
	} else if status == "slashed" {
		return "<b>Slashed</b>"
	}
	return "<b>Unknown</b>"
}

// FormatValidator will return html formatted text for a validator
func FormatValidator(validator uint64) template.HTML {
	return template.HTML(fmt.Sprintf("<i class=\"fas fa-male\"></i> <a href=\"/validator/%v\">%v</a>", validator, validator))
}

func FormatValidatorWithName(validator uint64, name string) template.HTML {
	if name != "" {
		return template.HTML(fmt.Sprintf("<i class=\"fas fa-male\"></i> <a href=\"/validator/%v\">%v (<span class=\"text-truncate\">"+html.EscapeString(name)+"</span>)</a>", validator, validator))
	} else {
		return template.HTML(fmt.Sprintf("<i class=\"fas fa-male\"></i> <a href=\"/validator/%v\">%v</a>", validator, validator))
	}
}

// FormatValidatorInt64 will return html formatted text for a validator (for an int64 validator-id)
func FormatValidatorInt64(validator int64) template.HTML {
	return FormatValidator(uint64(validator))
}

// FormatSlashedValidatorInt64 will return html formatted text for a slashed validator
func FormatSlashedValidatorInt64(validator int64) template.HTML {
	return template.HTML(fmt.Sprintf("<i class=\"fas fa-user-slash text-danger\"></i> <a href=\"/validator/%v\">%v</a>", validator, validator))
}

// FormatSlashedValidator will return html formatted text for a slashed validator
func FormatSlashedValidator(validator uint64) template.HTML {
	return template.HTML(fmt.Sprintf("<i class=\"fas fa-user-slash text-danger\"></i> <a href=\"/validator/%v\">%v</a>", validator, validator))
}

// FormatSlashedValidatorsInt64 will return html formatted text for slashed validators
func FormatSlashedValidatorsInt64(validators []int64) template.HTML {
	str := ""
	for i, v := range validators {
		if i == len(validators)+1 {
			str += fmt.Sprintf("<i class=\"fas fa-user-slash text-danger\"></i> <a href=\"/validator/%v\">%v</a>", v, v)
		} else {
			str += fmt.Sprintf("<i class=\"fas fa-user-slash text-danger\"></i> <a href=\"/validator/%v\">%v</a>, ", v, v)
		}
	}
	return template.HTML(str)
}

// FormatSlashedValidators will return html formatted text for slashed validators
func FormatSlashedValidators(validators []uint64) template.HTML {
	vals := make([]string, 0, len(validators))
	for _, v := range validators {
		vals = append(vals, fmt.Sprintf("<i class=\"fas fa-user-slash text-danger\"></i> <a href=\"/validator/%v\">%v</a>", v, v))
	}
	return template.HTML(strings.Join(vals, ","))
}

// FormatYesNo will return yes or no formated as html
func FormatYesNo(yes bool) template.HTML {
	if yes {
		return `<span class="badge bg-success text-white">Yes</span>`
	}
	return `<span class="badge bg-warning text-dark">No</span>`
}

func FormatValidatorName(name string) template.HTML {
	str := strings.Map(fixUtf, template.HTMLEscapeString(name))
	return template.HTML(fmt.Sprintf("<b><abbr title=\"This name has been set by the owner of this validator\">%s</abbr></b>", str))
}
