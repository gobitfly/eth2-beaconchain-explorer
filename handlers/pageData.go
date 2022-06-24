package handlers

import (
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

	"github.com/gorilla/sessions"
)

func InitPageData(w http.ResponseWriter, r *http.Request, active, path, title string) *types.PageData {
	user := getUser(r)
	data := &types.PageData{
		HeaderAd: false,
		Meta: &types.Meta{
			Title:       fmt.Sprintf("%v - %v - beaconcha.in - %v", utils.Config.Frontend.SiteName, title, time.Now().Year()),
			Description: "beaconcha.in makes the Ethereum 2.0. beacon chain accessible to non-technical end users",
			Path:        path,
			GATag:       utils.Config.Frontend.GATag,
			NoTrack:     false,
		},
		Active:                active,
		Data:                  &types.Empty{},
		User:                  user,
		Version:               version.Version,
		ChainSlotsPerEpoch:    utils.Config.Chain.Config.SlotsPerEpoch,
		ChainSecondsPerSlot:   utils.Config.Chain.Config.SecondsPerSlot,
		ChainGenesisTimestamp: utils.Config.Chain.GenesisTimestamp,
		CurrentEpoch:          services.LatestEpoch(),
		CurrentSlot:           services.LatestSlot(),
		FinalizationDelay:     services.FinalizationDelay(),
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
		Mainnet:               utils.Config.Chain.Config.ConfigName == "mainnet",
		DepositContract:       utils.Config.Indexer.Eth1DepositContractAddress,
		Currency:              GetCurrency(r),
		CurrentPriceFormatted: GetCurrentPriceFormatted(r),
		CurrentSymbol:         GetCurrencySymbol(r),
		ClientsUpdated:        ethclients.ClientsUpdated(),
		ChainConfig:           utils.Config.Chain.Config,
		Lang:                  "en-US",
		NoAds:                 user.Authenticated && user.Subscription != "",
	}
	data.EthPrice = price.GetEthPrice(data.Currency)
	data.ExchangeRate = price.GetEthPrice(data.Currency)
	data.EthRoundPrice = price.GetEthRoundPrice(data.EthPrice)
	data.EthTruncPrice = utils.KFormatterEthPrice(data.EthRoundPrice)
	data.UsdTruncPrice = utils.KFormatterEthPrice(data.UsdRoundPrice)
	data.EurTruncPrice = utils.KFormatterEthPrice(data.EurRoundPrice)
	data.GbpTruncPrice = utils.KFormatterEthPrice(data.GbpRoundPrice)
	data.CnyTruncPrice = utils.KFormatterEthPrice(data.CnyRoundPrice)
	data.RubTruncPrice = utils.KFormatterEthPrice(data.RubRoundPrice)
	data.CadTruncPrice = utils.KFormatterEthPrice(data.CadRoundPrice)
	data.AudTruncPrice = utils.KFormatterEthPrice(data.AudRoundPrice)
	data.JpyTruncPrice = utils.KFormatterEthPrice(data.JpyRoundPrice)

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

func getUserSession(r *http.Request) (*types.User, *sessions.Session, error) {
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
	u.Authenticated, ok = session.Values["authenticated"].(bool)
	if !ok {
		u.Authenticated = false
		return u, session, nil
	}
	u.UserID, ok = session.Values["user_id"].(uint64)
	if !ok {
		u.Authenticated = false
		return u, session, nil
	}
	u.Subscription, ok = session.Values["subscription"].(string)
	if !ok {
		u.Subscription = ""
		return u, session, nil
	}
	return u, session, nil
}
