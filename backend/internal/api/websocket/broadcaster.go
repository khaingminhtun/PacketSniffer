package websocket

import (
	"encoding/json"
	"packet-sniffer/internal/core/bus"
)

func StartBroadcaster(
	eventBus *bus.EventBus,
	hub *Hub,
) {
	packetCh := eventBus.Subscribe(bus.EventPacket)
	streamCh := eventBus.Subscribe(bus.EventStream)

	// packet events
	go func() {
		for evt := range packetCh {
			data, err := json.Marshal(evt)
			if err != nil {
				continue
			}
			hub.broadcast <- data
		}
	}()

	// stream events
	go func() {
		for evt := range streamCh {
			data, err := json.Marshal(evt)
			if err != nil {
				continue
			}
			hub.broadcast <- data
		}
	}()
}
