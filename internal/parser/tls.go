package parser

import (
	"fmt"

	"github.com/google/gopacket"
)

type StreamParser struct {
	buffer  map[string][]byte
	tlsSeen map[string]bool
}

func NewStreamParser() *StreamParser {
	return &StreamParser{
		buffer:  make(map[string][]byte),
		tlsSeen: make(map[string]bool),
	}
}

// build unique flow key
func flowKey(net, transport gopacket.Flow) string {
	return net.String() + "->" + transport.String()
}

func (p *StreamParser) HandleStream(net, transport gopacket.Flow, data []byte) {

	key := flowKey(net, transport)

	// append stream data
	p.buffer[key] = append(p.buffer[key], data...)

	stream := p.buffer[key]

	fmt.Printf("STREAM %s → %d bytes total\n", key, len(stream))

	// not enough data yet
	if len(stream) < 5 {
		return
	}

	// TLS detection (safe check)
	// TLS record starts with:
	// 0x16 = handshake
	// 0x03 = version
	if stream[0] == 0x16 && stream[1] == 0x03 {
		if !p.tlsSeen[key] {
			fmt.Println("TLS handshake detected (stream-level)")
			p.tlsSeen[key] = true
		}
	}

	// OPTIONAL: prevent unlimited memory growth
	if len(stream) > 50*1024 {
		delete(p.buffer, key)
	}
}
