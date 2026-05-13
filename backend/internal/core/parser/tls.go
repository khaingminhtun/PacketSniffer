package parser

import (
	"packet-sniffer/internal/models"

	"github.com/google/gopacket"
)

type StreamParser struct {
	buffer  map[string][]byte
	sniSeen map[string]bool
	handler func(models.StreamEvent)
}

func NewStreamParser() *StreamParser {
	return &StreamParser{
		buffer:  make(map[string][]byte),
		sniSeen: make(map[string]bool),
	}
}

// allow main to receive events
func (p *StreamParser) SetHandler(h func(models.StreamEvent)) {
	p.handler = h
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

	event := models.StreamEvent{
		Key:     key,
		DataLen: len(stream),
		// Bytes:   stream,
	}

	event.SrcIP = net.Src().String()
	event.DstIP = net.Dst().String()

	event.SrcPort = transport.Src().String()
	event.DstPort = transport.Dst().String()

	if !p.sniSeen[key] {

		if sni, ok := extractSNI(stream); ok {
			event.IsTLS = true
			event.SNI = sni
			p.sniSeen[key] = true
		}
	}

	// // TLS detection
	// if len(stream) >= 2 && stream[0] == 0x16 && stream[1] == 0x03 {
	// 	if !p.tlsSeen[key] {
	// 		event.IsTLS = true
	// 		p.tlsSeen[key] = true
	// 	}
	// }

	// send event to main
	if p.handler != nil {
		p.handler(event)
	}

	// memory safety
	if len(stream) > 50*1024 {
		delete(p.buffer, key)
		delete(p.sniSeen, key)

	}
}

func extractSNI(data []byte) (string, bool) {
	if len(data) < 5 || data[0] != 0x16 {
		return "", false
	}

	recordLen := int(data[3])<<8 | int(data[4])
	if len(data) < 5+recordLen {
		return "", false
	}

	payload := data[5 : 5+recordLen]
	if len(payload) < 4 || payload[0] != 0x01 {
		return "", false
	}

	hsLen := int(payload[1])<<16 | int(payload[2])<<8 | int(payload[3])
	if len(payload) < 4+hsLen {
		return "", false
	}

	hs := payload[4 : 4+hsLen]
	if len(hs) < 34 {
		return "", false
	}

	offset := 34
	if offset+1 > len(hs) {
		return "", false
	}

	sessionIDLen := int(hs[offset])
	offset++
	if offset+sessionIDLen > len(hs) {
		return "", false
	}
	offset += sessionIDLen

	if offset+2 > len(hs) {
		return "", false
	}

	cipherSuitesLen := int(hs[offset])<<8 | int(hs[offset+1])
	offset += 2
	if offset+cipherSuitesLen > len(hs) {
		return "", false
	}
	offset += cipherSuitesLen

	if offset+1 > len(hs) {
		return "", false
	}

	compressionMethodsLen := int(hs[offset])
	offset++
	if offset+compressionMethodsLen > len(hs) {
		return "", false
	}
	offset += compressionMethodsLen

	if offset+2 > len(hs) {
		return "", false
	}

	extensionsLen := int(hs[offset])<<8 | int(hs[offset+1])
	offset += 2
	if offset+extensionsLen > len(hs) {
		return "", false
	}

	endExtensions := offset + extensionsLen
	for offset+4 <= endExtensions {
		extType := int(hs[offset])<<8 | int(hs[offset+1])
		extLen := int(hs[offset+2])<<8 | int(hs[offset+3])
		offset += 4
		if offset+extLen > endExtensions {
			return "", false
		}

		if extType == 0 {
			if extLen < 2 || offset+2 > len(hs) {
				return "", false
			}

			serverNameListLen := int(hs[offset])<<8 | int(hs[offset+1])
			offset += 2
			if offset+serverNameListLen > endExtensions {
				return "", false
			}

			endServerNameList := offset + serverNameListLen
			for offset+3 <= endServerNameList {
				nameType := hs[offset]
				nameLen := int(hs[offset+1])<<8 | int(hs[offset+2])
				offset += 3
				if offset+nameLen > endServerNameList {
					return "", false
				}

				if nameType == 0 {
					return string(hs[offset : offset+nameLen]), true
				}
				offset += nameLen
			}

			return "", false
		}

		offset += extLen
	}

	return "", false
}
