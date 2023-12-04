package handlers

import (
	"eth2-exporter/mail"
	"eth2-exporter/templates"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"html/template"
	"net/http"
)

func AdvertiseWithUs(w http.ResponseWriter, r *http.Request) {
	templateFiles := append(layoutTemplateFiles, "advertisewithus.html")
	var advertisewithusTemplate = templates.GetTemplate(templateFiles...)

	var err error

	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "advertisewithus", "/advertisewithus", "Adverstise With Us", templateFiles)

	pageData := &types.AdvertiseWithUsPageData{}
	pageData.RecaptchaKey = utils.Config.Frontend.RecaptchaSiteKey

	pageData.FlashMessage, err = utils.GetFlash(w, r, "ad_flash")
	if err != nil {
		logger.Errorf("error retrieving flashes for advertisewithusform %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data.Data = pageData
	if handleTemplateError(w, r, "advertisewithus.go", "AdvertiseWithUs", "", advertisewithusTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}

func AdvertiseWithUsPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		utils.LogError(err, "error parsing ad request form", 0, map[string]interface{}{
			"route": r.URL.String(),
		})
		utils.SetFlash(w, r, "ad_flash", "Error: invalid form submitted")
		http.Redirect(w, r, "/advertisewithus", http.StatusSeeOther)
		return
	}

	if err := utils.HandleRecaptcha(w, r, "/advertisewithus"); err != nil {
		return
	}

	name := r.FormValue("name")
	email := r.FormValue("email")
	url := r.FormValue("url")
	company := r.FormValue("company")
	ad := r.FormValue("ad")
	comments := r.FormValue("comments")

	msg := fmt.Sprintf(`New ad inquiry:
								Name: %s
								Email: %s
								Url: %s
								Company: %s
								Ad: %s
								Comments: %s`, name, email, url, company, ad, comments)
	// escape html
	msg = template.HTMLEscapeString(msg)

	err = mail.SendTextMail(utils.Config.Frontend.Mail.Contact.InquiryEmail, "New ad inquiry", msg, []types.EmailAttachment{})
	if err != nil {
		logger.Errorf("error sending ad form: %v", err)
		utils.SetFlash(w, r, "ad_flash", "Error: unable to submit ad request")
		http.Redirect(w, r, "/advertisewithus", http.StatusSeeOther)
		return
	}

	utils.SetFlash(w, r, "ad_flash", "Thank you for your inquiry, we will get back to you as soon as possible.")
	http.Redirect(w, r, "/advertisewithus", http.StatusSeeOther)
}
