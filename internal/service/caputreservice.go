package service

import (
	"sync"

	"packet-sniffer/internal/core/bus"
	"packet-sniffer/internal/core/capture"
	"packet-sniffer/internal/core/filter"
	"packet-sniffer/internal/core/parser"
	"packet-sniffer/internal/core/reassembly"
	"packet-sniffer/internal/models"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

type CaptureService struct {
	mu  sync.Mutex
	bus *bus.EventBus

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

	// ENGINE
	filterManager *filter.Manager
	parser        *parser.StreamParser
	assembler     *reassembly.Assembler
}

func NewCaptureService() *CaptureService {

	p := parser.NewStreamParser()

	s := &CaptureService{
		rawChan: make(chan []byte, 1000),

		parseChan:    make(chan []byte, 1000),
		parsedChan:   make(chan models.Packet, 1000),
		filteredChan: make(chan models.Packet, 1000),

		assemblerChan: make(chan []byte, 1000),

		filterManager: filter.NewManager(),
		parser:        p,
		bus:           bus.NewEventBus(),
	}

	// IMPORTANT: StreamParser → EventBus directly
	p.SetHandler(func(ev models.StreamEvent) {
		s.bus.Publish(bus.Event{
			Type: bus.EventStream,
			Data: ev,
		})
	})

	// IMPORTANT: initialize assembler correctly
	s.assembler = reassembly.New(reassembly.NewStreamFactory(p))

	return s
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

	// =========================
	// 1. RAW CAPTURE
	// =========================
	go capture.StartCapture(handle, s.rawChan, s.stop)

	// =========================
	// 2. SAFE FANOUT
	// =========================
	go func() {
		for {
			select {
			case <-s.stop:
				return

			case raw, ok := <-s.rawChan:
				if !ok {
					return
				}

				s.parseChan <- raw
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
		for {
			select {
			case <-s.stop:
				return

			case raw := <-s.assemblerChan:
				packet := gopacket.NewPacket(
					raw,
					layers.LayerTypeEthernet,
					gopacket.Default,
				)

				s.assembler.Process(packet)
			}
		}
	}()

	// =========================
	// 5. EVENTBUS OUTPUT (PACKETS)
	// =========================
	go func() {
		for pkt := range s.filteredChan {
			s.bus.Publish(bus.Event{
				Type: bus.EventPacket,
				Data: pkt,
			})
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
