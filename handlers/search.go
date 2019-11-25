package handlers

import (
	"net/http"
	"strconv"
	"strings"
)

func Search(w http.ResponseWriter, r *http.Request) {

	search := r.FormValue("search")

	_, err := strconv.Atoi(search)

	if err == nil {
		http.Redirect(w, r, "/block/" + search, 301)
		return
	}

	search = strings.Replace(search,"0x", "", -1)

	if len(search) == 64 {
		http.Redirect(w, r, "/block/" + search, 301)
	} else if len(search) == 96 {
		http.Redirect(w, r, "/validator/" + search, 301)
	} else {
		http.Error(w, "Not found", 404)
	}
}