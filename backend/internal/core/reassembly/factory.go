package reassembly

import (
	"github.com/google/gopacket"
	"github.com/google/gopacket/tcpassembly"
)

type StreamHandler interface {
	HandleStream(net, transport gopacket.Flow, data []byte)
}

type StreamFactory struct {
	handler StreamHandler
}

func NewStreamFactory(h StreamHandler) *StreamFactory {
	return &StreamFactory{handler: h}
}

func (f *StreamFactory) New(net, transport gopacket.Flow) tcpassembly.Stream {
	s := NewTCPStream(net, transport, f.handler)
	return &s.reader
}
