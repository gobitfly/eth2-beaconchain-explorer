package utils

import "time"

const GROUP_API = "api"
const GROUP_MOBILE = "mobile"
const GROUP_ADDON = "addon"

var ProductsGroups = map[string]string{
	"sapphire":             GROUP_API,
	"emerald":              GROUP_API,
	"diamond":              GROUP_API,
	"plankton":             GROUP_MOBILE,
	"goldfish":             GROUP_MOBILE,
	"whale":                GROUP_MOBILE,
	"guppy":                GROUP_MOBILE,
	"dolphin":              GROUP_MOBILE,
	"orca":                 GROUP_MOBILE,
	"guppy.yearly":         GROUP_MOBILE,
	"dolphin.yearly":       GROUP_MOBILE,
	"orca.yearly":          GROUP_MOBILE,
	"vdb_addon_1k":         GROUP_ADDON,
	"vdb_addon_1k.yearly":  GROUP_ADDON,
	"vdb_addon_10k":        GROUP_ADDON,
	"vdb_addon_10k.yearly": GROUP_ADDON,
}

var ProductsMapV1ToV2 = map[string]string{
	"plankton": "guppy",
	"goldfish": "guppy",
	"whale":    "dolphin",
}

var ProductsMapV2ToV1 = map[string]string{
	"guppy":                "goldfish",
	"dolphin":              "whale",
	"orca":                 "whale",
	"guppy.yearly":         "goldfish",
	"dolphin.yearly":       "whale",
	"orca.yearly":          "whale",
	"vdb_addon_1k":         "",
	"vdb_addon_1k.yearly":  "",
	"vdb_addon_10k":        "",
	"vdb_addon_10k.yearly": "",
}

func GetPurchaseGroup(priceId string) string {
	productId := PriceIdToProductId(priceId)
	if group, exists := ProductsGroups[productId]; exists {
		return group
	}
	return ""
}

func EffectiveProductId(productId string) string {
	if Config.Frontend.OldProductsDeadlineUnix > 0 && time.Now().Unix() > Config.Frontend.OldProductsDeadlineUnix {
		return MapProductV1ToV2(productId)
	}
	return productId
}

func EffectiveProductName(productId string) string {
	productId = EffectiveProductId(productId)
	switch productId {
	case "plankton":
		return "Plankton"
	case "goldfish":
		return "Goldfish"
	case "whale":
		return "Whale"
	case "guppy":
		return "Guppy"
	case "dolphin":
		return "Dolphin"
	case "orca":
		return "Orca"
	case "guppy.yearly":
		return "Guppy (yearly)"
	case "dolphin.yearly":
		return "Dolphin (yearly)"
	case "orca.yearly":
		return "Orca (yearly)"
	default:
		return ""
	}
}

func ProductIsEligibleForAddons(productId string) bool {
	switch productId {
	case "orca", "orca.yearly":
		return true
	default:
		return false
	}
}

func PriceIdToProductId(priceId string) string {
	switch priceId {
	case Config.Frontend.Stripe.Sapphire:
		return "sapphire"
	case Config.Frontend.Stripe.Emerald:
		return "emerald"
	case Config.Frontend.Stripe.Diamond:
		return "diamond"
	case Config.Frontend.Stripe.Plankton:
		return "plankton"
	case Config.Frontend.Stripe.Goldfish:
		return "goldfish"
	case Config.Frontend.Stripe.Whale:
		return "whale"
	case Config.Frontend.Stripe.Guppy:
		return "guppy"
	case Config.Frontend.Stripe.Dolphin:
		return "dolphin"
	case Config.Frontend.Stripe.Orca:
		return "orca"
	case Config.Frontend.Stripe.GuppyYearly:
		return "guppy.yearly"
	case Config.Frontend.Stripe.DolphinYearly:
		return "dolphin.yearly"
	case Config.Frontend.Stripe.OrcaYearly:
		return "orca.yearly"
	case Config.Frontend.Stripe.VdbAddon1k:
		return "vdb_addon_1k"
	case Config.Frontend.Stripe.VdbAddon1kYearly:
		return "vdb_addon_1k.yearly"
	case Config.Frontend.Stripe.VdbAddon10k:
		return "vdb_addon_10k"
	case Config.Frontend.Stripe.VdbAddon10kYearly:
		return "vdb_addon_10k.yearly"
	default:
		return ""
	}
}

func MapProductV1ToV2(product string) string {
	if v, exists := ProductsMapV1ToV2[product]; exists {
		return v
	}
	if _, exists := ProductsMapV2ToV1[product]; exists {
		// just return it if its a v2 product
		return product
	}
	return ""
}

func MapProductV2ToV1(product string) string {
	if v, exists := ProductsMapV2ToV1[product]; exists {
		return v
	}
	if _, exists := ProductsMapV1ToV2[product]; exists {
		// just return it if its a v1 product
		return product
	}
	return ""
}
