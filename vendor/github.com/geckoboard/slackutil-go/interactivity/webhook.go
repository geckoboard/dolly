package interactivity

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/geckoboard/slackutil-go/messaging"
)

var slackClient = http.Client{}

const (
	// The response should be visible to everyone in the channel
	ResponseInChannel = "in_channel"
	// The response should only be visible to the user who typed the command
	ResponseEphemeral = "ephemeral"
)

type Request struct {
	// this helps identify which type of interactive component sent the
	// payload; An interactive element in a block will have a type of
	// `block_actions`, whereas an interactive element in an attachment will
	// have a type of `interactive_message`.
	Type        string               `json:"type"`
	TriggerID   string               `json:"trigger_id"`
	ResponseURL string               `json:"response_url"`
	Actions     []StaticSelectAction `json:"actions"`
}

// Action contains data from the specific interactive component that was used
type Action struct {
	// this will identify the block within a message that contained the
	// interactive component that was used
	BlockID string `json:"block_id"`
	// this identifies the interactive component itself. Some blocks can
	// contain multiple interactive components, so the block_id alone may
	// not be specific enough to identify the source component
	ActionID string `json:"action_id"`
	// set by your app when you composed the message, this is the value
	// that was specified in the interactive component when an interaction
	// happened. For example, a select menu will have multiple possible
	// values depending on what the user picks from the menu, and value
	// will identify the chosen option
	Value string `json:"value"`
}

type StaticSelectAction struct {
	Action
	SelectedOption messaging.MenuOption `json:"selected_option"`
}

type MessageResponder struct {
	r Request
}

func (m MessageResponder) EphemeralResponse(resp messaging.CommonPayload) {
	payload := struct {
		ResponseType string `json:"response_type"`
		messaging.CommonPayload
	}{ResponseEphemeral, resp}

	b, err := json.Marshal(&payload)
	if err != nil {
		panic(err)
	}

	r, err := http.NewRequest("POST", m.r.ResponseURL, bytes.NewReader(b))
	if err != nil {
		panic(err)
	}

	_, err = slackClient.Do(r)
	if err != nil {
		panic(err)
	}
}

func (m MessageResponder) PublicResponse(resp messaging.CommonPayload) {
	payload := struct {
		ResponseType string `json:"response_type"`
		messaging.CommonPayload
	}{ResponseInChannel, resp}

	d, _ := json.MarshalIndent(payload, "", "  ")
	fmt.Println(string(d))
	b, err := json.Marshal(&payload)
	if err != nil {
		panic(err)
	}

	r, err := http.NewRequest("POST", m.r.ResponseURL, bytes.NewReader(b))
	if err != nil {
		panic(err)
	}

	apiResp, err := slackClient.Do(r)
	if err != nil {
		panic(err)
	}

	ioutil.ReadAll(apiResp.Body)
}

func Handler(with func(Request, MessageResponder)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var sr Request

		json.Unmarshal([]byte(r.FormValue("payload")), &sr)

		with(sr, MessageResponder{sr})
	}
}
