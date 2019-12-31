package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/andig/evcc/api"
	"github.com/andig/evcc/core"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

//go:generate esc -o assets.go -pkg server -modtime 1566640112 -prefix ../assets ../assets

const (
	liveAssets = false
)

type errorModeJson struct {
	Error error `json:"error"`
}

type chargeModeJson struct {
	Mode string `json:"mode"`
}

type route struct {
	Methods     []string
	Pattern     string
	HandlerFunc http.HandlerFunc
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

func IndexHandler(liveAssets bool) http.HandlerFunc {
	template, err := FSString(liveAssets, "/index.html")
	if err != nil {
		log.Fatal("httpd: failed to load embedded template: " + err.Error())
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-8")
		w.WriteHeader(http.StatusOK)

		_, err := fmt.Fprint(w, template)
		if err != nil {
			log.Println("httpd: failed to render main page: ", err.Error())
		}
	})
}

// JSONHandler is a middleware that decorates responses with JSON and CORS headers
func JSONHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		h.ServeHTTP(w, r)
	})
}

// CurrentChargeModeHandler returns current charge mode
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

// ChargeModeHandler updates charge mode
func ChargeModeHandler(lp api.LoadPoint) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		mode, ok := vars["mode"]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if err := lp.ChargeMode(api.ChargeMode(mode)); err != nil {
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

// SocketHandler attaches websocket handler to uri
func SocketHandler(hub *SocketHub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ServeWebsocket(hub, w, r)
	}
}

// NewHttpd creates HTTP server with configured routes for loadpoint
func NewHttpd(url string, lp *core.LoadPoint, hub *SocketHub) *http.Server {
	var routes = []route{
		route{
			[]string{"GET"},
			"/mode",
			CurrentChargeModeHandler(lp),
		},
		route{
			[]string{"PUT", "POST", "OPTIONS"},
			"/mode/{mode:[a-z]+}",
			ChargeModeHandler(lp),
		},
	}

	router := mux.NewRouter().StrictSlash(true)

	// static
	router.HandleFunc("/", IndexHandler(liveAssets))

	// individual handlers per folder
	for _, folder := range []string{"js", "css", "webfonts"} {
		prefix := fmt.Sprintf("/%s/", folder)
		router.PathPrefix(prefix).Handler(http.StripPrefix(prefix, http.FileServer(Dir(liveAssets, prefix))))
	}

	// api
	api := router.PathPrefix("/api").Subrouter()
	api.Use(JSONHandler)
	for _, r := range routes {
		api.
			Methods(r.Methods...).
			Path(r.Pattern).
			Handler(r.HandlerFunc) // RouteLogger(r.HandlerFunc)
	}

	// websocket
	router.HandleFunc("/ws", SocketHandler(hub))

	// add handlers
	handler := handlers.CompressHandler(router)
	handler = handlers.CORS(
		handlers.AllowedHeaders([]string{
			"Accept", "Accept-Language", "Content-Language", "Content-Type", "Origin",
		}),
	)(handler)

	srv := &http.Server{
		Addr:         url,
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		// ErrorLog:     log.New(DebugLogger{}, "", 0),
	}
	srv.SetKeepAlivesEnabled(true)

	return srv
}
