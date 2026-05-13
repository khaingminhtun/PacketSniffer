package iface

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/google/gopacket/pcap"
)

func SelectInterface(devices []pcap.Interface) (string, error) {
	for i, dev := range devices {
		fmt.Printf("[%d] %s\n", i, dev.Name)
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Select interface index: ")
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	input = strings.TrimSpace(input)

	index, err := strconv.Atoi(input)
	if err != nil {
		return "", fmt.Errorf("invalid number")
	}

	if index < 0 || index >= len(devices) {
		return "", fmt.Errorf("index out of range")
	}

	return devices[index].Name, nil
}
