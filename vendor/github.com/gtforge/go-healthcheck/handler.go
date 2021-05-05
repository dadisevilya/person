package healthcheck

import (
	"encoding/json"
	"net/http"
)

// Make an http handler to serve /alive calls.
func MakeHealthcheckHandler(hc Checker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response, err := hc.Check(r.Context())
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		data, err := json.Marshal(response)
		if err != nil {
			_, err = w.Write([]byte(`{"alive": true, "error": "Response marshall error"}`))
			if err != nil {
				return
			}
		}
		_, err = w.Write(data)
		if err != nil {
			return
		}
	})
}
