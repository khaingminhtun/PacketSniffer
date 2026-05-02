package filter

import "packet-sniffer/pkg/models"

type Filter interface {
	Allow(pkt models.Packet) bool
}
