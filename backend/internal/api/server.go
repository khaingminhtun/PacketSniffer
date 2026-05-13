package httpserver

import (
	"fmt"
	"net/http"

	"packet-sniffer/internal/api/handlers"
	"packet-sniffer/internal/api/middleware"
	"packet-sniffer/internal/api/routes"
	"packet-sniffer/internal/api/websocket"
	"packet-sniffer/internal/service"
)

func StartServer() error {

	mux := http.NewServeMux()

	// services
	interfaceService := service.NewInterfaceService()
	captureService := service.NewCaptureService()

	// websocket hub
	hub := websocket.NewHub()
	go hub.Run()

	websocket.StartBroadcaster(
		captureService.GetBus(),
		hub,
	)

	interfaceHandler := handlers.NewCentralHandler(
		interfaceService,
		captureService,
	)

	// routes
	routes.RegisterRoutes(
		mux,
		interfaceHandler,
	)

	// websocket
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		websocket.ServeWS(hub, w, r)
	})

	fmt.Println("HTTP server running on :8080")

	// ✅ WRAP WITH MIDDLEWARE HERE (IMPORTANT PART)
	handler := middleware.CORS(mux)

	return http.ListenAndServe(":8080", handler)
}
