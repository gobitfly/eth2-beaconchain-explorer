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

func GetPurchaseGroup(priceId string) string {
	if priceId == Config.Frontend.Stripe.Sapphire || priceId == Config.Frontend.Stripe.Emerald || priceId == Config.Frontend.Stripe.Diamond || priceId == Config.Frontend.Stripe.Iron || priceId == Config.Frontend.Stripe.Silver || priceId == Config.Frontend.Stripe.Gold || priceId == Config.Frontend.Stripe.IronYearly || priceId == Config.Frontend.Stripe.SilverYearly || priceId == Config.Frontend.Stripe.GoldYearly {
		return GROUP_API
	}
	if priceId == Config.Frontend.Stripe.Whale || priceId == Config.Frontend.Stripe.Goldfish || priceId == Config.Frontend.Stripe.Plankton || priceId == Config.Frontend.Stripe.Orca || priceId == Config.Frontend.Stripe.Dolphin || priceId == Config.Frontend.Stripe.Guppy || priceId == Config.Frontend.Stripe.OrcaYearly || priceId == Config.Frontend.Stripe.DolphinYearly || priceId == Config.Frontend.Stripe.GuppyYearly {
		return GROUP_MOBILE
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
