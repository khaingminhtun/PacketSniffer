package parser

import (
	"fmt"
	"log"

	"packet-sniffer/internal/analyzer"
	"packet-sniffer/internal/utils"
	"packet-sniffer/pkg/models"

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
			tcp, _ := tcpLayer.(*layers.TCP)
			parsed.SrcPort = uint16(tcp.SrcPort)
			parsed.DstPort = uint16(tcp.DstPort)
			parsed.Protocol = "TCP"

			// app protocol from destination port
			parsed.AppProtocol = analyzer.DetectAppProtocol(parsed.SrcPort, parsed.DstPort)

			// NEW: TCP flags
			parsed.TCPFlags = utils.FormatTCPFlags(tcp)
			fmt.Println("TCP payload:", len(tcp.Payload))

			// 🔥 IMPORTANT: only try TLS if payload exists
			if len(tcp.Payload) > 0 {
				fmt.Println("HAS TLS DATA")
				if sni, ok := ExtractTLSSNI(tcp.Payload); ok {
					parsed.IsTLS = true
					parsed.TLSSNI = sni
					parsed.AppProtocol = "TLS"
				}
			}
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
