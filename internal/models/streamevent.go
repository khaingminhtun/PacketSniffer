package models

type StreamEvent struct {
	Key     string
	SrcIP   string
	DstIP   string
	DataLen int

	IsTLS  bool
	IsHTTP bool

	Bytes []byte
	SNI   string
}
