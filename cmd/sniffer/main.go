package main

import (
	"fmt"
	"log"
	"packet-sniffer/internal/capture"
	"packet-sniffer/internal/filter"
	iface "packet-sniffer/internal/interface"
	"packet-sniffer/internal/parser"
	"packet-sniffer/pkg/models"
)

func main() {

	// 1. List available interfaces
	devices, err := iface.ListInterfaces()
	if err != nil {
		log.Fatal(err)
	}

	device, err := iface.SelectInterface(devices)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Using interface:", device)

	//2. capture
	handle, err := capture.OpenDevice(device)
	if err != nil {
		log.Fatal(err)
	}
	defer handle.Close()

	//3. channels
	rawChan := make(chan []byte, 1000)
	parsedChan := make(chan models.Packet, 1000)
	filteredChan := make(chan models.Packet, 1000)
	stop := make(chan struct{})

	// Create filter
	myFilter := &filter.BasicFilter{
		AllowedProtocols: map[string]bool{
			"HTTPS": true,
			"DNS":   true,
		},
		BlockedPorts: map[uint16]bool{
			5353: true, // block mDNS noise
		},
	}

	//4 pipelines
	go capture.StartCapture(handle, rawChan, stop)
	go parser.StartParser(rawChan, parsedChan)
	go filter.StartFilter(parsedChan, filteredChan, myFilter)

	for pkt := range filteredChan {

		// ---------------------------
		// DNS
		// ---------------------------
		if pkt.IsDNS {
			fmt.Printf("[DNS] %s → %s | Query: %s\n",
				pkt.SrcIP,
				pkt.DstIP,
				pkt.DNSQuery,
			)
			continue
		}

		// ---------------------------
		// TLS (SNI)
		// ---------------------------
		if pkt.IsTLS {
			fmt.Printf("[TLS] %s → %s | SNI: %s\n",
				pkt.SrcIP,
				pkt.DstIP,
				pkt.TLSSNI,
			)
			continue
		}

		// ---------------------------
		// Normal traffic
		// ---------------------------
		fmt.Printf("[%s] %s:%d → %s:%d | %s | %d bytes %s\n",
			pkt.AppProtocol,
			pkt.SrcIP,
			pkt.SrcPort,
			pkt.DstIP,
			pkt.DstPort,
			pkt.Protocol,
			pkt.Length,
			pkt.TCPFlags,
		)
	}
}
