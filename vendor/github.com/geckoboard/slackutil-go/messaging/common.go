package messaging

type CommonPayload struct {
	Text   string  `json:"text"`
	Blocks []Block `json:"blocks"`
}
