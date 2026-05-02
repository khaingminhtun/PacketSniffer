package utils

import "github.com/google/gopacket/layers"

func FormatTCPFlags(tcp *layers.TCP) string {
	flags := ""

	if tcp.SYN {
		flags += "SYN "
	}
	if tcp.ACK {
		flags += "ACK "
	}
	if tcp.FIN {
		flags += "FIN "
	}
	if tcp.RST {
		flags += "RST "
	}
	if tcp.PSH {
		flags += "PSH "
	}
	if tcp.URG {
		flags += "URG "
	}

	if flags == "" {
		return "NONE"
	}

	return flags[:len(flags)-1] // remove trailing space
}
