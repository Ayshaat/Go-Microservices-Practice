package event

type KafkaMessage struct {
	Type      string      `json:"type"`
	Service   string      `json:"service"`
	Timestamp string      `json:"timestamp"`
	Payload   interface{} `json:"payload"`
}

type CartItemAddedPayload struct {
	CartID string `json:"cartId"`
	SKU    string `json:"sku"`
	Count  int    `json:"count"`
	Status string `json:"status"`
}

type CartItemFailedPayload struct {
	CartID string `json:"cartId"`
	SKU    string `json:"sku"`
	Count  int    `json:"count"`
	Status string `json:"status"`
	Reason string `json:"reason"`
}
