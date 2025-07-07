// Package handlers contains all the routes for the API
package handlers

import "rimcs/metafor-blueprint/workerpool"

// Handler contains all the routes as methods.
// This makes it easy to spread api keys and secrets between your routes.
// In case you need to add one of those said common parts, you just need to add them to your struct definition.
type Handler struct {
	JobQueue chan workerpool.Job
}

// NewHandler creates and returns a Handler struct
func NewHandler() *Handler {
	jobQueue := workerpool.StartWorkerPool(4, 10) //TODO: get these as environment params / command line args
	return &Handler{JobQueue: jobQueue}
}
