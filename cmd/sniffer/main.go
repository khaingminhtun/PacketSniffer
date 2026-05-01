package main

import (
	"fmt"
	"log"
	"packet-sniffer/internal/capture"
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
	stop := make(chan struct{})
	//4 pipelines
	go capture.StartCapture(handle, rawChan, stop)
	go parser.StartParser(rawChan, parsedChan)

	// 5. Output
	for pkt := range parsedChan {
		fmt.Printf("%s | %s → %s | %d bytes\n",
			pkt.Protocol,
			pkt.SrcIP,
			pkt.DstIP,
			pkt.Length,
		)
	}
}
