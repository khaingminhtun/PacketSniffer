package routes

import (
	"net/http"
	"packet-sniffer/internal/api/handlers"
)

func RegisterRoutes(
	mux *http.ServeMux,
	centralHandler *handlers.CentralHandler,
) {

	mux.HandleFunc(
		"GET /interfaces",
		centralHandler.GetInterfaces,
	)

	mux.HandleFunc(
		"POST /filter",
		centralHandler.UpdateFilter,
	)

	mux.HandleFunc(
		"POST /capture/start/{id}",
		centralHandler.StartCapture,
	)

	mux.HandleFunc(
		"POST /capture/stop",
		centralHandler.StopCapture,
	)
}
