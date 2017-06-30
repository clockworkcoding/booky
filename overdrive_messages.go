package main

import (
	"fmt"

	"github.com/clockworkcoding/slack"
)

func overdriveAuthMessage(action action, token, text string) {
	params := slack.NewResponseMessageParameters()
	params.ResponseType = "ephemeral"
	params.ReplaceOriginal = false
	params.Text = text
	params.Attachments = []slack.Attachment{
		slack.Attachment{
			ThumbURL:  "https://developerportaldev.blob.core.windows.net/media/Default/images/newLogos/OverDrive_Logo_42x42_rgb.jpg",
			Title:     "Connect to your library's Overdrive catalog",
			TitleLink: fmt.Sprintf("%s/odadd?team=%s&user=%s", config.URL, action.Team.ID, action.User.ID),
			Text:      "Try the action again when you're done",
		},
	}

	api := slack.New(token)
	err := api.PostResponse(action.ResponseURL, params)
	if err != nil {
		responseError(action.ResponseURL, err.Error(), token)
	}
}
