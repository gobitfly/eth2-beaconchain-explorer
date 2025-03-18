package handlers

import (
	"net/http"

	"github.com/gobitfly/eth2-beaconchain-explorer/templates"
	"github.com/gobitfly/eth2-beaconchain-explorer/types"
	"github.com/gobitfly/eth2-beaconchain-explorer/utils"
)

func ApiDocs(w http.ResponseWriter, r *http.Request) {
	templateFiles := append(layoutTemplateFiles, "api_docs.html")
	var advertisewithusTemplate = templates.GetTemplate(templateFiles...)

	w.Header().Set("Content-Type", "text/html")

	data := InitPageData(w, r, "api_docs", "/api_docs", "API Documentation", templateFiles)

	pageData := &types.AdvertiseWithUsPageData{}
	pageData.RecaptchaKey = utils.Config.Frontend.RecaptchaSiteKey

	data.Data = pageData
	if handleTemplateError(w, r, "api_docs.go.go", "AdvertiseWithUs", "", advertisewithusTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}
