package main

import (
	"fmt"
	"log"

	"packet-sniffer/internal/capture"
	"packet-sniffer/internal/filter"
	iface "packet-sniffer/internal/interface"
	"packet-sniffer/internal/parser"
	"packet-sniffer/internal/reassembly"
	"packet-sniffer/pkg/models"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func main() {

	devices, err := iface.ListInterfaces()
	if err != nil {
		log.Fatal(err)
	}

	device, err := iface.SelectInterface(devices)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Using interface:", device)

	handle, err := capture.OpenDevice(device)
	if err != nil {
		log.Fatal(err)
	}
	defer handle.Close()

	// channels
	rawChan := make(chan []byte, 1000)
	parsedChan := make(chan models.Packet, 1000)
	filteredChan := make(chan models.Packet, 1000)
	stop := make(chan struct{})

	// stream parser (TLS)
	streamParser := parser.NewStreamParser()

	// reassembly
	streamFactory := reassembly.NewStreamFactory(streamParser)
	assembler := reassembly.New(streamFactory)

	myFilter := &filter.BasicFilter{
		AllowedProtocols: map[string]bool{
			"TCP": true,
			"UDP": true,
			"DNS": true,
		},
	}

	// capture
	go capture.StartCapture(handle, rawChan, stop)

	// packet pipeline
	go parser.StartParser(rawChan, parsedChan)
	go filter.StartFilter(parsedChan, filteredChan, myFilter)

	// stream pipeline (IMPORTANT)
	go func() {
		for raw := range rawChan {

			packet := gopacket.NewPacket(raw, layers.LayerTypeEthernet, gopacket.Default)

			assembler.Process(packet)
		}
	}()
	// UI
	for pkt := range filteredChan {

		if pkt.IsDNS {
			fmt.Printf("[DNS] %s → %s | %s\n",
				pkt.SrcIP, pkt.DstIP, pkt.DNSQuery)
			continue
		}

		if pkt.IsTLS {
			fmt.Printf("[TLS] %s → %s | SNI: %s\n",
				pkt.SrcIP, pkt.DstIP, pkt.TLSSNI)
			continue
		}

		fmt.Printf("[%s] %s:%d → %s:%d | %d bytes\n",
			pkt.AppProtocol,
			pkt.SrcIP, pkt.SrcPort,
			pkt.DstIP, pkt.DstPort,
			pkt.Length,
		)
	}
}
