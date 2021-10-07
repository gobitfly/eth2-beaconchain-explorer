package handlers

import (
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"eth2-exporter/db"
	"eth2-exporter/mail"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/context"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

var userTemplate = template.Must(template.New("user").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/user/settings.html"))
var notificationTemplate = template.Must(template.New("user").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/user/notifications.html"))
var notificationsCenterTemplate = template.Must(template.New("user").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/user/notificationsCenter.html"))
var authorizeTemplate = template.Must(template.New("user").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/user/authorize.html"))

func UserAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := getUser(r)
		if !user.Authenticated {
			logger.Errorf("User not authorized")
			utils.SetFlash(w, r, authSessionName, "Error: Please login first")
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// UserSettings renders the user-template
func UserSettings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	userSettingsData := &types.UserSettingsPageData{}

	user, session, err := getUserSession(r)
	if err != nil {
		logger.Errorf("error retrieving session: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	premiumSubscription, err := db.GetUserPremiumSubscription(user.UserID)
	if err != nil && err != sql.ErrNoRows {
		logger.Errorf("Error retrieving the premium subscriptions for user: %v %v", user.UserID, err)
		session.Flashes("Error: Something went wrong.")
		session.Save(r, w)
		http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
		return
	}

	subscription, err := db.StripeGetUserSubscription(user.UserID, utils.GROUP_API)
	if err != nil && err != sql.ErrNoRows {
		logger.Errorf("Error retrieving the subscriptions for user: %v %v", user.UserID, err)
		session.Flashes("Error: Something went wrong.")
		session.Save(r, w)
		http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
		return
	}

	var pairedDevices []types.PairedDevice = nil
	pairedDevices, err = db.GetUserDevicesByUserID(user.UserID)
	if err != nil && err != sql.ErrNoRows {
		logger.Errorf("Error retrieving the paired devices for user: %v %v", user.UserID, err)
		pairedDevices = nil
	}
	statsSharing, err := db.GetUserMonitorSharingSetting(user.UserID)
	if err != nil {
		logger.Errorf("Error retrieving stats sharing setting: %v %v", user.UserID, err)
		statsSharing = false
	}

	maxDaily := 10000
	maxMonthly := 30000
	if subscription.PriceID != nil {
		if *subscription.PriceID == utils.Config.Frontend.Stripe.Sapphire {
			maxDaily = 100000
			maxMonthly = 500000
		} else if *subscription.PriceID == utils.Config.Frontend.Stripe.Emerald {
			maxDaily = 200000
			maxMonthly = 1000000
		} else if *subscription.PriceID == utils.Config.Frontend.Stripe.Diamond {
			maxDaily = -1
			maxMonthly = 4000000
		}
	}

	userSettingsData.ApiStatistics = &types.ApiStatistics{}

	if subscription.ApiKey != nil && len(*subscription.ApiKey) > 0 {
		apiStats, err := db.GetUserAPIKeyStatistics(subscription.ApiKey)
		if err != nil {
			logger.Errorf("Error retrieving user api key usage: %v %v", user.UserID, err)
		}
		if apiStats != nil {
			userSettingsData.ApiStatistics = apiStats
		}
	}

	userSettingsData.ApiStatistics.MaxDaily = &maxDaily
	userSettingsData.ApiStatistics.MaxMonthly = &maxMonthly

	userSettingsData.PairedDevices = pairedDevices
	userSettingsData.Subscription = subscription
	userSettingsData.Premium = premiumSubscription
	userSettingsData.Sapphire = &utils.Config.Frontend.Stripe.Sapphire
	userSettingsData.Emerald = &utils.Config.Frontend.Stripe.Emerald
	userSettingsData.Diamond = &utils.Config.Frontend.Stripe.Diamond
	userSettingsData.ShareMonitoringData = statsSharing
	userSettingsData.Flashes = utils.GetFlashes(w, r, authSessionName)
	userSettingsData.CsrfField = csrf.TemplateField(r)

	data := InitPageData(w, r, "user", "/user", "User Settings")
	data.HeaderAd = true
	data.Data = userSettingsData
	data.User = user

	var premiumPkg = ""
	if premiumSubscription.Active {
		premiumPkg = premiumSubscription.Package
	}

	session.Values["subscription"] = premiumPkg
	session.Save(r, w)

	err = userTemplate.ExecuteTemplate(w, "layout", data)
	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// GenerateAPIKey generates an API key for users that do not yet have a key.
func GenerateAPIKey(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	user := getUser(r)

	err := db.CreateAPIKey(user.UserID)
	if err != nil {
		logger.WithError(err).Error("Could not create API key for user")
		http.Error(w, "Internal server error", 503)
		return
	}

	http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
}

// UserAuthorizeConfirm renders the user-authorize template
func UserAuthorizeConfirm(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	authorizeData := &types.UserAuthorizeConfirmPageData{}

	user, session, err := getUserSession(r)
	if err != nil {
		logger.Errorf("error retrieving session: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	q := r.URL.Query()
	redirectURI := q.Get("redirect_uri")
	clientID := q.Get("client_id")
	state := q.Get("state")

	session.Values["state"] = state
	session.Values["client_id"] = clientID
	session.Values["oauth_redirect_uri"] = redirectURI
	session.Save(r, w)

	if !user.Authenticated {
		logger.Errorf("User not authorized")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	appData, err := db.GetAppDataFromRedirectUri(redirectURI)

	if err != nil {
		logger.Errorf("error app not found: %v: %v: %v", user.UserID, appData, err)
		utils.SetFlash(w, r, authSessionName, "Error: App not found. Is your redirect_uri correct and registered?")
		session.Save(r, w)
	} else {
		authorizeData.AppData = appData
	}

	authorizeData.State = state
	authorizeData.CsrfField = csrf.TemplateField(r)
	authorizeData.Flashes = utils.GetFlashes(w, r, authSessionName)

	data := InitPageData(w, r, "user", "/user", "")
	data.Data = authorizeData
	data.Meta.NoTrack = true

	err = authorizeTemplate.ExecuteTemplate(w, "layout", data)
	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		callback := appData.RedirectURI + "?error=temporarily_unaviable&error_description=err_template&state=" + state
		http.Redirect(w, r, callback, http.StatusSeeOther)
		return
	}
}

// UserAuthorizationCancel cancels oauth authorization session states and redirects to frontpage
func UserAuthorizationCancel(w http.ResponseWriter, r *http.Request) {
	_, session, err := getUserSession(r)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	delete(session.Values, "oauth_redirect_uri")
	delete(session.Values, "state")
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusSeeOther)
	return
}

func UserNotifications(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	userNotificationsData := &types.UserNotificationsPageData{}

	user := getUser(r)

	userNotificationsData.Flashes = utils.GetFlashes(w, r, authSessionName)
	userNotificationsData.CsrfField = csrf.TemplateField(r)

	var watchlistIndices []uint64
	err := db.DB.Select(&watchlistIndices, `
	SELECT validators.validatorindex as index
	FROM users_validators_tags
	INNER JOIN validators
	ON
	  users_validators_tags.validator_publickey = validators.pubkey
	WHERE user_id = $1 and tag = $2
	`, user.UserID, types.ValidatorTagsWatchlist)
	if err != nil {
		logger.Errorf("error retrieving watchlist validator count %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	var countSubscriptions int
	err = db.FrontendDB.Get(&countSubscriptions, `
	SELECT count(*) as count
	FROM users_subscriptions
	WHERE user_id = $1
	`, user.UserID)
	if err != nil {
		logger.Errorf("error retrieving subscription count %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	userNotificationsData.CountSubscriptions = countSubscriptions
	userNotificationsData.WatchlistIndices = watchlistIndices
	userNotificationsData.CountWatchlist = len(watchlistIndices)
	link := "/dashboard?validators="
	for _, i := range watchlistIndices {
		link += strconv.FormatUint(i, 10) + ","
	}

	link = link[:len(link)-1]
	userNotificationsData.DashboardLink = link

	data := InitPageData(w, r, "user", "/user", "")
	data.Data = userNotificationsData
	data.User = user

	err = notificationTemplate.ExecuteTemplate(w, "layout", data)
	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func getUserMetrics(userId uint64) (interface{}, error) {
	metricsdb := struct {
		Validators         uint64 `db:"validators"`
		Notifications      uint64 `db:"notifications"`
		AttestationsMissed uint64 `db:"attestations_missed"`
		ProposalsMissed    uint64 `db:"proposals_missed"`
		ProposalsSubmitted uint64 `db:"proposals_submitted"`
	}{}
	net := strings.ToLower(utils.GetNetwork())
	err := db.FrontendDB.Get(&metricsdb, `
		SELECT COUNT(uvt.user_id) as validators,
		(SELECT COUNT(event_name) FROM users_subscriptions WHERE user_id=$1 AND last_sent_ts > NOW() - INTERVAL '1 MONTH' AND COUNT(uvt.user_id)>0) AS notifications,
		(SELECT COUNT(event_name) FROM users_subscriptions WHERE user_id=$1 AND last_sent_ts > NOW() - INTERVAL '1 MONTH' AND event_name=$2 AND COUNT(uvt.user_id)>0) AS attestations_missed,
		(SELECT COUNT(event_name) FROM users_subscriptions WHERE user_id=$1 AND last_sent_ts > NOW() - INTERVAL '1 MONTH' AND event_name=$3 AND COUNT(uvt.user_id)>0) AS proposals_missed,
		(SELECT COUNT(event_name) FROM users_subscriptions WHERE user_id=$1 AND last_sent_ts > NOW() - INTERVAL '1 MONTH' AND event_name=$4 AND COUNT(uvt.user_id)>0) AS proposals_submitted
		FROM users_validators_tags  uvt
		WHERE user_id=$1;
		`, userId, net+":"+string(types.ValidatorMissedAttestationEventName),
		net+":"+string(types.ValidatorMissedProposalEventName),
		net+":"+string(types.ValidatorExecutedProposalEventName))
	return metricsdb, err
}

func getValidatorTableData(userId uint64) (interface{}, error) {
	validatordb := []struct {
		Pubkey       string  `db:"pubkey"`
		Notification *string `db:"event_name"`
		LastSent     *uint64 `db:"last_sent_ts"`
		Threshold    *string `db:"event_threshold"`
	}{}

	err := db.FrontendDB.Select(&validatordb, `
	SELECT ENCODE(uvt.validator_publickey::bytea, 'hex') AS pubkey, us.event_name, extract( epoch from last_sent_ts)::Int as last_sent_ts, us.event_threshold 
	FROM users_validators_tags uvt 
	LEFT JOIN users_subscriptions us ON us.event_filter = ENCODE(uvt.validator_publickey::bytea, 'hex') AND us.user_id = uvt.user_id
	WHERE uvt.user_id = $1;`, userId)

	if err != nil {
		return validatordb, err
	}

	type notification struct {
		Notification string
		Timestamp    uint64
		Threshold    string
	}

	type validatorDetails struct {
		Index  uint64
		Pubkey string
	}

	type validator struct {
		Validator     validatorDetails
		Notifications []notification
	}

	result_map := map[uint64]validator{}

	for _, item := range validatordb {
		var index uint64

		err = db.DB.Get(&index, `
		SELECT validatorindex
		FROM validators WHERE pubkeyhex=$1
			`, item.Pubkey)

		if err != nil {
			return validatordb, err
		}

		if _, exists := result_map[index]; !exists {
			result_map[index] = validator{Validator: validatorDetails{Pubkey: item.Pubkey, Index: index}, Notifications: []notification{}}
		}

		if item.Notification != nil {
			map_item := result_map[index]
			var ts uint64 = 0
			if item.LastSent != nil {
				ts = *item.LastSent
			}
			map_item.Notifications = append(map_item.Notifications, notification{Notification: *item.Notification, Timestamp: ts, Threshold: *item.Threshold})
			result_map[index] = map_item
		}

	}

	valiadtors := []validator{}
	for _, item := range result_map {
		valiadtors = append(valiadtors, item)
	}

	return valiadtors, err
}

func getUserNetworkEvents(userId uint64) (interface{}, error) {
	type result struct {
		Notification string
		Network      string
		Timestamp    uint64
	}
	net := struct {
		IsSubscribed bool
		Events_ts    []result
	}{Events_ts: []result{}}

	c := 0
	err := db.FrontendDB.Get(&c, `
		SELECT count(user_id)                 
		FROM users_subscriptions      
		WHERE user_id=$1 AND event_name=$2;
	`, userId, strings.ToLower(utils.GetNetwork())+":"+string(types.NetworkLivenessIncreasedEventName))

	if c > 0 {
		net.IsSubscribed = true
		n := []uint64{}
		err = db.DB.Select(&n, `select extract( epoch from ts)::Int as ts from network_liveness where (headepoch-finalizedepoch)!=2 AND ts > now() - interval '1 year';`)

		resp := []result{}
		for _, item := range n {
			resp = append(resp, result{Notification: "Finality issue", Network: utils.Config.Chain.Network, Timestamp: item * 1000})
		}
		net.Events_ts = resp

	}

	return net, err
}

func RemoveAllValidatorsAndUnsubscribe(w http.ResponseWriter, r *http.Request) {

	SetAutoContentType(w, r) //w.Header().Set("Content-Type", "text/html")

	user := getUser(r)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logger.Errorf("error reading body of request: %v, %v", r.URL.String(), err)
		ErrorOrJSONResponse(w, r, "Internal server error", http.StatusInternalServerError)
		return
	}

	pubkeys := make([]string, 0)
	err = json.Unmarshal(body, &pubkeys)
	if err != nil {
		logger.Errorf("error parsing request body: %v, %v", r.URL.String(), err)
		ErrorOrJSONResponse(w, r, "Internal server error", http.StatusInternalServerError)
		return
	}
	for _, item := range pubkeys {
		err = db.RemoveFromWatchlist(user.UserID, item, utils.GetNetwork())
		if err != nil {
			logger.Errorf("error removing from  watchlist: %v, %v", r.URL.String(), err)
			continue
		}
	}
}

type valEvent struct {
	Event string `json:"event"`
	Email bool   `json:"email"`
	Push  bool   `json:"push"`
	Web   bool   `json:"web"`
}

func addValidatorAndSubscribe(user_id uint64, pubkey string, events []valEvent, w http.ResponseWriter, r *http.Request) bool {
	err := db.AddToWatchlist([]db.WatchlistEntry{{UserId: user_id, Validator_publickey: pubkey}}, utils.GetNetwork())
	if err != nil {
		logger.Errorf("error adding to watchlist: %v", err)
		return false
	}

	result := true
	m := map[string]bool{}
	for _, item := range events {
		if item.Email {
			if m[item.Event] && pubkey == "" {
				continue
			}

			result = result && internUserNotificationsSubscribe(item.Event, pubkey, 0, w, r)
			m[item.Event] = true
			if !result {
				break
			}
		}
	}

	return result
}

func AddValidatorsAndSubscribe(w http.ResponseWriter, r *http.Request) {

	SetAutoContentType(w, r) //w.Header().Set("Content-Type", "text/html")

	user := getUser(r)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logger.Errorf("error reading body of request: %v, %v", r.URL.String(), err)
		ErrorOrJSONResponse(w, r, "Internal server error", http.StatusInternalServerError)
		return
	}

	reqData := struct {
		Indices []uint64   `json:"indices"`
		Events  []valEvent `json:"events"`
	}{}

	// pubkeys := make([]string, 0)
	err = json.Unmarshal(body, &reqData)
	if err != nil {
		logger.Errorf("error parsing request body: %v, %v", r.URL.String(), err)
		ErrorOrJSONResponse(w, r, "Internal server error", http.StatusInternalServerError)
		return
	}

	if len(reqData.Indices) == 0 {
		logger.Errorf("error invalid indices: %v, %v", r.URL.String(), err)
		ErrorOrJSONResponse(w, r, "Internal server error", http.StatusInternalServerError)
		return
	}

	pubkeys := []string{}
	err = db.DB.Select(&pubkeys, `
		SELECT pubkeyhex FROM validators WHERE validatorindex=ANY($1);
		`, pq.Array(reqData.Indices))
	if err != nil {
		logger.Errorf("error retrieving pubkeys: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}

	result := true
	for _, pk := range pubkeys {
		result = addValidatorAndSubscribe(user.UserID, pk, reqData.Events, w, r)
	}

	if result {
		OKResponse(w, r)
	}
}

func UserUpdateSubscriptions(w http.ResponseWriter, r *http.Request) {

	SetAutoContentType(w, r) //w.Header().Set("Content-Type", "text/html")

	user := getUser(r)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logger.Errorf("error reading body of request: %v, %v", r.URL.String(), err)
		ErrorOrJSONResponse(w, r, "Internal server error", http.StatusInternalServerError)
		return
	}

	reqData := struct {
		Pubkeys []string `json:"pubkeys"`
		Events  []struct {
			Event     string  `json:"event"`
			Email     bool    `json:"email"`
			Push      bool    `json:"push"`
			Web       bool    `json:"web"`
			Threshold float64 `json:"threshold"`
		} `json:"events"`
	}{}

	// pubkeys := make([]string, 0)
	err = json.Unmarshal(body, &reqData)
	if err != nil {
		logger.Errorf("error parsing request body: %v, %v", r.URL.String(), err)
		ErrorOrJSONResponse(w, r, "Internal server error", http.StatusInternalServerError)
		return
	}

	if len(reqData.Pubkeys) == 0 {
		logger.Errorf("error invalid pubkey: %v, %v", r.URL.String(), err)
		ErrorOrJSONResponse(w, r, "Internal server error", http.StatusInternalServerError)
		return
	}

	net := strings.ToLower(utils.GetNetwork())
	pqPubkeys := pq.Array(reqData.Pubkeys)
	pqEventNames := pq.Array([]string{net + ":" + string(types.ValidatorMissedAttestationEventName),
		net + ":" + string(types.ValidatorMissedProposalEventName),
		net + ":" + string(types.ValidatorExecutedProposalEventName),
		net + ":" + string(types.ValidatorGotSlashedEventName)})

	_, err = db.FrontendDB.Exec(`
			DELETE FROM users_subscriptions WHERE user_id=$1 AND event_filter=ANY($2) AND event_name=ANY($3);
		`, user.UserID, pqPubkeys, pqEventNames)
	if err != nil {
		logger.Errorf("error removing old events: %v, %v", r.URL.String(), err)
		ErrorOrJSONResponse(w, r, "Internal server error", http.StatusInternalServerError)
		return
	}

	all_success := true
	for _, pubkey := range reqData.Pubkeys {
		result := true
		m := map[string]bool{}
		for _, item := range reqData.Events {
			if item.Email {
				if m[item.Event] && pubkey == "" {
					continue
				}

				result = result && internUserNotificationsSubscribe(item.Event, pubkey, item.Threshold, w, r)
				m[item.Event] = true
				if !result {
					break
				}
			}
		}
		if !result {
			all_success = false
			break
		}
	}

	if all_success {
		OKResponse(w, r)
	}
}

func UserUpdateMonitoringSubscriptions(w http.ResponseWriter, r *http.Request) {

	SetAutoContentType(w, r) //w.Header().Set("Content-Type", "text/html")

	user := getUser(r)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logger.Errorf("error reading body of request: %v, %v", r.URL.String(), err)
		ErrorOrJSONResponse(w, r, "Internal server error", http.StatusInternalServerError)
		return
	}

	reqData := struct {
		Pubkeys []string `json:"pubkeys"`
		Events  []struct {
			Event    string  `json:"event"`
			Email    bool    `json:"email"`
			Push     bool    `json:"push"`
			Web      bool    `json:"web"`
			Treshold float64 `json:"treshold"`
		} `json:"events"`
	}{}

	// pubkeys := make([]string, 0)
	err = json.Unmarshal(body, &reqData)
	if err != nil {
		logger.Errorf("error parsing request body: %v, %v", r.URL.String(), err)
		ErrorOrJSONResponse(w, r, "Internal server error", http.StatusInternalServerError)
		return
	}

	if len(reqData.Pubkeys) == 0 {
		logger.Errorf("error invalid pubkey: %v, %v", r.URL.String(), err)
		ErrorOrJSONResponse(w, r, "Internal server error", http.StatusInternalServerError)
		return
	}

	net := strings.ToLower(utils.GetNetwork())
	pqPubkeys := pq.Array(reqData.Pubkeys)
	pqEventNames := pq.Array([]string{net + ":" + string(types.MonitoringMachineCpuLoadEventName),
		net + ":" + string(types.MonitoringMachineDiskAlmostFullEventName),
		net + ":" + string(types.MonitoringMachineOfflineEventName),
		net + ":" + string(types.MonitoringMachineSwitchedToETH1FallbackEventName),
		net + ":" + string(types.MonitoringMachineSwitchedToETH2FallbackEventName)})

	_, err = db.FrontendDB.Exec(`
			DELETE FROM users_subscriptions WHERE user_id=$1 AND event_filter=ANY($2) AND event_name=ANY($3);
		`, user.UserID, pqPubkeys, pqEventNames)
	if err != nil {
		logger.Errorf("error removing old events: %v, %v", r.URL.String(), err)
		ErrorOrJSONResponse(w, r, "Internal server error", http.StatusInternalServerError)
		return
	}

	all_success := true
	for _, pubkey := range reqData.Pubkeys {
		result := true
		m := map[string]bool{}
		for _, item := range reqData.Events {
			if item.Email {
				if m[item.Event] && pubkey == "" {
					continue
				}

				result = result && internUserNotificationsSubscribe(item.Event, pubkey, item.Treshold, w, r)
				m[item.Event] = true
				if !result {
					break
				}
			}
		}
		if !result {
			all_success = false
			break
		}
	}

	if all_success {
		OKResponse(w, r)
	}
}

// UserNotificationsCenter renders the notificationsCenter template
func UserNotificationsCenter(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	userNotificationsCenterData := &types.UserNotificationsCenterPageData{}

	user := getUser(r)

	userNotificationsCenterData.Flashes = utils.GetFlashes(w, r, authSessionName)
	userNotificationsCenterData.CsrfField = csrf.TemplateField(r)
	//add notification-center data
	// type metrics
	metricsdb, err := getUserMetrics(user.UserID)
	if err != nil {
		logger.Errorf("error retrieving metrics data for users: %v ", user.UserID, err)
		http.Error(w, "Internal server error", 503)
		return
	}

	validatorTableData, err := getValidatorTableData(user.UserID)
	if err != nil {
		logger.Errorf("error retrieving validators table data for users: %v ", user.UserID, err)
		http.Error(w, "Internal server error", 503)
		return
	}

	networkData, err := getUserNetworkEvents(user.UserID)
	if err != nil {
		logger.Errorf("error retrieving network data for users: %v ", user.UserID, err)
		http.Error(w, "Internal server error", 503)
		return
	}

	data := InitPageData(w, r, "user", "/user", "")
	userNotificationsCenterData.Metrics = metricsdb
	userNotificationsCenterData.Validators = validatorTableData
	userNotificationsCenterData.Network = networkData
	data.Data = userNotificationsCenterData
	data.User = user

	err = notificationsCenterTemplate.ExecuteTemplate(w, "layout", data)
	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func UserNotificationsData(w http.ResponseWriter, r *http.Request) {
	currency := GetCurrency(r)
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()

	search := q.Get("search[value]")
	search = strings.Replace(search, "0x", "", -1)

	draw, err := strconv.ParseUint(q.Get("draw"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables data parameter from string to int: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
	// start, err := strconv.ParseUint(q.Get("start"), 10, 64)
	// if err != nil {
	// 	logger.Errorf("error converting datatables start parameter from string to int: %v", err)
	// 	http.Error(w, "Internal server error", 503)
	// 	return
	// }
	length, err := strconv.ParseUint(q.Get("length"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables length parameter from string to int: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
	if length > 100 {
		length = 100
	}

	user := getUser(r)

	type watchlistSubscription struct {
		Index     *uint64 // consider validators that only have deposited but do not have an index yet
		Publickey []byte
		Balance   uint64
		Events    *pq.StringArray
	}

	wl := []watchlistSubscription{}
	err = db.DB.Select(&wl, `
		SELECT 
			validators.validatorindex as index,
			users_validators_tags.validator_publickey as publickey,
			COALESCE (MAX(validators.balance), 0) as balance,
			ARRAY_REMOVE(ARRAY_AGG(users_subscriptions.event_name), NULL) as events
		FROM users_validators_tags
		LEFT JOIN users_subscriptions
			ON users_validators_tags.user_id = users_subscriptions.user_id
			AND ENCODE(users_validators_tags.validator_publickey::bytea, 'hex') = users_subscriptions.event_filter
		LEFT JOIN validators
			ON users_validators_tags.validator_publickey = validators.pubkey
		WHERE users_validators_tags.user_id = $1
		GROUP BY users_validators_tags.user_id, users_validators_tags.validator_publickey, validators.validatorindex;
		`, user.UserID)
	if err != nil {
		logger.Errorf("error retrieving subscriptions for users: %v validators: %v", user.UserID, err)
		http.Error(w, "Internal server error", 503)
		return
	}

	tableData := make([][]interface{}, 0, len(wl))
	for _, entry := range wl {
		index := template.HTML("-")
		if entry.Index != nil {
			index = utils.FormatValidator(*entry.Index)
		}
		tableData = append(tableData, []interface{}{
			index,
			utils.FormatPublicKey(entry.Publickey),
			utils.FormatBalance(entry.Balance, currency),
			entry.Events,
		})
	}

	data := &types.DataTableResponse{
		Draw:            draw,
		RecordsTotal:    uint64(len(wl)),
		RecordsFiltered: uint64(len(wl)),
		Data:            tableData,
	}

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}

}

func UserSubscriptionsData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()

	search := q.Get("search[value]")
	search = strings.Replace(search, "0x", "", -1)

	draw, err := strconv.ParseUint(q.Get("draw"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables data parameter from string to int: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
	// start, err := strconv.ParseUint(q.Get("start"), 10, 64)
	// if err != nil {
	// 	logger.Errorf("error converting datatables start parameter from string to int: %v", err)
	// 	http.Error(w, "Internal server error", 503)
	// 	return
	// }
	length, err := strconv.ParseUint(q.Get("length"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables length parameter from string to int: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
	if length > 100 {
		length = 100
	}

	user := getUser(r)

	subs := []types.Subscription{}
	err = db.FrontendDB.Select(&subs, `
			SELECT *
			FROM users_subscriptions
			WHERE user_id = $1
	`, user.UserID)
	if err != nil {
		logger.Errorf("error retrieving subscriptions for users %v: %v", user.UserID, err)
		http.Error(w, "Internal server error", 503)
		return
	}

	tableData := make([][]interface{}, 0, len(subs))
	for _, sub := range subs {
		ls := template.HTML("N/A")
		pubkey := template.HTML(sub.EventFilter)
		if sub.LastSent != nil {
			ls = utils.FormatTimestamp(sub.LastSent.Unix())
		}

		if len(sub.EventFilter) == 96 {
			h, err := hex.DecodeString(sub.EventFilter)
			if err != nil {
				logger.Errorf("Could not decode Pubkey %v", err)
			} else {
				pubkey = utils.FormatPublicKey(h)
			}
		} else if sub.EventName == types.TaxReportEventName {
			pubkey = template.HTML(`<a href="/rewards">report</a>`)
		} else if strings.HasPrefix(string(sub.EventName), "monitoring_") {
			pubkey = utils.FormatMachineName(sub.EventFilter)
		}
		if sub.EventName != types.ValidatorBalanceDecreasedEventName {
			tableData = append(tableData, []interface{}{
				pubkey,
				sub.EventName,
				utils.FormatTimestamp(sub.CreatedTime.Unix()),
				ls,
			})
		}

	}

	// log.Println("COUNT", len(watchlist))
	data := &types.DataTableResponse{
		Draw:            draw,
		RecordsTotal:    uint64(len(subs)),
		RecordsFiltered: uint64(len(subs)),
		Data:            tableData,
	}

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}

}

func UserAuthorizeConfirmPost(w http.ResponseWriter, r *http.Request) {
	logger := logger.WithField("route", r.URL.String())

	redirectURI := r.FormValue("redirect_uri")
	state := r.FormValue("state")
	var stateAppend string = ""
	if state != "" {
		stateAppend = "&state=" + state
	}

	appData, err := db.GetAppDataFromRedirectUri(redirectURI)
	if err != nil {
		logger.Errorf("error app no found: %v %v", appData, err)
		callback := redirectURI + "?error=invalid_request&error_description=missing_redirect_uri" + stateAppend
		http.Redirect(w, r, callback, http.StatusSeeOther)
		return
	}

	user, session, err := getUserSession(r)
	if err != nil {
		logger.Errorf("error retrieving session: %v", err)
		callback := appData.RedirectURI + "?error=access_denied&error_description=no_session" + stateAppend
		http.Redirect(w, r, callback, http.StatusSeeOther)
		return
	}

	if user.Authenticated == true {
		codeBytes, err1 := utils.GenerateRandomBytesSecure(32)
		if err1 != nil {
			logger.Errorf("error creating secure random bytes for user: %v %v", user.UserID, err1)
			callback := appData.RedirectURI + "?error=server_error&error_description=err_random_number" + stateAppend
			http.Redirect(w, r, callback, http.StatusSeeOther)
			return
		}

		code := hex.EncodeToString(codeBytes)   // return to user
		codeHashed := utils.HashAndEncode(code) // save hashed code in db
		clientID := session.Values["client_id"].(string)

		err2 := db.AddAuthorizeCode(user.UserID, codeHashed, clientID, appData.ID)
		if err2 != nil {
			logger.Errorf("error adding authorization code for user: %v %v", user.UserID, err2)
			callback := appData.RedirectURI + "?error=server_error&error_description=err_db_storefail" + stateAppend
			http.Redirect(w, r, callback, http.StatusSeeOther)
			return
		}

		callbackTemplate := appData.RedirectURI + "?code="

		callback := callbackTemplate + code + stateAppend
		http.Redirect(w, r, callback, http.StatusSeeOther)
		return
	} else {
		logger.Error("Not authorized")
		callback := appData.RedirectURI + "?error=access_denied&error_description=no_authentication" + stateAppend
		http.Redirect(w, r, callback, http.StatusSeeOther)
		return
	}
}

func UserDeletePost(w http.ResponseWriter, r *http.Request) {
	logger := logger.WithField("route", r.URL.String())
	user, session, err := getUserSession(r)
	if err != nil {
		logger.Errorf("error retrieving session: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if user.Authenticated == true {
		err := db.DeleteUserById(user.UserID)
		if err != nil {
			logger.Errorf("error deleting user by email for user: %v %v", user.UserID, err)
			http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
			session.Flashes("Error: Could not delete user.")
			session.Save(r, w)
			http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
			return
		}

		Logout(w, r)
	} else {
		logger.Error("Trying to delete a unauthenticated user")
		http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
		return
	}
}

func UserUpdateFlagsPost(w http.ResponseWriter, r *http.Request) {
	user, _, err := getUserSession(r)
	if err != nil {
		logger.Errorf("error retrieving session: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	shareStats := FormValueOrJSON(r, "shareStats")

	logger.Errorf("shareStats: %v", shareStats)

	err = db.SetUserMonitorSharingSetting(user.UserID, shareStats == "true")

	http.Redirect(w, r, "/user/settings#app", http.StatusOK)
}

func UserUpdatePasswordPost(w http.ResponseWriter, r *http.Request) {
	var GenericUpdatePasswordError = "Error: Something went wrong updating your password 😕. If this error persists please contact <a href=\"https://support.bitfly.at/support/home\">support</a>"

	user, session, err := getUserSession(r)
	if err != nil {
		logger.Errorf("error retrieving session: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = r.ParseForm()
	if err != nil {
		logger.Errorf("error parsing form: %v", err)
		session.AddFlash(authInternalServerErrorFlashMsg)
		session.Save(r, w)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	pwdNew := r.FormValue("password")
	pwdOld := r.FormValue("old-password")

	currentUser := struct {
		ID        int64  `db:"id"`
		Email     string `db:"email"`
		Password  string `db:"password"`
		Confirmed bool   `db:"email_confirmed"`
	}{}

	err = db.FrontendDB.Get(&currentUser, "SELECT id, email, password, email_confirmed FROM users WHERE id = $1", user.UserID)
	if err != nil {
		logger.Errorf("error retrieving password for user %v: %v", user.UserID, err)
		session.AddFlash("Error: Invalid password!")
		session.Save(r, w)
		http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
		return
	}

	if !currentUser.Confirmed {
		session.AddFlash("Error: Email has not been confirmed, please click the link in the email we sent you or <a href='/resend'>resend link</a>!")
		session.Save(r, w)
		http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(currentUser.Password), []byte(pwdOld))
	if err != nil {
		logger.Errorf("error verifying password for user %v: %v", currentUser.Email, err)
		session.AddFlash("Error: Invalid password!")
		session.Save(r, w)
		http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
		return
	}

	pHash, err := bcrypt.GenerateFromPassword([]byte(pwdNew), 10)
	if err != nil {
		logger.Errorf("error generating hash for password: %v", err)
		session.AddFlash(GenericUpdatePasswordError)
		session.Save(r, w)
		http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
		return
	}

	err = db.UpdatePassword(user.UserID, pHash)
	if err != nil {
		logger.Errorf("error updating password for user: %v", err)
		session.AddFlash("Error: Something went wrong updating your password 😕. If this error persists please contact <a href=\"https://support.bitfly.at/support/home\">support</a>")
		session.Save(r, w)
		http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
		return
	}
	session.AddFlash("Password Updated Successfully ✔️")
	session.Save(r, w)
	http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
}

// UserUpdateEmailPost gets called from the settings page to request a new email update. Only once the update link is pressed does the email actually change.
func UserUpdateEmailPost(w http.ResponseWriter, r *http.Request) {
	user, session, err := getUserSession(r)
	if err != nil {
		logger.Errorf("error retrieving session: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = r.ParseForm()
	if err != nil {
		logger.Errorf("error parsing form: %v", err)
		session.AddFlash(authInternalServerErrorFlashMsg)
		session.Save(r, w)
		http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
		return
	}
	email := r.FormValue("email")

	if !utils.IsValidEmail(email) {
		session.AddFlash("Error: Invalid email format!")
		session.Save(r, w)
		http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
		return
	}

	var existingEmails struct {
		Count int
		Email string
	}
	err = db.FrontendDB.Get(&existingEmails, "SELECT email FROM users WHERE email = $1", email)

	if existingEmails.Email == email {
		http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
		return
	} else if existingEmails.Email != "" {
		session.AddFlash("Error: Email already exists please choose a unique email")
		session.Save(r, w)
		http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
		return
	}

	var rateLimitError *types.RateLimitError
	err = sendEmailUpdateConfirmation(user.UserID, email)
	if err != nil {
		logger.Errorf("error sending confirmation-email: %v", err)
		if errors.As(err, &rateLimitError) {
			session.AddFlash(fmt.Sprintf("Error: The ratelimit for sending emails has been exceeded, please try again in %v.", err.(*types.RateLimitError).TimeLeft.Round(time.Second)))
		} else {
			session.AddFlash(authInternalServerErrorFlashMsg)
		}
		session.Save(r, w)
		http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
		return
	}

	session.AddFlash("Verification link sent to your new email " + email)
	session.Save(r, w)
	http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
}

// ConfirmUpdateEmail confirms and updates the email address of the user. Given an update link the email in the db is changed.
func UserConfirmUpdateEmail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hash := vars["hash"]

	_, session, err := getUserSession(r)
	if err != nil {
		logger.Errorf("error retrieving session: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	q := r.URL.Query()

	newEmail := q.Get("email")

	if !utils.IsValidEmail(newEmail) {
		utils.SetFlash(w, r, authSessionName, "Error: Could not update your email because the new email is invalid, please try again.")
		http.Redirect(w, r, "/confirmation", http.StatusSeeOther)
		return
	}

	user := struct {
		ID        int64     `db:"id"`
		Email     string    `db:"email"`
		ConfirmTs time.Time `db:"email_confirmation_ts"`
		Confirmed bool      `db:"email_confirmed"`
	}{}

	err = db.FrontendDB.Get(&user, "SELECT id, email, email_confirmation_ts, email_confirmed FROM users WHERE email_confirmation_hash = $1", hash)
	if err != nil {
		logger.Errorf("error retreiveing email for confirmation_hash %v %v", hash, err)
		utils.SetFlash(w, r, authSessionName, "Error: This confirmation link is invalid / outdated.")
		http.Redirect(w, r, "/confirmation", http.StatusSeeOther)
		return
	}

	if user.Confirmed != true {
		utils.SetFlash(w, r, authSessionName, "Error: Cannot update email for an unconfirmed address.")
		http.Redirect(w, r, "/confirmation", http.StatusSeeOther)
		return
	}

	if user.ConfirmTs.Add(time.Minute * 30).Before(time.Now()) {
		utils.SetFlash(w, r, authSessionName, "Error: This confirmation link has expired.")
		http.Redirect(w, r, "/confirmation", http.StatusSeeOther)
		return
	}

	var emailExists string
	err = db.FrontendDB.Get(&emailExists, "SELECT email FROM users WHERE email = $1", newEmail)
	if emailExists != "" {
		utils.SetFlash(w, r, authSessionName, "Error: Email already exists. We could not update your email.")
		http.Redirect(w, r, "/confirmation", http.StatusSeeOther)
		return
	}

	_, err = db.FrontendDB.Exec(`UPDATE users SET email = $1 WHERE id = $2`, newEmail, user.ID)
	if err != nil {
		logger.Errorf("error: updating email for user: %v", err)
		utils.SetFlash(w, r, authSessionName, "Error: Could not Update Email.")
		http.Redirect(w, r, "/confirmation", http.StatusSeeOther)
		return
	}

	session.Values["subscription"] = ""
	session.Values["authenticated"] = false
	delete(session.Values, "user_id")

	utils.SetFlash(w, r, authSessionName, "Your email has been updated successfully! <br> You can log in with your new email.")
	http.Redirect(w, r, "/confirmation", http.StatusSeeOther)
}

func sendEmailUpdateConfirmation(userId uint64, newEmail string) error {
	now := time.Now()
	emailConfirmationHash := utils.RandomString(40)

	tx, err := db.FrontendDB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var lastTs *time.Time
	err = tx.Get(&lastTs, "SELECT email_confirmation_ts FROM users WHERE id = $1", userId)
	if err != nil {
		return fmt.Errorf("error getting confirmation-ts: %w", err)
	}
	if lastTs != nil && (*lastTs).Add(authConfirmEmailRateLimit).After(now) {
		return &types.RateLimitError{(*lastTs).Add(authConfirmEmailRateLimit).Sub(now)}
	}

	_, err = tx.Exec("UPDATE users SET email_confirmation_hash = $1 WHERE id = $2", emailConfirmationHash, userId)
	if err != nil {
		return fmt.Errorf("error updating confirmation-hash: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error commiting db-tx: %w", err)
	}

	subject := fmt.Sprintf("%s: Verify your email-address", utils.Config.Frontend.SiteDomain)
	msg := fmt.Sprintf(`To update your email on %[1]s please verify it by clicking this link:

https://%[1]s/settings/email/%[2]s?email=%[3]s

Best regards,

%[1]s
`, utils.Config.Frontend.SiteDomain, emailConfirmationHash, url.QueryEscape(newEmail))
	err = mail.SendMail(newEmail, subject, msg, []types.EmailAttachment{})
	if err != nil {
		return err
	}

	_, err = db.FrontendDB.Exec("UPDATE users SET email_confirmation_ts = TO_TIMESTAMP($1) WHERE id = $2", time.Now().Unix(), userId)
	if err != nil {
		return fmt.Errorf("error updating confirmation-ts: %w", err)
	}

	return nil
}

// UserValidatorWatchlistAdd godoc
// @Summary  subscribes a user to get notifications from a specific validator
// @Tags User
// @Produce  json
// @Param pubKey query string true "Public Key of validator you want to subscribe to"
// @Param balance_decreases body string false "Submit \"on\" to enable notifications for this event"
// @Param validator_slashed body string false "Submit \"on\" to enable notifications for this event"
// @Success 200 {object} types.ApiResponse
// @Failure 400 {object} types.ApiResponse
// @Failure 500 {object} types.ApiResponse
// @Security ApiKeyAuth
// @Router /api/v1/user/validator/{pubkey}/add [post]
func UserValidatorWatchlistAdd(w http.ResponseWriter, r *http.Request) {
	SetAutoContentType(w, r)
	user := getUser(r)
	vars := mux.Vars(r)

	pubKey := strings.Replace(vars["pubkey"], "0x", "", -1)
	if !user.Authenticated {
		FlashRedirectOrJSONErrorResponse(w, r,
			validatorEditFlash,
			"Error: You need a user account to follow a validator <a href=\"/login\">Login</a> or <a href=\"/register\">Sign up</a>",
			"/validator/"+pubKey,
			http.StatusSeeOther,
		)
		return
	}

	balance := FormValueOrJSON(r, "balance_decreases")
	if balance == "on" {
		err := db.AddSubscription(user.UserID, utils.Config.Chain.Phase0.ConfigName, types.ValidatorBalanceDecreasedEventName, pubKey, 0)
		if err != nil {
			logger.Errorf("error could not ADD subscription for user %v eventName %v eventfilter %v: %v", user.UserID, types.ValidatorBalanceDecreasedEventName, pubKey, err)
			ErrorOrJSONResponse(w, r, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
	slashed := FormValueOrJSON(r, "validator_slashed")
	if slashed == "on" {
		err := db.AddSubscription(user.UserID, utils.Config.Chain.Phase0.ConfigName, types.ValidatorGotSlashedEventName, pubKey, 0)
		if err != nil {
			logger.Errorf("error could not ADD subscription for user %v eventName %v eventfilter %v: %v", user.UserID, types.ValidatorGotSlashedEventName, pubKey, err)
			ErrorOrJSONResponse(w, r, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
	proposalSubmitted := FormValueOrJSON(r, "validator_proposal_submitted")
	if proposalSubmitted == "on" {
		err := db.AddSubscription(user.UserID, utils.Config.Chain.Phase0.ConfigName, types.ValidatorExecutedProposalEventName, pubKey, 0)
		if err != nil {
			logger.Errorf("error could not ADD subscription for user %v eventName %v eventfilter %v: %v", user.UserID, types.ValidatorGotSlashedEventName, pubKey, err)
			ErrorOrJSONResponse(w, r, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
	proposalMissed := FormValueOrJSON(r, "validator_proposal_missed")
	if proposalMissed == "on" {
		err := db.AddSubscription(user.UserID, utils.Config.Chain.Phase0.ConfigName, types.ValidatorMissedProposalEventName, pubKey, 0)
		if err != nil {
			logger.Errorf("error could not ADD subscription for user %v eventName %v eventfilter %v: %v", user.UserID, types.ValidatorGotSlashedEventName, pubKey, err)
			ErrorOrJSONResponse(w, r, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
	attestationMissed := FormValueOrJSON(r, "validator_attestation_missed")
	if attestationMissed == "on" {
		err := db.AddSubscription(user.UserID, utils.Config.Chain.Phase0.ConfigName, types.ValidatorMissedAttestationEventName, pubKey, 0)
		if err != nil {
			logger.Errorf("error could not ADD subscription for user %v eventName %v eventfilter %v: %v", user.UserID, types.ValidatorGotSlashedEventName, pubKey, err)
			ErrorOrJSONResponse(w, r, "Internal server error", http.StatusInternalServerError)
			return
		}
	}

	if len(pubKey) != 96 {
		FlashRedirectOrJSONErrorResponse(w, r,
			validatorEditFlash,
			"Error: Validator not found",
			"/validator/"+pubKey,
			http.StatusSeeOther,
		)
		return
	}

	watchlistEntries := []db.WatchlistEntry{
		{
			UserId:              user.UserID,
			Validator_publickey: pubKey,
		},
	}
	err := db.AddToWatchlist(watchlistEntries, utils.GetNetwork())
	if err != nil {
		logger.Errorf("error adding validator to watchlist to db: %v", err)
		FlashRedirectOrJSONErrorResponse(w, r,
			validatorEditFlash,
			"Error: Could not follow validator.",
			"/validator/"+pubKey,
			http.StatusSeeOther,
		)
		return
	}

	RedirectOrJSONOKResponse(w, r, "/validator/"+pubKey, http.StatusSeeOther)
}

// UserDashboardWatchlistAdd godoc
// @Summary  subscribes a user to get notifications from a specific validator via index
// @Tags User
// @Produce  json
// @Param pubKey body []string true "Index of validator you want to subscribe to"
// @Success 200 {object} types.ApiResponse
// @Failure 400 {object} types.ApiResponse
// @Failure 500 {object} types.ApiResponse
// @Security ApiKeyAuth
// @Router /api/v1/user/validator/{pubkey}/remove [post]
func UserDashboardWatchlistAdd(w http.ResponseWriter, r *http.Request) {
	SetAutoContentType(w, r) //w.Header().Set("Content-Type", "text/html")
	user := getUser(r)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logger.Errorf("error reading body of request: %v, %v", r.URL.String(), err)
		ErrorOrJSONResponse(w, r, "Internal server error", http.StatusInternalServerError)
		return
	}

	indices := make([]string, 0)
	err = json.Unmarshal(body, &indices)
	if err != nil {
		logger.Errorf("error parsing request body: %v, %v", r.URL.String(), err)
		ErrorOrJSONResponse(w, r, "Internal server error", http.StatusInternalServerError)
		return
	}
	indicesParsed := make([]int64, 0)
	for _, i := range indices {
		parsed, err := strconv.ParseInt(i, 10, 64)
		if err != nil {
			logger.Errorf("error could not parse validator indices: %v, %v", r.URL.String(), err)
			ErrorOrJSONResponse(w, r, "Internal server error", http.StatusInternalServerError)
			return
		}
		indicesParsed = append(indicesParsed, parsed)
	}

	publicKeys := make([]string, 0)
	db.DB.Select(&publicKeys, `
	SELECT pubkeyhex as pubkey
	FROM validators
	WHERE validatorindex = ANY($1)
	`, pq.Int64Array(indicesParsed))

	watchListEntries := []db.WatchlistEntry{}

	for _, key := range publicKeys {
		watchListEntries = append(watchListEntries, db.WatchlistEntry{
			UserId:              user.UserID,
			Validator_publickey: key,
		})
	}
	err = db.AddToWatchlist(watchListEntries, utils.GetNetwork())
	if err != nil {
		logger.Errorf("error could not add validators to watchlist: %v, %v", r.URL.String(), err)
		ErrorOrJSONResponse(w, r, "Internal server error", http.StatusInternalServerError)
		return
	}

	OKResponse(w, r)
}

// UserValidatorWatchlistRemove godoc
// @Summary  unsubscribes a user from a specific validator
// @Tags User
// @Produce  json
// @Param pubKey query string true "Public Key of validator you want to subscribe to"
// @Success 200 {object} types.ApiResponse
// @Failure 400 {object} types.ApiResponse
// @Failure 500 {object} types.ApiResponse
// @Security ApiKeyAuth
// @Router /api/v1/user/validator/{pubkey}/remove [post]
func UserValidatorWatchlistRemove(w http.ResponseWriter, r *http.Request) {
	SetAutoContentType(w, r)

	user := getUser(r)
	vars := mux.Vars(r)

	pubKey := strings.Replace(vars["pubkey"], "0x", "", -1)
	if !user.Authenticated {
		FlashRedirectOrJSONErrorResponse(w, r,
			validatorEditFlash,
			"Error: You need a user account to follow a validator <a href=\"/login\">Login</a> or <a href=\"/register\">Sign up</a>",
			"/validator/"+pubKey,
			http.StatusSeeOther,
		)
		return
	}

	if len(pubKey) != 96 {
		FlashRedirectOrJSONErrorResponse(w, r,
			validatorEditFlash,
			"Error: Validator not found",
			"/validator/"+pubKey,
			http.StatusSeeOther,
		)
		return
	}

	err := db.RemoveFromWatchlist(user.UserID, pubKey, utils.GetNetwork())
	if err != nil {
		logger.Errorf("error deleting subscription: %v", err)
		FlashRedirectOrJSONErrorResponse(w, r,
			validatorEditFlash,
			"Error: Could not remove bookmark.",
			"/validator/"+pubKey,
			http.StatusSeeOther,
		)
		return
	}

	RedirectOrJSONOKResponse(w, r, "/validator/"+pubKey, http.StatusSeeOther)
}

func UserNotificationsSubscribe(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	event := q.Get("event")
	filter := q.Get("filter")
	thresholdString := q.Get("threshold")
	var threshold float64 = 0
	threshold, _ = strconv.ParseFloat(thresholdString, 64)

	if internUserNotificationsSubscribe(event, filter, threshold, w, r) {
		OKResponse(w, r)
	}
}

func MultipleUsersNotificationsSubscribe(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	j := json.NewEncoder(w)

	type SubIntent struct {
		EventName      string  `json:"event_name"`
		EventFilter    string  `json:"event_filter"`
		EventThreshold float64 `json:"event_threshold"`
	}

	var jsonObjects []SubIntent
	err := json.Unmarshal(context.Get(r, utils.JsonBodyNakedKey).([]byte), &jsonObjects)
	if err != nil {
		logger.Errorf("Could not parse multiple notification subscription intent | %v", err)
		sendErrorResponse(j, r.URL.String(), "could not parse request")
		return
	}

	if len(jsonObjects) > 100 {
		logger.Errorf("Max number bundle subscribe is 100", err)
		sendErrorResponse(j, r.URL.String(), "Max number bundle subscribe is 100")
		return
	}

	var result bool = true
	m := make(map[string]bool)
	for i := 0; i < len(jsonObjects); i++ {
		obj := jsonObjects[i]

		// make sure expensive operations without filter can only be done once per request
		if m[obj.EventName] && obj.EventFilter == "" {
			continue
		}

		result = result && internUserNotificationsSubscribe(obj.EventName, obj.EventFilter, obj.EventThreshold, w, r)
		m[obj.EventName] = true
		if !result {
			break
		}
	}

	if result {
		OKResponse(w, r)
	}
}

func internUserNotificationsSubscribe(event, filter string, threshold float64, w http.ResponseWriter, r *http.Request) bool {
	w.Header().Set("Content-Type", "text/html")
	user := getUser(r)
	filter = strings.Replace(filter, "0x", "", -1)

	eventName, err := types.EventNameFromString(event)
	if err != nil {
		logger.Errorf("error invalid event name: %v", err)
		ErrorOrJSONResponse(w, r, "Internal server error", http.StatusInternalServerError)
		return false
	}

	isPkey := !pkeyRegex.MatchString(filter)
	filterLen := len(filter)

	if filterLen != 96 && filterLen != 0 && isPkey {
		logger.Errorf("error invalid pubkey characters or length: %v", err)
		ErrorOrJSONResponse(w, r, "Internal server error", http.StatusInternalServerError)
		return false
	}

	userPremium := getUserPremium(r)

	if filterLen == 0 && !strings.HasPrefix(string(eventName), "monitoring_") { // no filter = add all my watched validators

		filter := db.WatchlistFilter{
			UserId:         user.UserID,
			Validators:     nil,
			Tag:            types.ValidatorTagsWatchlist,
			JoinValidators: true,
			Network:        utils.GetNetwork(),
		}

		myValidators, err2 := db.GetTaggedValidators(filter)
		if err2 != nil {
			ErrorOrJSONResponse(w, r, "could not retrieve db results", http.StatusInternalServerError)
			return false
		}

		maxValidators := userPremium.MaxValidators

		// not quite happy performance wise, placing a TODO here for future me
		for i, v := range myValidators {
			err = db.AddSubscription(user.UserID, utils.GetNetwork(), eventName, fmt.Sprintf("%v", hex.EncodeToString(v.PublicKey)), 0)
			if err != nil {
				logger.Errorf("error could not ADD subscription for user %v eventName %v eventfilter %v: %v", user.UserID, eventName, filter, err)
				ErrorOrJSONResponse(w, r, "Internal server error", http.StatusInternalServerError)
				return false
			}

			if i >= maxValidators {
				break
			}
		}
	} else { // add filtered one

		if !userPremium.NotificationThresholds {
			if eventName == types.MonitoringMachineDiskAlmostFullEventName {
				threshold = 0.1
			} else if eventName == types.MonitoringMachineCpuLoadEventName {
				threshold = 0.6
			} else if eventName == types.MonitoringMachineMemoryUsageEventName {
				threshold = 0.8
			}
		}

		err = db.AddSubscription(user.UserID, utils.GetNetwork(), eventName, filter, threshold)
		if err != nil {
			logger.Errorf("error could not ADD subscription for user %v eventName %v eventfilter %v: %v", user.UserID, eventName, filter, err)
			ErrorOrJSONResponse(w, r, "Internal server error", http.StatusInternalServerError)
			return false
		}
	}

	return true
}

func MultipleUsersNotificationsUnsubscribe(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	j := json.NewEncoder(w)

	type UnSubIntent struct {
		EventName   string `json:"event_name"`
		EventFilter string `json:"event_filter"`
	}

	var jsonObjects []UnSubIntent
	err := json.Unmarshal(context.Get(r, utils.JsonBodyNakedKey).([]byte), &jsonObjects)
	if err != nil {
		logger.Errorf("Could not parse multiple notification subscription intent | %v", err)
		sendErrorResponse(j, r.URL.String(), "could not parse request")
		return
	}

	if len(jsonObjects) > 100 {
		logger.Errorf("Max number bundle unsubscribe is 100", err)
		sendErrorResponse(j, r.URL.String(), "Max number bundle unsubscribe is 100")
		return
	}

	var result bool = true
	m := make(map[string]bool)
	for i := 0; i < len(jsonObjects); i++ {
		obj := jsonObjects[i]

		// make sure expensive operations without filter can only be done once per request
		if m[obj.EventName] && obj.EventFilter == "" {
			continue
		}

		result = result && internUserNotificationsUnsubscribe(jsonObjects[i].EventName, jsonObjects[i].EventFilter, w, r)
		m[obj.EventName] = true

		if !result {
			break
		}
	}

	if result {
		OKResponse(w, r)
	}
}

func internUserNotificationsUnsubscribe(event, filter string, w http.ResponseWriter, r *http.Request) bool {
	w.Header().Set("Content-Type", "text/html")
	user := getUser(r)

	filter = strings.Replace(filter, "0x", "", -1)

	eventName, err := types.EventNameFromString(event)
	if err != nil {
		logger.Errorf("error invalid event name: %v", err)
		ErrorOrJSONResponse(w, r, "Internal server error", http.StatusInternalServerError)
		return false
	}

	isPkey := !pkeyRegex.MatchString(filter)
	filterLen := len(filter)

	if len(filter) != 96 && filterLen != 0 && isPkey {
		logger.Errorf("error invalid pubkey characters or length: %v", err)
		ErrorOrJSONResponse(w, r, "Internal server error", http.StatusInternalServerError)
		return false
	}

	if filterLen == 0 && !strings.HasPrefix(string(eventName), "monitoring_") { // no filter = add all my watched validators

		filter := db.WatchlistFilter{
			UserId:         user.UserID,
			Validators:     nil,
			Tag:            types.ValidatorTagsWatchlist,
			JoinValidators: true,
			Network:        utils.GetNetwork(),
		}

		myValidators, err2 := db.GetTaggedValidators(filter)
		if err2 != nil {
			ErrorOrJSONResponse(w, r, "could not retrieve db results", http.StatusInternalServerError)
			return false
		}

		maxValidators := getUserPremium(r).MaxValidators
		// not quite happy performance wise, placing a TODO here for future me
		for i, v := range myValidators {
			err = db.DeleteSubscription(user.UserID, utils.GetNetwork(), eventName, fmt.Sprintf("%v", hex.EncodeToString(v.PublicKey)))
			if err != nil {
				logger.Errorf("error could not REMOVE subscription for user %v eventName %v eventfilter %v: %v", user.UserID, eventName, filter, err)
				ErrorOrJSONResponse(w, r, "Internal server error", http.StatusInternalServerError)
				return false
			}

			if i >= maxValidators {
				break
			}
		}
	} else {
		// filtered one only
		err = db.DeleteSubscription(user.UserID, utils.GetNetwork(), eventName, filter)
		if err != nil {
			logger.Errorf("error could not REMOVE subscription for user %v eventName %v eventfilter %v: %v", user.UserID, eventName, filter, err)
			ErrorOrJSONResponse(w, r, "Internal server error", http.StatusInternalServerError)
			return false
		}
	}

	return true
}

func UserNotificationsUnsubscribe(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	user := getUser(r)
	q := r.URL.Query()
	event := q.Get("event")
	filter := q.Get("filter")
	filter = strings.Replace(filter, "0x", "", -1)

	eventName, err := types.EventNameFromString(event)
	if err != nil {
		logger.Errorf("error invalid event name: %v", err)
		ErrorOrJSONResponse(w, r, "Internal server error", http.StatusInternalServerError)
		return
	}

	isPkey := !pkeyRegex.MatchString(filter)
	filterLen := len(filter)

	if len(filter) != 96 && filterLen != 0 && isPkey {
		logger.Errorf("error invalid pubkey characters or length: %v", err)
		ErrorOrJSONResponse(w, r, "Internal server error", http.StatusInternalServerError)
		return
	}

	if filterLen == 0 && !strings.HasPrefix(string(eventName), "monitoring_") { // no filter = add all my watched validators

		filter := db.WatchlistFilter{
			UserId:         user.UserID,
			Validators:     nil,
			Tag:            types.ValidatorTagsWatchlist,
			JoinValidators: true,
			Network:        utils.GetNetwork(),
		}

		myValidators, err2 := db.GetTaggedValidators(filter)
		if err2 != nil {
			ErrorOrJSONResponse(w, r, "could not retrieve db results", http.StatusInternalServerError)
			return
		}

		maxValidators := getUserPremium(r).MaxValidators

		// not quite happy performance wise, placing a TODO here for future me
		for i, v := range myValidators {
			err = db.DeleteSubscription(user.UserID, utils.GetNetwork(), eventName, fmt.Sprintf("%v", hex.EncodeToString(v.PublicKey)))
			if err != nil {
				logger.Errorf("error could not REMOVE subscription for user %v eventName %v eventfilter %v: %v", user.UserID, eventName, filter, err)
				ErrorOrJSONResponse(w, r, "Internal server error", http.StatusInternalServerError)
				return
			}

			if i >= maxValidators {
				break
			}
		}
	} else {
		// filtered one only
		err = db.DeleteSubscription(user.UserID, utils.GetNetwork(), eventName, filter)
		if err != nil {
			logger.Errorf("error could not REMOVE subscription for user %v eventName %v eventfilter %v: %v", user.UserID, eventName, filter, err)
			ErrorOrJSONResponse(w, r, "Internal server error", http.StatusInternalServerError)
			return
		}
	}

	OKResponse(w, r)
}

func MobileDeviceDeletePOST(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	j := json.NewEncoder(w)

	claims := getAuthClaims(r)
	var userDeviceID uint64
	var userID uint64

	if claims == nil {
		customDeviceID := FormValueOrJSON(r, "id")
		temp, err := strconv.ParseUint(customDeviceID, 10, 64)
		if err != nil {
			logger.Errorf("error parsing id %v | err: %v", customDeviceID, err)
			sendErrorResponse(j, r.URL.String(), "could not parse id")
			return
		}
		userDeviceID = temp
		sessionUser := getUser(r)
		if !sessionUser.Authenticated {
			sendErrorResponse(j, r.URL.String(), "not authenticated")
			return
		}
		userID = sessionUser.UserID
	} else {
		sendErrorResponse(j, r.URL.String(), "you can not delete the device you are currently signed in with")
		return
	}

	err := db.MobileDeviceDelete(userID, userDeviceID)
	if err != nil {
		logger.Errorf("could not retrieve db results err: %v", err)
		sendErrorResponse(j, r.URL.String(), "could not retrieve db results")
		return
	}

	sendOKResponse(j, r.URL.String(), nil)
}
