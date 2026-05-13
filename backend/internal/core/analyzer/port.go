package analyzer

func DetectAppProtocol(srcPort, dstPort uint16) string {
	// First check source port (important for responses)
	if proto := mapPortToProtocol(srcPort); proto != "UNKNOWN" {
		return proto
	}

	// Then check destination port (important for requests)
	return mapPortToProtocol(dstPort)
}

func mapPortToProtocol(port uint16) string {
	switch port {
	case 80:
		return "HTTP"
	case 443:
		return "HTTPS"
	case 22:
		return "SSH"
	case 53:
		return "DNS"
	case 3306:
		return "MySQL"
	case 5432:
		return "PostgreSQL"
	case 5353:
		return "mDNS"
	default:
		return "UNKNOWN"
	}
}
