// Package server contains all the logic regarding the http server, CORS, and route registering.
package server

import (
	"net/http"
	"os"
	"rimcs/metafor-blueprint/handlers"
	"time"

	"github.com/julienschmidt/httprouter"
)

// NewServer creates and return a http.Server struct, sets timeouts, register routes and optionally adds CORS handling.
func NewServer(h *handlers.Handler) *http.Server {

	router := httprouter.New()

	// health check
	router.GET("/info", h.Info())
	router.GET("/workload", h.Workload())

	// Enables CORS. If the functionality is needed just add this section back.
	// be sure download the module before running your server!
	/*	corsConfig := cors.New(cors.Options{
			AllowedOrigins:   []string{"http://localhost:3000"},
			AllowedMethods:   []string{"POST", "GET", "DELETE", "PUT", "PATCH"},
			AllowCredentials: true,
		})
		handlerWithCORS := corsConfig.Handler(router)*/

	//TODO: What to do about these?
	srvr := &http.Server{
		Addr:         ":" + os.Getenv("PORT"),
		Handler:      router, // replace it with handlerWithCORS if cors is needed
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return srvr
}
