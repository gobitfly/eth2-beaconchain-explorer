package handlers

import (
	"encoding/json"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"net/http"

	"github.com/gorilla/mux"
)

func ResolveEnsDomain(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	ens := vars["domain"]
	data := &types.EnsResolveDomainResponse{}
	j := json.NewEncoder(w)

	address, err := utils.ResolveEnsDomain(ens)

	if err != nil {
		logger.Warnf("failed to resolve ens \"%v\": %v", ens, err)
		sendErrorResponse(j, r.URL.String(), "failed to resolve ens")
		return
	}

	data.Domain = ens
	data.Address = address

	sendOKResponse(j, r.URL.String(), []interface{}{data})
}
