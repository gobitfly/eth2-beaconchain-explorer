package handlers

import (
	"context"
	"encoding/hex"
	"eth2-exporter/db"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"
)

// UsersModalAddValidator a validator to the watchlist and subscribes to events
func UsersModalAddValidator(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	user := getUser(r)

	err := r.ParseForm()
	if err != nil {
		utils.LogError(err, "error parsing form", 0)
		utils.SetFlash(w, r, authSessionName, "Error: Something went wrong, please try again in a bit.")
		http.Redirect(w, r, "/user/notifications", http.StatusSeeOther)
		return
	}

	validatorForm := r.FormValue("validator")

	validators := []string{}
	invalidValidators := []string{}
	for _, userInput := range strings.Split(validatorForm, ",") {
		if utils.IsValidEnsDomain(userInput) || utils.IsEth1Address(userInput) {
			searchResult, err := FindValidatorIndicesByEth1Address(userInput)
			if err != nil {
				invalidValidators = append(invalidValidators, userInput)
				continue
			}
			for _, res := range searchResult {
				for _, index := range res.ValidatorIndices {
					validators = append(validators, fmt.Sprintf("%v", index))
				}
			}
		} else if _, err := strconv.ParseUint(userInput, 10, 32); err == nil {
			validators = append(validators, userInput)
		} else {
			invalidValidators = append(invalidValidators, userInput)
		}
	}
	if len(invalidValidators) > 0 {
		desc := "validator"
		if len(invalidValidators) > 1 {
			desc = "validators"
		}
		logger.Warn("Invalid validators when adding to watchlist: ", invalidValidators)
		utils.SetFlash(w, r, authSessionName, fmt.Sprintf("Error: Invalid %s %v. No validators added to the watchlist, please try again in a bit.", desc, strings.Join(invalidValidators, ", ")))
		http.Redirect(w, r, "/user/notifications", http.StatusSeeOther)
		return
	}

	errorMsg := "Error: Something went wrong. No validators added to the watchlist, please try again in a bit."

	pubkeys, err := GetValidatorKeysFrom(validators)
	if err != nil {
		logger.Warnf("Could not find validators when trying to add to watchlist: %v", err)
		utils.SetFlash(w, r, authSessionName, "Error: Could not find validator all validators. No validators added to the watchlist, please try again")
		http.Redirect(w, r, "/user/notifications", http.StatusSeeOther)
		return
	}

	pubKeyStrings := []string{}
	entries := []db.WatchlistEntry{}
	for _, key := range pubkeys {
		keyString := hex.EncodeToString(key)
		pubKeyStrings = append(pubKeyStrings, keyString)
		entries = append(entries, db.WatchlistEntry{UserId: user.UserID, Validator_publickey: keyString})
	}
	err = db.AddToWatchlist(entries, utils.GetNetwork())
	if err != nil {
		logger.WithError(err).Errorf("error adding validators to watchlist: %v", user.UserID)
		utils.SetFlash(w, r, authSessionName, errorMsg)
		http.Redirect(w, r, "/user/notifications", http.StatusSeeOther)
		return
	}

	if err = handleEventSubscriptions(w, r, user, pubKeyStrings); err != nil {
		return
	}

	http.Redirect(w, r, "/user/notifications", http.StatusSeeOther)
}

// UserModalAddNetworkEvent subscribes the user for a network notification
func UserModalAddNetworkEvent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	user := getUser(r)

	err := r.ParseForm()
	if err != nil {
		utils.LogError(err, "error parsing form", 0)
		utils.SetFlash(w, r, authSessionName, "Error: Something went wrong updating your network subscriptions, please try again in a bit.")
		http.Redirect(w, r, "/user/notifications", http.StatusSeeOther)
		return
	}

	for _, ev := range types.NetworkNotificationEvents {
		if r.FormValue(string(ev.Event)) == "on" || r.FormValue("all") == "on" {
			err := db.AddSubscription(user.UserID, utils.GetNetwork(), ev.Event, string(ev.Event), 0)
			if err != nil {
				logger.WithError(err).Errorf("error adding subscription for user: %v", user.UserID)
				utils.SetFlash(w, r, authSessionName, "Error: Something went wrong adding a network subscription, please try again in a bit.")
				http.Redirect(w, r, "/user/notifications", http.StatusSeeOther)
				return
			}
		} else {
			err := db.DeleteSubscription(user.UserID, utils.GetNetwork(), ev.Event, string(ev.Event))
			if err != nil {
				logger.WithError(err).Errorf("error deleting subscription for user: %v", user.UserID)
				utils.SetFlash(w, r, authSessionName, "Error: Something went wrong updating a network subscription, please try again in a bit.")
				http.Redirect(w, r, "/user/notifications", http.StatusSeeOther)
				return
			}
		}
	}

	http.Redirect(w, r, "/user/notifications", http.StatusSeeOther)
}

// UserModalRemoveSelectedValidator a validator to the watchlist and subscribes to events
// Takes the POST of a form with an input field attr name = validators and value = <comam separated list of validator pubkeys>
func UserModalRemoveSelectedValidator(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	user := getUser(r)

	err := r.ParseForm()
	if err != nil {
		utils.LogError(err, "error parsing form", 0)
		utils.SetFlash(w, r, authSessionName, "Error: Something went wrong removing your validators from the watchlist, please try again in a bit.")
		http.Redirect(w, r, "/user/notifications", http.StatusSeeOther)
		return
	}

	validatorsInput := r.FormValue("validators")
	validators := strings.Split(validatorsInput, ",")

	err = db.RemoveFromWatchlistBatch(user.UserID, validators, utils.GetNetwork())
	if err != nil {
		logger.WithError(err).Errorf("error removing validator from watchlist")
		utils.SetFlash(w, r, authSessionName, "Error: Could not remove one or more of your validators.")
	}

	http.Redirect(w, r, "/user/notifications", http.StatusSeeOther)
}

// UserModalManageNotificationModal a validator to the watchlist and subscribes to events
func UserModalManageNotificationModal(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	user := getUser(r)

	err := r.ParseForm()
	if err != nil {
		utils.LogError(err, "error parsing form", 0)
		utils.SetFlash(w, r, authSessionName, "Error: Something went wrong adding your validator to the watchlist, please try again in a bit.")
		http.Redirect(w, r, "/user/notifications", http.StatusSeeOther)
		return
	}

	validatorsForm := r.FormValue("validators")

	validators := strings.Split(validatorsForm, ",")

	events := make(map[types.EventName]bool, 0)
	for _, ev := range types.AddWatchlistEvents {
		events[ev.Event] = r.FormValue(string(ev.Event)) == "on"
	}
	publicKeys, err := GetValidatorKeysFrom(validators)
	if err != nil {
		utils.LogError(err, "error getting validator keys", 0)
		utils.SetFlash(w, r, authSessionName, "Error: Something went wrong updating the validators in your watchlist, please try again in a bit.")
		http.Redirect(w, r, "/user/notifications", http.StatusSeeOther)
		return
	}
	pubKeyStrings := []string{}
	for _, key := range publicKeys {
		pubKeyStrings = append(pubKeyStrings, hex.EncodeToString(key))
	}

	if err = handleEventSubscriptions(w, r, user, pubKeyStrings); err != nil {
		return
	}

	http.Redirect(w, r, "/user/notifications", http.StatusSeeOther)
}

func handleEventSubscriptions(w http.ResponseWriter, r *http.Request, user *types.User, pubKeyStrings []string) error {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Minute*10))
	defer cancel()

	g, gCtx := errgroup.WithContext(ctx)
	events := make(map[types.EventName]bool, 0)
	for _, ev := range types.AddWatchlistEvents {
		events[ev.Event] = r.FormValue(string(ev.Event)) == "on"
	}
	for n, a := range events {
		eventName := n
		active := a
		g.Go(func() error {
			select {
			case <-gCtx.Done():
				return gCtx.Err()
			default:
			}
			logger.Infof("eventName: %v, active: %v", eventName, active)
			if active {
				err := db.AddSubscriptionBatch(user.UserID, utils.GetNetwork(), eventName, pubKeyStrings, 0)
				if err != nil {
					logger.WithError(err).Errorf("error adding subscription for user: %v", user.UserID)
					return err
				}
			} else {
				err := db.DeleteSubscriptionBatch(user.UserID, utils.GetNetwork(), eventName, pubKeyStrings)
				if err != nil {
					logger.WithError(err).Errorf("error deleting subscription for user: %v", user.UserID)
					return err
				}
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		utils.SetFlash(w, r, authSessionName, "Error: Something went wrong updating the validators in your watchlist, please try again in a bit.")
		http.Redirect(w, r, "/user/notifications", http.StatusSeeOther)
		return err
	}
	return nil
}
