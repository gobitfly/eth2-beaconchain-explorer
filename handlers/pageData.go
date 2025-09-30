package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gobitfly/eth2-beaconchain-explorer/db"
	"github.com/gobitfly/eth2-beaconchain-explorer/price"
	"github.com/gobitfly/eth2-beaconchain-explorer/services"
	"github.com/gobitfly/eth2-beaconchain-explorer/types"
	"github.com/gobitfly/eth2-beaconchain-explorer/utils"
	"github.com/gobitfly/eth2-beaconchain-explorer/version"
)

var layoutTemplateFiles = []string{
	"layout.html",
	"layout/mainnavigation.html",
	"layout/ad_handler.html",
}

func InitPageData(w http.ResponseWriter, r *http.Request, active, path, title string, mainTemplates []string) *types.PageData {
	fullTitle := fmt.Sprintf("%v - %v - beaconcha.in - %v", title, utils.Config.Frontend.SiteName, time.Now().Year())

	if title == "" {
		fullTitle = fmt.Sprintf("%v - beaconcha.in - %v", utils.Config.Frontend.SiteName, time.Now().Year())
	}

	isMainnet := utils.Config.Chain.ClConfig.ConfigName == "mainnet"

	user, session, _ := getUserSession(r)
	user = checkForV1Notifications(r.Context(), user, session)

	data := &types.PageData{
		Meta: &types.Meta{
			Title:       fullTitle,
			Description: "beaconcha.in makes Ethereum accessible to non-technical end users",
			Path:        path,
			GATag:       utils.Config.Frontend.GATag,
			NoTrack:     false,
			Templates:   strings.Join(mainTemplates, ","),
		},

		Active:                active,
		Data:                  &types.Empty{},
		User:                  user,
		Version:               version.Version,
		Year:                  time.Now().UTC().Year(),
		ChainSlotsPerEpoch:    utils.Config.Chain.ClConfig.SlotsPerEpoch,
		ChainSecondsPerSlot:   utils.Config.Chain.ClConfig.SecondsPerSlot,
		ChainGenesisTimestamp: utils.Config.Chain.GenesisTimestamp,
		CurrentEpoch:          services.LatestEpoch(),
		LatestFinalizedEpoch:  services.LatestFinalizedEpoch(),
		CurrentSlot:           services.LatestSlot(),
		FinalizationDelay:     services.FinalizationDelay(),
		Rates:                 services.GetRates(GetCurrency(r)),
		Mainnet:               utils.Config.Chain.ClConfig.ConfigName == "mainnet" || utils.Config.Chain.ClConfig.ConfigName == "gnosis",
		DepositContract:       utils.Config.Chain.ClConfig.DepositContractAddress,
		ChainConfig:           utils.Config.Chain.ClConfig,
		Lang:                  "en-US",
		NoAds:                 user.Authenticated && user.Subscription != "",
		Debug:                 utils.Config.Frontend.Debug,
		GasNow:                services.LatestGasNowData(),
		ShowSyncingMessage:    services.IsSyncing(),
		GlobalNotification:    services.GlobalNotificationMessage(),
		AvailableCurrencies:   price.GetAvailableCurrencies(),
		MainMenuItems:         createMenuItems(active, isMainnet, user.HasV1Notifications),
		TermsOfServiceUrl:     utils.Config.Frontend.Legal.TermsOfServiceUrl,
		PrivacyPolicyUrl:      utils.Config.Frontend.Legal.PrivacyPolicyUrl,
	}

	adConfigurations, err := db.GetAdConfigurationsForTemplate(mainTemplates, data.NoAds)
	if err != nil {
		utils.LogError(err, fmt.Sprintf("error loading the ad configurations for template %v", path), 0)
	} else {
		data.AdConfigurations = adConfigurations
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

func checkForV1Notifications(ctx context.Context, user *types.User, session *utils.CustomSession) *types.User {
	if !user.Authenticated || user.UserID <= 0 || user.HasV1Notifications != types.UserV1Notification_Unknown {
		return user
	}

	hasV1Notifications, err := hasUserV1NotificationSubscriptions(ctx, user.UserID)
	if err != nil {
		logger.Warnf("error checking v1 notifications for user %v: %v", user.UserID, err)
		return user
	}

	if hasV1Notifications {
		user.HasV1Notifications = types.UserV1Notification_True
	} else {
		user.HasV1Notifications = types.UserV1Notification_False
	}
	session.SetValue("has_v1_notifications", user.HasV1Notifications)

	return user
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
	u.HasV1Notifications, ok = session.GetValue("has_v1_notifications").(types.UserV1Notification)
	if !ok {
		u.HasV1Notifications = types.UserV1Notification_Unknown
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

func createMenuItems(active string, isMain bool, hasV1Notifications types.UserV1Notification) []types.MainMenuItem {
	notificationItems := []types.MainMenuItem{}

	v2NotificationText := "Notifications"
	if hasV1Notifications == types.UserV1Notification_True {
		notificationItems = append(notificationItems, types.MainMenuItem{
			Label:    "v1 Notifications",
			IsActive: false,
			Path:     "/user/notifications",
		})
		v2NotificationText = "v2 Notifications"
	}
	notificationItems = append(notificationItems, types.MainMenuItem{
		Label:    v2NotificationText,
		IsActive: false,
		Path:     utils.Config.V2NotificationURL,
	})

	if utils.Config.Chain.Name == "gnosis" {
		return createMenuItemsGnosis(active, isMain, notificationItems)
	}

	hiddenFor := []string{"confirmation", "login", "register"}

	if utils.SliceContains(hiddenFor, active) {
		return []types.MainMenuItem{}
	}

	composed := []types.MainMenuItem{
		{
			Label:    "Blockchain",
			IsActive: active == "blockchain",
			Groups: []types.NavigationGroup{
				{
					Links: []types.NavigationLink{
						{
							Label: "Epochs",
							Path:  "/epochs",
							Icon:  "fa-history",
						},
						{
							Label: "Slots",
							Path:  "/slots",
							Icon:  "fa-cube",
						},
					},
				}, {
					Links: []types.NavigationLink{
						{
							Label: "Blocks",
							Path:  "/blocks",
							Icon:  "fa-cubes",
						},
						{
							Label: "Txs",
							Path:  "/transactions",
							Icon:  "fa-credit-card",
						},
						{
							Label: "Mempool",
							Path:  "/mempool",
							Icon:  "fa-upload",
						},
					},
				},
			},
		},
		{
			Label:    "Validators",
			IsActive: active == "validators",
			Groups: []types.NavigationGroup{
				{
					Links: []types.NavigationLink{
						{
							Label: "Entities",
							Path:  "/entities",
							Icon:  "fa-building",
						},
					},
				},
				{
					Links: []types.NavigationLink{
						{
							Label: "Overview",
							Path:  "/validators",
							Icon:  "fa-table",
						},
						{
							Label: "Slashings",
							Path:  "/validators/slashings",
							Icon:  "fa-user-slash",
						},
					},
				}, {
					Links: []types.NavigationLink{
						{
							Label: "Validator Leaderboard",
							Path:  "/validators/leaderboard",
							Icon:  "fa-medal",
						},
						{
							Label: "Deposit Leaderboard",
							Path:  "/validators/deposit-leaderboard",
							Icon:  "fa-file-import",
						},
					},
				}, {
					Links: []types.NavigationLink{
						{
							Label: "Deposits",
							Path:  "/validators/deposits",
							Icon:  "fa-file-signature",
						},
						{
							Label: "Withdrawals",
							Path:  "/validators/withdrawals",
							Icon:  "fa-money-bill",
						},
					},
				},
			},
		},
		{
			Label:    "Dashboard",
			IsActive: active == "dashboard",
			Path:     "/dashboard",
		},
	}

	composed = append(composed, notificationItems...)

	composed = append(composed, types.MainMenuItem{
		Label:        "More",
		IsActive:     active == "more",
		HasBigGroups: true,
		Groups: []types.NavigationGroup{
			{
				Label: "Staking Pools",
				Links: []types.NavigationLink{
					{
						Label:         "Run a Validator!",
						Path:          "https://ethpool.org/",
						CustomIcon:    "ethermine_staking_logo_svg",
						IsHighlighted: true,
					},
					{
						Label:      "ETH.STORE®",
						Path:       "/ethstore",
						CustomIcon: "ethermine_stake_logo_svg",
					},
					{
						Label: "Staking Services",
						Path:  "/stakingServices",
						Icon:  "fa-drumstick-bite",
					},
					{
						Label: "Rocket Pool Stats",
						Path:  "/pools/rocketpool",
						Icon:  "fa-rocket",
					},
				},
			},
			{
				Label: "Stats",
				Links: []types.NavigationLink{
					{
						Label: "Charts",
						Path:  "/charts",
						Icon:  "fa-chart-bar",
					},
					{
						Label: "Profit Calculator",
						Path:  "/calculator",
						Icon:  "fa-calculator",
					},
					{
						Label: "Block Viz",
						Path:  "/vis",
						Icon:  "fa-project-diagram",
					},
					{
						Label: "Relays",
						Path:  "/relays",
						Icon:  "fa-robot",
					},
					{
						Label: "EIP-1559 Burn",
						Path:  "/burn",
						Icon:  "fa-burn",
					},
					{
						Label:    "Correlations",
						Path:     "/correlations",
						Icon:     "fa-chart-line",
						IsHidden: !isMain,
					},
				},
			}, {
				Label: "Tools",
				Links: []types.NavigationLink{
					{
						Label: "beaconcha.in App",
						Path:  "/mobile",
						Icon:  "fa-mobile-alt",
					},
					{
						Label: "beaconcha.in Premium",
						Path:  "/premium",
						Icon:  "fa-gem",
					},
					{
						Label:      "Webhooks",
						Path:       "/user/webhooks",
						CustomIcon: "webhook_logo_svg",
					},
					{
						Label: "API Docs",
						Path:  "/api/v1/docs",
						Icon:  "fa-book-reader",
					},
					{
						Label: "API Pricing",
						Path:  "/pricing",
						Icon:  "fa-laptop-code",
					},
					{
						Label: "Unit Converter",
						Path:  "/tools/unitConverter",
						Icon:  "fa-sync",
					},
					{
						Label: "GasNow",
						Path:  "/gasnow",
						Icon:  "fa-gas-pump",
					},
					{
						Label: "Broadcast Signed Messages",
						Path:  "/tools/broadcast",
						Icon:  "fa-bullhorn",
					},
				},
			}, {
				Label: "Services",
				Links: []types.NavigationLink{
					{
						Label: "Knowledge Base",
						Path:  "https://kb.beaconcha.in",
						Icon:  "fa-external-link-alt",
					},
					{
						Label: "Notifications",
						Path:  "/user/notifications",
						Icon:  "fa-bell",
					},
					{
						Label: "Graffiti Wall",
						Path:  "/graffitiwall",
						Icon:  "fa-paint-brush",
					},
					{
						Label: "Ethereum Clients",
						Path:  "/ethClients",
						Icon:  "fa-desktop",
					},
					{
						Label: "Slot Finder",
						Path:  "/slots/finder",
						Icon:  "fa-cube",
					},
					{
						Label: "Report a scam",
						Path:  "https://www.chainabuse.com/report?source=bitfly",
						Icon:  "fa-flag",
					},
				},
			},
		},
	})
	return composed
}

func createMenuItemsGnosis(active string, isMain bool, notificationItems []types.MainMenuItem) []types.MainMenuItem {
	hiddenFor := []string{"confirmation", "login", "register"}

	if utils.SliceContains(hiddenFor, active) {
		return []types.MainMenuItem{}
	}
	composed := []types.MainMenuItem{
		{
			Label:    "Blockchain",
			IsActive: active == "blockchain",
			Groups: []types.NavigationGroup{
				{
					Links: []types.NavigationLink{
						{
							Label: "Epochs",
							Path:  "/epochs",
							Icon:  "fa-history",
						},
						{
							Label: "Slots",
							Path:  "/slots",
							Icon:  "fa-cube",
						},
					},
				}, {
					Links: []types.NavigationLink{
						{
							Label: "Blocks",
							Path:  "/blocks",
							Icon:  "fa-cubes",
						},
						{
							Label: "Txs",
							Path:  "/transactions",
							Icon:  "fa-credit-card",
						},
						{
							Label: "Mempool",
							Path:  "/mempool",
							Icon:  "fa-upload",
						},
					},
				},
			},
		},
		{
			Label:    "Validators",
			IsActive: active == "validators",
			Groups: []types.NavigationGroup{
				{
					Links: []types.NavigationLink{
						{
							Label: "Overview",
							Path:  "/validators",
							Icon:  "fa-table",
						},
						{
							Label: "Slashings",
							Path:  "/validators/slashings",
							Icon:  "fa-user-slash",
						},
					},
				}, {
					Links: []types.NavigationLink{
						{
							Label: "Validator Leaderboard",
							Path:  "/validators/leaderboard",
							Icon:  "fa-medal",
						},
						{
							Label: "Deposit Leaderboard",
							Path:  "/validators/deposit-leaderboard",
							Icon:  "fa-file-import",
						},
					},
				}, {
					Links: []types.NavigationLink{
						{
							Label: "Deposits",
							Path:  "/validators/deposits",
							Icon:  "fa-file-signature",
						},
						{
							Label: "Withdrawals",
							Path:  "/validators/withdrawals",
							Icon:  "fa-money-bill",
						},
					},
				},
			},
		},
		{
			Label:    "Dashboard",
			IsActive: active == "dashboard",
			Path:     "/dashboard",
		},
	}
	composed = append(composed, notificationItems...)
	composed = append(composed, types.MainMenuItem{
		Label:        "More",
		IsActive:     active == "more",
		HasBigGroups: true,
		Groups: []types.NavigationGroup{
			{
				Label: "Stats",
				Links: []types.NavigationLink{
					{
						Label: "Charts",
						Path:  "/charts",
						Icon:  "fa-chart-bar",
					},
					{
						Label: "Block Viz",
						Path:  "/vis",
						Icon:  "fa-project-diagram",
					},
					{
						Label:    "Correlations",
						Path:     "/correlations",
						Icon:     "fa-chart-line",
						IsHidden: !isMain,
					},
				},
			},
			{
				Label: "Tools",
				Links: []types.NavigationLink{
					{
						Label: "beaconcha.in App",
						Path:  "/mobile",
						Icon:  "fa-mobile-alt",
					},
					{
						Label: "beaconcha.in Premium",
						Path:  "/premium",
						Icon:  "fa-gem",
					},
					{
						Label:      "Webhooks",
						Path:       "/user/webhooks",
						CustomIcon: "webhook_logo_svg",
					},
					{
						Label: "API Docs",
						Path:  "/api/v1/docs",
						Icon:  "fa-book-reader",
					},
					{
						Label: "API Pricing",
						Path:  "/pricing",
						Icon:  "fa-laptop-code",
					},
					{
						Label: "Broadcast Signed Messages",
						Path:  "/tools/broadcast",
						Icon:  "fa-bullhorn",
					},
				},
			},
			{
				Label: "Services",
				Links: []types.NavigationLink{
					{
						Label: "Knowledge Base",
						Path:  "https://kb.beaconcha.in",
						Icon:  "fa-external-link-alt",
					},
					{
						Label: "Notifications",
						Path:  "/user/notifications",
						Icon:  "fa-bell",
					},
					{
						Label: "Graffiti Wall",
						Path:  "/graffitiwall",
						Icon:  "fa-paint-brush",
					},
				},
			},
		},
	})
	return composed
}
