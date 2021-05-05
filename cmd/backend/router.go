package main

import (
	"context"
	"database/sql"
	"github.com/gtforge/go-skeleton-draft/structure/pkg/person"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gtforge/global_services_common_go/gett-storages"
	"github.com/gtforge/go-healthcheck"
	"github.com/gtforge/go-skeleton-draft/core"
)

func createHealthCheckHandler(pingers []healthcheck.Pinger) http.Handler {
	hc := healthcheck.NewHealthCheck(pingers...)
	return healthcheck.MakeHealthcheckHandler(hc)
}

func healthCheckPingers(db *sql.DB) []healthcheck.Pinger {
	return []healthcheck.Pinger{
		// Example of the built-in db pinger
		healthcheck.MakeDbPinger(db, "main"),

		// Example of a custom pinger
		func(ctx context.Context) (map[string]interface{}, error) {
			values := map[string]interface{}{}
			values["custom"] = map[string]interface{}{
				"active_rides": 5050,
			}
			return values, nil
		},
	}
}

func createRouter() *mux.Router {
	router := mux.NewRouter()

	router.Handle("/alive", createHealthCheckHandler(healthCheckPingers(gettStorages.DB.DB()))).Methods(http.MethodGet)
	router.PathPrefix("/debug/pprof/").Handler(skeleton.BasicAuthMiddleware(http.DefaultServeMux))

	s := router.PathPrefix("/api/v1").Subrouter()

	ridesHandler := person.NewHandler(person.NewPersonService(person.NewRepo(gettStorages.DB)))
	ridesHandler.RegisterRoutes(s)

	return router
}
