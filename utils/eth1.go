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

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/shopspring/decimal"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var Erc20TransferEventHash = common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")
var Erc1155TransferSingleEventHash = common.HexToHash("0xc3d58168c5ae7397731d063d5bbf3d657854427343f4c083240f7aacaa2d0f62")

func Eth1BlockReward(blockNumber uint64, difficulty []byte) *big.Int {

	// no block rewards for PoS blocks
	// holesky genesis block has difficulty 1 and zero block reward (launched with pos)
	if len(difficulty) == 0 || (len(difficulty) == 1 && difficulty[0] == 1) {
		return big.NewInt(0)
	}

	if blockNumber < Config.Chain.ElConfig.ByzantiumBlock.Uint64() {
		return big.NewInt(5e+18)
	} else if blockNumber < Config.Chain.ElConfig.ConstantinopleBlock.Uint64() {
		return big.NewInt(3e+18)
	} else if Config.Chain.ClConfig.DepositChainID == 5 { // special case for goerli: https://github.com/eth-clients/goerli
		return big.NewInt(0)
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
	return WeiBytesToEther(b).InexactFloat64()
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

func FormatTransactionHash(hash []byte, successful bool) template.HTML {
	if len(hash) < 20 {
		return template.HTML("N/A")
	}
	failedStr := ""
	if !successful {
		failedStr = `<span data-toggle="tooltip" title="Transaction failed">❗</span>`
	}
	return template.HTML(fmt.Sprintf(`<a class="text-monospace" href="/tx/0x%x">0x%x…%x</a>%s`, hash, hash[:3], hash[len(hash)-3:], failedStr))
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
		return formatAddress(address, token, name, isContract, "address", 17, 0, false)
	}
	return formatAddress(address, token, name, isContract, "", 17, 0, false)
}

func FormatBuilder(pubkey []byte) template.HTML {
	name := ""
	if bytes.Equal(pubkey, common.Hex2Bytes("aa1488eae4b06a1fff840a2b6db167afc520758dc2c8af0dfb57037954df3431b747e2f900fe8805f05d635e9a29717b")) {
		name = "MEV-geth Default"
	}
	return FormatAddress(pubkey, nil, name, false, false, false)
}

func FormatBytes(b []byte, addCopyToClipboard bool, link string) template.HTML {
	bStr := fmt.Sprintf("%#x", b)
	ret := ""
	if len(bStr) <= 10 {
		ret += fmt.Sprintf(`<span class="text-monospace">%s</span>`, bStr)
	} else {
		ret += fmt.Sprintf(`<span class="text-monospace" data-html="true" data-toggle="tooltip" data-placement="bottom" title="" data-original-title="%s" data-container="body">%s…%s</span>`, bStr, bStr[0:6], bStr[len(bStr)-4:])
	}
	if len(link) > 0 {
		ret = fmt.Sprintf(`<a href="%s" target="_parent">%s</a>`, link, ret)
	}
	if addCopyToClipboard {
		ret += ` <i class="fa fa-copy text-muted p-1" role="button" data-toggle="tooltip" title="Copy to clipboard" data-clipboard-text="` + bStr + `"></i>`
	}
	return template.HTML(ret)
}

func FormatBlobVersionedHash(h []byte) template.HTML {
	if Config.Frontend.BlobProviderUrl == "" {
		return FormatBytes(h, true, "")
	}
	return FormatBytes(h, true, fmt.Sprintf("%s/%#x", Config.Frontend.BlobProviderUrl, h))
}

func FormatAddressWithLimits(address []byte, name string, isContract bool, link string, digitsLimit int, nameLimit int, addCopyToClipboard bool) template.HTML {
	return formatAddress(address, nil, name, isContract, link, digitsLimit, nameLimit, addCopyToClipboard)
}

func FormatAddressAll(address []byte, name string, isContract bool, link string, digitsLimit int, nameLimit int, addCopyToClipboard bool) template.HTML {
	return formatAddress(address, nil, name, isContract, link, digitsLimit, nameLimit, addCopyToClipboard)
}

// wrapper function of FormatAddressWithLimits used to format addresses in the address page's tables for txs
//
// no link to the given txAddress will be added if it is mainAddress of the page is formatted
//
//	otherwise, "address" will be passed as link
func FormatAddressWithLimitsInAddressPageTable(mainAddress []byte, txAddress []byte, name string, isContract bool, digitsLimit int, nameLimit int, addCopyToClipboard bool) template.HTML {
	link := "address"
	if bytes.Equal(mainAddress, txAddress) {
		link = ""
	}

	return FormatAddressWithLimits(txAddress, name, isContract, link, digitsLimit, nameLimit, addCopyToClipboard)
}

// digitsLimit will limit the address output to that amount of total digits (including 0x & …)
// nameLimit will limit the name, if existing to giving amount of letters, a limit of 0 will display the full name
func formatAddress(address []byte, token []byte, name string, isContract bool, link string, digitsLimit int, nameLimit int, addCopyToClipboard bool) template.HTML {
	name = template.HTMLEscapeString(name)

	// we need at least 5 digits for 0x & …
	if digitsLimit < 5 {
		digitsLimit = 5
	}

	// setting tooltip & limit name/address if necessary

	addressString := fmt.Sprintf("0x%x", address)
	if IsEth1Address(addressString) {
		addressString = FixAddressCasing(addressString)
	}
	tooltip := ""
	if len(name) == 0 { // no name set
		tooltip = addressString

		l := len(address) * 2 // len will be twice address size, as 1 byte hex is 2 digits
		if l <= digitsLimit { // len inside digitsLimits, not much to do
			name = addressString
		} else { // reduce to digits limit
			digitsLimit -= 5     // we will need 5 digits for 0x & …
			name = addressString // get hex bytes as string
			f := digitsLimit / 2 // as this int devision will always cut, we at an odd limit, we will have more digits at the end
			name = fmt.Sprintf("%s…%s", name[:(f+2)], name[(l-(digitsLimit-f)+2):])
		}
		name = fmt.Sprintf(`<span class="text-monospace">%s</span>`, name)
	} else { // name set
		addCopyToClipboard = true
		tooltip = fmt.Sprintf("%s\n%s", name, addressString) // set tool tip first, as we will change name
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

	if len(link) == 0 {
		// not a link
		ret += fmt.Sprintf(`<span data-html="true" data-toggle="tooltip" data-placement="bottom" title="" data-original-title="%s" data-container="body">%s</span>`, tooltip, name)
	} else {
		if token != nil {
			// link & token
			ret += fmt.Sprintf(`<a href="/%s/0x%x#erc20Txns" target="_parent" data-html="true" data-toggle="tooltip" data-placement="bottom" title="" data-original-title="%s">%s</a>`, link, address, tooltip, name)
		} else {
			// just link
			ret += fmt.Sprintf(`<a href="/%s/0x%x" target="_parent" data-html="true" data-toggle="tooltip" data-placement="bottom" title="" data-original-title="%s">%s</a>`, link, address, tooltip, name)
		}
	}

	// copy to clipboard
	if addCopyToClipboard {
		ret += ` <i class="fa fa-copy text-muted p-1" role="button" data-toggle="tooltip" title="Copy to clipboard" data-clipboard-text="` + addressString + `"></i>`
	}

	return template.HTML(ret)
}

func FormatAddressAsLink(address []byte, name string, isContract bool) template.HTML {
	ret := ""
	name = template.HTMLEscapeString(name)
	addressString := FixAddressCasing(fmt.Sprintf("%x", address))

	if len(name) > 0 {
		ret = fmt.Sprintf("<a class=\"text-monospace\" href=\"/address/%s\">%s</a> %v", addressString, name, CopyButton(addressString))
	} else {
		ret = fmt.Sprintf("<a class=\"text-monospace\" href=\"/address/%s\">%s…%s</a> %v", addressString, addressString[:8], addressString[len(addressString)-6:], CopyButton(addressString))
	}

	if isContract {
		ret = "<i class=\"fas fa-file-contract mr-1\"></i>" + ret
	}
	return template.HTML(ret)
}

func FormatAddressAsTokenLink(token, address []byte, name string, verified bool, isContract bool) template.HTML {
	ret := ""
	name = template.HTMLEscapeString(name)
	addressString := FixAddressCasing(fmt.Sprintf("%x", address))

	if len(name) > 0 {
		if verified {
			ret = fmt.Sprintf("<a class=\"text-monospace\" href=\"/token/%x?a=%s\">✔ %s (%s…%s)</a> %v", token, addressString, name, addressString[:8], addressString[len(addressString)-6:], CopyButton(addressString))
		} else {
			ret = fmt.Sprintf("<a class=\"text-monospace\" href=\"/token/%x?a=%s\">%s %s…%s</a> %v", token, addressString, name, addressString[:8], addressString[len(addressString)-6:], CopyButton(addressString))
		}
	} else {
		ret = fmt.Sprintf("<a class=\"text-monospace\" href=\"/token/%x?a=%s\">%s…%s</a> %v", token, addressString, addressString[:8], addressString[len(addressString)-6:], CopyButton(addressString))
	}

	if isContract {
		ret = "<i class=\"fas fa-file-contract mr-1\"></i>" + ret
	}
	return template.HTML(ret)
}

func FormatHashLong(hash common.Hash) template.HTML {
	address := hash.String()
	if len(address) > 4 {
		htmlFormat := `
		<div class="d-flex text-monospace">
			%s
			<span class="flex-shrink-1 text-truncate">%s</span>
			%s
		</div>`

		return template.HTML(fmt.Sprintf(htmlFormat, address[:4], address[4:len(address)-4], address[len(address)-4:]))
	}

	return template.HTML(address)
}

func FormatAddressLong(address string) template.HTML {
	if IsValidEnsDomain(address) {
		return template.HTML(fmt.Sprintf(`<span data-truncate-middle="%s"></span>.eth`, strings.TrimSuffix(address, ".eth")))
	}
	address = FixAddressCasing(address)
	if len(address) > 4 {
		htmlFormat := `
		<span class="text-monospace mw-100">%s<span class="text-primary">%s</span>%s<span class="text-primary">%s</span></span>`

		return template.HTML(fmt.Sprintf(htmlFormat, address[:2], address[2:6], address[6:len(address)-4], address[len(address)-4:]))
	}

	return template.HTML(address)

}

func FormatAmountFormatted(amount *big.Int, unit string, digits int, maxPreCommaDigitsBeforeTrim int, fullAmountTooltip bool, smallUnit bool, newLineForUnit bool) template.HTML {
	return formatAmount(amount, unit, digits, maxPreCommaDigitsBeforeTrim, fullAmountTooltip, smallUnit, newLineForUnit)
}
func FormatAmount(amount *big.Int, unit string, digits int) template.HTML {
	return formatAmount(amount, unit, digits, 0, true, false, false)
}
func FormatBigAmount(amount *hexutil.Big, unit string, digits int) template.HTML {
	return FormatAmount((*big.Int)(amount), unit, digits)
}
func FormatBytesAmount(amount []byte, unit string, digits int) template.HTML {
	return FormatAmount(new(big.Int).SetBytes(amount), unit, digits)
}
func formatAmount(amount *big.Int, unit string, digits int, maxPreCommaDigitsBeforeTrim int, fullAmountTooltip bool, smallUnit bool, newLineForUnit bool) template.HTML {
	// define display unit & digits used per unit max
	displayUnit := " " + unit
	var unitDigits int
	if unit == "ETH" || unit == "Ether" || unit == "xDAI" || unit == "GNO" {
		unitDigits = 18
	} else if unit == "GWei" {
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

	trimmedAmount, fullAmount := trimAmount(amount, unitDigits, maxPreCommaDigitsBeforeTrim, digits, false)
	tooltip := ""
	if fullAmountTooltip {
		tooltip = fmt.Sprintf(` data-toggle="tooltip" data-placement="top" title="%s"`, fullAmount)
	}

	// done, convert to HTML & return
	return template.HTML(fmt.Sprintf("<span%s>%s%s</span>", tooltip, trimmedAmount, displayUnit))
}

func trimAmount(amount *big.Int, unitDigits int, maxPreCommaDigitsBeforeTrim int, digits int, addPositiveSign bool) (trimmedAmount, fullAmount string) {
	// Initialize trimmedAmount and postComma variables to "0"
	trimmedAmount = "0"
	postComma := "0"
	proceed := ""

	if amount != nil {
		s := amount.String()
		if amount.Sign() > 0 && addPositiveSign {
			proceed = "+"
		} else if amount.Sign() < 0 {
			proceed = "-"
			s = strings.Replace(s, "-", "", 1)
		}
		l := len(s)

		// Check if there is a part of the amount before the decimal point
		if l > int(unitDigits) {
			// Calculate length of preComma part
			l -= unitDigits
			// Set preComma to part of the string before the decimal point
			trimmedAmount = s[:l]
			// Set postComma to part of the string after the decimal point, after removing trailing zeros
			postComma = strings.TrimRight(s[l:], "0")

			// Check if the preComma part exceeds the maximum number of digits before the decimal point
			if maxPreCommaDigitsBeforeTrim > 0 && l > maxPreCommaDigitsBeforeTrim {
				// Reduce the number of digits after the decimal point by the excess number of digits in the preComma part
				l -= maxPreCommaDigitsBeforeTrim
				if digits < l {
					digits = 0
				} else {
					digits -= l
				}
			}
			// Check if there is only a part of the amount after the decimal point, and no leading zeros need to be added
		} else if l == unitDigits {
			// Set postComma to part of the string after the decimal point, after removing trailing zeros
			postComma = strings.TrimRight(s, "0")
			// Check if there is only a part of the amount after the decimal point, and leading zeros need to be added
		} else if l != 0 {
			// Use fmt package to add leading zeros to the string
			d := fmt.Sprintf("%%0%dd", unitDigits-l)
			// Set postComma to resulting string, after removing trailing zeros
			postComma = strings.TrimRight(fmt.Sprintf(d, 0)+s, "0")
		}

		fullAmount = trimmedAmount
		if len(postComma) > 0 {
			fullAmount += "." + postComma
		}

		// limit floating part
		if len(postComma) > digits {
			postComma = postComma[:digits]
		}

		// set floating point
		if len(postComma) > 0 {
			trimmedAmount += "." + postComma
		}
	}
	return proceed + trimmedAmount, proceed + fullAmount
}

func FormatMethod(method string) template.HTML {
	return template.HTML(fmt.Sprintf(`<span class="badge badge-light text-truncate mw-100" truncate-tooltip="%s">%s</span>`, method, method))
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
	return fmt.Sprintf("%.1f T", decimal.NewFromBigInt(number, -12).InexactFloat64())
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
