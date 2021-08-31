package handlers

import (
	"database/sql"
	"eth2-exporter/db"
	"eth2-exporter/mail"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"html/template"
	"net/http"

	"github.com/gorilla/csrf"
)

var pricingTemplate = template.Must(template.New("pricing").Funcs(utils.GetTemplateFuncs()).ParseFiles(
	"templates/layout.html",
	"templates/payment/pricing.html",
	"templates/svg/pricing.html",
))

var mobilePricingTemplate = template.Must(template.New("mobilepricing").Funcs(utils.GetTemplateFuncs()).ParseFiles(
	"templates/layout.html",
	"templates/payment/mobilepricing.html",
	"templates/svg/mobilepricing.html",
))

var successTemplate = template.Must(template.New("success").Funcs(utils.GetTemplateFuncs()).ParseFiles(
	"templates/layout.html",
	"templates/payment/success.html",
))

var cancelTemplate = template.Must(template.New("cancled").Funcs(utils.GetTemplateFuncs()).ParseFiles(
	"templates/layout.html",
	"templates/payment/cancled.html",
))

func Pricing(w http.ResponseWriter, r *http.Request) {
	var err error

	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "pricing", "/pricing", "API Pricing")

	pageData := &types.ApiPricing{}
	pageData.RecaptchaKey = utils.Config.Frontend.RecaptchaSiteKey
	pageData.CsrfField = csrf.TemplateField(r)

	pageData.User = data.User
	pageData.FlashMessage, err = utils.GetFlash(w, r, "pricing_flash")
	if err != nil {
		logger.Errorf("error retrieving flashes for advertisewithusform %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	if data.User.Authenticated {
		subscription, err := db.StripeGetUserSubscription(data.User.UserID, utils.GROUP_API)
		if err != nil {
			logger.Errorf("error retrieving user subscriptions %v", err)
			http.Error(w, "Internal server error", 503)
			return
		}
		pageData.Subscription = subscription
	}

	pageData.StripePK = utils.Config.Frontend.Stripe.PublicKey
	pageData.Sapphire = utils.Config.Frontend.Stripe.Sapphire
	pageData.Emerald = utils.Config.Frontend.Stripe.Emerald
	pageData.Diamond = utils.Config.Frontend.Stripe.Diamond

	data.Data = pageData

	err = pricingTemplate.ExecuteTemplate(w, "layout", data)
	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}

func MobilePricing(w http.ResponseWriter, r *http.Request) {
	var err error

	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "premium", "/premium", "Premium Pricing")

	pageData := &types.MobilePricing{}
	pageData.RecaptchaKey = utils.Config.Frontend.RecaptchaSiteKey
	pageData.CsrfField = csrf.TemplateField(r)

	pageData.User = data.User
	pageData.FlashMessage, err = utils.GetFlash(w, r, "pricing_flash")
	if err != nil {
		logger.Errorf("error retrieving flashes for advertisewithusform %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	if data.User.Authenticated {
		subscription, err := db.StripeGetUserSubscription(data.User.UserID, utils.GROUP_MOBILE)
		if err != nil {
			logger.Errorf("error retrieving user subscriptions %v", err)
			http.Error(w, "Internal server error", 503)
			return
		}
		pageData.Subscription = subscription

		premiumSubscription, err := db.GetUserPremiumSubscription(data.User.UserID)
		if err != nil && err != sql.ErrNoRows {
			logger.Errorf("error retrieving user subscriptions %v", err)
			http.Error(w, "Internal server error", 503)
			return
		}
		pageData.ActiveMobileStoreSub = premiumSubscription.Active
	}

	pageData.StripePK = utils.Config.Frontend.Stripe.PublicKey
	pageData.Plankton = utils.Config.Frontend.Stripe.Plankton
	pageData.Goldfish = utils.Config.Frontend.Stripe.Goldfish
	pageData.Whale = utils.Config.Frontend.Stripe.Whale

	data.Data = pageData

	err = mobilePricingTemplate.ExecuteTemplate(w, "layout", data)
	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}

// PricingPost sends an email for a user request for an api subscription
func PricingPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		logger.Errorf("error parsing form: %v", err)
		utils.SetFlash(w, r, "pricing_flash", "Error: invalid form submitted")
		logger.Errorf("error parsing pricing request form for %v route: %v", r.URL.String(), err)
		http.Redirect(w, r, "/pricing", http.StatusSeeOther)
		return
	}

	if len(utils.Config.Frontend.RecaptchaSecretKey) > 0 && len(utils.Config.Frontend.RecaptchaSiteKey) > 0 {
		if len(r.FormValue("g-recaptcha-response")) == 0 {
			utils.SetFlash(w, r, "pricing_flash", "Error: Failed to create request")
			logger.Errorf("error no recaptca response present %v route: %v", r.URL.String(), r.FormValue("g-recaptcha-response"))
			http.Redirect(w, r, "/pricing", http.StatusSeeOther)
			return
		}

		valid, err := utils.ValidateReCAPTCHA(r.FormValue("g-recaptcha-response"))
		if err != nil || !valid {
			utils.SetFlash(w, r, "pricing_flash", "Error: Failed to create request")
			logger.Errorf("error validating recaptcha %v route: %v", r.URL.String(), err)
			http.Redirect(w, r, "/pricing", http.StatusSeeOther)
			return
		}
	}

	name := r.FormValue("name")
	email := r.FormValue("email")
	url := r.FormValue("url")
	company := r.FormValue("company")
	plan := r.FormValue("plan")
	comments := r.FormValue("comments")

	msg := fmt.Sprintf(`New API usage inquiry:
	Name: %s
	Email: %s
	Url: %s
	Company: %s
	Interested in plan: %s
	Comments: %s`, name, email, url, company, plan, comments)
	// escape html
	msg = template.HTMLEscapeString(msg)

	err = mail.SendMail("support@beaconcha.in", "New API usage inquiry", msg, []types.EmailAttachment{})
	if err != nil {
		logger.Errorf("error sending ad form: %v", err)
		utils.SetFlash(w, r, "pricing_flash", "Error: unable to submit api request")
		http.Redirect(w, r, "/pricing", http.StatusSeeOther)
		return
	}

	utils.SetFlash(w, r, "pricing_flash", "Thank you for your inquiry, we will get back to you as soon as possible.")
	http.Redirect(w, r, "/pricing", http.StatusSeeOther)
}
