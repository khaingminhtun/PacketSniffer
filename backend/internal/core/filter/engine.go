package filter

import "packet-sniffer/internal/models"

func StartFilter(
	in <-chan models.Packet,
	out chan<- models.Packet,
	m *Manager,
) {
	for pkt := range in {
		if m.Allow(pkt) {
			out <- pkt
		}
	}
}
