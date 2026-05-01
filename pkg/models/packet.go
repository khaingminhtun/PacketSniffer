package models

import "time"

type Packet struct {
	Timestamp time.Time

	SrcIP string
	DstIP string

	Protocol string

	SrcPort uint16
	DstPort uint16

	Length int
}
