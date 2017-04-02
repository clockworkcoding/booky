package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/clockworkcoding/goodreads"
	"github.com/clockworkcoding/slack"
)

func goodreadsButton(w http.ResponseWriter, action action, token string) {
	var values goodreadsButtonValues
	err := values.decodeValues(action.Actions[0].Value)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	auth, err := getGoodreadsAuth(goodreadsAuth{slackUserID: action.User.ID, teamID: action.Team.ID})
	if err != nil {
		if err.Error() == "User not found" {
			goodreadsAuthMessage(w, action, token)
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
	}
	c := goodreads.NewClientWithToken(config.Goodreads.Key, config.Goodreads.Secret, auth.token, auth.secret)
	shelves, err := c.GetUserShelves(auth.goodreadsUserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var attachments []slack.Attachment
	var shelfButtons []slack.AttachmentAction
	for i, shelf := range shelves.Shelf_user_shelf {
		if (i+1)%5 == 0 {
			attachments = append(attachments, newGoodreadsButtonGroup(shelfButtons))
			shelfButtons = []slack.AttachmentAction{}
		}
		values.shelfID = shelf.Shelf_id.Text
		button := slack.AttachmentAction{
			Name:  "selectedShelf",
			Text:  shelf.Shelf_name.Text,
			Type:  "button",
			Value: values.encodeValues(),
		}
		shelfButtons = append(shelfButtons, button)
	}
	params := slack.NewResponseMessageParameters()
	params.ResponseType = "ephemeral"
	params.ReplaceOriginal = false
	params.Text = "Which shelf?"
	params.Attachments = attachments
	params.Attachments = append(params.Attachments, newGoodreadsButtonGroup(shelfButtons))

	api := slack.New(token)
	err = api.PostResponse(action.ResponseURL, params.Text, params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
	}
}

func goodreadsAuthMessage(w http.ResponseWriter, action action, token string) {
	params := slack.NewResponseMessageParameters()
	params.ResponseType = "ephemeral"
	params.ReplaceOriginal = false
	params.Text = "You have to connect Booky to your Goodreads account to do that"
	params.Attachments = []slack.Attachment{
		slack.Attachment{
			Title:     "Connect to Goodreads",
			TitleLink: fmt.Sprintf("%s/gradd?team=%s&user=%s", config.URL, action.Team.ID, action.User.ID),
			Text:      "Try the action again when you're done",
		},
	}

	api := slack.New(token)
	err := api.PostResponse(action.ResponseURL, params.Text, params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
	}
}

func newGoodreadsButtonGroup(buttons []slack.AttachmentAction) slack.Attachment {
	return slack.Attachment{
		CallbackID: "goodreads",
		Fallback:   "Something went wrong, try again later",
		Actions:    buttons,
	}
}

type goodreadsButtonValues struct {
	bookID    string `json:"bookID"`
	shelfID   string `json:"shelfID"`
	shelfName string `json:"shelfName"`
}

func (values *goodreadsButtonValues) encodeValues() string {
	return fmt.Sprintf("%v|+|%v|+|%v", values.bookID, values.shelfID, values.shelfName)
}
func (values *goodreadsButtonValues) decodeValues(valueString string) (err error) {
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
