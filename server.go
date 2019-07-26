package main

import (
	"context"
	"net/http"

	"github.com/geckoboard/slackutil-go/interactivity"
	"github.com/geckoboard/slackutil-go/messaging"
	slashcommand "github.com/geckoboard/slackutil-go/slashcommand"

	"github.com/julienschmidt/httprouter"
)

const SelectPresetActionID = "choose_preset"
const SelectPresetBlockID = "preset_dropdown"

var presets = []preset{
	{"entire room", "0"},
	{"whiteboard", "1"},
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

		http.Post("http://192.168.40.108:8000/presets/"+desiredPreset+"/recall", "application/json", nil)
	}
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
