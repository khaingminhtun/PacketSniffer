package main

import "fmt"

func new() {
	fmt.Println("Hello")
}

// import (
// 	"fmt"
// 	"log"

// 	"packet-sniffer/internal/capture"
// 	"packet-sniffer/internal/filter"
// 	iface "packet-sniffer/internal/interface"
// 	"packet-sniffer/internal/parser"
// 	"packet-sniffer/internal/reassembly"
// 	"packet-sniffer/pkg/models"

// 	"github.com/google/gopacket"
// 	"github.com/google/gopacket/layers"
// )

// func cli() {

// 	// =========================
// 	// INTERFACE SELECTION
// 	// =========================

// 	devices, err := iface.ListInterfaces()
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	device, err := iface.SelectInterface(devices)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	fmt.Println("Using interface:", device)

// 	// =========================
// 	// OPEN DEVICE
// 	// =========================

// 	handle, err := capture.OpenDevice(device)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer handle.Close()

// 	// =========================
// 	// CHANNELS
// 	// =========================

// 	// raw capture channel
// 	rawChan := make(chan []byte, 1000)

// 	// fan-out channels
// 	parseChan := make(chan []byte, 1000)
// 	assemblerChan := make(chan []byte, 1000)

// 	// parsed packet channels
// 	parsedChan := make(chan models.Packet, 1000)
// 	filteredChan := make(chan models.Packet, 1000)

// 	// stream events
// 	streamChan := make(chan models.StreamEvent, 1000)

// 	stop := make(chan struct{})

// 	// =========================
// 	// STREAM PARSER
// 	// =========================

// 	streamParser := parser.NewStreamParser()

// 	streamParser.SetHandler(func(ev models.StreamEvent) {
// 		streamChan <- ev
// 	})

// 	// =========================
// 	// TCP REASSEMBLY
// 	// =========================

// 	streamFactory := reassembly.NewStreamFactory(streamParser)

// 	assembler := reassembly.New(streamFactory)

// 	// =========================
// 	// FILTERS
// 	// =========================

// 	myFilter := &filter.BasicFilter{
// 		AllowedProtocols: map[string]bool{
// 			"TCP": true,
// 			"UDP": true,
// 			"DNS": true,
// 		},
// 	}

// 	// =========================
// 	// START PACKET CAPTURE
// 	// =========================

// 	go capture.StartCapture(
// 		handle,
// 		rawChan,
// 		stop,
// 	)

// 	// =========================
// 	// FAN-OUT DISPATCHER
// 	// =========================

// 	go func() {

// 		for raw := range rawChan {

// 			// send to parser pipeline
// 			parseChan <- raw

// 			// send to TCP reassembly pipeline
// 			assemblerChan <- raw
// 		}
// 	}()

// 	// =========================
// 	// PARSER PIPELINE
// 	// =========================

// 	go parser.StartParser(
// 		parseChan,
// 		parsedChan,
// 	)

// 	go filter.StartFilter(
// 		parsedChan,
// 		filteredChan,
// 		myFilter,
// 	)

// 	// =========================
// 	// TCP STREAM REASSEMBLY
// 	// =========================

// 	go func() {

// 		for raw := range assemblerChan {

// 			packet := gopacket.NewPacket(
// 				raw,
// 				layers.LayerTypeEthernet,
// 				gopacket.Default,
// 			)

// 			assembler.Process(packet)
// 		}
// 	}()

// 	// =========================
// 	// STREAM OUTPUT
// 	// =========================

// 	go func() {

// 		for ev := range streamChan {

// 			if ev.IsTLS {

// 				fmt.Printf(
// 					"[TLS] %s | SNI: %s | %d bytes\n",
// 					ev.Key,
// 					ev.SNI,
// 					ev.DataLen,
// 				)

// 				continue
// 			}

// 			fmt.Printf(
// 				"[STREAM] %s | %d bytes\n",
// 				ev.Key,
// 				ev.DataLen,
// 			)
// 		}
// 	}()

// 	// =========================
// 	// PACKET OUTPUT
// 	// =========================

// 	for pkt := range filteredChan {

// 		// DNS
// 		if pkt.IsDNS {

// 			fmt.Printf(
// 				"[DNS] %s → %s | %s\n",
// 				pkt.SrcIP,
// 				pkt.DstIP,
// 				pkt.DNSQuery,
// 			)

// 			continue
// 		}

// 		// TLS
// 		if pkt.IsTLS {

// 			fmt.Printf(
// 				"[TLS] %s → %s | SNI: %s\n",
// 				pkt.SrcIP,
// 				pkt.DstIP,
// 				pkt.TLSSNI,
// 			)

// 			continue
// 		}

// 		// Generic packets
// 		fmt.Printf(
// 			"[%s] %s:%d → %s:%d | %d bytes\n",
// 			pkt.AppProtocol,
// 			pkt.SrcIP,
// 			pkt.SrcPort,
// 			pkt.DstIP,
// 			pkt.DstPort,
// 			pkt.Length,
// 		)
// 	}
// }
