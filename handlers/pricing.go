package handlers

import (
	"database/sql"
	"eth2-exporter/db"
	"eth2-exporter/mail"
	"eth2-exporter/templates"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"html/template"
	"net/http"

	"github.com/gorilla/csrf"
)

func Pricing(w http.ResponseWriter, r *http.Request) {
	templateFiles := append(layoutTemplateFiles, "payment/pricing.html", "svg/pricing.html")
	var pricingTemplate = templates.GetTemplate(templateFiles...)
	var err error

	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "pricing", "/pricing", "API Pricing", templateFiles)

	pageData := &types.ApiPricing{}
	pageData.RecaptchaKey = utils.Config.Frontend.RecaptchaSiteKey
	pageData.CsrfField = csrf.TemplateField(r)

	pageData.User = data.User
	pageData.FlashMessage, err = utils.GetFlash(w, r, "pricing_flash")
	if err != nil {
		logger.Errorf("error retrieving flashes for advertisewithusform %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if data.User.Authenticated {
		subscription, err := db.StripeGetUserSubscription(data.User.UserID, utils.GROUP_API)
		if err != nil {
			logger.Errorf("error retrieving user subscriptions %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		pageData.Subscription = subscription
	}

	pageData.StripePK = utils.Config.Frontend.Stripe.PublicKey
	pageData.Sapphire = utils.Config.Frontend.Stripe.Sapphire
	pageData.Emerald = utils.Config.Frontend.Stripe.Emerald
	pageData.Diamond = utils.Config.Frontend.Stripe.Diamond

	data.Data = pageData

	if handleTemplateError(w, r, "pricing.go", "Pricing", "", pricingTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}

func MobilePricing(w http.ResponseWriter, r *http.Request) {

	templateFiles := append(layoutTemplateFiles, "payment/mobilepricing.html", "svg/mobilepricing.html")
	var mobilePricingTemplate = templates.GetTemplate(templateFiles...)

	var err error

	w.Header().Set("Content-Type", "text/html")
	data := InitPageData(w, r, "premium", "/premium", "Premium Pricing", templateFiles)

	pageData := &types.MobilePricing{}
	pageData.RecaptchaKey = utils.Config.Frontend.RecaptchaSiteKey
	pageData.CsrfField = csrf.TemplateField(r)

	pageData.User = data.User
	pageData.FlashMessage, err = utils.GetFlash(w, r, "pricing_flash")
	if err != nil {
		logger.Errorf("error retrieving flashes for advertisewithusform %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if data.User.Authenticated {
		subscription, err := db.StripeGetUserSubscription(data.User.UserID, utils.GROUP_MOBILE)
		if err != nil {
			logger.Errorf("error retrieving user subscriptions %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		pageData.Subscription = subscription

		premiumSubscription, err := db.GetUserPremiumSubscription(data.User.UserID)
		if err != nil && err != sql.ErrNoRows {
			logger.Errorf("error retrieving user subscriptions %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		pageData.ActiveMobileStoreSub = premiumSubscription.Active
	}

	pageData.StripePK = utils.Config.Frontend.Stripe.PublicKey
	pageData.Plankton = utils.Config.Frontend.Stripe.Plankton
	pageData.Goldfish = utils.Config.Frontend.Stripe.Goldfish
	pageData.Whale = utils.Config.Frontend.Stripe.Whale

	data.Data = pageData

	if handleTemplateError(w, r, "pricing.go", "MobilePricing", "", mobilePricingTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}

// PricingPost sends an email for a user request for an api subscription
func PricingPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		utils.LogError(err, "error parsing form for pricing request", 0, map[string]interface{}{
			"route": r.URL.String(),
		})
		utils.SetFlash(w, r, "pricing_flash", "Error: invalid form submitted")
		http.Redirect(w, r, "/pricing", http.StatusSeeOther)
		return
	}

	if err := utils.HandleRecaptcha(w, r, "/pricing"); err != nil {
		return
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

	err = mail.SendTextMail(utils.Config.Frontend.Mail.Contact.InquiryEmail, "New API usage inquiry", msg, []types.EmailAttachment{})
	if err != nil {
		logger.Errorf("error sending ad form: %v", err)
		utils.SetFlash(w, r, "pricing_flash", "Error: unable to submit api request")
		http.Redirect(w, r, "/pricing", http.StatusSeeOther)
		return
	}

	utils.SetFlash(w, r, "pricing_flash", "Thank you for your inquiry, we will get back to you as soon as possible.")
	http.Redirect(w, r, "/pricing", http.StatusSeeOther)
}
