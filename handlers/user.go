package handlers

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"eth2-exporter/db"
	"eth2-exporter/mail"
	"eth2-exporter/services"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"eth2-exporter/version"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

var userTemplate = template.Must(template.New("user").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/user/settings.html"))
var notificationTemplate = template.Must(template.New("user").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/user/notifications.html"))

func UserAuthMiddleware(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	user := getUser(w, r)
	if !user.Authenticated {
		logger.Errorf("User not authorized")
		utils.SetFlash(w, r, authSessionName, "Error: Please login first")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	next(w, r)
}

// UserSettings renders the user-template
func UserSettings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	userSettingsData := &types.UserSettingsPageData{}

	user, session, err := getUserSession(w, r)
	if err != nil {
		logger.Errorf("error retrieving session: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	email, err := db.GetUserEmailById(user.UserID)
	if err != nil {
		logger.Errorf("Error retrieving the email for user: %v %v", user.UserID, err)
		session.Flashes("Error: Something went wrong.")
		session.Save(r, w)
		http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
		return
	}

	userSettingsData.Email = email
	userSettingsData.Flashes = utils.GetFlashes(w, r, authSessionName)

	data := &types.PageData{
		HeaderAd: true,
		Meta: &types.Meta{
			Description: "beaconcha.in makes the Ethereum 2.0. beacon chain accessible to non-technical end users",
			Path:        "/user",
			GATag:       utils.Config.Frontend.GATag,
		},
		Active:                "user",
		Data:                  userSettingsData,
		User:                  user,
		Version:               version.Version,
		ChainSlotsPerEpoch:    utils.Config.Chain.SlotsPerEpoch,
		ChainSecondsPerSlot:   utils.Config.Chain.SecondsPerSlot,
		ChainGenesisTimestamp: utils.Config.Chain.GenesisTimestamp,
		CurrentEpoch:          services.LatestEpoch(),
		CurrentSlot:           services.LatestSlot(),
		FinalizationDelay:     services.FinalizationDelay(),
	}
	err = userTemplate.ExecuteTemplate(w, "layout", data)
	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func UserNotifications(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	userNotificationsData := &types.UserNotificationsPageData{}

	user := getUser(w, r)

	userNotificationsData.Flashes = utils.GetFlashes(w, r, authSessionName)

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
	err = db.DB.Get(&countSubscriptions, `
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

	data := &types.PageData{
		HeaderAd: true,
		Meta: &types.Meta{
			Description: "beaconcha.in makes the Ethereum 2.0. beacon chain accessible to non-technical end users",
			Path:        "/user",
			GATag:       utils.Config.Frontend.GATag,
		},
		Active:                "user",
		Data:                  userNotificationsData,
		User:                  user,
		Version:               version.Version,
		ChainSlotsPerEpoch:    utils.Config.Chain.SlotsPerEpoch,
		ChainSecondsPerSlot:   utils.Config.Chain.SecondsPerSlot,
		ChainGenesisTimestamp: utils.Config.Chain.GenesisTimestamp,
		CurrentEpoch:          services.LatestEpoch(),
		CurrentSlot:           services.LatestSlot(),
		FinalizationDelay:     services.FinalizationDelay(),
	}

	err = notificationTemplate.ExecuteTemplate(w, "layout", data)
	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func UserNotificationsData(w http.ResponseWriter, r *http.Request) {
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

	user := getUser(w, r)

	type watchlistSubscription struct {
		Publickey []byte
		Balance   uint64
		Events    *pq.StringArray
	}

	wl := []watchlistSubscription{}
	err = db.DB.Select(&wl, `
	SELECT 
			users_validators_tags.validator_publickey as publickey,
			COALESCE (MAX(validators.balance), 0) as balance,
			ARRAY_REMOVE(ARRAY_AGG(users_subscriptions.event_name), NULL) as events
		FROM users_validators_tags
		LEFT JOIN users_subscriptions
		ON 
			users_validators_tags.user_id = users_subscriptions.user_id
		AND 
			ENCODE(users_validators_tags.validator_publickey::bytea, 'hex') = users_subscriptions.event_filter
		LEFT JOIN validators
		ON
			users_validators_tags.validator_publickey = validators.pubkey
		WHERE users_validators_tags.user_id = $1
		GROUP BY users_validators_tags.user_id, users_validators_tags.validator_publickey;
	`, user.UserID)
	if err != nil {
		logger.Errorf("error retrieving subscriptions for users: %v validators: %v", user.UserID, err)
		http.Error(w, "Internal server error", 503)
		return
	}

	tableData := make([][]interface{}, 0, len(wl))
	for _, entry := range wl {
		tableData = append(tableData, []interface{}{
			utils.FormatPublicKey(entry.Publickey),
			utils.FormatBalance(entry.Balance),
			entry.Events,
			// utils.FormatBalance(item.Balance),
			// item.Events[0],
		})
	}

	// log.Println("COUNT", len(watchlist))
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

	user := getUser(w, r)

	subs := []types.Subscription{}
	err = db.DB.Select(&subs, `
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
		}

		tableData = append(tableData, []interface{}{
			pubkey,
			sub.EventName,
			utils.FormatTimestamp(sub.CreatedTime.Unix()),
			ls,
		})
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

func UserDeletePost(w http.ResponseWriter, r *http.Request) {
	logger := logger.WithField("route", r.URL.String())
	user, session, err := getUserSession(w, r)
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

func UserUpdatePasswordPost(w http.ResponseWriter, r *http.Request) {
	var GenericUpdatePasswordError = "Error: Something went wrong updating your password 😕. If this error persists please contact <a href=\"https://support.bitfly.at/support/home\">support</a>"

	user, session, err := getUserSession(w, r)
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
		session.AddFlash("Error: Email has not been comfirmed, please click the link in the email we sent you or <a href='/resend'>resend link</a>!")
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
	user, session, err := getUserSession(w, r)
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

	_, session, err := getUserSession(w, r)
	if err != nil {
		logger.Errorf("error retrieving session: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	q := r.URL.Query()
	newEmail, err := url.QueryUnescape(q.Get("email"))
	if err != nil {
		utils.SetFlash(w, r, authSessionName, "Error: Could not update your email please try again.")
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
	err = mail.SendMail(newEmail, subject, msg)
	if err != nil {
		return err
	}

	_, err = db.FrontendDB.Exec("UPDATE users SET email_confirmation_ts = TO_TIMESTAMP($1) WHERE id = $2", time.Now().Unix(), userId)
	if err != nil {
		return fmt.Errorf("error updating confirmation-ts: %w", err)
	}

	return nil
}

// UserValidatorWatchlistAdd subscribes a user to get notifications from a specific validator
func UserValidatorWatchlistAdd(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	user := getUser(w, r)
	vars := mux.Vars(r)

	pubKey := strings.Replace(vars["pubkey"], "0x", "", -1)
	if !user.Authenticated {
		utils.SetFlash(w, r, validatorEditFlash, "Error: You need a user account to follow a validator <a href=\"/login\">Login</a> or <a href=\"/register\">Sign up</a>")
		http.Redirect(w, r, "/validator/"+pubKey, http.StatusSeeOther)
		return
	}

	balance := r.FormValue("balance_decreases")
	if balance == "on" {
		err := db.AddSubscription(user.UserID, types.ValidatorBalanceDecreasedEventName, pubKey)
		if err != nil {
			logger.Errorf("error could not ADD subscription for user %v eventName %v eventfilter %v: %v", user.UserID, types.ValidatorBalanceDecreasedEventName, pubKey, err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
	slashed := r.FormValue("validator_slashed")
	if slashed == "on" {
		err := db.AddSubscription(user.UserID, types.ValidatorGotSlashedEventName, pubKey)
		if err != nil {
			logger.Errorf("error could not ADD subscription for user %v eventName %v eventfilter %v: %v", user.UserID, types.ValidatorGotSlashedEventName, pubKey, err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}

	if len(pubKey) != 96 {
		utils.SetFlash(w, r, validatorEditFlash, "Error: Validator not found")
		http.Redirect(w, r, "/validator/"+pubKey, http.StatusSeeOther)
		return
	}

	watchlistEntries := []db.WatchlistEntry{
		{
			UserId:              user.UserID,
			Validator_publickey: pubKey,
		},
	}

	err := db.AddToWatchlist(watchlistEntries)
	if err != nil {
		logger.Errorf("error adding validator to watchlist to db: %v", err)
		utils.SetFlash(w, r, validatorEditFlash, "Error: Could not follow validator.")
		http.Redirect(w, r, "/validator/"+pubKey, http.StatusSeeOther)
		return
	}

	// utils.SetFlash(w, r, validatorEditFlash, "Subscribed to this validator")
	http.Redirect(w, r, "/validator/"+pubKey, http.StatusSeeOther)
}

// UserValidatorWatchlistAdd subscribes a user to get notifications from a specific validator
func UserDashboardWatchlistAdd(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	user := getUser(w, r)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logger.Errorf("error reading body of request: %v, %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	indices := make([]string, 0)
	err = json.Unmarshal(body, &indices)
	if err != nil {
		logger.Errorf("error parsing request body: %v, %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	indicesParsed := make([]int64, 0)
	for _, i := range indices {
		parsed, err := strconv.ParseInt(i, 10, 64)
		if err != nil {
			logger.Errorf("error could not parse validator indices: %v, %v", r.URL.String(), err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		indicesParsed = append(indicesParsed, parsed)
	}

	publicKeys := make([]string, 0)
	db.DB.Select(&publicKeys, `
	SELECT Encode(pubkey::bytea, 'hex') as pubkey
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

	err = db.AddToWatchlist(watchListEntries)
	if err != nil {
		logger.Errorf("error could not add validators to watchlist: %v, %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(200)
}

// UserValidatorWatchlistRemove unsubscribes a user from a specific validator
func UserValidatorWatchlistRemove(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	user := getUser(w, r)
	vars := mux.Vars(r)

	pubKey := strings.Replace(vars["pubkey"], "0x", "", -1)
	if !user.Authenticated {
		utils.SetFlash(w, r, validatorEditFlash, "Error: You need a user account to follow a validator <a href=\"/login\">Login</a> or <a href=\"/register\">Sign up</a>")
		http.Redirect(w, r, "/validator/"+pubKey, http.StatusSeeOther)
		return
	}

	if len(pubKey) != 96 {
		utils.SetFlash(w, r, validatorEditFlash, "Error: Validator not found")
		http.Redirect(w, r, "/validator/"+pubKey, http.StatusSeeOther)
		return
	}

	err := db.RemoveFromWatchlist(user.UserID, pubKey)
	if err != nil {
		logger.Errorf("error deleting subscription: %v", err)
		utils.SetFlash(w, r, validatorEditFlash, "Error: Could not remove bookmark.")
		http.Redirect(w, r, "/validator/"+pubKey, http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/validator/"+pubKey, http.StatusSeeOther)
}

func UserNotificationsSubscribe(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	user := getUser(w, r)
	q := r.URL.Query()
	event := q.Get("event")
	filter := q.Get("filter")
	filter = strings.Replace(filter, "0x", "", -1)

	eventName, err := types.EventNameFromString(event)
	if err != nil {
		logger.Errorf("error invalid event name: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	isPkey := !pkeyRegex.MatchString(filter)

	if len(filter) != 96 && isPkey {
		logger.Errorf("error invalid pubkey characters or length: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = db.AddSubscription(user.UserID, eventName, filter)
	if err != nil {
		logger.Errorf("error could not ADD subscription for user %v eventName %v eventfilter %v: %v", user.UserID, eventName, filter, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(200)
}

func UserNotificationsUnsubscribe(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	user := getUser(w, r)
	q := r.URL.Query()
	event := q.Get("event")
	filter := q.Get("filter")
	filter = strings.Replace(filter, "0x", "", -1)

	eventName, err := types.EventNameFromString(event)
	if err != nil {
		logger.Errorf("error invalid event name: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	isPkey := !pkeyRegex.MatchString(filter)

	if len(filter) != 96 && isPkey {
		logger.Errorf("error invalid pubkey characters or length: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = db.DeleteSubscription(user.UserID, eventName, filter)
	if err != nil {
		logger.Errorf("error could not REMOVE subscription for user %v eventName %v eventfilter %v: %v", user.UserID, eventName, filter, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(200)
}
