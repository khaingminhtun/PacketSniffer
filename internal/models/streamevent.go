package models

type StreamEvent struct {
	Key string

	SrcIP   string
	DstIP   string
	SrcPort string
	DstPort string

	Bytes   []byte
	DataLen int

	IsTLS bool
	SNI   string
}
