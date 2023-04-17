package handlers

import (
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"eth2-exporter/db"
	"eth2-exporter/mail"
	"eth2-exporter/services"
	"eth2-exporter/templates"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	ctxt "context"

	"github.com/gorilla/context"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

var notificationCenterParts []string = append(layoutTemplateFiles, "user/notificationsCenter.html", "modals.html")

func UserAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := getUser(r)
		if !user.Authenticated {
			utils.SetFlash(w, r, authSessionName, "Error: Please login first")
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// UserSettings renders the user-template
func UserSettings(w http.ResponseWriter, r *http.Request) {
	templateFiles := append(layoutTemplateFiles, "user/settings.html")
	var userTemplate = templates.GetTemplate(templateFiles...)

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
		utils.SetFlash(w, r, "", "Error: Something went wrong.")
		http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
		return
	}

	subscription, err := db.StripeGetUserSubscription(user.UserID, utils.GROUP_API)
	if err != nil && err != sql.ErrNoRows {
		logger.Errorf("Error retrieving the subscriptions for user: %v %v", user.UserID, err)
		utils.SetFlash(w, r, "", "Error: Something went wrong.")
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

	data := InitPageData(w, r, "user", "/user", "User Settings", templateFiles)
	data.Data = userSettingsData
	data.User = user

	var premiumPkg = ""
	if premiumSubscription.Active {
		premiumPkg = premiumSubscription.Package
	}

	session.SetValue("subscription", premiumPkg)
	session.Save(r, w)

	if handleTemplateError(w, r, "user.go", "UserSettings", "", userTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}

// GenerateAPIKey generates an API key for users that do not yet have a key.
func GenerateAPIKey(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	user := getUser(r)

	err := db.CreateAPIKey(user.UserID)
	if err != nil {
		logger.WithError(err).Error("Could not create API key for user")
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}

	http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
}

// UserAuthorizeConfirm renders the user-authorize template
func UserAuthorizeConfirm(w http.ResponseWriter, r *http.Request) {
	templateFiles := append(layoutTemplateFiles, "user/authorize.html")
	var authorizeTemplate = templates.GetTemplate(templateFiles...)

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

	session.SetValue("state", state)
	session.SetValue("client_id", clientID)
	session.SetValue("oauth_redirect_uri", redirectURI)
	session.Save(r, w)

	if !user.Authenticated {
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

	data := InitPageData(w, r, "user", "/user", "", templateFiles)
	data.Data = authorizeData
	data.Meta.NoTrack = true

	err = authorizeTemplate.ExecuteTemplate(w, "layout", data)
	if err != nil {
		logger.Errorf("error executing template [user.go / UserAuthorizeConfirm] for %v route: %v", r.URL.String(), err)
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

	session.DeleteValue("oauth_redirect_uri")
	session.DeleteValue("state")
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func UserNotifications(w http.ResponseWriter, r *http.Request) {
	templateFiles := append(layoutTemplateFiles, "user/notifications.html")
	var notificationTemplate = templates.GetTemplate(templateFiles...)

	w.Header().Set("Content-Type", "text/html")
	userNotificationsData := &types.UserNotificationsPageData{}

	user := getUser(r)

	userNotificationsData.Flashes = utils.GetFlashes(w, r, authSessionName)
	userNotificationsData.CsrfField = csrf.TemplateField(r)

	var watchlistIndices []uint64
	err := db.WriterDb.Select(&watchlistIndices, `
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
	err = db.FrontendWriterDB.Get(&countSubscriptions, `
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

	data := InitPageData(w, r, "user", "/user", "", templateFiles)
	data.Data = userNotificationsData
	data.User = user

	if handleTemplateError(w, r, "user.go", "UserNotifications", "", notificationTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}

// func getUserMetrics(userId uint64) (interface{}, error) {
// 	metricsdb := struct {
// 		Validators         uint64 `db:"validators"`
// 		Notifications      uint64 `db:"notifications"`
// 		AttestationsMissed uint64 `db:"attestations_missed"`
// 		ProposalsMissed    uint64 `db:"proposals_missed"`
// 		ProposalsSubmitted uint64 `db:"proposals_submitted"`
// 	}{}
// 	net := strings.ToLower(utils.GetNetwork())
// 	err := db.FrontendWriterDB.Get(&metricsdb, `

// 		SELECT COUNT(uvt.user_id) as validators,
// 		(SELECT COUNT(event_name) FROM users_subscriptions WHERE user_id=$1 AND last_sent_ts > NOW() - INTERVAL '1 MONTH' AND COUNT(uvt.user_id)>0) AS notifications,
// 		(SELECT COUNT(event_name) FROM users_subscriptions WHERE user_id=$1 AND last_sent_ts > NOW() - INTERVAL '1 MONTH' AND event_name=$2 AND COUNT(uvt.user_id)>0) AS attestations_missed,
// 		(SELECT COUNT(event_name) FROM users_subscriptions WHERE user_id=$1 AND last_sent_ts > NOW() - INTERVAL '1 MONTH' AND event_name=$3 AND COUNT(uvt.user_id)>0) AS proposals_missed,
// 		(SELECT COUNT(event_name) FROM users_subscriptions WHERE user_id=$1 AND last_sent_ts > NOW() - INTERVAL '1 MONTH' AND event_name=$4 AND COUNT(uvt.user_id)>0) AS proposals_submitted
// 		FROM users_validators_tags  uvt
// 		WHERE user_id=$1 and tag LIKE $5;
// 		`, userId, net+":"+string(types.ValidatorMissedAttestationEventName),
// 		net+":"+string(types.ValidatorMissedProposalEventName),
// 		net+":"+string(types.ValidatorExecutedProposalEventName),
// 		net+":%")
// 	return metricsdb, err
// }

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

	net.IsSubscribed = true
	n := []uint64{}
	err := db.ReaderDb.Select(&n, `select extract( epoch from ts)::Int as ts from network_liveness where (headepoch-finalizedepoch)!=2 AND ts > now() - interval '1 year';`)

	resp := []result{}
	for _, item := range n {
		resp = append(resp, result{Notification: "Finality issue", Network: utils.Config.Chain.Config.ConfigName, Timestamp: item * 1000})
	}
	net.Events_ts = resp

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

// AddValidatorsAndSubscribe adds a validator to a users watchlist and subscribes to the validator
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
		Index  uint64 `json:"index"`
		Pubkey string `json:"pubkey"`
		Events []struct {
			Event string `json:"event"`
			Email bool   `json:"email"`
			Push  bool   `json:"push"`
			Web   bool   `json:"web"`
		} `json:"events"`
	}{}

	// pubkeys := make([]string, 0)
	err = json.Unmarshal(body, &reqData)
	if err != nil {
		logger.Errorf("error parsing request body: %v, %v", r.URL.String(), err)
		ErrorOrJSONResponse(w, r, "Internal server error", http.StatusInternalServerError)
		return
	}
	if reqData.Pubkey == "" {
		err := db.WriterDb.Get(&reqData.Pubkey, "SELECT ENCODE(pubkey, 'hex') as pubkey from validators where validatorindex = $1", reqData.Index)
		if err != nil {
			logger.Errorf("error getting pubkey from validator index route: %v, %v", r.URL.String(), err)
			ErrorOrJSONResponse(w, r, "Internal server error", http.StatusInternalServerError)
			return
		}
	}

	if reqData.Pubkey == "" {
		logger.Errorf("error invalid pubkey: %v, %v", r.URL.String(), err)
		ErrorOrJSONResponse(w, r, "Internal server error", http.StatusInternalServerError)
		return
	}
	err = db.AddToWatchlist([]db.WatchlistEntry{{UserId: user.UserID, Validator_publickey: reqData.Pubkey}}, utils.GetNetwork())
	if err != nil {
		logger.Errorf("error adding to watchlist: %v, %v", r.URL.String(), err)
		return
	}

	result := true
	m := map[string]bool{}
	for _, item := range reqData.Events {
		if item.Email {
			if m[item.Event] && reqData.Pubkey == "" {
				continue
			}

			result = result && internUserNotificationsSubscribe(item.Event, reqData.Pubkey, 0, w, r)
			m[item.Event] = true
			if !result {
				break
			}
		}
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
			Filter    string  `json:"filter"`
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
		net + ":" + string(types.ValidatorGotSlashedEventName),
		net + ":" + string(types.SyncCommitteeSoon)})

	_, err = db.FrontendWriterDB.Exec(`
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

	_, err = db.FrontendWriterDB.Exec(`
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
	var notificationsCenterTemplate = templates.GetTemplate(notificationCenterParts...)

	w.Header().Set("Content-Type", "text/html")
	userNotificationsCenterData := &types.UserNotificationsCenterPageData{}
	data := InitPageData(w, r, "user", "/user", "", notificationCenterParts)

	user := getUser(r)

	userNotificationsCenterData.Flashes = utils.GetFlashes(w, r, authSessionName)
	userNotificationsCenterData.CsrfField = csrf.TemplateField(r)
	var watchlistPubkeys [][]byte
	err := db.FrontendWriterDB.Select(&watchlistPubkeys, `
	SELECT validator_publickey
	FROM users_validators_tags
	WHERE user_id = $1 and tag = $2
	`, user.UserID, utils.GetNetwork()+":"+string(types.ValidatorTagsWatchlist))
	if err != nil && err != sql.ErrNoRows {
		logger.Errorf("error retrieving pubkeys from watchlist validator count %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	type watchlistValidators struct {
		Index  uint64 `db:"index"`
		Pubkey string `db:"pubkey"`
	}

	watchlist := []watchlistValidators{}
	err = db.WriterDb.Select(&watchlist, `
	SELECT 
		validatorindex as index,
		ENCODE(pubkey, 'hex') as pubkey
	FROM validators 
	WHERE pubkey = ANY($1)
	`, pq.ByteaArray(watchlistPubkeys))
	if err != nil {
		logger.Errorf("error retrieving watchlist indices validator count %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	var subscriptions []types.Subscription
	err = db.FrontendWriterDB.Select(&subscriptions, `
	SELECT 
		event_name, event_filter, last_sent_ts, last_sent_epoch, created_ts, created_epoch, event_threshold
	FROM users_subscriptions
	WHERE user_id = $1 AND (event_name like $2 OR event_name like 'monitoring%') AND event_name != $3
	`, user.UserID, utils.GetNetwork()+":%", utils.GetNetwork()+":"+"validator_balance_decreased")
	if err != nil {
		logger.Errorf("error retrieving subscriptions for user %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	validatorMap := make(map[string]types.UserValidatorNotificationTableData, len(watchlist))
	link := "/dashboard?validators="

	validatorCount := 0

	for _, val := range watchlist {
		validatorCount += 1
		link += strconv.FormatUint(val.Index, 10) + ","
		validatorMap[val.Pubkey] = types.UserValidatorNotificationTableData{
			Index:  val.Index,
			Pubkey: val.Pubkey,
		}
	}
	link = link[:len(link)-1]

	monitoringSubscriptions := make([]types.Subscription, 0)
	networkSubscriptions := make([]types.Subscription, 0)

	type metrics struct {
		Validators         uint64
		Subscriptions      uint64
		Notifications      uint64
		AttestationsMissed uint64
		ProposalsSubmitted uint64
		ProposalsMissed    uint64
	}

	var metricsMonth metrics = metrics{
		Validators:    uint64(validatorCount),
		Subscriptions: uint64(len(subscriptions)),
	}

	var networkData interface{}

	for _, sub := range subscriptions {
		monthAgo := time.Now().Add(time.Hour * 24 * 31 * -1)
		if sub.LastSent != nil && sub.LastSent.After(monthAgo) {
			metricsMonth.Notifications += 1
			switch sub.EventName {
			case utils.GetNetwork() + ":" + string(types.ValidatorMissedAttestationEventName):
				metricsMonth.AttestationsMissed += 1
			case utils.GetNetwork() + ":" + string(types.ValidatorExecutedProposalEventName):
				metricsMonth.ProposalsSubmitted += 1
			case utils.GetNetwork() + ":" + string(types.ValidatorMissedProposalEventName):
				metricsMonth.ProposalsMissed += 1
			}
		}
		event := strings.TrimPrefix(sub.EventName, utils.GetNetwork()+":")
		if strings.HasPrefix(event, "network_") {
			networkSubscriptions = append(networkSubscriptions, sub)
			if event == string(types.NetworkLivenessIncreasedEventName) {
				networkData, err = getUserNetworkEvents(user.UserID)
				if err != nil {
					logger.Errorf("error retrieving network data for user %v: %v ", user.UserID, err)
					http.Error(w, "Internal server error", http.StatusServiceUnavailable)
					return
				}
			}
		}

		val, ok := validatorMap[sub.EventFilter]
		if !ok {
			if (utils.GetNetwork() == "mainnet" && strings.HasPrefix(string(sub.EventName), "monitoring_")) || strings.HasPrefix(string(sub.EventName), utils.GetNetwork()+":"+"monitoring_") {
				monitoringSubscriptions = append(monitoringSubscriptions, sub)
			}
			continue
		}

		if sub.LastSent == nil {
			zeroTime := time.Unix(0, 0)
			sub.LastSent = &zeroTime
		}

		val.Notification = append(val.Notification, struct {
			Notification string
			Timestamp    uint64
			Threshold    string
		}{
			Notification: string(sub.EventName),
			Timestamp:    uint64(sub.LastSent.Unix()),
			Threshold:    fmt.Sprintf("%.f", sub.EventThreshold),
		})

		validatorMap[sub.EventFilter] = val
	}
	validatorTableData := make([]types.UserValidatorNotificationTableData, 0, len(validatorMap))
	for _, val := range validatorMap {
		validatorTableData = append(validatorTableData, val)
	}

	//add notification-center data
	// type metrics
	// metricsdb, err := getUserMetrics(user.UserID)
	// if err != nil {
	// 	logger.Errorf("error retrieving metrics data for users: %v ", user.UserID, err)
	// 	http.Error(w, "Internal server error", http.StatusServiceUnavailable)
	// 	return
	// }

	machines, err := db.BigtableClient.GetMachineMetricsMachineNames(user.UserID)
	if err != nil {
		logger.Errorf("error retrieving user machines for user %v: %v ", user.UserID, err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}

	// validatorTableData, err := getValidatorTableData(user.UserID)
	// if err != nil {
	// 	logger.Errorf("error retrieving validators table data for users: %v ", user.UserID, err)
	// 	http.Error(w, "Internal server error", http.StatusServiceUnavailable)
	// 	return
	// }

	var notificationChannels []types.UserNotificationChannels

	err = db.FrontendReaderDB.Select(&notificationChannels, `
		SELECT
			channel,
			active
		FROM
			users_notification_channels
		WHERE
			user_id = $1
	`, user.UserID)
	if err != nil {
		logger.Errorf("error retrieving notification channels for user %v: %v ", user.UserID, err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
	email := false
	push := false
	webhook := false
	for _, ch := range notificationChannels {
		if ch.Channel == types.EmailNotificationChannel {
			email = true
		}
		if ch.Channel == types.PushNotificationChannel {
			push = true
		}
		if ch.Channel == types.WebhookNotificationChannel {
			webhook = true
		}
	}

	if !email {
		notificationChannels = append(notificationChannels, types.UserNotificationChannels{
			Channel: types.EmailNotificationChannel,
			Active:  true,
		})
	}
	if !push {
		notificationChannels = append(notificationChannels, types.UserNotificationChannels{
			Channel: types.PushNotificationChannel,
			Active:  true,
		})
	}
	if !webhook {
		notificationChannels = append(notificationChannels, types.UserNotificationChannels{
			Channel: types.WebhookNotificationChannel,
			Active:  true,
		})
	}

	events := make([]types.EventNameCheckbox, 0)
	for _, ev := range types.AddWatchlistEvents {
		events = append(events, types.EventNameCheckbox{
			EventLabel: ev.Desc,
			EventName:  ev.Event,
			Active:     false,
			Warning:    ev.Warning,
			Info:       ev.Info,
		})
	}

	networkEvents := make([]types.EventNameCheckbox, 0)
	for _, ev := range types.NetworkNotificationEvents {
		networkEvents = append(networkEvents, types.EventNameCheckbox{
			EventLabel: ev.Desc,
			EventName:  ev.Event,
			Active:     false,
		})
	}

	for i, nEvent := range networkEvents {
		for _, nSub := range networkSubscriptions {
			if nSub.EventName == utils.GetNetwork()+":"+string(nEvent.EventName) {
				networkEvents[i].Active = true
			}
		}
	}

	userNotificationsCenterData.ManageNotificationModal = types.ManageNotificationModal{
		CsrfField: csrf.TemplateField(r),
		Events:    events,
	}

	userNotificationsCenterData.AddValidatorWatchlistModal = types.AddValidatorWatchlistModal{
		CsrfField: csrf.TemplateField(r),
		Events:    events,
	}

	userNotificationsCenterData.NotificationChannelsModal = types.NotificationChannelsModal{
		CsrfField:            csrf.TemplateField(r),
		NotificationChannels: notificationChannels,
	}
	userNotificationsCenterData.NetworkEventModal = types.NetworkEventModal{
		CsrfField: csrf.TemplateField(r),
		Events:    networkEvents,
	}

	userNotificationsCenterData.DashboardLink = link
	userNotificationsCenterData.Metrics = metricsMonth
	userNotificationsCenterData.Validators = validatorTableData
	userNotificationsCenterData.Network = networkData
	userNotificationsCenterData.MonitoringSubscriptions = monitoringSubscriptions
	userNotificationsCenterData.Machines = machines
	data.Data = userNotificationsCenterData
	data.User = user

	if data.Debug {
		data.DebugTemplates = notificationCenterParts
	}

	if handleTemplateError(w, r, "user.go", "UserNotificationsCenter", "", notificationsCenterTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}

func UserNotificationsData(w http.ResponseWriter, r *http.Request) {
	currency := GetCurrency(r)
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()

	draw, err := strconv.ParseUint(q.Get("draw"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables data parameter from string to int: %v", err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
	// start, err := strconv.ParseUint(q.Get("start"), 10, 64)
	// if err != nil {
	// 	logger.Errorf("error converting datatables start parameter from string to int: %v", err)
	// 	http.Error(w, "Internal server error", http.StatusServiceUnavailable)
	// 	return
	// }
	// length, err := strconv.ParseUint(q.Get("length"), 10, 64)
	// if err != nil {
	// 	logger.Errorf("error converting datatables length parameter from string to int: %v", err)
	// 	http.Error(w, "Internal server error", http.StatusServiceUnavailable)
	// 	return
	// }
	// if length > 100 {
	// 	length = 100
	// }

	user := getUser(r)

	type watchlistSubscription struct {
		Index     *uint64 // consider validators that only have deposited but do not have an index yet
		Publickey []byte
		Balance   uint64
		Events    *pq.StringArray
	}

	wl := []watchlistSubscription{}
	err = db.WriterDb.Select(&wl, `
		SELECT 
			validators.validatorindex as index,
			users_validators_tags.validator_publickey as publickey,
			ARRAY_REMOVE(ARRAY_AGG(users_subscriptions.event_name order by users_subscriptions.event_name asc), NULL) as events
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
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}

	indices := make([]uint64, 0, len(wl))
	for _, vali := range wl {
		if vali.Index != nil {
			indices = append(indices, *vali.Index)
		}
	}

	balances, err := db.BigtableClient.GetValidatorBalanceHistory(indices, services.LatestEpoch(), services.LatestEpoch())
	if err != nil {
		logger.WithError(err).WithField("route", r.URL.String()).Errorf("error retrieving validator balance data")
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}

	for _, validator := range wl {
		for balanceIndex, balance := range balances {
			if len(balance) == 0 {
				continue
			}
			if *validator.Index == balanceIndex {
				validator.Balance = balance[0].Balance
			}
		}
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
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}

}

func UserSubscriptionsData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()

	draw, err := strconv.ParseUint(q.Get("draw"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables data parameter from string to int: %v", err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
	// start, err := strconv.ParseUint(q.Get("start"), 10, 64)
	// if err != nil {
	// 	logger.Errorf("error converting datatables start parameter from string to int: %v", err)
	// 	http.Error(w, "Internal server error", http.StatusServiceUnavailable)
	// 	return
	// }
	length, err := strconv.ParseUint(q.Get("length"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables length parameter from string to int: %v", err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
	if length > 100 {
		length = 100
	}

	user := getUser(r)

	subs := []types.Subscription{}
	err = db.FrontendWriterDB.Select(&subs, `
			SELECT id, user_id, event_name, event_filter, last_sent_ts, last_sent_epoch, created_ts, created_epoch, event_threshold, unsubscribe_hash, internal_state
			FROM users_subscriptions
			WHERE user_id = $1
	`, user.UserID)
	if err != nil {
		logger.Errorf("error retrieving subscriptions for users %v: %v", user.UserID, err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
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
		} else if sub.EventName == string(types.TaxReportEventName) {
			pubkey = template.HTML(`<a href="/rewards">report</a>`)
		} else if strings.HasPrefix(string(sub.EventName), "monitoring_") {
			pubkey = utils.FormatMachineName(sub.EventFilter)
		}
		if sub.EventName != string(types.ValidatorBalanceDecreasedEventName) {
			tableData = append(tableData, []interface{}{
				pubkey,
				sub.EventName,
				utils.FormatTimestamp(sub.CreatedTime.Unix()),
				ls,
			})
		}

	}

	data := &types.DataTableResponse{
		Draw:            draw,
		RecordsTotal:    uint64(len(subs)),
		RecordsFiltered: uint64(len(subs)),
		Data:            tableData,
	}

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
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

	if user.Authenticated {
		codeBytes, err1 := utils.GenerateRandomBytesSecure(32)
		if err1 != nil {
			logger.Errorf("error creating secure random bytes for user: %v %v", user.UserID, err1)
			callback := appData.RedirectURI + "?error=server_error&error_description=err_random_number" + stateAppend
			http.Redirect(w, r, callback, http.StatusSeeOther)
			return
		}

		code := hex.EncodeToString(codeBytes)   // return to user
		codeHashed := utils.HashAndEncode(code) // save hashed code in db
		clientID := session.GetValue("client_id").(string)

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
		utils.LogError(nil, "Not authorized", 0)
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
	if user.Authenticated {
		err := db.DeleteUserById(user.UserID)
		if err != nil {
			logger.Errorf("error deleting user by email for user: %v %v", user.UserID, err)
			http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
			utils.SetFlash(w, r, "", "Error: Could not delete user.")
			session.Save(r, w)
			http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
			return
		}

		Logout(w, r)
	} else {
		utils.LogError(nil, "Trying to delete an unauthenticated user", 0)
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
	if err != nil {
		logger.Errorf("error setting user monitor sharing settings: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

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

	err = db.FrontendWriterDB.Get(&currentUser, "SELECT id, email, password, email_confirmed FROM users WHERE id = $1", user.UserID)
	if err != nil {
		if err != sql.ErrNoRows {
			logger.Errorf("error retrieving password for user %v: %v", user.UserID, err)
		}
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

	err = purgeAllSessionsForUser(r.Context(), user.UserID)
	if err != nil {
		logger.Errorf("error purging sessions for user %v: %v", user.UserID, err)
		session.AddFlash(authInternalServerErrorFlashMsg)
		session.Save(r, w)
		http.Redirect(w, r, "/confirmation", http.StatusSeeOther)
		return
	}

	err = session.SCS.RenewToken(r.Context())
	if err != nil {
		logger.Errorf("error renewing session token for user: %v", err)
		session.AddFlash(authInternalServerErrorFlashMsg)
		session.Save(r, w)
		http.Redirect(w, r, "/confirmation", http.StatusSeeOther)
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
	db.FrontendWriterDB.Get(&existingEmails, "SELECT email FROM users WHERE email = $1", email)

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

	user := struct {
		ID        int64     `db:"id"`
		Email     string    `db:"email"`
		ConfirmTs time.Time `db:"email_confirmation_ts"`
		Confirmed bool      `db:"email_confirmed"`
		NewEmail  string    `db:"email_change_to_value"`
	}{}

	err = db.FrontendWriterDB.Get(&user, "SELECT id, email, email_confirmation_ts, email_confirmed FROM users WHERE email_confirmation_hash = $1", hash)
	if err != nil {
		logger.Errorf("error retreiveing email for confirmation_hash %v %v", hash, err)
		utils.SetFlash(w, r, authSessionName, "Error: This confirmation link is invalid / outdated.")
		http.Redirect(w, r, "/confirmation", http.StatusSeeOther)
		return
	}

	if !user.Confirmed {
		utils.SetFlash(w, r, authSessionName, "Error: Cannot update email for an unconfirmed address.")
		http.Redirect(w, r, "/confirmation", http.StatusSeeOther)
		return
	}

	if user.ConfirmTs.Add(time.Minute * 30).Before(time.Now()) {
		utils.SetFlash(w, r, authSessionName, "Error: This confirmation link has expired.")
		http.Redirect(w, r, "/confirmation", http.StatusSeeOther)
		return
	}

	if !utils.IsValidEmail(user.NewEmail) {
		utils.SetFlash(w, r, authSessionName, "Error: Could not update your email because the new email is invalid, please try again.")
		http.Redirect(w, r, "/confirmation", http.StatusSeeOther)
		return
	}

	var emailExists string
	db.FrontendWriterDB.Get(&emailExists, "SELECT email FROM users WHERE email = $1", user.NewEmail)
	if emailExists != "" {
		utils.SetFlash(w, r, authSessionName, "Error: Email already exists. We could not update your email.")
		http.Redirect(w, r, "/confirmation", http.StatusSeeOther)
		return
	}

	_, err = db.FrontendWriterDB.Exec(`UPDATE users SET email = $1, email_confirmation_hash = '' WHERE id = $2`, user.NewEmail, user.ID)
	if err != nil {
		logger.Errorf("error: updating email for user: %v", err)
		utils.SetFlash(w, r, authSessionName, "Error: Could not Update Email.")
		http.Redirect(w, r, "/confirmation", http.StatusSeeOther)
		return
	}

	session.SetValue("subscription", "")
	session.SetValue("authenticated", false)
	session.DeleteValue("user_id")

	err = purgeAllSessionsForUser(r.Context(), uint64(user.ID))
	if err != nil {
		logger.Errorf("error: purging sessions for user %v: %v", user.ID, err)
		utils.SetFlash(w, r, authSessionName, "Error: Could not Update Email.")
		http.Redirect(w, r, "/confirmation", http.StatusSeeOther)
		return
	}

	utils.SetFlash(w, r, authSessionName, "Your email has been updated successfully! <br> You can log in with your new email.")
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func sendEmailUpdateConfirmation(userId uint64, newEmail string) error {
	now := time.Now()
	emailConfirmationHash := utils.RandomString(40)

	tx, err := db.FrontendWriterDB.Beginx()
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
		return &types.RateLimitError{TimeLeft: (*lastTs).Add(authConfirmEmailRateLimit).Sub(now)}
	}

	_, err = tx.Exec("UPDATE users SET email_confirmation_hash = $1, email_change_to_value = $2 WHERE id = $3", emailConfirmationHash, newEmail, userId)
	if err != nil {
		return fmt.Errorf("error updating confirmation-hash: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error committing db-tx: %w", err)
	}

	subject := fmt.Sprintf("%s: Verify your email-address", utils.Config.Frontend.SiteDomain)
	msg := fmt.Sprintf(`To update your email on %[1]s please verify it by clicking this link:

https://%[1]s/settings/email/%[2]s

Best regards,

%[1]s
`, utils.Config.Frontend.SiteDomain, emailConfirmationHash)
	err = mail.SendTextMail(newEmail, subject, msg, []types.EmailAttachment{})
	if err != nil {
		return err
	}

	_, err = db.FrontendWriterDB.Exec("UPDATE users SET email_confirmation_ts = TO_TIMESTAMP($1) WHERE id = $2", time.Now().Unix(), userId)
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
		err := db.AddSubscription(user.UserID, utils.GetNetwork(), types.ValidatorBalanceDecreasedEventName, pubKey, 0)
		if err != nil {
			logger.Errorf("error could not ADD subscription for user %v eventName %v eventfilter %v: %v", user.UserID, types.ValidatorBalanceDecreasedEventName, pubKey, err)
			ErrorOrJSONResponse(w, r, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
	slashed := FormValueOrJSON(r, "validator_slashed")
	if slashed == "on" {
		err := db.AddSubscription(user.UserID, utils.GetNetwork(), types.ValidatorGotSlashedEventName, pubKey, 0)
		if err != nil {
			logger.Errorf("error could not ADD subscription for user %v eventName %v eventfilter %v: %v", user.UserID, types.ValidatorGotSlashedEventName, pubKey, err)
			ErrorOrJSONResponse(w, r, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
	proposalSubmitted := FormValueOrJSON(r, "validator_proposal_submitted")
	if proposalSubmitted == "on" {
		err := db.AddSubscription(user.UserID, utils.GetNetwork(), types.ValidatorExecutedProposalEventName, pubKey, 0)
		if err != nil {
			logger.Errorf("error could not ADD subscription for user %v eventName %v eventfilter %v: %v", user.UserID, types.ValidatorGotSlashedEventName, pubKey, err)
			ErrorOrJSONResponse(w, r, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
	proposalMissed := FormValueOrJSON(r, "validator_proposal_missed")
	if proposalMissed == "on" {
		err := db.AddSubscription(user.UserID, utils.GetNetwork(), types.ValidatorMissedProposalEventName, pubKey, 0)
		if err != nil {
			logger.Errorf("error could not ADD subscription for user %v eventName %v eventfilter %v: %v", user.UserID, types.ValidatorGotSlashedEventName, pubKey, err)
			ErrorOrJSONResponse(w, r, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
	attestationMissed := FormValueOrJSON(r, "validator_attestation_missed")
	if attestationMissed == "on" {
		err := db.AddSubscription(user.UserID, utils.GetNetwork(), types.ValidatorMissedAttestationEventName, pubKey, 0)
		if err != nil {
			logger.Errorf("error could not ADD subscription for user %v eventName %v eventfilter %v: %v", user.UserID, types.ValidatorGotSlashedEventName, pubKey, err)
			ErrorOrJSONResponse(w, r, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
	syncCommittee := FormValueOrJSON(r, "validator_synccommittee_soon")
	if syncCommittee == "on" {
		err := db.AddSubscription(user.UserID, utils.GetNetwork(), types.SyncCommitteeSoon, pubKey, 0)
		if err != nil {
			logger.Errorf("error could not ADD subscription for user %v eventName %v eventfilter %v: %v", user.UserID, types.SyncCommitteeSoon, pubKey, err)
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
// @Router /api/v1/user/dashboard/save [post]
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
	db.WriterDb.Select(&publicKeys, `
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

	type SubIntent struct {
		EventName      string  `json:"event_name"`
		EventFilter    string  `json:"event_filter"`
		EventThreshold float64 `json:"event_threshold"`
	}

	var jsonObjects []SubIntent
	err := json.Unmarshal(context.Get(r, utils.JsonBodyNakedKey).([]byte), &jsonObjects)
	if err != nil {
		logger.Errorf("Could not parse multiple notification subscription intent | %v", err)
		sendErrorResponse(w, r.URL.String(), "could not parse request")
		return
	}

	if len(jsonObjects) > 100 {
		utils.LogError(nil, "Multiple notification subscription: max number bundle subscribe is 100", 0)
		sendErrorResponse(w, r.URL.String(), "Max number bundle subscribe is 100")
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

func MultipleUsersNotificationsSubscribeWeb(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	type SubIntent struct {
		EventName      string  `json:"event_name"`
		EventFilter    string  `json:"event_filter"`
		EventThreshold float64 `json:"event_threshold"`
	}

	var jsonObjects []SubIntent
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logger.Errorf("error reading body %v URL: %v", err, r.URL.String())
		sendErrorResponse(w, r.URL.String(), "could not parse body")
		return
	}

	err = json.Unmarshal(b, &jsonObjects)
	if err != nil {
		logger.Errorf("Could not parse multiple notification subscription intent | %v", err)
		sendErrorResponse(w, r.URL.String(), "could not parse request")
		return
	}

	if len(jsonObjects) > 100 {
		utils.LogError(nil, "Multiple notification subscription web: max number bundle subscribe is 100", 0)
		sendErrorResponse(w, r.URL.String(), "Max number bundle subscribe is 100")
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

	event = strings.TrimPrefix(event, utils.GetNetwork()+":")

	eventName, err := types.EventNameFromString(event)
	if err != nil {
		logger.Errorf("error invalid event name: %v event: %v", err, event)
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

	filterWatchlist := db.WatchlistFilter{
		UserId:         user.UserID,
		Validators:     nil,
		Tag:            types.ValidatorTagsWatchlist,
		JoinValidators: true,
		Network:        utils.GetNetwork(),
	}
	if !userPremium.NotificationThresholds {
		if eventName == types.MonitoringMachineDiskAlmostFullEventName {
			threshold = 0.1
		} else if eventName == types.MonitoringMachineCpuLoadEventName {
			threshold = 0.6
		} else if eventName == types.MonitoringMachineMemoryUsageEventName {
			threshold = 0.8
		} else if eventName == types.ValidatorIsOfflineEventName {
			threshold = 3
		}
		// rocketpool thresholds are free
	}

	if filterLen == 0 && !strings.HasPrefix(string(eventName), "monitoring_") && !strings.HasPrefix(string(eventName), "rocketpool_") { // no filter = add all my watched validators
		myValidators, err2 := db.GetTaggedValidators(filterWatchlist)
		if err2 != nil {
			ErrorOrJSONResponse(w, r, "could not retrieve db results", http.StatusInternalServerError)
			return false
		}

		maxValidators := userPremium.MaxValidators

		// not quite happy performance wise, placing a TODO here for future me
		for i, v := range myValidators {
			err = db.AddSubscription(
				user.UserID,
				utils.GetNetwork(),
				eventName,
				fmt.Sprintf("%v", hex.EncodeToString(v.ValidatorPublickey)),
				threshold,
			)
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

		network := utils.GetNetwork()
		if eventName == types.EthClientUpdateEventName || strings.HasPrefix(string(eventName), "monitoring_") {
			network = ""
		}

		if filterLen == 0 && (eventName == types.RocketpoolColleteralMaxReached || eventName == types.RocketpoolColleteralMinReached) {

			myValidators, err2 := db.GetTaggedValidators(filterWatchlist)
			if err2 != nil {
				ErrorOrJSONResponse(w, r, "could not retrieve db results", http.StatusInternalServerError)
				return false
			}

			maxValidators := userPremium.MaxValidators

			var pubkeys [][]byte
			for _, v := range myValidators {
				pubkeys = append(pubkeys, v.ValidatorPublickey)
			}

			var rocketpoolNodes []string
			err = db.WriterDb.Select(&rocketpoolNodes, `
				SELECT DISTINCT(ENCODE(node_address, 'hex')) as node_address FROM rocketpool_minipools WHERE pubkey = ANY($1)
			`, pq.ByteaArray(pubkeys))
			if err != nil {
				ErrorOrJSONResponse(w, r, "could not retrieve db results", http.StatusInternalServerError)
				return false
			}

			for i, v := range rocketpoolNodes {
				err = db.AddSubscription(user.UserID, utils.GetNetwork(), eventName, v, threshold)
				if err != nil {
					logger.Errorf("error could not ADD subscription for user %v eventName %v eventfilter %v: %v", user.UserID, eventName, filter, err)
					ErrorOrJSONResponse(w, r, "Internal server error", http.StatusInternalServerError)
					return false
				}

				if i >= maxValidators {
					break
				}
			}
		} else {
			err = db.AddSubscription(user.UserID, network, eventName, filter, threshold)
			if err != nil {
				logger.Errorf("error could not ADD subscription for user %v eventName %v eventfilter %v: %v", user.UserID, eventName, filter, err)
				ErrorOrJSONResponse(w, r, "Internal server error", http.StatusInternalServerError)
				return false
			}
		}

	}

	return true
}

func MultipleUsersNotificationsUnsubscribe(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	type UnSubIntent struct {
		EventName   string `json:"event_name"`
		EventFilter string `json:"event_filter"`
	}

	var jsonObjects []UnSubIntent
	err := json.Unmarshal(context.Get(r, utils.JsonBodyNakedKey).([]byte), &jsonObjects)
	if err != nil {
		logger.Errorf("Could not parse multiple notification subscription intent | %v", err)
		sendErrorResponse(w, r.URL.String(), "could not parse request")
		return
	}

	if len(jsonObjects) > 100 {
		utils.LogError(nil, "Max number bundle unsubscribe is 100", 0)
		sendErrorResponse(w, r.URL.String(), "Max number bundle unsubscribe is 100")
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
	event = strings.TrimPrefix(event, utils.GetNetwork()+":")

	eventName, err := types.EventNameFromString(event)
	if err != nil {
		logger.Errorf("error invalid event name: %v event: %v", err, event)
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

	filterWatchlist := db.WatchlistFilter{
		UserId:         user.UserID,
		Validators:     nil,
		Tag:            types.ValidatorTagsWatchlist,
		JoinValidators: true,
		Network:        utils.GetNetwork(),
	}

	if filterLen == 0 && !strings.HasPrefix(string(eventName), "monitoring_") && !strings.HasPrefix(string(eventName), "rocketpool_") { // no filter = add all my watched validators

		myValidators, err2 := db.GetTaggedValidators(filterWatchlist)
		if err2 != nil {
			ErrorOrJSONResponse(w, r, "could not retrieve db results", http.StatusInternalServerError)
			return false
		}

		maxValidators := getUserPremium(r).MaxValidators
		// not quite happy performance wise, placing a TODO here for future me
		for i, v := range myValidators {
			err = db.DeleteSubscription(user.UserID, utils.GetNetwork(), eventName, fmt.Sprintf("%v", hex.EncodeToString(v.ValidatorPublickey)))
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
		if filterLen == 0 && (eventName == types.RocketpoolColleteralMaxReached || eventName == types.RocketpoolColleteralMinReached) {

			err = db.DeleteAllSubscription(user.UserID, utils.GetNetwork(), eventName)
			if err != nil {
				logger.Errorf("error could not REMOVE subscription for user %v eventName %v eventfilter %v: %v", user.UserID, eventName, filter, err)
				ErrorOrJSONResponse(w, r, "Internal server error", http.StatusInternalServerError)
				return false
			}

		} else {
			network := utils.GetNetwork()
			if eventName == types.EthClientUpdateEventName || strings.HasPrefix(string(eventName), "monitoring_") {
				network = ""
			}
			// filtered one only
			err = db.DeleteSubscription(user.UserID, network, eventName, filter)
			if err != nil {
				logger.Errorf("error could not REMOVE subscription for user %v eventName %v eventfilter %v: %v", user.UserID, eventName, filter, err)
				ErrorOrJSONResponse(w, r, "Internal server error", http.StatusInternalServerError)
				return false
			}
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

	event = strings.TrimPrefix(event, utils.GetNetwork()+":")

	eventName, err := types.EventNameFromString(event)
	if err != nil {
		logger.Errorf("error invalid event name: %v event: %v", err, event)
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

	if filterLen == 0 && !types.IsUserIndexed(eventName) { // no filter = add all my watched validators

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
			err = db.DeleteSubscription(user.UserID, utils.GetNetwork(), eventName, fmt.Sprintf("%v", hex.EncodeToString(v.ValidatorPublickey)))
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
		network := utils.GetNetwork()
		if eventName == types.EthClientUpdateEventName || strings.HasPrefix(string(eventName), "monitoring_") {
			network = ""
		}
		// filtered one only
		err = db.DeleteSubscription(user.UserID, network, eventName, filter)
		if err != nil {
			logger.Errorf("error could not REMOVE subscription for user %v eventName %v eventfilter %v: %v", user.UserID, eventName, filter, err)
			ErrorOrJSONResponse(w, r, "Internal server error", http.StatusInternalServerError)
			return
		}
	}

	OKResponse(w, r)
}

func UserNotificationsUnsubscribeByHash(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	q := r.URL.Query()

	ctx, done := ctxt.WithTimeout(ctxt.Background(), time.Second*30)
	defer done()

	hashes, ok := q["hash"]
	if !ok {
		logger.Errorf("no query params given")
		http.Error(w, "invalid request", 400)
		return
	}

	tx, err := db.FrontendWriterDB.Beginx()
	if err != nil {
		//  return fmt.Errorf("error beginning transaction")
		logger.WithError(err).Errorf("error committing transacton")
		http.Error(w, "error processing request", 500)
		return
	}
	defer tx.Rollback()

	bHashes := make([][]byte, 0, len(hashes))
	for _, hash := range hashes {
		hash = strings.Replace(hash, "0x", "", -1)
		if !utils.HashLikeRegex.MatchString(hash) {
			logger.Errorf("error validating unsubscribe digest hashes")
			http.Error(w, "error processing request", 500)
		}
		b, _ := hex.DecodeString(hash)
		bHashes = append(bHashes, b)
	}

	_, err = tx.ExecContext(ctx, `DELETE from users_subscriptions where unsubscribe_hash = ANY($1)`, pq.ByteaArray(bHashes))
	if err != nil {
		logger.Errorf("error deleting from users_subscriptions %v", err)
		http.Error(w, "error processing request", 500)
		return
	}

	err = tx.Commit()
	if err != nil {
		logger.WithError(err).Errorf("error committing transacton")
		http.Error(w, "error processing request", 500)
		return
	}

	fmt.Fprintf(w, "successfully unsubscribed from %v event(s)", len(hashes))
}

type UsersNotificationsRequest struct {
	EventNames    []string `json:"event_names"`
	EventFilters  []string `json:"event_filters"`
	Search        string   `json:"search"`
	Limit         uint64   `json:"limit"`
	Offset        uint64   `json:"offset"`
	JoinValidator bool     `json:"join_validator"`
}

// UserNotificationsSubscribed godoc
// @Summary Get a set of events a user is subscribed to
// @Tags User
// @Param requestFilter body types.UsersNotificationsRequest false "An object that filters through the active subscriptions"
// @Produce json
// @Success 200 {object} types.ApiResponse{data=[]types.Subscription}
// @Failure 400 {object} types.ApiResponse
// @Failure 500 {object} types.ApiResponse
// @Security ApiKeyAuth
// @Router /api/v1/user/notifications [post]
func UserNotificationsSubscribed(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	j := json.NewEncoder(w)
	q := r.URL.Query()
	sessionUser := getUser(r)
	if !sessionUser.Authenticated {
		sendErrorResponse(w, r.URL.String(), "not authenticated")
		return
	}

	decoder := json.NewDecoder(r.Body)
	req := &types.UsersNotificationsRequest{}

	err := decoder.Decode(req)
	if err != nil && err != io.EOF {
		logger.WithError(err).Error("error decoding request body")
		sendErrorResponse(w, r.URL.String(), "error decoding request body")
		return
	}

	joinValidators := false

	filters := req.EventFilters
	names := req.EventNames
	limit := req.Limit
	offset := req.Offset
	search := req.Search
	joinValidators = req.JoinValidator

	name := q.Get("name")
	filter := q.Get("filter")
	lim := q.Get("limit")
	off := q.Get("offset")

	if q.Get("search") != "" {
		search = q.Get("search")
	}
	join := q.Get("join")

	if join != "" {
		joinValidators = true
	}

	if lim != "" {
		limit, err = strconv.ParseUint(lim, 10, 64)
		if err != nil {
			sendErrorResponse(w, r.URL.String(), "error parsing limit")
		}
	}

	if off != "" {
		offset, err = strconv.ParseUint(off, 10, 64)
		if err != nil {
			sendErrorResponse(w, r.URL.String(), "error parsing offset")
		}
	}

	if name != "" {
		names = strings.Split(name, ",")
	}

	if filter != "" {
		filters = strings.Split(filter, ",")
	}

	eventNames := make([]types.EventName, 0, len(names))
	for _, en := range names {
		n, err := types.EventNameFromString(en)
		if err != nil {
			logger.WithError(err).Errorf("error parsing provided event %v to a known event name type", en)
			sendErrorResponse(w, r.URL.String(), "error invalid event name provided")
		}
		eventNames = append(eventNames, n)
	}

	users := make([]uint64, 1)
	users[0] = sessionUser.UserID

	queryFilter := db.GetSubscriptionsFilter{
		EventNames:    &eventNames,
		EventFilters:  &filters,
		UserIDs:       &users,
		Limit:         limit,
		Offset:        offset,
		Search:        search,
		JoinValidator: joinValidators,
	}

	subs, err := db.GetSubscriptions(queryFilter)
	if err != nil {
		sendErrorResponse(w, r.URL.String(), "not authenticated")
		return
	}

	sendOKResponse(j, r.URL.String(), []interface{}{subs})
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
			sendErrorResponse(w, r.URL.String(), "could not parse id")
			return
		}
		userDeviceID = temp
		sessionUser := getUser(r)
		if !sessionUser.Authenticated {
			sendErrorResponse(w, r.URL.String(), "not authenticated")
			return
		}
		userID = sessionUser.UserID
	} else {
		sendErrorResponse(w, r.URL.String(), "you can not delete the device you are currently signed in with")
		return
	}

	err := db.MobileDeviceDelete(userID, userDeviceID)
	if err != nil {
		logger.Errorf("could not retrieve db results err: %v", err)
		sendErrorResponse(w, r.URL.String(), "could not retrieve db results")
		return
	}

	sendOKResponse(j, r.URL.String(), nil)
}

// Imprint will show the imprint data using a go template
func NotificationWebhookPage(w http.ResponseWriter, r *http.Request) {
	templateFiles := append(layoutTemplateFiles, "user/webhooks.html")
	var webhookTemplate = templates.GetTemplate(templateFiles...)

	w.Header().Set("Content-Type", "text/html")
	user := getUser(r)

	data := InitPageData(w, r, "webhook", "/webhook", "Webhook configuration", templateFiles)
	pageData := types.WebhookPageData{}

	ctx, done := ctxt.WithTimeout(ctxt.Background(), time.Second*30)
	defer done()

	pageData.CsrfField = csrf.TemplateField(r)

	var webhookCount uint64
	err := db.FrontendReaderDB.GetContext(ctx, &webhookCount, `SELECT count(*) from users_webhooks where user_id = $1`, user.UserID)
	if err != nil {
		logger.WithError(err).Errorf("error getting webhook count")
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}

	pageData.WebhookCount = webhookCount

	allowed := uint64(1)

	var activeAPP uint64
	err = db.FrontendReaderDB.GetContext(ctx, &activeAPP, `SELECT count(*) from users_app_subscriptions where active = 't' and user_id = $1;`, user.UserID)
	if err != nil {
		logger.WithError(err).Errorf("error getting app subscription count")
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}

	if activeAPP > 0 {
		allowed = 2
	}

	var activeAPI uint64
	err = db.FrontendReaderDB.GetContext(ctx, &activeAPI, `SELECT count(*) from users_stripe_subscriptions us join users u on u.stripe_customer_id = us.customer_id where active = 't' and u.id = $1;`, user.UserID)
	if err != nil {
		logger.WithError(err).Errorf("error getting api subscription count")
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}

	if activeAPI > 0 {
		allowed = 5
	}

	pageData.Allowed = allowed

	webhooks := []types.UserWebhook{}
	err = db.FrontendReaderDB.SelectContext(ctx, &webhooks, `
		SELECT 
			id,
			url,
			retries,
			last_sent,
			event_names,
			destination,
			request,
			response
		FROM users_webhooks
		WHERE user_id = $1;
	`, user.UserID)
	if err != nil {
		logger.Errorf("error querying for webhooks for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}

	webhookRows := make([]types.UserWebhookRow, 0)
	for _, wh := range webhooks {

		url, err := r.URL.Parse(wh.Url)
		if err != nil {
			logger.WithError(err).Error("error parsing URL for webhook")
			wh.Url = "Invalid URL"
		}
		// events := template.HTML{}

		// for _, ev := range wh.EventNames {

		// }

		events := make([]types.EventNameCheckbox, 0, 7)

		events = append(events, types.EventNameCheckbox{
			EventLabel: "Validator is Offline",
			EventName:  types.ValidatorIsOfflineEventName,
			Active:     utils.ElementExists(wh.EventNames, string(types.ValidatorIsOfflineEventName)),
		})
		events = append(events, types.EventNameCheckbox{
			EventLabel: "Proposal Missed",
			EventName:  types.ValidatorMissedProposalEventName,
			Active:     utils.ElementExists(wh.EventNames, string(types.ValidatorMissedProposalEventName)),
		})
		events = append(events, types.EventNameCheckbox{
			EventLabel: "Proposal Submitted",
			EventName:  types.ValidatorExecutedProposalEventName,
			Active:     utils.ElementExists(wh.EventNames, string(types.ValidatorExecutedProposalEventName)),
		})
		events = append(events, types.EventNameCheckbox{
			EventLabel: "Slashed",
			EventName:  types.ValidatorGotSlashedEventName,
			Active:     utils.ElementExists(wh.EventNames, string(types.ValidatorGotSlashedEventName)),
		})
		events = append(events, types.EventNameCheckbox{
			EventLabel: "Sync Committee Soon",
			EventName:  types.SyncCommitteeSoon,
			Active:     utils.ElementExists(wh.EventNames, string(types.SyncCommitteeSoon)),
		})
		events = append(events, types.EventNameCheckbox{
			EventLabel: "Attestation Missed",
			EventName:  types.ValidatorMissedAttestationEventName,
			Active:     utils.ElementExists(wh.EventNames, string(types.ValidatorMissedAttestationEventName)),
		})
		events = append(events, types.EventNameCheckbox{
			EventLabel: "Machine Offline",
			EventName:  types.MonitoringMachineOfflineEventName,
			Active:     utils.ElementExists(wh.EventNames, string(types.MonitoringMachineOfflineEventName)),
		})
		events = append(events, types.EventNameCheckbox{
			EventLabel: "Machine Disk Full",
			EventName:  types.MonitoringMachineDiskAlmostFullEventName,
			Active:     utils.ElementExists(wh.EventNames, string(types.MonitoringMachineDiskAlmostFullEventName)),
		})
		events = append(events, types.EventNameCheckbox{
			EventLabel: "Machine CPU",
			EventName:  types.MonitoringMachineCpuLoadEventName,
			Active:     utils.ElementExists(wh.EventNames, string(types.MonitoringMachineCpuLoadEventName)),
		})

		isDiscord := false

		if wh.Destination.Valid && wh.Destination.String == "webhook_discord" {
			isDiscord = true
		}

		ls := template.HTML(`N/A`)

		if wh.LastSent.Valid {
			ls = utils.FormatTimestampTs(wh.LastSent.Time)
		}

		whErr := types.UserWebhookRowError{}

		if wh.Retries > 0 && wh.Request.Valid {
			whErr.SummaryRequest = template.HTML("Request Sent")
			whErr.ContentRequest = template.HTML(fmt.Sprintf(`<pre><code>%v</code></pre>`, wh.Request.String))

		}

		if wh.Retries > 0 && wh.Response.Valid {
			whErr.SummaryResponse = template.HTML("Response Received")
			whErr.ContentResponse = template.HTML(fmt.Sprintf(`<pre><code>%v</code></pre>`, wh.Response.String))
		}

		hostname := ""
		if url != nil {
			hostname = url.Hostname()
		} else {
			hostname = wh.Url
		}

		webhookRows = append(webhookRows, types.UserWebhookRow{
			ID:           wh.ID,
			Retries:      template.HTML(fmt.Sprintf("%d", wh.Retries)),
			UrlFull:      wh.Url,
			Url:          template.HTML(fmt.Sprintf(`<span>%v</span><span style="margin-left: .5rem;">%v</span>`, hostname, utils.CopyButtonText(wh.Url))),
			LastSent:     ls,
			Events:       events,
			Discord:      isDiscord,
			CsrfField:    csrf.TemplateField(r),
			WebhookError: whErr,
		})

	}

	pageData.Webhooks = webhooks
	pageData.WebhookRows = webhookRows

	// logger.Infof("events: %+v", webhooks)

	events := make([]types.EventNameCheckbox, 0, 7)

	events = append(events, types.EventNameCheckbox{
		EventLabel: "Validator is Offline",
		EventName:  types.ValidatorIsOfflineEventName,
	})
	events = append(events, types.EventNameCheckbox{
		EventLabel: "Proposal Missed",
		EventName:  types.ValidatorMissedProposalEventName,
	})
	events = append(events, types.EventNameCheckbox{
		EventLabel: "Proposal Submitted",
		EventName:  types.ValidatorExecutedProposalEventName,
	})
	events = append(events, types.EventNameCheckbox{
		EventLabel: "Got Slashed",
		EventName:  types.ValidatorGotSlashedEventName,
	})
	events = append(events, types.EventNameCheckbox{
		EventLabel: "Sync Committee Soon",
		EventName:  types.SyncCommitteeSoon,
	})
	events = append(events, types.EventNameCheckbox{
		EventLabel: "Attestation Missed",
		EventName:  types.ValidatorMissedAttestationEventName,
	})
	events = append(events, types.EventNameCheckbox{
		EventLabel: "Machine Offline",
		EventName:  types.MonitoringMachineOfflineEventName,
	})
	events = append(events, types.EventNameCheckbox{
		EventLabel: "Machine Disk Full",
		EventName:  types.MonitoringMachineDiskAlmostFullEventName,
	})
	events = append(events, types.EventNameCheckbox{
		EventLabel: "Machine CPU",
		EventName:  types.MonitoringMachineCpuLoadEventName,
	})

	pageData.Events = events

	pageData.Flashes = utils.GetFlashes(w, r, authSessionName)

	data.Data = pageData

	if handleTemplateError(w, r, "user.go", "NotificationWebhookPage", "", webhookTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}

func UsersAddWebhook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	user := getUser(r)

	err := r.ParseForm()
	if err != nil {
		logger.WithError(err).Errorf("error parsing form")
		utils.SetFlash(w, r, authSessionName, "Error: Something went wrong adding your webhook, please try again in a bit.")
		http.Redirect(w, r, "/user/webhooks", http.StatusSeeOther)
		return
	}

	// const VALIDATOR_EVENTS = ['validator_attestation_missed', 'validator_proposal_missed', 'validator_proposal_submitted', 'validator_got_slashed', 'validator_synccommittee_soon']
	// const MONITORING_EVENTS = ['monitoring_machine_offline', 'monitoring_hdd_almostfull', 'monitoring_cpu_load']

	urlForm := r.FormValue("url")

	if !utils.IsValidUrl(urlForm) {
		utils.SetFlash(w, r, authSessionName, "Error: The URL provided is invalid.")
		http.Redirect(w, r, "/user/webhooks", http.StatusSeeOther)
		return
	}

	destination := "webhook"

	validatorIsOffline := r.FormValue(string(types.ValidatorIsOfflineEventName)) == "on"
	validatorProposalMissed := r.FormValue(string(types.ValidatorMissedProposalEventName)) == "on"
	validatorProposalSubmitted := r.FormValue(string(types.ValidatorExecutedProposalEventName)) == "on"
	validatorGotSlashed := r.FormValue(string(types.ValidatorGotSlashedEventName)) == "on"
	validatorSyncCommiteeSoon := r.FormValue(string(types.SyncCommitteeSoon)) == "on"
	validatorAttestationMissed := r.FormValue(string(types.ValidatorMissedAttestationEventName)) == "on"
	monitoringMachineOffline := r.FormValue(string(types.MonitoringMachineOfflineEventName)) == "on"
	monitoringHddAlmostfull := r.FormValue(string(types.MonitoringMachineDiskAlmostFullEventName)) == "on"
	monitoringCpuLoad := r.FormValue(string(types.MonitoringMachineCpuLoadEventName)) == "on"
	discord := r.FormValue("discord") == "on"

	if discord {
		destination = "webhook_discord"
	}

	all := r.FormValue("all") == "on"

	events := make(map[string]bool, 0)

	events[string(types.ValidatorIsOfflineEventName)] = validatorIsOffline
	events[string(types.ValidatorMissedProposalEventName)] = validatorProposalMissed
	events[string(types.ValidatorExecutedProposalEventName)] = validatorProposalSubmitted
	events[string(types.ValidatorGotSlashedEventName)] = validatorGotSlashed
	events[string(types.SyncCommitteeSoon)] = validatorSyncCommiteeSoon
	events[string(types.ValidatorMissedAttestationEventName)] = validatorAttestationMissed
	events[string(types.MonitoringMachineOfflineEventName)] = monitoringMachineOffline
	events[string(types.MonitoringMachineDiskAlmostFullEventName)] = monitoringHddAlmostfull
	events[string(types.MonitoringMachineCpuLoadEventName)] = monitoringCpuLoad

	eventNames := make([]string, 0)

	for eventName, active := range events {
		if active || all {
			eventNames = append(eventNames, eventName)
		}
	}

	ctx, done := ctxt.WithTimeout(ctxt.Background(), time.Second*30)
	defer done()

	tx, err := db.FrontendWriterDB.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		logger.WithError(err).Errorf("error beginning transaction")
		utils.SetFlash(w, r, authSessionName, "Error: Something went wrong adding your webhook, please try again in a bit.")
		http.Redirect(w, r, "/user/webhooks", http.StatusSeeOther)
		return
	}
	defer tx.Rollback()

	var webhookCount uint64
	err = tx.Get(&webhookCount, `SELECT count(*) from users_webhooks where user_id = $1`, user.UserID)
	if err != nil {
		logger.WithError(err).Errorf("error getting webhook count")
		utils.SetFlash(w, r, authSessionName, "Error: Something went wrong adding your webhook, please try again in a bit.")
		http.Redirect(w, r, "/user/webhooks", http.StatusSeeOther)
		return
	}

	allowed := uint64(1)

	var activeAPP uint64
	err = tx.Get(&activeAPP, `SELECT count(*) from users_app_subscriptions where active = 't' and user_id = $1;`, user.UserID)
	if err != nil {
		logger.WithError(err).Errorf("error getting app subscription count")
		utils.SetFlash(w, r, authSessionName, "Error: Something went wrong adding your webhook, please try again in a bit.")
		http.Redirect(w, r, "/user/webhooks", http.StatusSeeOther)
		return
	}

	if activeAPP > 0 {
		allowed = 2
	}

	var activeAPI uint64
	err = db.FrontendWriterDB.GetContext(ctx, &activeAPI, `SELECT count(*) from users_stripe_subscriptions us join users u on u.stripe_customer_id = us.customer_id where active = 't' and u.id = $1;`, user.UserID)
	if err != nil {
		logger.WithError(err).Errorf("error getting api subscription count")
		utils.SetFlash(w, r, authSessionName, "Error: Something went wrong adding your webhook, please try again in a bit.")
		http.Redirect(w, r, "/user/webhooks", http.StatusSeeOther)
		return
	}

	if activeAPI > 0 {
		allowed = 5
	}

	if webhookCount >= allowed {
		http.Error(w, "Too man webhooks exist already", 400)
		utils.SetFlash(w, r, authSessionName, "Error: We could not add another webhook, you've already reached the maximum allowed, which.")
		http.Redirect(w, r, "/user/webhooks", http.StatusSeeOther)
		return
	}

	_, err = tx.Exec(`INSERT INTO users_webhooks (user_id, url, event_names, destination) VALUES ($1, $2, $3, $4)`, user.UserID, urlForm, pq.StringArray(eventNames), destination)
	if err != nil {
		logger.WithError(err).Errorf("error inserting a new webhook for user")
		utils.SetFlash(w, r, authSessionName, "Error: Something went wrong adding your webhook, please try again in a bit.")
		http.Redirect(w, r, "/user/webhooks", http.StatusSeeOther)
		return
	}

	err = tx.Commit()
	if err != nil {
		logger.WithError(err).Errorf("error for %v route: %v", r.URL.String(), err)
		http.Redirect(w, r, "/user/webhooks", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/user/webhooks", http.StatusSeeOther)
}

func UsersEditWebhook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	user := getUser(r)

	err := r.ParseForm()
	if err != nil {
		logger.WithError(err).Errorf("error parsing form")
		utils.SetFlash(w, r, authSessionName, "Error: Something went wrong editing your webhook, please try again in a bit.")
		http.Redirect(w, r, "/user/webhooks", http.StatusSeeOther)
		return
	}

	vars := mux.Vars(r)

	webhookID := vars["webhookID"]

	// const VALIDATOR_EVENTS = ['validator_attestation_missed', 'validator_proposal_missed', 'validator_proposal_submitted', 'validator_got_slashed', 'validator_synccommittee_soon']
	// const MONITORING_EVENTS = ['monitoring_machine_offline', 'monitoring_hdd_almostfull', 'monitoring_cpu_load']

	urlForm := r.FormValue("url")

	destination := "webhook"

	validatorIsOffline := r.FormValue(string(types.ValidatorIsOfflineEventName)) == "on"
	validatorProposalMissed := r.FormValue(string(types.ValidatorMissedProposalEventName)) == "on"
	validatorProposalSubmitted := r.FormValue(string(types.ValidatorExecutedProposalEventName)) == "on"
	validatorGotSlashed := r.FormValue(string(types.ValidatorGotSlashedEventName)) == "on"
	validatorSyncCommiteeSoon := r.FormValue(string(types.SyncCommitteeSoon)) == "on"
	validatorAttestationMissed := r.FormValue(string(types.ValidatorMissedAttestationEventName)) == "on"
	monitoringMachineOffline := r.FormValue(string(types.MonitoringMachineOfflineEventName)) == "on"
	monitoringHddAlmostfull := r.FormValue(string(types.MonitoringMachineDiskAlmostFullEventName)) == "on"
	monitoringCpuLoad := r.FormValue(string(types.MonitoringMachineCpuLoadEventName)) == "on"
	discord := r.FormValue("discord") == "on"

	if discord {
		destination = "webhook_discord"
	}

	all := r.FormValue("all") == "on"

	events := make(map[string]bool, 0)

	events[string(types.ValidatorIsOfflineEventName)] = validatorIsOffline
	events[string(types.ValidatorMissedProposalEventName)] = validatorProposalMissed
	events[string(types.ValidatorExecutedProposalEventName)] = validatorProposalSubmitted
	events[string(types.ValidatorGotSlashedEventName)] = validatorGotSlashed
	events[string(types.SyncCommitteeSoon)] = validatorSyncCommiteeSoon
	events[string(types.ValidatorMissedAttestationEventName)] = validatorAttestationMissed
	events[string(types.MonitoringMachineOfflineEventName)] = monitoringMachineOffline
	events[string(types.MonitoringMachineDiskAlmostFullEventName)] = monitoringHddAlmostfull
	events[string(types.MonitoringMachineCpuLoadEventName)] = monitoringCpuLoad

	eventNames := make([]string, 0)

	for eventName, active := range events {
		if active || all {
			eventNames = append(eventNames, eventName)
		}
	}

	ctx, done := ctxt.WithTimeout(ctxt.Background(), time.Second*30)
	defer done()

	tx, err := db.FrontendWriterDB.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		logger.WithError(err).Errorf("error beginning transaction")
		utils.SetFlash(w, r, authSessionName, "Error: Something went wrong editing your webhook, please try again in a bit.")
		http.Redirect(w, r, "/user/webhooks", http.StatusSeeOther)
		return
	}
	defer tx.Rollback()

	urlValid := ""

	urlParsed, err := url.Parse(urlForm)
	if err != nil {
		logger.WithError(err).Errorf("could not parse url: %v", urlForm)
		utils.SetFlash(w, r, authSessionName, "Error: the URL you have provided is invalid.")
		http.Redirect(w, r, "/user/webhooks", http.StatusSeeOther)
		return
	}

	if urlParsed != nil {
		urlValid = urlForm
	}

	_, err = tx.Exec(`UPDATE users_webhooks set url = $1, event_names = $2, destination = $3 where user_id = $4 and id = $5`, urlValid, pq.StringArray(eventNames), destination, user.UserID, webhookID)
	if err != nil {
		logger.WithError(err).Errorf("error update webhook for user")
		utils.SetFlash(w, r, authSessionName, "Error: Something went wrong editing your webhook, please try again in a bit.")
		http.Redirect(w, r, "/user/webhooks", http.StatusSeeOther)
		return
	}

	err = tx.Commit()
	if err != nil {
		logger.WithError(err).Errorf("error for %v route: %v", r.URL.String(), err)
		utils.SetFlash(w, r, authSessionName, "Error: Something went wrong editing your webhook, please try again in a bit.")
		http.Redirect(w, r, "/user/webhooks", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/user/webhooks", http.StatusSeeOther)
}

func UsersDeleteWebhook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	user := getUser(r)

	vars := mux.Vars(r)

	webhookID := vars["webhookID"]
	ctx, done := ctxt.WithTimeout(ctxt.Background(), time.Second*30)
	defer done()

	tx, err := db.FrontendWriterDB.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		logger.WithError(err).Errorf("error beginning transaction")
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec(`DELETE FROM users_webhooks where user_id = $1 and id = $2`, user.UserID, webhookID)
	if err != nil {
		logger.WithError(err).Errorf("error update webhook for user")
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}

	err = tx.Commit()
	if err != nil {
		logger.WithError(err).Errorf("error for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
	http.Redirect(w, r, "/user/webhooks", http.StatusSeeOther)
}

// UsersNotificationChannel
// Accepts form encoded values channel and active to set the global notification settings for a user
func UsersNotificationChannels(w http.ResponseWriter, r *http.Request) {
	user := getUser(r)

	err := r.ParseForm()
	if err != nil {
		logger.Errorf("error parsing form: %v", err)
		// session.AddFlash(authInternalServerErrorFlashMsg)
		// session.Save(r, w)
		http.Redirect(w, r, "/user/notifications", http.StatusSeeOther)
		return
	}

	channelEmail := r.FormValue(string(types.EmailNotificationChannel))
	channelPush := r.FormValue(string(types.PushNotificationChannel))
	channelWebhook := r.FormValue(string(types.WebhookNotificationChannel))

	tx, err := db.FrontendWriterDB.Beginx()
	if err != nil {
		logger.WithError(err).Error("error beginning transaction")
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec(`INSERT INTO users_notification_channels (user_id, channel, active) VALUES ($1, $2, $3) ON CONFLICT (user_id, channel) DO UPDATE SET active = $3`, user.UserID, types.EmailNotificationChannel, channelEmail == "on")
	if err != nil {
		logger.WithError(err).Error("error updating users_notification_channels")
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
	_, err = tx.Exec(`INSERT INTO users_notification_channels (user_id, channel, active) VALUES ($1, $2, $3) ON CONFLICT (user_id, channel) DO UPDATE SET active = $3`, user.UserID, types.PushNotificationChannel, channelPush == "on")
	if err != nil {
		logger.WithError(err).Error("error updating users_notification_channels")
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
	_, err = tx.Exec(`INSERT INTO users_notification_channels (user_id, channel, active) VALUES ($1, $2, $3) ON CONFLICT (user_id, channel) DO UPDATE SET active = $3`, user.UserID, types.WebhookNotificationChannel, channelWebhook == "on")
	if err != nil {
		logger.WithError(err).Error("error updating users_notification_channels")
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}

	err = tx.Commit()
	if err != nil {
		logger.WithError(err).Error("error committing transaction")
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}

	http.Redirect(w, r, "/user/notifications", http.StatusSeeOther)
}

// UserSettings renders the user-template
func UserGlobalNotification(w http.ResponseWriter, r *http.Request) {
	isAdmin, user := handleAdminPermissions(w, r)
	if !isAdmin {
		return
	}

	templateFiles := append(layoutTemplateFiles, "user/global_notification.html")
	var userTemplate = templates.GetTemplate(templateFiles...)

	w.Header().Set("Content-Type", "text/html")

	type notificationConfig struct {
		Target  string
		Content string
		Enabled bool
	}

	var configs []*notificationConfig

	err := db.WriterDb.Select(&configs, "SELECT target, content, enabled FROM global_notifications WHERE target = $1 ORDER BY target", utils.Config.Chain.Name)
	if err != nil {
		logger.Errorf("error retrieving globalNotificationMessage: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if len(configs) == 0 {
		_, err = db.WriterDb.Exec("INSERT INTO global_notifications VALUES ($1, '', false)", utils.Config.Chain.Name)
		if err != nil {
			logger.Errorf("error creating default global notification entry: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		configs = append(configs, &notificationConfig{
			Target:  utils.Config.Chain.Name,
			Enabled: false,
			Content: "",
		})
	}

	pageData := &struct {
		CsrfField                  template.HTML
		GlobalNotificationMessages []*notificationConfig
	}{}
	pageData.GlobalNotificationMessages = configs
	pageData.CsrfField = csrf.TemplateField(r)

	data := InitPageData(w, r, "user", "/user/global_notification", "Global Notification", templateFiles)
	data.Data = pageData
	data.User = user

	if handleTemplateError(w, r, "user.go", "UserGlobalNotification", "", userTemplate.ExecuteTemplate(w, "layout", data)) != nil {
		return // an error has occurred and was processed
	}
}

// LoginPost handles authenticating the user.
func UserGlobalNotificationPost(w http.ResponseWriter, r *http.Request) {
	isAdmin, _ := handleAdminPermissions(w, r)
	if !isAdmin {
		return
	}

	err := r.ParseForm()
	if err != nil {
		logger.Errorf("error parsing form: %v", err)
		http.Redirect(w, r, "/user/global_notification", http.StatusSeeOther)
		return
	}

	var targets []string
	err = db.WriterDb.Select(&targets, "SELECT target FROM global_notifications WHERE target = $1", utils.Config.Chain.Name)
	if err != nil {
		logger.Errorf("error retrieving targets: %v", err)
		http.Redirect(w, r, "/user/global_notification", http.StatusSeeOther)
		return
	}

	for _, target := range targets {
		content := r.FormValue("content_" + target)
		enabledText := r.FormValue("enabled_" + target)

		enabled := false
		if enabledText == "on" {
			enabled = true
		}
		_, err = db.WriterDb.Exec("UPDATE global_notifications SET content = $1, enabled = $2 WHERE target = $3", content, enabled, target)

		if err != nil {
			logger.Errorf("error setting globalNotificationMessage: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}

	// Index(w, r)
	http.Redirect(w, r, "/user/global_notification", http.StatusSeeOther)
}

// returns true if admin permissions are available, otherwise http.Error is called and false is returned
func handleAdminPermissions(w http.ResponseWriter, r *http.Request) (bool, *types.User) {
	user, _, err := getUserSession(r)
	if err != nil {
		utils.LogError(err, "error retrieving session", 0)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return false, user
	}

	if user.UserGroup != "ADMIN" {
		http.Error(w, "Insufficient privileges", http.StatusUnauthorized)
		return false, user
	}

	return true, user
}
