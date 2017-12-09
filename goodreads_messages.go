package main

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/clockworkcoding/goodreads"
	"github.com/clockworkcoding/slack"
)

func goodreadsButton(action action, token string) {
	auth, err := getGoodreadsAuth(goodreadsAuth{slackUserID: action.User.ID, teamID: action.Team.ID})
	if err != nil {
		if err.Error() == "User not found" {
			goodreadsAuthMessage(action, token, "You have to connect Booky to your Goodreads account to do that")
			return
		}
		responseError(action.ResponseURL, err.Error(), token)
		return
	}
	if auth.goodreadsUserID == "" {
		goodreadsAuthMessage(action, token, "Something went wrong, make sure your Goodreads account is connected to Booky")
	}
	switch action.Actions[0].Name {
	case "selectedShelf":
		go goodreadsAddToShelf(action, token, auth)
	case "addToShelf":
		go goodreadsShowShelves(action, token, auth)
	}

}

func goodreadsAddToShelf(action action, token string, auth goodreadsAuth) {
	var values goodreadsButtonValues
	err := values.decodeValues(action.Actions[0].SelectedOptions[0].Value)
	if err != nil {
		responseError(action.ResponseURL, err.Error(), token)
		return
	}
	c := goodreads.NewClientWithToken(config.Goodreads.Key, config.Goodreads.Secret, auth.token, auth.secret)
	err = c.AddBookToShelf(values.bookID, values.shelfName)
	if err != nil {
		log.Println(err.Error())
		goodreadsAuthMessage(action, token, "Something went wrong, make sure your Goodreads account is connected to Booky")
		return
	}
	simpleResponse(action.ResponseURL, "Adding...", true, token)
	if title := checkIfBookAdded(auth, values, c); title != "" {
		params := slack.NewResponseMessageParameters()
		params.ResponseType = "ephemeral"
		params.ReplaceOriginal = true
		params.Text = fmt.Sprintf("Succesfully added %s to shelf \"%s.\"", title, values.shelfName)

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

func goodreadsShowShelves(action action, token string, auth goodreadsAuth) {

	var values goodreadsButtonValues
	err := values.decodeValues(action.Actions[0].Value)
	if err != nil {
		responseError(action.ResponseURL, err.Error(), token)
		return
	}
	c := goodreads.NewClientWithToken(config.Goodreads.Key, config.Goodreads.Secret, auth.token, auth.secret)
	shelves, err := c.GetUserShelves(auth.goodreadsUserID)
	if err != nil {
		responseError(action.ResponseURL, err.Error(), token)
		return
	}

	commonShelves := slack.AttachmentActionOptionGroup{Text: "Non-exclusive Shelves"}
	exclusiveShelves := slack.AttachmentActionOptionGroup{Text: "Exclusive Shelves"}
	for _, shelf := range shelves.Shelf_user_shelf {
		values.shelfID = shelf.Shelf_id.Text
		values.shelfName = shelf.Shelf_name.Text
		option := slack.AttachmentActionOption{
			Text:  shelf.Shelf_name.Text,
			Value: values.encodeValues(),
		}

		if shelf.Shelf_exclusive_flag.Text == "false" {
			commonShelves.Options = append(commonShelves.Options, option)
		} else {
			exclusiveShelves.Options = append(exclusiveShelves.Options, option)
		}
	}
	params := slack.NewResponseMessageParameters()
	params.ResponseType = "ephemeral"
	params.ReplaceOriginal = false
	params.Attachments = []slack.Attachment{
		slack.Attachment{
			Text:       "Which shelf should *" + values.bookName + "* be added to?",
			CallbackID: "goodreads",
			Fallback:   "Something went wrong, try again later",
			MarkdownIn: []string{"text", "fields"},
			Actions: []slack.AttachmentAction{
				slack.AttachmentAction{
					Type: "select",
					Name: "selectedShelf",
					OptionGroups: []slack.AttachmentActionOptionGroup{
						exclusiveShelves,
						commonShelves,
					},
				},
			},
		},
	}

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
	bookName  string
}

func (values *goodreadsButtonValues) encodeValues() string {
	return fmt.Sprintf("%v|+|%v|+|%v|+|%v", values.bookID, values.shelfID, values.shelfName, values.bookName)
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
	if len(valueStrings) >= 4 {
		values.bookName = valueStrings[3]
	}
	if len(valueStrings) < 3 {
		err = errors.New("not enough values")
		return
	}
	return
}
