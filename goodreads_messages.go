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
			goodreadsAuthMessage(action, token, "You have to connect Booky to your Goodreads account to do that")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
	}
	switch action.Actions[0].Name {
	case "selectedShelf":
		go goodreadsAddToShelf(action, token, values, auth)
	case "addToShelf":
		go goodreadsShowShelves(action, token, values, auth)
	}

}

func goodreadsAddToShelf(action action, token string, values goodreadsButtonValues, auth goodreadsAuth) {
	c := goodreads.NewClientWithToken(config.Goodreads.Key, config.Goodreads.Secret, auth.token, auth.secret)
	err := c.AddBookToShelf(values.bookID, values.shelfName)
	if err != nil {
		responseError(action.ResponseURL, err.Error(), token)
		return
	}
	simpleResponse(action.ResponseURL, "Adding...", true, token)
	if title := checkIfBookAdded(auth, values, c); title != "" {
		params := slack.NewResponseMessageParameters()
		params.ResponseType = "ephemeral"
		params.ReplaceOriginal = true
		params.Text = fmt.Sprintf("Succesfully added %s to shelf %s.", title, values.shelfName)

		api := slack.New(token)
		err = api.PostResponse(action.ResponseURL, params)
		if err != nil {
			responseError(action.ResponseURL, err.Error(), token)
		}
	} else {
		goodreadsAuthMessage(action, token, "Something went wrong, make sure your Goodreads account is connected to Booky")
	}
}

//checkIfBookAdded returns the title of the book if it is found on the shelf, otherwise, an empty string
func checkIfBookAdded(auth goodreadsAuth, values goodreadsButtonValues, c *goodreads.Client) (title string) {
	for i := 1; ; i++ {
		shelf, err := c.ReviewList(auth.goodreadsUserID, goodreads.ReviewListParameters{Shelf: values.shelfName, Page: i, PerPage: 50})
		if err != nil {
			return
		}

		for _, rev := range shelf.Reviews_review {
			if rev.Reviews_book.Reviews_id.Text == values.bookID {
				return rev.Reviews_book.Reviews_title.Text
			}
		}
		if shelf.Attr_end == shelf.Attr_total {
			return
		}
	}
}

func goodreadsShowShelves(action action, token string, values goodreadsButtonValues, auth goodreadsAuth) {
	c := goodreads.NewClientWithToken(config.Goodreads.Key, config.Goodreads.Secret, auth.token, auth.secret)
	shelves, err := c.GetUserShelves(auth.goodreadsUserID)
	if err != nil {
		responseError(action.ResponseURL, err.Error(), token)
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
	err = api.PostResponse(action.ResponseURL, params)
	if err != nil {
		responseError(action.ResponseURL, err.Error(), token)
	}
}

func goodreadsAuthMessage(action action, token, text string) {
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
	err := api.PostResponse(action.ResponseURL, params)
	if err != nil {
		responseError(action.ResponseURL, err.Error(), token)
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
