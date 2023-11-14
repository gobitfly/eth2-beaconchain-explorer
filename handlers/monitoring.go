package handlers

import (
	"database/sql"
	"errors"
	"eth2-exporter/db"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func Monitoring(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/plain")

	vars := mux.Vars(r)

	module := ""

	switch vars["module"] {
	case "app":
		module = "monitoring_app"
	case "el-data":
		module = "monitoring_el_data"
	case "services":
		module = "monitoring_services"
	case "cl-data":
		module = "monitoring_cl_data"
	case "api":
		module = "monitoring_api"
	case "redis":
		module = "monitoring_redis"
	default:
		http.Error(w, "Invalid monitoring module provided", http.StatusNotFound)
		return
	}

	status := ""
	err := db.WriterDb.Get(&status, "SELECT status FROM service_status WHERE name = $1 AND last_update > NOW() - INTERVAL '5 MINUTES' ORDER BY last_update DESC LIMIT 1", module)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "No monitoring data available", http.StatusNotFound)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	if status == "OK" {
		_, err = fmt.Fprint(w, status)

		if err != nil {
			logger.Debugf("error writing status: %v", err)
		}
	} else {
		http.Error(w, status, http.StatusInternalServerError)
		return
	}

}
