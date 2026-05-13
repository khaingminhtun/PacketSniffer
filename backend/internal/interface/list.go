package iface

import (
	"fmt"

	"github.com/google/gopacket/pcap"
)

func ListInterfaces() ([]pcap.Interface, error) {
	devices, err := pcap.FindAllDevs()
	if err != nil {
		return nil, err
	}

	for i, dev := range devices {
		fmt.Printf("[%d] Name: %s\n", i, dev.Name)
		fmt.Printf("    Description: %s\n", dev.Description)
	}

	return devices, nil
}
