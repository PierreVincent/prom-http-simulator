package main

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"github.com/PierreVincent/prom-http-simulator"
	"github.com/gorilla/mux"
	"encoding/json"
)

func main() {
	s := prom_http_simulator.NewSimulator(prom_http_simulator.SimulatorOpts{
		Endpoints: []string{
			"/login", "/login", "/login", "/login", "/login", "/login", "/login",
			"/users", "/users", "/users",
			"/users/{id}",
			"/register", "/register",
			"/logout", "/logout", "/logout", "/logout",
		},

		RequestRate:            1000,
		RequestRateUncertainty: 70,

		LatencyMin:         10,
		LatencyP50:         25,
		LatencyP90:         150,
		LatencyP99:         750,
		LatencyMax:         10000,
		LatencyUncertainty: 70,

		ErrorRate: 1,

		SpikeStartChance: 5,
		SpikeEndChance:   30,
	})
	go s.Run()
	startHTTPServer(s)
}

func startHTTPServer(s *prom_http_simulator.Simulator) {

	router := mux.NewRouter()
	router.KeepContext = true
	router.Handle("/spike/{mode}", newSpikeModeHandler(s)).Methods("POST")
	router.Handle("/error_rate", newErrorRateHandler(s)).Methods("PUT")

	router.Handle("/metrics", promhttp.Handler())

	panic(http.ListenAndServe(":8080", router))
}

func newSpikeModeHandler(s *prom_http_simulator.Simulator) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		mode := mux.Vars(r)["mode"]
		s.SetSpikeMode(mode)
	}
}

func newErrorRateHandler(s *prom_http_simulator.Simulator) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		defer r.Body.Close()

		var t struct{ ErrorRate int `json:"error_rate"` }
		err := decoder.Decode(&t)
		if err != nil {
			rw.WriteHeader(400)
		}
		s.SetErrorRate(t.ErrorRate)
	}
}
