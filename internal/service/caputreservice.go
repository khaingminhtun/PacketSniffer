package service

import (
	"packet-sniffer/internal/core/capture"
	"packet-sniffer/internal/core/filter"
	"packet-sniffer/internal/core/parser"
	"packet-sniffer/internal/core/reassembly"
	"packet-sniffer/internal/models"
	"sync"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

type CaptureService struct {
	mu sync.Mutex

	handle  *pcap.Handle
	stop    chan struct{}
	running bool

	activeInterface string

	// RAW INPUT
	rawChan chan []byte

	// PACKET PIPELINE
	parseChan    chan []byte
	parsedChan   chan models.Packet
	filteredChan chan models.Packet

	// STREAM PIPELINE
	assemblerChan chan []byte
	streamChan    chan models.StreamEvent

	// ENGINE
	filterManager *filter.Manager
	parser        *parser.StreamParser
	assembler     *reassembly.Assembler

	// OUTPUT (optional websocket hook)
	packetHandler func(models.Packet)
	streamHandler func(models.StreamEvent)
}

func NewCaptureService() *CaptureService {

	p := parser.NewStreamParser()

	return &CaptureService{
		rawChan: make(chan []byte, 1000),

		parseChan:    make(chan []byte, 1000),
		parsedChan:   make(chan models.Packet, 1000),
		filteredChan: make(chan models.Packet, 1000),

		assemblerChan: make(chan []byte, 1000),
		streamChan:    make(chan models.StreamEvent, 1000),

		filterManager: filter.NewManager(),
		parser:        p,
	}
}

func (s *CaptureService) StartCapture(iface string) error {

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		s.stopInternal()
	}

	handle, err := capture.OpenDevice(iface)
	if err != nil {
		return err
	}
	s.handle = handle
	s.activeInterface = iface
	s.stop = make(chan struct{})
	s.running = true

	// STREAM PARSER OUTPUT
	s.parser.SetHandler(func(ev models.StreamEvent) {
		s.streamChan <- ev
	})

	// =========================
	// 1. RAW CAPTURE
	// =========================
	go capture.StartCapture(handle, s.rawChan, s.stop)

	// =========================
	// 2. FANOUT RAW → TWO PATHS
	// =========================
	go func() {
		for {
			select {
			case <-s.stop:
				return

			case raw := <-s.rawChan:

				// packet pipeline
				s.parseChan <- raw

				// stream pipeline
				s.assemblerChan <- raw
			}
		}
	}()

	// =========================
	// 3. PACKET PIPELINE
	// =========================
	go parser.StartParser(s.parseChan, s.parsedChan)

	go filter.StartFilter(
		s.parsedChan,
		s.filteredChan,
		s.filterManager,
	)

	// =========================
	// 4. STREAM PIPELINE
	// =========================
	go func() {
		for raw := range s.assemblerChan {

			packet := gopacket.NewPacket(
				raw,
				layers.LayerTypeEthernet,
				gopacket.Default,
			)

			s.assembler.Process(packet)
		}
	}()

	// =========================
	// 5. OUTPUT LAYER (PACKETS)
	// =========================
	go func() {
		for pkt := range s.filteredChan {
			if s.packetHandler != nil {
				s.packetHandler(pkt)
			}
		}
	}()

	// =========================
	// 6. OUTPUT LAYER (STREAMS)
	// =========================
	go func() {
		for ev := range s.streamChan {
			if s.streamHandler != nil {
				s.streamHandler(ev)
			}
		}
	}()

	return nil

}

func (s *CaptureService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.stopInternal()
}

func (s *CaptureService) stopInternal() {
	if !s.running {
		return
	}

	close(s.stop)

	if s.handle != nil {
		s.handle.Close()
	}

	s.running = false
}

func (s *CaptureService) UpdateFilter(rule models.PacketFilterRule) {
	s.filterManager.Update(rule)
}
