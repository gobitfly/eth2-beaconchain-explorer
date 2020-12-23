package handlers

import (
	ethclients "eth2-exporter/ethClients"
	"eth2-exporter/price"
	"eth2-exporter/services"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"eth2-exporter/version"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/sessions"
)

func InitPageData(w http.ResponseWriter, r *http.Request, active, path, title string) *types.PageData {
	data := &types.PageData{
		HeaderAd: false,
		Meta: &types.Meta{
			Title:       fmt.Sprintf("%v - %v - beaconcha.in - %v", utils.Config.Frontend.SiteName, title, time.Now().Year()),
			Description: "beaconcha.in makes the Ethereum 2.0. beacon chain accessible to non-technical end users",
			Path:        path,
			GATag:       utils.Config.Frontend.GATag,
		},
		Active:                active,
		Data:                  &types.Empty{},
		User:                  getUser(w, r),
		Version:               version.Version,
		ChainSlotsPerEpoch:    utils.Config.Chain.SlotsPerEpoch,
		ChainSecondsPerSlot:   utils.Config.Chain.SecondsPerSlot,
		ChainGenesisTimestamp: utils.Config.Chain.GenesisTimestamp,
		CurrentEpoch:          services.LatestEpoch(),
		CurrentSlot:           services.LatestSlot(),
		FinalizationDelay:     services.FinalizationDelay(),
		EthPrice:              price.GetEthPrice("USD"),
		Mainnet:               utils.Config.Chain.Mainnet,
		DepositContract:       utils.Config.Indexer.Eth1DepositContractAddress,
		Currency:              GetCurrency(r),
		InfoBanner:            ethclients.GetBannerClients(),
	}
	data.ExchangeRate = price.GetEthPrice(data.Currency)

	return data
}

func getUser(w http.ResponseWriter, r *http.Request) *types.User {
	if IsMobileAuth(r) {
		claims := getAuthClaims(r)
		u := &types.User{}
		u.UserID = claims.UserID
		u.Authenticated = true
		return u
	} else {
		return getUserFromSessionStore(w, r)
	}
}

func getUserFromSessionStore(w http.ResponseWriter, r *http.Request) *types.User {
	u := &types.User{}
	session, err := utils.SessionStore.Get(r, authSessionName)
	if err != nil {
		logger.Errorf("error getting session from sessionStore: %v", err)
		return u
	}
	ok := false
	u.Authenticated, ok = session.Values["authenticated"].(bool)
	if !ok {
		u.Authenticated = false
		return u
	}
	u.UserID, ok = session.Values["user_id"].(uint64)
	if !ok {
		u.Authenticated = false
		return u
	}
	return u
}

func getUserSession(w http.ResponseWriter, r *http.Request) (*types.User, *sessions.Session, error) {
	u := &types.User{}
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
	return u, session, nil
}
