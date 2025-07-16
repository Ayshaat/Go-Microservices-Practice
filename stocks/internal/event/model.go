package event

type KafkaMessage struct {
	Type      string      `json:"type"`
	Service   string      `json:"service"`
	Timestamp string      `json:"timestamp"`
	Payload   interface{} `json:"payload"`
}

type SKUCreatedPayload struct {
	SKU   string  `json:"sku"`
	Price float64 `json:"price"`
	Count int     `json:"count"`
}

type StockChangedPayload struct {
	SKU   string  `json:"sku"`
	Count int     `json:"count"`
	Price float64 `json:"price"`
}
