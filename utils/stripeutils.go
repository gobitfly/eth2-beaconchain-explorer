package utils

const GROUP_API = "api"
const GROUP_MOBILE = "mobile"

func GetPurchaseGroup(priceId string) string {
	if priceId == Config.Frontend.Stripe.Sapphire || priceId == Config.Frontend.Stripe.Emerald || priceId == Config.Frontend.Stripe.Diamond {
		return GROUP_API
	}
	if priceId == Config.Frontend.Stripe.Whale || priceId == Config.Frontend.Stripe.Goldfish || priceId == Config.Frontend.Stripe.Plankton {
		return GROUP_MOBILE
	}
	return ""
}
