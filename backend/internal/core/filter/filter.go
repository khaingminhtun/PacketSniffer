package filter

import "packet-sniffer/internal/models"

type Filter interface {
	Allow(pkt models.Packet) bool
}
