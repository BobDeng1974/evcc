package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/andig/ulm/api"
	"github.com/andig/ulm/core"
	"github.com/andig/ulm/server"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

const (
	url = "127.1:7070"
)

type Route struct {
	Methods     []string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

func main() {
	m := &meter{}
	c := &charger{}

	lp := &core.LoadPoint{
		Name:       "lp1",
		Mode:       api.ModeNow,
		GridMeter:  m,
		Charger:    c,
		Phases:     2,
		MinCurrent: 6,  // A
		MaxCurrent: 16, // A
		Log:        log.New(os.Stdout, "", log.LstdFlags),
	}

	loadPoints := []*core.LoadPoint{lp}

	var routes = []Route{
		Route{
			[]string{"GET"},
			"/modes",
			server.AllChargeModesHandler(),
		},
		Route{
			[]string{"GET"},
			"/mode",
			server.CurrentChargeModeHandler(lp),
		},
		Route{
			[]string{"PUT", "POST"},
			"/mode/{mode:[a-z]+}",
			server.ChargeModeHandler(lp),
		},
	}

	router := mux.NewRouter().StrictSlash(true)

	// static
	// router.HandleFunc("/", h.mkIndexHandler())

	// individual handlers per folder
	// for _, folder := range []string{"js", "css"} {
	// 	prefix := fmt.Sprintf("/%s/", folder)
	// 	router.PathPrefix(prefix).Handler(http.StripPrefix(prefix, http.FileServer(_escDir(devAssets, prefix))))
	// }

	// api
	api := router.PathPrefix("/api").Subrouter()
	api.Use(server.JsonHandler)
	for _, r := range routes {
		api.
			Methods(r.Methods...).
			Path(r.Pattern).
			Handler(server.RouteLogger(r.HandlerFunc))
	}

	srv := http.Server{
		Addr:         url,
		Handler:      handlers.CompressHandler(router),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		// ErrorLog: log.New(server.DebugLogger{}, "", 0),
	}
	srv.SetKeepAlivesEnabled(true)

	go func() {
		for range time.Tick(time.Second) {
			for _, lp := range loadPoints {
				go lp.Update()
			}
		}
	}()

	log.Fatal(srv.ListenAndServe())
}
