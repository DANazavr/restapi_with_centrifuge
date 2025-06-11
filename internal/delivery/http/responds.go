package delivery

import (
	"encoding/json"
	"net/http"
)

func HendleRespond(w http.ResponseWriter, r *http.Request, status int, data interface{}) error {
	w.WriteHeader(status)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			return err
		}
	}
	return nil
}

func HendleError(w http.ResponseWriter, r *http.Request, status int, err error) error {
	HendleRespond(w, r, status, map[string]string{"error": err.Error()})
	return err
}
