package server

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/andig/ulm/api"
	"github.com/gorilla/mux"
)

type errorModeJson struct {
	Error error `json:"error"`
}

type chargeModeJson struct {
	Mode string `json:"mode"`
}

// JsonHandler is a middleware that decorates responses with JSON and CORS headers
func JsonHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		h.ServeHTTP(w, r)
	})
}

func RouteLogger(inner http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		inner.ServeHTTP(w, r)
		log.Printf(
			"%s\t%s\t%s",
			r.Method,
			r.RequestURI,
			time.Since(start),
		)
	}
}

type DebugLogger struct {
	pattern string
}

func (d DebugLogger) Write(p []byte) (n int, err error) {
	s := string(p)
	if strings.Contains(s, d.pattern) {
		debug.PrintStack()
	}
	return os.Stderr.Write(p)
}

func AllChargeModesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res := []string{"now", "pv", "minpv"}
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(res); err != nil {
			log.Printf("httpd: failed to encode JSON: %s", err.Error())
		}
	}
}

func CurrentChargeModeHandler(lp api.LoadPoint) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res := chargeModeJson{
			Mode: string(lp.CurrentChargeMode()),
		}

		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(res); err != nil {
			log.Printf("httpd: failed to encode JSON: %s", err.Error())
		}
	}
}

func ChargeModeHandler(lp api.LoadPoint) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		mode, ok := vars["mode"]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if err := lp.SetChargeMode(api.ChargeMode(mode)); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			res := errorModeJson{
				Error: err,
			}
			if err := json.NewEncoder(w).Encode(res); err != nil {
				log.Printf("httpd: failed to encode JSON: %s", err.Error())
			}
			return
		}

		res := chargeModeJson{
			Mode: string(lp.CurrentChargeMode()),
		}

		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(res); err != nil {
			log.Printf("httpd: failed to encode JSON: %s", err.Error())
		}
	}
}
