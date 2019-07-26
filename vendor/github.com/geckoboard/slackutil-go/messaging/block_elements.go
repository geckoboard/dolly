package messaging

import "encoding/json"

type MenuOption struct {
	Text  Text   `json:"text"`
	Value string `json:"value"`
	URL   string `json:"url,omitempty"`
}

type StaticSelect struct {
	Placeholder   Text         `json:"placeholder"`
	ActionID      string       `json:"action_id"`
	Options       []MenuOption `json:"options,omitempty"`
	InitialOption *MenuOption  `json:"initial_option,omitempty"`
}

// create an alias that does not implement encoding/json.Marshaller
// so that we can encode the alias without invoking the custom marshaller
// on the original type
type encodingSpecificStaticSelect StaticSelect

// MarshalJSON is a custom marshaller that adds static type information to the
// json payload
func (s StaticSelect) MarshalJSON() ([]byte, error) {
	payload := struct {
		Type string `json:"type"`
		encodingSpecificStaticSelect
	}{"static_select", encodingSpecificStaticSelect(s)}

	return json.Marshal(payload)
}
