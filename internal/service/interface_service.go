package service

import (
	"packet-sniffer/internal/core/netfaces"
	"packet-sniffer/internal/models"
)

type InterfaceService struct{}

func NewInterfaceService() *InterfaceService {
	return &InterfaceService{}
}

func (s *InterfaceService) GetAllInterfaces() ([]models.NetworkInterface, error) {
	return netfaces.GetInterfaces()
}
