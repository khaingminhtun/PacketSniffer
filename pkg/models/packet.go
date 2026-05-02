package models

import "time"

type Packet struct {
	Timestamp time.Time

	SrcIP string
	DstIP string

	SrcPort uint16
	DstPort uint16

	Protocol    string // TCP , UDp
	AppProtocol string // HTTP, DNS, TLS

	TCPFlags string // for TCP packets, store flags like SYN, ACK, etc.

	Length int
}
