package handlers

import (
	"encoding/json"
	"net/http"
	"packet-sniffer/internal/service"
)

type InterfaceHandler struct {
	service *service.InterfaceService
}

func NewInterfaceHandler(service *service.InterfaceService) *InterfaceHandler {
	return &InterfaceHandler{
		service: service,
	}
}

func (h *InterfaceHandler) GetInterfaces(w http.ResponseWriter, r *http.Request) {

	interfaces, err := h.service.GetAllInterfaces()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(interfaces)
}
