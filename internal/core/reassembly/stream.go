package reassembly

import (
	"io"

	"github.com/google/gopacket"
	"github.com/google/gopacket/tcpassembly/tcpreader"
)

type TCPStream struct {
	net, transport gopacket.Flow
	reader         tcpreader.ReaderStream
	handler        StreamHandler
	srcIP          string
	dstIP          string
	srcPort        string
	dstPort        string
	buffer         []byte
}

func NewTCPStream(net, transport gopacket.Flow, h StreamHandler) *TCPStream {
	s := &TCPStream{
		net:       net,
		transport: transport,
		reader:    tcpreader.NewReaderStream(),
		handler:   h,
	}

	s.srcIP = net.Src().String()
	s.dstIP = net.Dst().String()
	s.srcPort = transport.Src().String()
	s.dstPort = transport.Dst().String()

	go s.run()
	return s
}

func (s *TCPStream) Reader() *tcpreader.ReaderStream {
	return &s.reader
}

func (s *TCPStream) run() {
	buf := make([]byte, 4096)

	for {
		n, err := s.reader.Read(buf)
		if err != nil {
			if err != io.EOF {
				// optional: log error
			}
			return
		}

		data := make([]byte, n)
		copy(data, buf[:n])

		// optional buffer (useful for TLS later)
		s.buffer = append(s.buffer, data...)

		// 🔥 send to handler (parser layer)
		if s.handler != nil {
			s.handler.HandleStream(s.net, s.transport, data)
		}
	}
}
