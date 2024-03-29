package slashcommand

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	bugsnag "github.com/bugsnag/bugsnag-go"
	"github.com/davecgh/go-spew/spew"
	"github.com/geckoboard/slackutil-go/messaging"
)

var slackClient = http.Client{}
var forceShowSlashCommandInChannelResponse = Response{ResponseType: ResponseInChannel}

type DelayedSlashResponse struct {
	// A mesage to send the user while we're preparing a response to
	PendingResponse messaging.CommonPayload

	// Should the command be visible to all other users in the channel?
	// Doing this changes how we respond to the slash command webhook
	ShowSlashCommandInChannel bool

	Handler func(context.Context, SlashCommandRequest, MessageResponder)
}

func (d DelayedSlashResponse) Run(w http.ResponseWriter, command SlashCommandRequest) {
	if d.ShowSlashCommandInChannel {
		// By default slack treats responses to slash commands as
		// "ephemeral", and will prevent the slash command from showing
		// up in chat history.
		// If we signal in our response that the slash command is not
		// ephemeral, then the original command will appear in the chat
		// history.
		// https://api.slack.com/slash-commands#responding_immediate_response
		RespondWith(w, forceShowSlashCommandInChannelResponse)
	}

	// We run the handler in a goroutine so that we can confirm receipt of slack's
	// slash command webhook (by returning 200 OK) as soon as possible
	go d.runHandler(command)
}

func (d DelayedSlashResponse) runHandler(command SlashCommandRequest) {
	ctx := context.Background()

	responder := MessageResponder{command}

	done := make(chan struct{})

	// Not using a waitgroup here as we don't really care about cleaning up this goroutine
	go func() {
		defer bugsnag.AutoNotify(ctx)
		d.Handler(ctx, command, responder)
		close(done)
	}()

	notifyUserTimeout := time.After(700 * time.Millisecond)

	for {
		select {
		case <-done:
			return
		case <-notifyUserTimeout:
			responder.EphemeralResponse(d.PendingResponse)
		}
	}
}

type MessageResponder struct {
	command SlashCommandRequest
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

	r, err := http.NewRequest("POST", m.command.ResponseURL, bytes.NewReader(b))
	if err != nil {
		panic(err)
	}

	apiResp, err := slackClient.Do(r)
	if err != nil {
		panic(err)
	}

	fmt.Println(apiResp)
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

	r, err := http.NewRequest("POST", m.command.ResponseURL, bytes.NewReader(b))
	if err != nil {
		panic(err)
	}

	apiResp, err := slackClient.Do(r)
	if err != nil {
		panic(err)
	}

	rawBody, err := ioutil.ReadAll(apiResp.Body)
	spew.Dump(string(rawBody))
}

type SlashCommandResponder interface {
	PublicResponse(Response)
}
