package handlers

import (
	"encoding/json"
	"net/http"
	"packet-sniffer/internal/models"
	"packet-sniffer/internal/service"
)

type CentralHandler struct {
	ifservice     *service.InterfaceService
	capfilservice *service.CaptureService
}

func NewCentralHandler(ifservice *service.InterfaceService, capfilservice *service.CaptureService) *CentralHandler {
	return &CentralHandler{
		ifservice:     ifservice,
		capfilservice: capfilservice,
	}
}

func (h *CentralHandler) GetInterfaces(w http.ResponseWriter, r *http.Request) {

	interfaces, err := h.ifservice.GetAllInterfaces()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(interfaces)
}

func (h *CentralHandler) UpdateFilter(w http.ResponseWriter, r *http.Request) {

	var rule models.PacketFilterRule
	json.NewDecoder(r.Body).Decode(&rule)

	h.capfilservice.UpdateFilter(rule)

	w.WriteHeader(http.StatusOK)
}

func (h *CentralHandler) StartCapture(w http.ResponseWriter, r *http.Request) {
	iface := r.PathValue("id")

	if iface == "" {
		http.Error(w, "missing interface id", http.StatusBadRequest)
		return
	}

	err := h.capfilservice.StartCapture(iface)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("capture started"))
}

func (h *CentralHandler) StopCapture(w http.ResponseWriter, r *http.Request) {

	h.capfilservice.Stop()

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("capture stopped"))
}
