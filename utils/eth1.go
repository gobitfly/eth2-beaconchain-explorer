package utils

import (
	"encoding/base64"
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
	return template.HTML(fmt.Sprintf("<a href=\"/block/%[1]d\">%[1]d</a>", number))
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

func FormatBlockHash(hash string) template.HTML {
	hash = strings.Replace(hash, "0x", "", -1)

	if len(hash) < 32 {
		return "N/A"
	}

	return template.HTML(fmt.Sprintf("<a href=\"/block/0x%s\">%s…%s</a>", hash, hash[:4], hash[28:]))
}

func FormatAddressAsLink(address string, name string, verified bool, isContract bool, length int) template.HTML {
	address = strings.Replace(address, "0x", "", -1)
	ret := ""
	name = template.HTMLEscapeString(name)

	if len(name) > 0 {
		if verified {
			ret = fmt.Sprintf("<a class=\"text-monospace\" href=\"/account/0x%s\">✔ %s (0x%s…)</a>", address, name, address[:4])
		} else {
			ret = fmt.Sprintf("<a class=\"text-monospace\" href=\"/account/0x%s\">%s 0x%s…</a>", address, name, address[:4])
		}
	} else {
		if length >= 40 {
			ret = fmt.Sprintf("<a class=\"text-monospace\" href=\"/account/0x%s\">0x%s</a>", address, address[:length])
		} else {
			ret = fmt.Sprintf("<a class=\"text-monospace\" href=\"/account/0x%s\">0x%s…</a>", address, address[:length])
		}
	}

	if isContract {
		ret = "<i class=\"fal fa-file-contract mr-1\"></i>" + ret
	}
	return template.HTML(ret)
}

func FormatAmount(amount float64, unit string, digits int) template.HTML {
	cssClass := "badge-success"
	if unit == "ETH" {
		amount = amount / 1e18
	} else if unit == "GWei" {
		amount = amount / 1e9
		cssClass = "badge-info"
	}
	return template.HTML(fmt.Sprintf("<span class=\"badge %s\">%."+strconv.Itoa(digits)+"f %s</span>", cssClass, amount, unit))
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

func FormatDifficulty(number float64) string {
	return fmt.Sprintf("%.1f T", number/1e12)
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
