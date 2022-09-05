package utils

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"html/template"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var Erc20TransferEventHash = common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")
var Erc1155TransferSingleEventHash = common.HexToHash("0xc3d58168c5ae7397731d063d5bbf3d657854427343f4c083240f7aacaa2d0f62")

func Eth1BlockReward(blockNumber uint64) *big.Int {
	if blockNumber < 4370000 {
		return big.NewInt(5e+18)
	} else if blockNumber < 7280000 {
		return big.NewInt(3e+18)
	} else {
		return big.NewInt(2e+18)
	}
}

func StripPrefix(hexStr string) string {
	return strings.Replace(hexStr, "0x", "", 1)
}

func EthBytesToFloat(b []byte) float64 {
	f, _ := new(big.Float).Quo(new(big.Float).SetInt(new(big.Int).SetBytes(b)), new(big.Float).SetInt(big.NewInt(params.Ether))).Float64()
	return f
}

func FormatBlockNumber(number uint64) template.HTML {
	return template.HTML(fmt.Sprintf("<a href=\"/execution/block/%[1]d\">%[1]d</a>", number))
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
	return template.HTML(fmt.Sprintf(`<a class="text-monospace" href="/execution/block/0x%x">0x%x…%x</a> %v`, hash, hash[:2], hash[len(hash)-2:], CopyButton(hex.EncodeToString(hash))))
}

func FormatTransactionHash(hash []byte) template.HTML {
	if len(hash) < 20 {
		return template.HTML("N/A")
	}
	return template.HTML(fmt.Sprintf(`<a class="text-monospace" href="/execution/tx/0x%x">0x%x…%x</a> %v`, hash, hash[:2], hash[len(hash)-2:], CopyButton(hex.EncodeToString(hash))))
}

func FormatInOutSelf(address, from, to []byte) template.HTML {
	if address == nil && len(address) == 0 {
		return ""
	}
	if !bytes.Equal(to, from) {
		if bytes.Equal(address, from) {
			return "OUT"
		} else {
			return "IN"
		}
	} else {
		return "SELF"
	}
}

func FormatAddress(address []byte, token []byte, name string, verified bool, isContract bool, link bool) template.HTML {

	name = template.HTMLEscapeString(name)

	tooltip := ""
	if len(name) == 0 {
		name = fmt.Sprintf("0x%x", address)
		tooltip = name
	} else {
		tooltip = fmt.Sprintf("%s\n0x%x", name, address)
	}

	ret := ""
	if isContract {
		ret = "<i class=\"fas fa-file-contract mr-1\"></i>" + ret
	}

	if !link {
		ret += fmt.Sprintf(`<span class="text-truncate" data-html="true" data-toggle="tooltip" data-placement="bottom" title="" data-original-title="%s">%s</span>`, tooltip, name)
		return template.HTML(ret)
	}

	if token != nil {
		ret += fmt.Sprintf(`<a class="text-truncate" href="/execution/address/0x%x#erc20Txns" target="_parent" data-html="true" data-toggle="tooltip" data-placement="bottom" title="" data-original-title="%s">%s</a>`, address, tooltip, name)
	} else {
		ret += fmt.Sprintf(`<a class="text-truncate" href="/execution/address/0x%x" target="_parent" data-html="true" data-toggle="tooltip" data-placement="bottom" title="" data-original-title="%s">%s</a>`, address, tooltip, name)
	}
	return template.HTML(ret)
}

func FormatAddressAsLink(address []byte, name string, verified bool, isContract bool) template.HTML {
	ret := ""
	name = template.HTMLEscapeString(name)

	if len(name) > 0 {
		if verified {
			ret = fmt.Sprintf("<a class=\"text-monospace\" href=\"/execution/address/0x%x\">✔ %s (0x%x…%x)</a> %v", address, name, address[:2], address[len(address)-2:], CopyButton(hex.EncodeToString(address)))
		} else {
			ret = fmt.Sprintf("<a class=\"text-monospace\" href=\"/execution/address/0x%x\">%s 0x%x…%x</a> %v", address, name, address[:2], address[len(address)-2:], CopyButton(hex.EncodeToString(address)))
		}
	} else {
		ret = fmt.Sprintf("<a class=\"text-monospace\" href=\"/execution/address/0x%x\">0x%x…%x</a> %v", address, address[:2], address[len(address)-2:], CopyButton(hex.EncodeToString(address)))
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
			ret = fmt.Sprintf("<a class=\"text-monospace\" href=\"/execution/token/0x%x?a=0x%x\">✔ %s (0x%x…%x)</a> %v", token, address, name, address[:2], address[len(address)-2:], CopyButton(hex.EncodeToString(address)))
		} else {
			ret = fmt.Sprintf("<a class=\"text-monospace\" href=\"/execution/token/0x%x?a=0x%x\">%s 0x%x…%x</a> %v", token, address, name, address[:2], address[len(address)-2:], CopyButton(hex.EncodeToString(address)))
		}
	} else {
		ret = fmt.Sprintf("<a class=\"text-monospace\" href=\"/execution/token/0x%x?a=0x%x\">0x%x…%x</a> %v", token, address, address[:2], address[len(address)-2:], CopyButton(hex.EncodeToString(address)))
	}

	if isContract {
		ret = "<i class=\"fas fa-file-contract mr-1\"></i>" + ret
	}
	return template.HTML(ret)
}

func FormatAddressLong(address string) template.HTML {
	if len(address) > 4 {
		return template.HTML(fmt.Sprintf(`<span class="text-monospace">0x<span class="text-primary">%s</span><span>%s</span><span class="text-primary">%s</span></span>`, address[:4], address[2:len(address)-4], address[len(address)-4:]))
	}

	return template.HTML(address)

}

func FormatAmount(amount *big.Int, unit string, digits int) template.HTML {
	// cssClass := "badge-success"
	amountF := new(big.Float).SetInt(amount)
	displayUnit := "Ether"
	if unit == "ETH" {
		amountF.Quo(amountF, big.NewFloat(1e18))
	} else if unit == "GWei" {
		displayUnit = "GWei"
		amountF.Quo(amountF, big.NewFloat(1e9))
		// cssClass = "badge-info"
	}

	return template.HTML(fmt.Sprintf("<span>%."+strconv.Itoa(digits)+"f %s</span>", amountF, displayUnit))
}

func FormatMethod(method string) template.HTML {
	return template.HTML(fmt.Sprintf(`<span class="badge badge-light">%s</span>`, method))
}

func FormatBlockUsage(gasUsage uint64, gasLimit uint64) template.HTML {
	percentage := uint64(0)
	if gasLimit != 0 {
		percentage = gasUsage * 100 / gasLimit
	}
	tpl := `<div>%[1]v of %[3]v gas used <small class="text-muted ml-3">(%[2]v%%)</small></div><div class="progress" style="height:5px;"><div class="progress-bar" role="progressbar" style="width: %[2]v%%;" aria-valuenow="%[2]v" aria-valuemin="0" aria-valuemax="100"></div></div>`
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
	return template.HTML(humanize.Time(t))
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

func BlockReward(blockNumber uint64) *big.Int {
	if blockNumber < 4370000 {
		return big.NewInt(5e+18)
	} else if blockNumber < 7280000 {
		return big.NewInt(3e+18)
	} else {
		return big.NewInt(2e+18)
	}

}
