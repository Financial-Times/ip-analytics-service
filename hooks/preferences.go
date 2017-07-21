package hooks

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type Preference struct {
	UUID                string `json:"uuid"`
	SuppressedMarketing string `json:"suppressedMarketing"`
}

func preferencesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
	}

	p := &Preference{}
	err = json.Unmarshal(b, p)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	if p.UUID == "" {
		http.Error(w, "Mandatory field missing: UUID", http.StatusBadRequest)
		return
	}

	successHandler(w, r)
}
