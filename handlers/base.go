package handlers

import (
	"encoding/json"
	"eth2-exporter/utils"
	"fmt"
	"net/http"

	"github.com/gorilla/context"
)

/**
base.go provides neat little helper functions for creating more abstract handlers.
By using these, you can write handlers that are able to handle both web and json api calls.

To migrate/write such handlers, replace X with Y:
	w.Header.Set()                     -> setAutoContentType()
	FormValue()	                       -> FormValueOrJSON()
	http.Error()                       -> ErrorOrJSONResponse()
	utils.SetFlash(); http.Redirect()  -> FlashRedirectOrJSONErrorResponse()
	http.Redirect()                    -> RedirectOrJSONOKResponse()
	w.WriteHeader(http.StatusOK)       -> OKResponse()

	getUser() handles auth accordingly, you can use it either way without changes.

Dont forget to register the route and set middleware accordingly:
	web auth   -> UserAuthMiddleware        -> authRouter
	api auth   -> AuthorizedAPIMiddleware   -> apiV1AuthRouter
*/

// IsMobileAuth false for web requests OR true for API (user authorized with authorization token)
func IsMobileAuth(r *http.Request) bool {
	mobile := context.Get(r, utils.MobileAuthorizedKey)
	if mobile == nil {
		return false
	}
	return mobile.(bool)
}

// SetAutoContentType text/html for web OR application/json for API
func SetAutoContentType(w http.ResponseWriter, r *http.Request) {
	if IsMobileAuth(r) {
		w.Header().Set("Content-Type", "application/json")
	} else {
		w.Header().Set("Content-Type", "text/html")
	}
}

// ErrorOrJSONResponse http.Error for web OR json error for API
func ErrorOrJSONResponse(w http.ResponseWriter, r *http.Request, errorText string, statusCode int) {
	if IsMobileAuth(r) {
		w.WriteHeader(statusCode)
		SendBadRequestResponse(w, r.URL.String(), errorText)
	} else {
		http.Error(w, errorText, statusCode)
	}
}

// FormValueOrJSON FormValue for web OR json value for API
func FormValueOrJSON(r *http.Request, key string) string {
	if IsMobileAuth(r) {
		jsonBody := context.Get(r, utils.JsonBodyKey)
		if jsonBody == nil {
			return ""
		}
		// In order to be consistent with FormValues string only return
		// (and to be able to pass non string values in json)
		// we convert every param to string. Handling/casting to
		// correct type is therefore escalated to handler
		value := jsonBody.(map[string]interface{})[key]
		if value == nil {
			return ""
		}
		return fmt.Sprintf("%v", value)
	}

	return r.FormValue(key)
}

// FlashRedirectOrJSONErrorResponse Set a flash message and redirect for web OR send an json error response with
// value as its data for API
func FlashRedirectOrJSONErrorResponse(w http.ResponseWriter, r *http.Request, name, value, url string, code int) {
	if IsMobileAuth(r) {
		ErrorOrJSONResponse(w, r, value, http.StatusBadRequest)
	} else {
		utils.SetFlash(w, r, name, value)
		http.Redirect(w, r, url, code)
	}
}

// RedirectOrJSONOKResponse Redirect for web OR send an OK json response with empty data for API
func RedirectOrJSONOKResponse(w http.ResponseWriter, r *http.Request, url string, code int) {
	if IsMobileAuth(r) {
		j := json.NewEncoder(w)
		w.WriteHeader(http.StatusOK)
		SendOKResponse(j, r.URL.String(), nil)
	} else {
		http.Redirect(w, r, url, code)
	}
}

// OKResponse writeHeader(200) for web OR writeHeader(200) + empty json OK response
func OKResponse(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	if IsMobileAuth(r) {
		j := json.NewEncoder(w)
		SendOKResponse(j, r.URL.String(), nil)
	}
}
