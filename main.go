package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/andig/ulm/api"
	"github.com/andig/ulm/core"
	"github.com/andig/ulm/exec"
	"github.com/andig/ulm/server"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

const (
	url     = "127.1:7070"
	timeout = 1 * time.Second
)

type route struct {
	Methods     []string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type charger struct {
	api.Charger
	api.ChargeController
}

var loadPoints []*core.LoadPoint

func updateLoadPoints() {
	for _, lp := range loadPoints {
		go lp.Update()
	}
}

func logEnabled() bool {
	env := strings.TrimSpace(os.Getenv("ULM_DEBUG"))
	return env != "" && env != "0"
}

func main() {
	if true || logEnabled() {
		logger := log.New(os.Stdout, "", log.LstdFlags)
		core.Logger = logger
		exec.Logger = logger
	}

	m := exec.NewMeter("/bin/bash -c 'echo $((RANDOM % 1000))'", timeout)
	c := &charger{
		exec.NewCharger(
			"/bin/bash -c 'echo C'",
			"/bin/bash -c 'echo $((RANDOM % 32))'",
			"/bin/bash -c 'echo true'",
			"/bin/bash -c 'echo true'",
			timeout,
		),
		exec.NewChargeController(
			"/bin/bash -c 'echo $((RANDOM % 1000))'",
			timeout,
		),
	}

	lp := &core.LoadPoint{
		Name:       "lp1",
		Mode:       api.ModePV,
		GridMeter:  m,
		Charger:    c,
		Phases:     2,
		Voltage:    230, // V
		MinCurrent: 5,   // A
		MaxCurrent: 16,  // A
	}

	loadPoints = append(loadPoints, lp)

	var routes = []route{
		route{
			[]string{"GET"},
			"/modes",
			server.AllChargeModesHandler(),
		},
		route{
			[]string{"GET"},
			"/mode",
			server.CurrentChargeModeHandler(lp),
		},
		route{
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
		updateLoadPoints()
		for range time.Tick(5 * time.Second) {
			core.Logger.Printf("---")
			updateLoadPoints()
		}
	}()

	log.Fatal(srv.ListenAndServe())
}
