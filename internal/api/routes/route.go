package routes

import (
	"net/http"
	"packet-sniffer/internal/api/handlers"
)

func RegisterRoutes(
	mux *http.ServeMux,
	interfaceHandler *handlers.InterfaceHandler,
) {
	mux.HandleFunc("/interfaces", interfaceHandler.GetInterfaces)
	mux.HandleFunc("/filter", interfaceHandler.UpdateFilter)
}
