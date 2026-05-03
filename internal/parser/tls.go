package parser

import (
	"encoding/binary"
)

// ExtractTLSSNI tries to parse SNI from TLS ClientHello
func ExtractTLSSNI(payload []byte) (string, bool) {

	// Minimum TLS record size
	if len(payload) < 5 {
		return "", false
	}

	// TLS Handshake (0x16)
	if payload[0] != 0x16 {
		return "", false
	}

	// Skip record header (5 bytes)
	data := payload[5:]

	// Must be handshake
	if len(data) < 42 {
		return "", false
	}

	// Handshake type = ClientHello (0x01)
	if data[0] != 0x01 {
		return "", false
	}

	// Skip:
	// Handshake header (4)
	// Version (2)
	// Random (32)
	offset := 4 + 2 + 32

	if len(data) < offset {
		return "", false
	}

	// Session ID
	sessionIDLen := int(data[offset])
	offset += 1 + sessionIDLen

	if len(data) < offset+2 {
		return "", false
	}

	// Cipher suites
	csLen := int(binary.BigEndian.Uint16(data[offset:]))
	offset += 2 + csLen

	if len(data) < offset+1 {
		return "", false
	}

	// Compression methods
	compLen := int(data[offset])
	offset += 1 + compLen

	if len(data) < offset+2 {
		return "", false
	}

	// Extensions
	extLen := int(binary.BigEndian.Uint16(data[offset:]))
	offset += 2

	end := offset + extLen
	if len(data) < end {
		return "", false
	}

	// Parse extensions
	for offset+4 <= end {
		extType := binary.BigEndian.Uint16(data[offset:])
		extSize := int(binary.BigEndian.Uint16(data[offset+2:]))
		offset += 4

		if extType == 0x0000 { // SNI
			serverNameData := data[offset : offset+extSize]

			// skip list length (2) + name type (1) + name length (2)
			if len(serverNameData) < 5 {
				return "", false
			}

			nameLen := int(binary.BigEndian.Uint16(serverNameData[3:]))

			if len(serverNameData) < 5+nameLen {
				return "", false
			}

			return string(serverNameData[5 : 5+nameLen]), true
		}

		offset += extSize
	}

	return "", false
}
