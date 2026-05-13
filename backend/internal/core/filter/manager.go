package filter

import (
	"packet-sniffer/internal/models"
	"sync"
)

type Manager struct {
	mu   sync.RWMutex
	rule models.PacketFilterRule
}

func NewManager() *Manager {
	return &Manager{
		rule: models.PacketFilterRule{
			AllowedProtocols: []string{},
			BlockedPorts:     []uint16{},
		},
	}
}

func (m *Manager) Update(rule models.PacketFilterRule) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.rule = rule
}

func (m *Manager) Allow(pkt models.Packet) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// blocked ports
	for _, p := range m.rule.BlockedPorts {
		if pkt.SrcPort == p || pkt.DstPort == p {
			return false
		}
	}

	// allowed protocols (if set)
	if len(m.rule.AllowedProtocols) > 0 {
		allowed := false

		for _, proto := range m.rule.AllowedProtocols {
			if proto == pkt.AppProtocol {
				allowed = true
				break
			}
		}

		if !allowed {
			return false
		}
	}

	return true
}
