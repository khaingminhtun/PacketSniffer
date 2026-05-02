package filter

import "packet-sniffer/pkg/models"

func StartFilter(in <-chan models.Packet, out chan<- models.Packet, f Filter) {
	for pkt := range in {
		if f.Allow(pkt) {
			out <- pkt
		}
	}
}
