package service

import (
	"fmt"
	"sync"

	"packet-sniffer/internal/core/netfaces"
	"packet-sniffer/internal/models"
)

type InterfaceService struct {
	mu   sync.RWMutex
	data map[int]models.NetworkInterface
}

func NewInterfaceService() *InterfaceService {
	s := &InterfaceService{
		data: make(map[int]models.NetworkInterface),
	}

	// preload interfaces at startup
	s.RefreshInterfaces()

	return s
}

func (s *InterfaceService) RefreshInterfaces() ([]models.NetworkInterface, error) {

	ifaces, err := netfaces.GetInterfaces()
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.data = make(map[int]models.NetworkInterface)

	for _, iface := range ifaces {
		s.data[iface.ID] = iface
	}

	return ifaces, nil
}

func (s *InterfaceService) GetAllInterfaces() ([]models.NetworkInterface, error) {

	// optional: refresh each time OR remove this if you want static cache
	return s.RefreshInterfaces()
}

func (s *InterfaceService) GetByID(id int) (models.NetworkInterface, error) {

	s.mu.RLock()
	defer s.mu.RUnlock()

	iface, ok := s.data[id]
	if !ok {
		return models.NetworkInterface{}, fmt.Errorf("interface not found")
	}

	return iface, nil
}
