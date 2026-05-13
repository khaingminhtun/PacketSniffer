package service

import (
	"context"
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

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	handle *pcap.Handle

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

	running bool
}

func NewCaptureService() *CaptureService {

	ctx, cancel := context.WithCancel(context.Background())

	p := parser.NewStreamParser()

	s := &CaptureService{
		ctx:    ctx,
		cancel: cancel,
		wg:     sync.WaitGroup{},

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
	// s.stop = make(chan struct{})
	s.running = true

	// =========================
	// 1. RAW CAPTURE
	// =========================
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		capture.StartCapture(handle, s.rawChan, s.ctx.Done())
	}()

	// =========================
	// 2. SAFE FANOUT
	// =========================
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		for {
			select {
			case <-s.ctx.Done():
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
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		parser.StartParser(s.parseChan, s.parsedChan)
	}()

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		filter.StartFilter(
			s.parsedChan,
			s.filteredChan,
			s.filterManager,
		)
	}()

	// =========================
	// 4. STREAM PIPELINE
	// =========================
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		for {
			select {
			case <-s.ctx.Done():
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
	// 5. EVENTBUS (PACKETS)
	// =========================
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		for {
			select {
			case <-s.ctx.Done():
				return

			case pkt := <-s.filteredChan:
				s.bus.Publish(bus.Event{
					Type: bus.EventPacket,
					Data: pkt,
				})
			}
		}
	}()

	return nil
}

func (s *CaptureService) GetBus() *bus.EventBus {
	return s.bus
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

	// 1. signal all goroutines
	s.cancel()

	// 2. close pcap capture
	if s.handle != nil {
		s.handle.Close()
		s.handle = nil
	}

	// 3. wait all goroutines finish
	s.wg.Wait()

	// 4. safe state reset
	s.running = false
}

func (s *CaptureService) UpdateFilter(rule models.PacketFilterRule) {
	s.filterManager.Update(rule)
}
