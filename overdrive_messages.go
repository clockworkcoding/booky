package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/clockworkcoding/slack"
)

func overdriveButton(action action, token string) {
	auth, err := getOverdriveAuth(overdriveAuth{slackUserID: action.User.ID, teamID: action.Team.ID})
	if err != nil {
		if err.Error() == "User not found" {
			overdriveAuthMessage(action, token, "You have to connect Booky to your Library's Overdrive catalog to do that")
			return
		}
		responseError(action.ResponseURL, err.Error(), token)
		return
	}
	if auth.overdriveAccountID == "" {
		overdriveAuthMessage(action, token, "Something went wrong, make sure your Library's Overdrive catalog is connected to Booky")
	}
	// switch action.Actions[0].Name {
	// case "selectedShelf":
	// 	go goodreadsAddToShelf(action, token, auth)
	// case "addToShelf":
	// 	go goodreadsShowShelves(action, token, auth)
	// }

}

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

func newOverdriveButtonGroup(buttons []slack.AttachmentAction) slack.Attachment {
	return slack.Attachment{
		CallbackID: "overdrive",
		Fallback:   "Something went wrong, try again later",
		Actions:    buttons,
	}
}

type overdriveButtonValues struct {
	bookID    string
	shelfID   string
	shelfName string
}

func (values *overdriveButtonValues) encodeOverdriveValues() string {
	return fmt.Sprintf("%v|+|%v|+|%v", values.bookID, values.shelfID, values.shelfName)
}
func (values *overdriveButtonValues) decodeOverdriveValues(valueString string) (err error) {
	valueStrings := strings.Split(valueString, "|+|")
	if len(valueStrings) < 3 {
		err = errors.New("not enough values")
		return
	}
	values.bookID = valueStrings[0]
	values.shelfID = valueStrings[1]
	values.shelfName = valueStrings[2]
	return
}
