package capture

import (
	"log"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

func StartCapture(handle *pcap.Handle, out chan<- []byte, stop <-chan struct{}) {

	defer close(out)
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

	for {
		select {
		case packet := <-packetSource.Packets():
			out <- packet.Data()
		case <-stop:
			log.Println("Stopping capture...")
			return
		}
	}
}
