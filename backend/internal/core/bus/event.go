package bus

type Event struct {
	Type EventType `json:"type"`
	Data any       `json:"data"`
}
