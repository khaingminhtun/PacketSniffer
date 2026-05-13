package models

type NetworkInterface struct {
	ID          int      `json:"id"`
	PcapID      string   `json:"pcap_id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Type        string   `json:"type"`
	IPs         []string `json:"ips"`
	IsLoopback  bool     `json:"is_loopback"`
	IsVirtual   bool     `json:"is_virtual"`
	IsUp        bool     `json:"is_up"`
}
