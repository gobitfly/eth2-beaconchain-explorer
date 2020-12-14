package handlers

import (
	"eth2-exporter/db"
	"eth2-exporter/mail"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"github.com/gorilla/csrf"
	"html/template"
	"net/http"
)

var pricingTemplate = template.Must(template.New("pricing").Funcs(utils.GetTemplateFuncs()).ParseFiles(
	"templates/layout.html",
	"templates/payment/pricing.html",
	"templates/svg/pricing.html",
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
	pageData.CsrfField = csrf.TemplateField(r)

	pageData.User = data.User
	pageData.FlashMessage, err = utils.GetFlash(w, r, "pricing_flash")
	if err != nil {
		logger.Errorf("error retrieving flashes for advertisewithusform %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	if data.User.Authenticated {
		subscription, err := db.GetUserSubscription(data.User.UserID)
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

func PricingPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		logger.Errorf("error parsing form: %v", err)
		utils.SetFlash(w, r, "pricing_flash", "Error: invalid form submitted")
		http.Redirect(w, r, "/pricing", http.StatusSeeOther)
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

	err = mail.SendMail("support@beaconcha.in", "New API usage inquiry", msg)
	if err != nil {
		logger.Errorf("error sending ad form: %v", err)
		utils.SetFlash(w, r, "pricing_flash", "Error: unable to submit api request")
		http.Redirect(w, r, "/pricing", http.StatusSeeOther)
		return
	}

	utils.SetFlash(w, r, "pricing_flash", "Thank you for your inquiry, we will get back to you as soon as possible.")
	http.Redirect(w, r, "/pricing", http.StatusSeeOther)
}

// page called when the checkout succeeds
func PricingSuccess(w http.ResponseWriter, r *http.Request) {
	var err error

	w.Header().Set("Content-Type", "text/html")
	data := InitPageData(w, r, "pricing", "/pricing", "API Pricing")

	pageData := &types.ApiPricing{}
	pageData.User = data.User
	pageData.FlashMessage, err = utils.GetFlash(w, r, "pricing_flash")
	if err != nil {
		logger.Errorf("error retrieving flashes for advertisewithusform %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
	data.Data = pageData

	err = successTemplate.ExecuteTemplate(w, "layout", data)
	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}

// PricingCancled page called when the checkout is calcned
func PricingCancled(w http.ResponseWriter, r *http.Request) {
	var err error

	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "pricing", "/pricing", "API Pricing")

	pageData := &types.ApiPricing{}
	pageData.User = data.User
	pageData.FlashMessage, err = utils.GetFlash(w, r, "pricing_flash")
	if err != nil {
		logger.Errorf("error retrieving flashes for advertisewithusform %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
	data.Data = pageData

	err = cancelTemplate.ExecuteTemplate(w, "layout", data)
	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}
