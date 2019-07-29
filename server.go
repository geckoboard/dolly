package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"

	"github.com/geckoboard/slackutil-go/interactivity"
	"github.com/geckoboard/slackutil-go/messaging"
	slashcommand "github.com/geckoboard/slackutil-go/slashcommand"

	"github.com/julienschmidt/httprouter"
)

const SelectPresetActionID = "choose_preset"
const SelectPresetBlockID = "preset_dropdown"

var presets = []preset{
	{"off position", "0"},
	{"whole room", "1"},
	{"whiteboard", "2"},
	{"presenters", "3"},
	{"Ben", "4"},
	{"position 5", "5"},
	{"position 6", "6"},
}

type preset struct {
	Name string
	ID   string
}

func makeHttpHandler() *httprouter.Router {
	router := httprouter.New()

	s := httpServer{}

	router.HandlerFunc("POST", "/slack/interactivity", interactivity.Handler(selectCallback))
	router.POST("/slack/slash-camera", s.slashCamera)

	return router
}

type httpServer struct {
}

func respondWithError(w http.ResponseWriter, statusCode int, msg string) {
	w.WriteHeader(statusCode)
	w.Write([]byte(msg))
}

func makePresetPrompt() messaging.Block {
	selectBlock := messaging.StaticSelect{
		Placeholder: messaging.PlainText("Choose a preset"),
		ActionID:    SelectPresetActionID,
		Options:     []messaging.MenuOption{},
	}
	for _, p := range presets {
		opt := messaging.MenuOption{
			Text:  messaging.PlainText(p.Name),
			Value: p.ID,
		}

		selectBlock.Options = append(selectBlock.Options, opt)
	}

	return messaging.Section{
		BlockID:   SelectPresetBlockID,
		Text:      messaging.PlainText("Change the camera preset"),
		Accessory: selectBlock,
	}
}

func selectCallback(req interactivity.Request, resp interactivity.MessageResponder) {
	if len(req.Actions) < 0 {
		return
	}

	for _, action := range req.Actions {
		if action.BlockID != SelectPresetBlockID {
			continue
		}
		if action.ActionID != SelectPresetActionID {
			continue
		}

		desiredPreset := action.SelectedOption.Value

		u, err := url.Parse(SyrupBaseURL)
		if err != nil {
			resp.EphemeralResponse(messaging.CommonPayload{
				Text: fmt.Sprintf("bad syrup url: %q", err),
			})
			return
		}

		u.Path = path.Join("presets", desiredPreset, "recall")
		http.Post(u.String(), "application/json", nil)

		// TODO: errors lol
		return
	}

	resp.EphemeralResponse(messaging.CommonPayload{
		Text: "could not handle interaction",
	})
}

func (h httpServer) slashCamera(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	command, err := slashcommand.ParseSlashCommandRequest(r)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "could not parse payload")
		return
	}

	slashCamera := slashcommand.DelayedSlashResponse{
		Handler: func(ctx context.Context, req slashcommand.SlashCommandRequest, resp slashcommand.MessageResponder) {
			resp.PublicResponse(messaging.CommonPayload{
				Blocks: []messaging.Block{makePresetPrompt()},
			})

		},

		ShowSlashCommandInChannel: false,
	}

	slashCamera.Run(w, *command)
}
