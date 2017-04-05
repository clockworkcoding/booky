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
			goodreadsAuthMessage(w, action, token, "You have to connect Booky to your Goodreads account to do that")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
	}
	switch action.Actions[0].Name {
	case "selectedShelf":
		goodreadsAddToShelf(w, action, token, values, auth)
	case "addToShelf":
		goodreadsShowShelves(w, action, token, values, auth)
	}

}

func goodreadsAddToShelf(w http.ResponseWriter, action action, token string, values goodreadsButtonValues, auth goodreadsAuth) {
	c := goodreads.NewClientWithToken(config.Goodreads.Key, config.Goodreads.Secret, auth.token, auth.secret)
	err := c.AddBookToShelf(values.bookID, values.shelfName)
	if err != nil {
		responseError(action.ResponseURL, err.Error(), token)
		return
	}
	shelf, err := c.ReviewList(auth.goodreadsUserID, goodreads.ReviewListParameters{Shelf: values.shelfName})
	if err != nil {
		responseError(action.ResponseURL, err.Error(), token)
		return
	}
	if title := checkIfBookAdded(shelf, values); title != "" {
		params := slack.NewResponseMessageParameters()
		params.ResponseType = "ephemeral"
		params.ReplaceOriginal = true
		params.Text = fmt.Sprintf("Succesfully added %s to shelf %s.", title, values.shelfName)

		api := slack.New(token)
		err = api.PostResponse(action.ResponseURL, params.Text, params)
		if err != nil {
			responseError(action.ResponseURL, err.Error(), token)
		}
	} else {
		goodreadsAuthMessage(w, action, token, "Something went wrong, make sure your Goodreads account is connected to Booky")
	}
}

func checkIfBookAdded(reviews goodreads.Reviews_reviews, values goodreadsButtonValues) (title string) {
	for _, rev := range reviews.Reviews_review {
		if rev.Reviews_book.Reviews_id.Text == values.bookID {
			return rev.Reviews_book.Reviews_title.Text
		}
	}
	return ""
}

func goodreadsShowShelves(w http.ResponseWriter, action action, token string, values goodreadsButtonValues, auth goodreadsAuth) {
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
		values.shelfName = shelf.Shelf_name.Text
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

func goodreadsAuthMessage(w http.ResponseWriter, action action, token, text string) {
	params := slack.NewResponseMessageParameters()
	params.ResponseType = "ephemeral"
	params.ReplaceOriginal = false
	params.Text = text
	params.Attachments = []slack.Attachment{
		slack.Attachment{
			ThumbURL:  "https://s.gr-assets.com/assets/icons/goodreads_icon_100x100-4a7d81b31d932cfc0be621ee15a14e70.png",
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
	bookID    string
	shelfID   string
	shelfName string
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
