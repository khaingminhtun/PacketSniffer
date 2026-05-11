package netfaces

import (
	"packet-sniffer/internal/models"
	"strings"

	"github.com/google/gopacket/pcap"
)

func GetInterfaces() ([]models.NetworkInterface, error) {

	devices, err := pcap.FindAllDevs()
	if err != nil {
		return nil, err
	}

	var result []models.NetworkInterface

	for _, dev := range devices {

		desc := strings.ToLower(dev.Description)

		interfaceType := "unknown"

		switch {
		case strings.Contains(desc, "wi-fi"):
			interfaceType = "wifi"

		case strings.Contains(desc, "wireless"):
			interfaceType = "wifi"

		case strings.Contains(desc, "ethernet"):
			interfaceType = "ethernet"

		case strings.Contains(desc, "loopback"):
			interfaceType = "loopback"

		case strings.Contains(desc, "virtual"):
			interfaceType = "virtual"

		case strings.Contains(desc, "vpn"):
			interfaceType = "vpn"

		}

		var ips []string

		for _, addr := range dev.Addresses {
			ips = append(ips, addr.IP.String())
		}

		result = append(result, models.NetworkInterface{
			ID:          dev.Name,
			Description: dev.Description, Name: dev.Description,

			Type:       interfaceType,
			IPs:        ips,
			IsLoopback: strings.Contains(desc, "loopback"),
			IsVirtual:  strings.Contains(desc, "virtual"),
			IsUp:       true,
		})
	}

	return result, nil
}
