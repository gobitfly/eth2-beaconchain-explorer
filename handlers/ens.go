package handlers

import (
	"encoding/json"
	"errors"
	"eth2-exporter/cache"
	"eth2-exporter/db"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
)

func ResolveEnsDomain(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	search := vars["domain"]

	data, err := GetEnsDomain(search)

	if err != nil {
		logger.Warnf("failed to resolve ens %v: %v", search, err)
		sendErrorResponse(w, r.URL.String(), "failed to resolve ens")
		return
	}

	j := json.NewEncoder(w)
	sendOKResponse(j, r.URL.String(), []interface{}{data})
}

func GetEnsDomain(search string) (*types.EnsDomainResponse, error) {
	data := &types.EnsDomainResponse{}
	var returnError error

	if utils.IsValidEnsDomain(search) {
		data.Domain = search

		cacheKey := fmt.Sprintf("%d:ens:address:%v", utils.Config.Chain.Config.DepositChainID, search)

		if address, err := cache.TieredCache.GetStringWithLocalTimeout(cacheKey, time.Minute); err != nil || len(address) == 0 {
			address, err := db.GetAddressForEnsName(search)

			if err == nil {
				data.Address = address.Hex()

				err := cache.TieredCache.SetString(cacheKey, data.Address, time.Minute)
				if err != nil {
					logger.Errorf("error caching ens address: %v", err)
				}
			} else {
				returnError = err
			}
		} else {
			data.Address = address
		}

	} else if utils.IsValidEth1Address(search) {
		data.Address = search

		cacheKey := fmt.Sprintf("%d:ens:domain:%v", utils.Config.Chain.Config.DepositChainID, search)

		if domain, err := cache.TieredCache.GetStringWithLocalTimeout(cacheKey, time.Minute); err != nil || len(domain) == 0 {
			name, err := db.GetEnsNameForAddress(common.HexToAddress(search))

			if err == nil {
				data.Domain = *name

				err := cache.TieredCache.SetString(cacheKey, data.Domain, time.Minute)
				if err != nil {
					logger.Errorf("error caching ens address: %v", err)
				}
			} else {
				returnError = err
			}
		} else {
			data.Domain = domain
		}
	} else {
		returnError = errors.New("not an ens domain or address")
	}
	return data, returnError //We always want to return the data if it was a valid address/domain even if there was an error getting data from the node. A valid address might be enough for the caller.
}
