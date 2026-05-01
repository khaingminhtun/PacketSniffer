package main

import (
	"fmt"
	"log"
	"packet-sniffer/internal/capture"
	iface "packet-sniffer/internal/interface"
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

	packetChan := make(chan []byte, 1000)
	stop := make(chan struct{})

	go capture.StartCapture(handle, packetChan, stop)

	for packet := range packetChan {
		fmt.Println("packet received:", len(packet))
	}
}
