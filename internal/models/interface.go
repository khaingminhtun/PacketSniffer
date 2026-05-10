package models

type NetworkInterface struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Type        string   `json:"type"`
	IPs         []string `json:"ips"`
	IsLoopback  bool     `json:"is_loopback"`
	IsVirtual   bool     `json:"is_virtual"`
	IsUp        bool     `json:"is_up"`
}
