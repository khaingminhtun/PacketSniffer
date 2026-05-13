package parser

import (
	"log"

	"packet-sniffer/internal/core/analyzer"
	"packet-sniffer/internal/models"
	"packet-sniffer/internal/utils"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func StartParser(in <-chan []byte, out chan<- models.Packet) {

	defer close(out)
	for data := range in {

		packet := gopacket.NewPacket(data, layers.LayerTypeEthernet, gopacket.Default)

		ipLayer := packet.Layer(layers.LayerTypeIPv4)
		if ipLayer == nil {
			continue
		}

		ip, _ := ipLayer.(*layers.IPv4)

		parsed := models.Packet{
			Timestamp: packet.Metadata().Timestamp,
			SrcIP:     ip.SrcIP.String(),
			DstIP:     ip.DstIP.String(),
			Protocol:  ip.Protocol.String(),
			Length:    len(data),
		}

		// TCP
		if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
			tcp, ok := tcpLayer.(*layers.TCP)
			if !ok {
				continue
			}
			parsed.SrcPort = uint16(tcp.SrcPort)
			parsed.DstPort = uint16(tcp.DstPort)
			parsed.Protocol = "TCP"

			// app protocol from destination port
			parsed.AppProtocol = analyzer.DetectAppProtocol(parsed.SrcPort, parsed.DstPort)

			// NEW: TCP flags
			parsed.TCPFlags = utils.FormatTCPFlags(tcp)
			// if len(tcp.Payload) > 0 {
			// 	// optional debug only
			// 	fmt.Println("TCP payload:", len(tcp.Payload))
			// }

		}

		// UDP
		if udpLayer := packet.Layer(layers.LayerTypeUDP); udpLayer != nil {
			udp, _ := udpLayer.(*layers.UDP)
			parsed.SrcPort = uint16(udp.SrcPort)
			parsed.DstPort = uint16(udp.DstPort)
			parsed.Protocol = "UDP"

			parsed.AppProtocol = analyzer.DetectAppProtocol(parsed.SrcPort, parsed.DstPort)
		}

		// ✅ NEW: DNS detection (AFTER transport layer)
		if dnsName, ok := ExtractDNS(packet); ok {
			parsed.IsDNS = true
			parsed.DNSQuery = dnsName
			parsed.AppProtocol = "DNS"
		}

		out <- parsed
	}

	log.Println("Parser stopped")
}
