package capture

import (
	"github.com/google/gopacket/pcap"
)

func OpenDevice(device string) (*pcap.Handle, error) {
	snapshotLen := int32(1600)
	promiscuous := true
	timeout := pcap.BlockForever

	handle, err := pcap.OpenLive(device, snapshotLen, promiscuous, timeout)
	if err != nil {
		return nil, err
	}

	// ADD bpf filter here if needed
	err = handle.SetBPFFilter("tcp port 443")
	if err != nil {
		return nil, err
	}

	return handle, nil
}
