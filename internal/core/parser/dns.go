package parser

import (
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func ExtractDNS(packet gopacket.Packet) (string, bool) {
	dnsLayer := packet.Layer(layers.LayerTypeDNS)
	if dnsLayer == nil {
		return "", false
	}

	dns, ok := dnsLayer.(*layers.DNS)
	if !ok {
		return "", false
	}

	// Only care about queries (not responses for now)
	if dns.QR {
		return "", false
	}

	if len(dns.Questions) > 0 {
		return string(dns.Questions[0].Name), true
	}

	return "", false
}
