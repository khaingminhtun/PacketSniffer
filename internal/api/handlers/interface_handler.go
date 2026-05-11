package handlers

import (
	"encoding/json"
	"net/http"
	"packet-sniffer/internal/models"
	"packet-sniffer/internal/service"
)

type InterfaceHandler struct {
	ifservice     *service.InterfaceService
	capfilservice *service.CaptureService
}

func NewInterfaceHandler(ifservice *service.InterfaceService, capfilservice *service.CaptureService) *InterfaceHandler {
	return &InterfaceHandler{
		ifservice:     ifservice,
		capfilservice: capfilservice,
	}
}

func (h *InterfaceHandler) GetInterfaces(w http.ResponseWriter, r *http.Request) {

	interfaces, err := h.ifservice.GetAllInterfaces()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(interfaces)
}

func (h *InterfaceHandler) UpdateFilter(w http.ResponseWriter, r *http.Request) {

	var rule models.PacketFilterRule
	json.NewDecoder(r.Body).Decode(&rule)

	h.capfilservice.UpdateFilter(rule)

	w.WriteHeader(http.StatusOK)
}
