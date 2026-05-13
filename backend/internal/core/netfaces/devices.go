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

	result := make([]models.NetworkInterface, 0, len(devices))

	idCounter := 1

	for _, dev := range devices {

		desc := strings.ToLower(dev.Description)
		name := dev.Name

		if desc == "" {
			desc = strings.ToLower(name)
		}

		interfaceType := detectType(desc)

		ips := extractIPs(dev.Addresses)

		result = append(result, models.NetworkInterface{
			ID:          idCounter,       // ✅ numeric frontend ID
			PcapID:      dev.Name,        // keep real capture id internally
			Name:        dev.Description, // readable name
			Description: dev.Description,

			Type: interfaceType,
			IPs:  ips,

			IsLoopback: isLoopback(desc, ips),
			IsVirtual:  isVirtual(desc),
			IsUp:       len(ips) > 0,
		})

		idCounter++
	}

	return result, nil
}

func detectType(desc string) string {

	switch {
	case strings.Contains(desc, "wi-fi"), strings.Contains(desc, "wireless"):
		return "wifi"

	case strings.Contains(desc, "ethernet"):
		return "ethernet"

	case strings.Contains(desc, "loopback"):
		return "loopback"

	case strings.Contains(desc, "vpn"), strings.Contains(desc, "tap"):
		return "vpn"

	case strings.Contains(desc, "virtual"):
		return "virtual"

	default:
		return "unknown"
	}
}

func extractIPs(addrs []pcap.InterfaceAddress) []string {

	var ips []string
	var fallback []string

	for _, addr := range addrs {

		if addr.IP == nil {
			continue
		}

		ip := addr.IP.String()

		// IPv4 valid
		if strings.Contains(ip, ".") {

			// skip APIPA unless needed
			if strings.HasPrefix(ip, "169.254") {
				fallback = append(fallback, ip)
				continue
			}

			ips = append(ips, ip)
			continue
		}

		// IPv6
		if strings.HasPrefix(ip, "fe80") {
			continue // skip noisy link-local
		}

		if ip != "::1" {
			ips = append(ips, ip)
		}
	}

	if len(ips) == 0 {
		return fallback
	}

	return ips
}

func isLoopback(desc string, ips []string) bool {
	if strings.Contains(desc, "loopback") {
		return true
	}

	for _, ip := range ips {
		if ip == "127.0.0.1" || ip == "::1" {
			return true
		}
	}

	return false
}

func isVirtual(desc string) bool {

	keywords := []string{
		"virtual",
		"vmware",
		"virtualbox",
		"hyper-v",
		"tap",
		"tunnel",
	}

	for _, k := range keywords {
		if strings.Contains(desc, k) {
			return true
		}
	}

	return false
}
