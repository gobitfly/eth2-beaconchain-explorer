package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gobitfly/eth2-beaconchain-explorer/cache"
	"github.com/gobitfly/eth2-beaconchain-explorer/db"
	"github.com/gobitfly/eth2-beaconchain-explorer/types"
	"github.com/gobitfly/eth2-beaconchain-explorer/utils"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
)

// ApiEnsLookup godoc
// @Tags ENS
// @Summary Resolve an ens name or address
// @Description Get the address for an ens name and vice versa. Returns and object with the ens name and address - if found.
// @Produce  json
// @Param domain path string true "domain can either be an ens name or an etherum address"
// @Success 200 {object} types.ApiResponse
// @Failure 400 {object} types.ApiResponse
// @Router /api/v1/ens/lookup/{domain} [get]
func ResolveEnsDomain(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	search := vars["domain"]

	data, err := GetEnsDomain(search)

	if err != nil {
		logger.Warnf("failed to resolve ens %v: %v", search, err)
		SendBadRequestResponse(w, r.URL.String(), "failed to resolve ens")
		return
	}

	j := json.NewEncoder(w)
	SendOKResponse(j, r.URL.String(), []interface{}{data})
}

func GetEnsDomain(search string) (*types.EnsDomainResponse, error) {
	data := &types.EnsDomainResponse{}
	var returnError error

	if utils.IsValidEnsDomain(search) {
		cacheKey := fmt.Sprintf("%d:ens:address:%v", utils.Config.Chain.ClConfig.DepositChainID, search)

		if address, err := cache.TieredCache.GetStringWithLocalTimeout(cacheKey, time.Minute); err == nil && len(address) > 0 {
			data.Address = address
			return data, nil
		}

		address, err := db.GetAddressForEnsName(search)
		if err != nil {
			data.Domain = search
			return data, err // We want to return the data if it was a valid domain even if there was an error getting the address from bigtable. A valid domain might be enough for the caller.
		}
		data.Address = address.Hex()

		name, err := db.GetEnsNameForAddress(*address)
		if err != nil && err != sql.ErrNoRows {
			return data, err // We want to return the data if it was a valid address even if there was an error getting the domain from bigtable. A valid address might be enough for the caller.
		}
		data.Domain = name

		err = cache.TieredCache.SetString(cacheKey, data.Address, time.Minute)
		if err != nil {
			logger.Errorf("error caching ens address: %v", err)
		}

	} else if utils.IsValidEth1Address(search) {
		data.Address = search

		cacheKey := fmt.Sprintf("%d:ens:domain:%v", utils.Config.Chain.ClConfig.DepositChainID, search)

		if domain, err := cache.TieredCache.GetStringWithLocalTimeout(cacheKey, time.Minute); err == nil && len(domain) > 0 {
			data.Domain = domain
			return data, nil
		}
		name, err := db.GetEnsNameForAddress(common.HexToAddress(search))
		if err != nil && err != sql.ErrNoRows {
			return data, err // We want to return the data if it was a valid address even if there was an error getting the domain from bigtable. A valid address might be enough for the caller.
		}
		data.Domain = name
		err = cache.TieredCache.SetString(cacheKey, data.Domain, time.Minute)
		if err != nil {
			logger.Errorf("error caching ens address: %v", err)
		}
	} else {
		returnError = errors.New("not an ens domain or address")
	}
	return data, returnError //We always want to return the data if it was a valid address/domain even if there was an error getting data. A valid address might be enough for the caller.
}

func ReplaceEnsNameWithAddress(search string) string {
	if utils.IsValidEnsDomain(search) {
		ensData, _ := GetEnsDomain(search)
		if len(ensData.Address) > 0 {
			search = strings.Replace(ensData.Address, "0x", "", -1)
		}
	}
	return search
}
