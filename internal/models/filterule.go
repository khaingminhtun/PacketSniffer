package models

type PacketFilterRule struct {
	AllowedProtocols []string `json:"allowed_protocols"`
	BlockedPorts     []uint16 `json:"blocked_ports"`
}
