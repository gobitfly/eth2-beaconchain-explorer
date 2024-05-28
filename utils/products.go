package utils

import "time"

var ProductsMapV1ToV2 = map[string]string{
	"plankton": "guppy",
	"goldfish": "guppy",
	"whale":    "dolphin",
}

var ProductsMapV2ToV1 = map[string]string{
	"guppy":          "goldfish",
	"dolphin":        "whale",
	"orca":           "whale",
	"guppy.yearly":   "goldfish",
	"dolphin.yearly": "whale",
	"orca.yearly":    "whale",
}

const GROUP_API = "api"
const GROUP_MOBILE = "mobile"
const GROUP_ADDON = "addon"

func GetPurchaseGroup(priceId string) string {
	switch priceId {
	case Config.Frontend.Stripe.Sapphire, Config.Frontend.Stripe.Emerald, Config.Frontend.Stripe.Diamond, Config.Frontend.Stripe.Iron, Config.Frontend.Stripe.Silver, Config.Frontend.Stripe.Gold, Config.Frontend.Stripe.IronYearly, Config.Frontend.Stripe.SilverYearly, Config.Frontend.Stripe.GoldYearly:
		return GROUP_API
	case Config.Frontend.Stripe.Whale, Config.Frontend.Stripe.Goldfish, Config.Frontend.Stripe.Plankton, Config.Frontend.Stripe.Orca, Config.Frontend.Stripe.Dolphin, Config.Frontend.Stripe.Guppy, Config.Frontend.Stripe.OrcaYearly, Config.Frontend.Stripe.DolphinYearly, Config.Frontend.Stripe.GuppyYearly:
		return GROUP_MOBILE
	case Config.Frontend.Stripe.VdbAddon1k, Config.Frontend.Stripe.VdbAddon1kYearly, Config.Frontend.Stripe.VdbAddon10k, Config.Frontend.Stripe.VdbAddon10kYearly:
		return GROUP_ADDON
	default:
		return ""
	}
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

func PriceIdToProductId(priceId string) string {
	switch priceId {
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
