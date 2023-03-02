package handlers

import (
	"context"
	"errors"
	ethclients "eth2-exporter/ethClients"
	"eth2-exporter/price"
	"eth2-exporter/services"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"eth2-exporter/version"
	"fmt"
	"net/http"
	"strings"
	"time"
)

func InitPageData(w http.ResponseWriter, r *http.Request, active, path, title string) *types.PageData {
	fullTitle := fmt.Sprintf("%v - %v - beaconcha.in - %v", title, utils.Config.Frontend.SiteName, time.Now().Year())

	if title == "" {
		fullTitle = fmt.Sprintf("%v - beaconcha.in - %v", utils.Config.Frontend.SiteName, time.Now().Year())
	}

	user := getUser(r)
	data := &types.PageData{
		HeaderAd: false,
		Meta: &types.Meta{
			Title:       fullTitle,
			Description: "beaconcha.in makes Ethereum accessible to non-technical end users",
			Path:        path,
			GATag:       utils.Config.Frontend.GATag,
			NoTrack:     false,
		},
		Active:                active,
		Data:                  &types.Empty{},
		User:                  user,
		Version:               version.Version,
		Year:                  time.Now().UTC().Year(),
		ChainSlotsPerEpoch:    utils.Config.Chain.Config.SlotsPerEpoch,
		ChainSecondsPerSlot:   utils.Config.Chain.Config.SecondsPerSlot,
		ChainGenesisTimestamp: utils.Config.Chain.GenesisTimestamp,
		CurrentEpoch:          services.LatestEpoch(),
		LatestFinalizedEpoch:  services.LatestFinalizedEpoch(),
		CurrentSlot:           services.LatestSlot(),
		FinalizationDelay:     services.FinalizationDelay(),
		Rates: types.PageRates{
			EthPrice:              0,
			EthRoundPrice:         0,
			EthTruncPrice:         "",
			UsdRoundPrice:         price.GetEthRoundPrice(price.GetEthPrice("USD")),
			UsdTruncPrice:         "",
			EurRoundPrice:         price.GetEthRoundPrice(price.GetEthPrice("EUR")),
			EurTruncPrice:         "",
			GbpRoundPrice:         price.GetEthRoundPrice(price.GetEthPrice("GBP")),
			GbpTruncPrice:         "",
			CnyRoundPrice:         price.GetEthRoundPrice(price.GetEthPrice("CNY")),
			CnyTruncPrice:         "",
			RubRoundPrice:         price.GetEthRoundPrice(price.GetEthPrice("RUB")),
			RubTruncPrice:         "",
			CadRoundPrice:         price.GetEthRoundPrice(price.GetEthPrice("CAD")),
			CadTruncPrice:         "",
			AudRoundPrice:         price.GetEthRoundPrice(price.GetEthPrice("AUD")),
			AudTruncPrice:         "",
			JpyRoundPrice:         price.GetEthRoundPrice(price.GetEthPrice("JPY")),
			JpyTruncPrice:         "",
			Currency:              GetCurrency(r),
			CurrentPriceFormatted: GetCurrentPriceFormatted(r),
			CurrentSymbol:         GetCurrencySymbol(r),
		},
		Mainnet:            utils.Config.Chain.Config.ConfigName == "mainnet",
		DepositContract:    utils.Config.Indexer.Eth1DepositContractAddress,
		ClientsUpdated:     ethclients.ClientsUpdated(),
		ChainConfig:        utils.Config.Chain.Config,
		Lang:               "en-US",
		NoAds:              user.Authenticated && user.Subscription != "",
		Debug:              utils.Config.Frontend.Debug,
		GasNow:             services.LatestGasNowData(),
		ShowSyncingMessage: services.IsSyncing(),
		GlobalNotification: services.GlobalNotificationMessage(),
	}

	if utils.Config.Frontend.Debug {
		_, session, err := getUserSession(r)
		if err != nil {
			logger.WithError(err).Error("error getting user session")
		}
		if session != nil {
			jsn := make(map[string]interface{})
			// convert map[interface{}]interface{} -> map[string]interface{}
			for sessionKey, sessionValue := range session.Values() {
				jsn[fmt.Sprintf("%v", sessionKey)] = sessionValue
			}
			data.DebugSession = jsn
		}
	}
	data.Rates.EthPrice = price.GetEthPrice(data.Rates.Currency)
	data.Rates.ExchangeRate = price.GetEthPrice(data.Rates.Currency)
	data.Rates.EthRoundPrice = price.GetEthRoundPrice(data.Rates.EthPrice)
	data.Rates.EthTruncPrice = utils.KFormatterEthPrice(data.Rates.EthRoundPrice)
	data.Rates.UsdTruncPrice = utils.KFormatterEthPrice(data.Rates.UsdRoundPrice)
	data.Rates.EurTruncPrice = utils.KFormatterEthPrice(data.Rates.EurRoundPrice)
	data.Rates.GbpTruncPrice = utils.KFormatterEthPrice(data.Rates.GbpRoundPrice)
	data.Rates.CnyTruncPrice = utils.KFormatterEthPrice(data.Rates.CnyRoundPrice)
	data.Rates.RubTruncPrice = utils.KFormatterEthPrice(data.Rates.RubRoundPrice)
	data.Rates.CadTruncPrice = utils.KFormatterEthPrice(data.Rates.CadRoundPrice)
	data.Rates.AudTruncPrice = utils.KFormatterEthPrice(data.Rates.AudRoundPrice)
	data.Rates.JpyTruncPrice = utils.KFormatterEthPrice(data.Rates.JpyRoundPrice)

	acceptedLangs := strings.Split(r.Header.Get("Accept-Language"), ",")
	if len(acceptedLangs) > 0 {
		if strings.Contains(acceptedLangs[0], "ru") || strings.Contains(acceptedLangs[0], "RU") {
			data.Lang = "ru-RU"
		}
	}

	for _, v := range r.Cookies() {
		if v.Name == "language" {
			data.Lang = v.Value
			break
		}
	}

	return data
}

func SetPageDataTitle(pageData *types.PageData, title string) {
	if title == "" {
		pageData.Meta.Title = fmt.Sprintf("%v - beaconcha.in - %v", utils.Config.Frontend.SiteName, time.Now().Year())
	} else {
		pageData.Meta.Title = fmt.Sprintf("%v - %v - beaconcha.in - %v", title, utils.Config.Frontend.SiteName, time.Now().Year())
	}
}

func getUser(r *http.Request) *types.User {
	if IsMobileAuth(r) {
		claims := getAuthClaims(r)
		u := &types.User{}
		u.UserID = claims.UserID
		u.Authenticated = true
		return u
	} else {
		return getUserFromSessionStore(r)
	}
}

func getUserFromSessionStore(r *http.Request) *types.User {
	u, _, _ := getUserSession(r)
	return u
}

func getUserSession(r *http.Request) (*types.User, *utils.CustomSession, error) {
	u := &types.User{}
	if utils.SessionStore == nil { // sanity check for production deployment where api runs independ of frontend and has no initialized sessionstore
		return u, nil, errors.New("sessionstore not initialized")
	}
	session, err := utils.SessionStore.Get(r, authSessionName)
	if err != nil {
		logger.Errorf("error getting session from sessionStore: %v", err)
		return u, session, err
	}
	ok := false
	u.Authenticated, ok = session.GetValue("authenticated").(bool)
	if !ok {
		u.Authenticated = false
		return u, session, nil
	}
	u.UserID, ok = session.GetValue("user_id").(uint64)
	if !ok {
		u.Authenticated = false
		return u, session, nil
	}
	u.Subscription, ok = session.GetValue("subscription").(string)
	if !ok {
		u.Subscription = ""
		return u, session, nil
	}
	u.UserGroup, ok = session.GetValue("user_group").(string)
	if !ok {
		u.UserGroup = ""
		return u, session, nil
	}
	return u, session, nil
}

func purgeAllSessionsForUser(ctx context.Context, userId uint64) error {
	// invalidate all sessions for this user
	err := utils.SessionStore.SCS.Iterate(ctx, func(ctx context.Context) error {
		sessionUserID, ok := utils.SessionStore.SCS.Get(ctx, "user_id").(uint64)
		if !ok {
			return nil
		}

		if userId == sessionUserID {
			return utils.SessionStore.SCS.Destroy(ctx)
		}

		return nil
	})

	return err

}
