package filter

import "packet-sniffer/pkg/models"

type BasicFilter struct {
	AllowedProtocols map[string]bool
	BlockedPorts     map[uint16]bool
}

func (f *BasicFilter) Allow(pkt models.Packet) bool {

	// Block unwanted ports
	if f.BlockedPorts[pkt.SrcPort] || f.BlockedPorts[pkt.DstPort] {
		return false
	}

	// Allow only selected protocols (if defined)
	if len(f.AllowedProtocols) > 0 {
		if !f.AllowedProtocols[pkt.AppProtocol] {
			return false
		}
	}

	return true
}
