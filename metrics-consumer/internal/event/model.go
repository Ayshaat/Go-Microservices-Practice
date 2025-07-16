package event

type KafkaMessage struct {
	Type      string      `json:"type"`
	Service   string      `json:"service"`
	Timestamp string      `json:"timestamp"`
	Payload   interface{} `json:"payload"`
}
