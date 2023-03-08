package utils

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"eth2-exporter/types"
	"fmt"
	"html/template"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/params"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var Erc20TransferEventHash = common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")
var Erc1155TransferSingleEventHash = common.HexToHash("0xc3d58168c5ae7397731d063d5bbf3d657854427343f4c083240f7aacaa2d0f62")

func Eth1BlockReward(blockNumber uint64, difficulty []byte) *big.Int {

	if len(difficulty) == 0 { // no block rewards for PoS blocks
		return big.NewInt(0)
	}

	if blockNumber < 4370000 {
		return big.NewInt(5e+18)
	} else if blockNumber < 7280000 {
		return big.NewInt(3e+18)
	} else {
		return big.NewInt(2e+18)
	}
}

func Eth1TotalReward(block *types.Eth1BlockIndexed) *big.Int {
	blockReward := Eth1BlockReward(block.GetNumber(), block.GetDifficulty())
	uncleReward := big.NewInt(0).SetBytes(block.GetUncleReward())
	txFees := big.NewInt(0).SetBytes(block.GetTxReward())

	totalReward := big.NewInt(0).Add(blockReward, txFees)
	return totalReward.Add(totalReward, uncleReward)
}

func StripPrefix(hexStr string) string {
	return strings.Replace(hexStr, "0x", "", 1)
}

func EthBytesToFloat(b []byte) float64 {
	f, _ := new(big.Float).Quo(new(big.Float).SetInt(new(big.Int).SetBytes(b)), new(big.Float).SetInt(big.NewInt(params.Ether))).Float64()
	return f
}

func FormatBlockNumber(number uint64) template.HTML {
	return template.HTML(fmt.Sprintf("<a href=\"/block/%[1]d\">%[2]s</a>", number, FormatAddCommas(number)))
}

func FormatTxHash(hash string) template.HTML {
	if len(hash) > 3 {
		return template.HTML(fmt.Sprintf("<a class=\"text-monospace\" href=\"/tx/%s\">%s…</a>", hash, hash[:5]))
	}
	return template.HTML(fmt.Sprintf("<a class=\"text-monospace\" href=\"/tx/%s\">%s…</a>", hash, hash))
}

// func FormatHash(hash string) template.HTML {
// 	hash = strings.Replace(hash, "0x", "", -1)
// 	if len(hash) > 3 {
// 		return template.HTML(fmt.Sprintf("<span class=\"text-monospace\">0x%#s…</span>", hash[:3]))
// 	}
// 	return template.HTML(fmt.Sprintf("<span class=\"text-monospace\">0x%#s</span>", hash))
// }

// func FormatTimestamp(ts int64) template.HTML {
// 	return template.HTML(fmt.Sprintf("<span class=\"timestamp\" title=\"%v\" data-toggle=\"tooltip\" data-placement=\"top\" data-timestamp=\"%d\"></span>", time.Unix(ts, 0), ts))
// }

func FormatBlockHash(hash []byte) template.HTML {
	if len(hash) < 20 {
		return template.HTML("N/A")
	}
	return template.HTML(fmt.Sprintf(`<a class="text-monospace" href="/block/0x%x">0x%x…%x</a> %v`, hash, hash[:2], hash[len(hash)-2:], CopyButton(hex.EncodeToString(hash))))
}

func FormatTransactionHash(hash []byte) template.HTML {
	if len(hash) < 20 {
		return template.HTML("N/A")
	}
	return template.HTML(fmt.Sprintf(`<a class="text-monospace" href="/tx/0x%x">0x%x…%x</a>`, hash, hash[:3], hash[len(hash)-3:]))
}

func FormatInOutSelf(address, from, to []byte) template.HTML {
	if address == nil && len(address) == 0 {
		return ""
	}
	if !bytes.Equal(to, from) {
		if bytes.Equal(address, from) {
			return template.HTML(`<span style="width: 45px;" class="font-weight-bold badge badge-warning text-white text-monospace">OUT</span>`)
		} else {
			return template.HTML(`<span style="width: 45px;" class="font-weight-bold badge badge-success text-white text-monospace">IN</span>`)
		}
	} else {
		return template.HTML(`<span style="width: 45px;" class="font-weight-bold badge badge-info text-white text-monospace">SELF</span>`)
	}
}

func FormatAddress(address []byte, token []byte, name string, verified bool, isContract bool, link bool) template.HTML {
	if link {
		return formatAddress(address, token, name, isContract, "address", "", 17, 0, false)
	}
	return formatAddress(address, token, name, isContract, "", "", 17, 0, false)
}

func FormatBuilder(pubkey []byte) template.HTML {
	name := ""
	if bytes.Equal(pubkey, common.Hex2Bytes("aa1488eae4b06a1fff840a2b6db167afc520758dc2c8af0dfb57037954df3431b747e2f900fe8805f05d635e9a29717b")) {
		name = "MEV-geth Default"
	}
	return FormatAddress(pubkey, nil, name, false, false, false)
}

func FormatAddressWithLimits(address []byte, name string, isContract bool, link string, digitsLimit int, nameLimit int, addCopyToClipboard bool) template.HTML {
	return formatAddress(address, nil, name, isContract, link, "", digitsLimit, nameLimit, addCopyToClipboard)
}

func FormatAddressAll(address []byte, name string, isContract bool, link string, urlFragment string, digitsLimit int, nameLimit int, addCopyToClipboard bool) template.HTML {
	return formatAddress(address, nil, name, isContract, link, urlFragment, digitsLimit, nameLimit, addCopyToClipboard)
}

// digitsLimit will limit the address output to that amount of total digits (including 0x & …)
// nameLimit will limit the name, if existing to giving amount of letters, a limit of 0 will display the full name
func formatAddress(address []byte, token []byte, name string, isContract bool, link string, urlFragment string, digitsLimit int, nameLimit int, addCopyToClipboard bool) template.HTML {
	name = template.HTMLEscapeString(name)

	// we need at least 5 digits for 0x & …
	if digitsLimit < 5 {
		digitsLimit = 5
	}

	// setting tooltip & limit name/address if necessary
	addressString := fmt.Sprintf("0x%x", address)
	tooltip := ""
	if len(name) == 0 { // no name set
		tooltip = addressString

		l := len(address) * 2 // len will be twice address size, as 1 byte hex is 2 digits
		if l <= digitsLimit { // len inside digitsLimits, not much to do
			name = addressString
		} else { // reduce to digits limit
			digitsLimit -= 5                  // we will need 5 digits for 0x & …
			name = fmt.Sprintf("%x", address) // get hex bytes as string
			f := digitsLimit / 2              // as this int devision will always cut, we at an odd limit, we will have more digits at the end
			name = fmt.Sprintf("0x%s…%s", name[:f], name[(l-(digitsLimit-f)):])
		}
		name = fmt.Sprintf(`<span class="text-monospace">%s</span>`, name)
	} else { // name set
		tooltip = fmt.Sprintf("%s\n0x%x", name, address) // set tool tip first, as we will change name
		// limit name if necessary
		if nameLimit > 0 && len(name) > nameLimit {
			name = name[:nameLimit-3] + "…"
		}
	}

	// contract
	ret := ""
	if isContract {
		ret = "<i class=\"fas fa-file-contract mr-1\"></i>" + ret
	}

	// not a link
	if len(link) < 1 {
		ret += fmt.Sprintf(`<span data-html="true" data-toggle="tooltip" data-placement="bottom" title="" data-original-title="%s" data-container="body">%s</span>`, tooltip, name)
	} else {
		// link & token
		if token != nil {
			ret += fmt.Sprintf(`<a href="/`+link+`/0x%x#erc20Txns" target="_parent" data-html="true" data-toggle="tooltip" data-placement="bottom" title="" data-original-title="%s">%s</a>`, address, tooltip, name)
		} else { // just link
			ret += fmt.Sprintf(`<a href="/`+link+`/0x%x`+urlFragment+`" target="_parent" data-html="true" data-toggle="tooltip" data-placement="bottom" title="" data-original-title="%s">%s</a>`, address, tooltip, name)
		}
	}

	// copy to clipboard
	if addCopyToClipboard {
		ret += ` <i class="fa fa-copy text-muted p-1" role="button" data-toggle="tooltip" title="Copy to clipboard" data-clipboard-text="` + addressString + `"></i>`
	}

	// done
	return template.HTML(ret)
}

func FormatAddressAsLink(address []byte, name string, verified bool, isContract bool) template.HTML {
	ret := ""
	name = template.HTMLEscapeString(name)

	if len(name) > 0 {
		if verified {
			ret = fmt.Sprintf("<a class=\"text-monospace\" href=\"/address/0x%x\">✔ %s (0x%x…%x)</a> %v", address, name, address[:3], address[len(address)-3:], CopyButton(hex.EncodeToString(address)))
		} else {
			ret = fmt.Sprintf("<a class=\"text-monospace\" href=\"/address/0x%x\">%s 0x%x…%x</a> %v", address, name, address[:3], address[len(address)-3:], CopyButton(hex.EncodeToString(address)))
		}
	} else {
		ret = fmt.Sprintf("<a class=\"text-monospace\" href=\"/address/0x%x\">0x%x…%x</a> %v", address, address[:3], address[len(address)-3:], CopyButton(hex.EncodeToString(address)))
	}

	if isContract {
		ret = "<i class=\"fas fa-file-contract mr-1\"></i>" + ret
	}
	return template.HTML(ret)
}

func FormatAddressAsTokenLink(token, address []byte, name string, verified bool, isContract bool) template.HTML {
	ret := ""
	name = template.HTMLEscapeString(name)

	if len(name) > 0 {
		if verified {
			ret = fmt.Sprintf("<a class=\"text-monospace\" href=\"/token/0x%x?a=0x%x\">✔ %s (0x%x…%x)</a> %v", token, address, name, address[:3], address[len(address)-3:], CopyButton(hex.EncodeToString(address)))
		} else {
			ret = fmt.Sprintf("<a class=\"text-monospace\" href=\"/token/0x%x?a=0x%x\">%s 0x%x…%x</a> %v", token, address, name, address[:3], address[len(address)-3:], CopyButton(hex.EncodeToString(address)))
		}
	} else {
		ret = fmt.Sprintf("<a class=\"text-monospace\" href=\"/token/0x%x?a=0x%x\">0x%x…%x</a> %v", token, address, address[:3], address[len(address)-3:], CopyButton(hex.EncodeToString(address)))
	}

	if isContract {
		ret = "<i class=\"fas fa-file-contract mr-1\"></i>" + ret
	}
	return template.HTML(ret)
}

func FormatHashLong(hash common.Hash) template.HTML {
	address := hash.String()
	test := `
	<div class="d-flex text-monospace">
		<span class="">%s</span>
		<span class="flex-shrink-1 text-truncate">%s</span>
		<span class="">%s</span>
	</div>`
	if len(address) > 4 {
		return template.HTML(fmt.Sprintf(test, address[:4], address[4:len(address)-4], address[len(address)-4:]))
	}

	return template.HTML(address)
}

func FormatAddressLong(address string) template.HTML {
	test := `
	<span class="text-monospace mw-100"><span class="text-primary">0x%s</span><span class="text-truncate">%s</span><span class="text-primary">%s</span></span>`
	if len(address) > 4 {
		return template.HTML(fmt.Sprintf(test, address[:4], address[4:len(address)-4], address[len(address)-4:]))
	}

	return template.HTML(address)

}

func FormatAmountFormated(amount *big.Int, unit string, digits int, maxPreCommaDigitsBeforeTrim int, fullAmountTooltip bool, smallUnit bool, newLineForUnit bool) template.HTML {
	return formatAmount(amount, unit, digits, maxPreCommaDigitsBeforeTrim, fullAmountTooltip, smallUnit, newLineForUnit)
}
func FormatAmount(amount *big.Int, unit string, digits int) template.HTML {
	return formatAmount(amount, unit, digits, 0, false, false, false)
}
func FormatBigAmount(amount *hexutil.Big, unit string, digits int) template.HTML {
	return FormatAmount((*big.Int)(amount), unit, digits)
}
func formatAmount(amount *big.Int, unit string, digits int, maxPreCommaDigitsBeforeTrim int, fullAmountTooltip bool, smallUnit bool, newLineForUnit bool) template.HTML {
	// define display unit & digits used per unit max
	var displayUnit string
	var unitDigits int
	if unit == "ETH" {
		displayUnit = " Ether"
		unitDigits = 18
	} else if unit == "GWei" {
		displayUnit = " GWei"
		unitDigits = 9
	} else {
		displayUnit = " ?"
		unitDigits = 0
	}

	// small unit & new line for unit handling
	{
		unit = displayUnit
		if newLineForUnit {
			displayUnit = "<BR />"
		} else {
			displayUnit = ""
		}
		if smallUnit {
			displayUnit += `<span style="font-size: .63rem;`
			if newLineForUnit {
				displayUnit += `color: grey;`
			}
			displayUnit += `">` + unit + `</span>`
		} else {
			displayUnit += unit
		}
	}

	// split in pre and post part
	preComma := "0"
	postComma := "0"
	if amount != nil {
		s := amount.String()
		l := len(s)
		if l > int(unitDigits) { // there is a pre comma part
			l -= unitDigits
			preComma = s[:l]
			postComma = strings.TrimRight(s[l:], "0")
			// reduce digits if precomma part exceeds limit
			if maxPreCommaDigitsBeforeTrim > 0 && l > maxPreCommaDigitsBeforeTrim {
				l -= maxPreCommaDigitsBeforeTrim
				if digits < l {
					digits = 0
				} else {
					digits -= l
				}
			}
		} else if l == unitDigits { // there is only post comma part and no leading zeros has to be added
			postComma = strings.TrimRight(s, "0")
		} else if l != 0 { // there is only post comma part and leading zeros as to be added
			d := fmt.Sprintf("%%0%dd", unitDigits-l)
			postComma = strings.TrimRight(fmt.Sprintf(d, 0)+s, "0")
		}
	}

	// tooltip
	var tooltip string
	if fullAmountTooltip {
		tooltip = ` data-toggle="tooltip" data-placement="top" title="` + preComma
		if len(postComma) > 0 {
			tooltip += `.` + postComma
		}
		tooltip += `"`
	}

	// limit floating part
	if len(postComma) > digits {
		postComma = postComma[:digits]
	}

	// set floating point
	if len(postComma) > 0 {
		preComma += "." + postComma
	}

	// done, convert to HTML & return
	return template.HTML(fmt.Sprintf("<span%s>%s%s</span>", tooltip, preComma, displayUnit))
}

func FormatMethod(method string) template.HTML {
	return template.HTML(fmt.Sprintf(`<span class="badge badge-light">%s</span>`, method))
}

func FormatBlockUsage(gasUsage uint64, gasLimit uint64) template.HTML {
	percentage := uint64(0)
	if gasLimit != 0 {
		percentage = gasUsage * 100 / gasLimit
	}
	tpl := `<div>%[1]v<small class="text-muted ml-2">(%[2]v%%)</small></div><div class="progress" style="height:5px;"><div class="progress-bar" role="progressbar" style="width: %[2]v%%;" aria-valuenow="%[2]v" aria-valuemin="0" aria-valuemax="100"></div></div>`
	p := message.NewPrinter(language.English)
	return template.HTML(p.Sprintf(tpl, gasUsage, percentage, gasLimit))
}

func FormatNumber(number interface{}) string {
	p := message.NewPrinter(language.English)
	return p.Sprintf("%.5f", number)
}

func FormatDifficulty(number *big.Int) string {
	f := new(big.Float).SetInt(number)
	f.Quo(f, big.NewFloat(1e12))
	r, _ := f.Float64()
	return fmt.Sprintf("%.1f T", r)
}

func FormatTime(t time.Time) template.HTML {
	return template.HTML(fmt.Sprintf("<span aria-ethereum-date=\"%v\">%v</span>", t.Unix(), t))
}

func FormatTimeFromNow(t time.Time) template.HTML {
	return template.HTML(HumanizeTime(t))
}

func FormatHashrate(h float64) template.HTML {
	if h > 1e12 {
		return template.HTML(fmt.Sprintf("%.1f TH/s", h/1e12))
	}
	return template.HTML(fmt.Sprintf("%.1f GH/s", h/1e9))
}

// func FormatPercentage(p float64, digits int) template.HTML {
// 	return template.HTML(fmt.Sprintf("%."+strconv.Itoa(digits)+"f %%", p))
// }

func FormatTokenIcon(icon []byte, size int) template.HTML {
	if icon == nil {
		return template.HTML("")
	}
	icon64 := base64.StdEncoding.EncodeToString(icon)
	return template.HTML(fmt.Sprintf("<img class=\"mb-1 mr-1\" src=\"data:image/gif;base64,%v\" width=\"%v\" height=\"%v\">", icon64, size, size))
}
