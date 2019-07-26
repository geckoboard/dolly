package messaging

import "encoding/json"

type Block interface {
	Block()
}

func PlainText(text string) Text {
	return Text{
		Text: text,
		Type: "plain_text",
	}
}

type Text struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type Section struct {
	Text      Text          `json:"text"`
	BlockID   string        `json:"block_id,omitempty"`
	Fields    []interface{} `json:"fields,omitempty"`
	Accessory interface{}   `json:"accessory,omitempty"`
}

func (s Section) Block() {}

type encodingSpecificSection Section

func (s Section) MarshalJSON() ([]byte, error) {
	payload := struct {
		Type string `json:"type"`
		encodingSpecificSection
	}{"section", encodingSpecificSection(s)}

	return json.Marshal(payload)
}
