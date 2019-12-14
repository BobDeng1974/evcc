package main

import (
	"log"
	"net/http"
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

type meter struct{}

func (m *meter) CurrentPower() (float64, error) {
	return 500, nil
}

type charger struct{}

func (c *charger) Enable(charge bool) error {
	return nil
}

func (c *charger) Status() (api.ChargeStatus, error) {
	return api.StatusC, nil
}

func (c *charger) Enabled() (bool, error) {
	return true, nil
}

func (c *charger) MaxPower(max float64) error {
	return nil
}

func control(c api.Charger) {
	if c, ok := c.(api.ChargeController); ok {
		log.Println("Maxpower")
		c.MaxPower(1)
	} else {
		log.Println("no Maxpower")
	}
}

type Route struct {
	Methods     []string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

var loadPoints []api.LoadPoint

func main() {
	m := &meter{}
	c := &charger{}
	// c.MaxPower(5)
	c.Enable(true)

	lp := &core.LoadPoint{
		Name:       "lp1",
		Mode:       api.ModeNow,
		GridMeter:  m,
		Charger:    c,
		Phases:     2,
		MinCurrent: 6,  // A
		MaxCurrent: 16, // A
		Debug:      core.LogEnabled(),
	}
	loadPoints = append(loadPoints, lp)

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
			for i, lp := range loadPoints {
				log.Printf("lp %d: update", i+1)
				go lp.Update()
			}
		}
	}()

	log.Fatal(srv.ListenAndServe())
}
