package bus

type EventType string

const (
	EventPacket EventType = "packet"
	EventStream EventType = "stream"
)
