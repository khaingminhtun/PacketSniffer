package httpserver

import (
	"fmt"
	"net/http"

	"packet-sniffer/internal/api/handlers"
	"packet-sniffer/internal/api/routes"
	"packet-sniffer/internal/service"
)

func StartServer() error {

	mux := http.NewServeMux()

	interfaceService := service.NewInterfaceService()

	interfaceHandler := handlers.NewInterfaceHandler(interfaceService)

	routes.RegisterRoutes(
		mux,
		interfaceHandler,
	)

	fmt.Println("HTTP server running on :8080")

	return http.ListenAndServe(":8080", mux)
}
