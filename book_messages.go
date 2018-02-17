package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/clockworkcoding/goodreads"
	"github.com/clockworkcoding/slack"
)

func createBookPost(values wrongBookButtonValues, wrongBookButtons bool) (params slack.PostMessageParameters, err error) {
	gr := goodreads.NewClient(config.Goodreads.Key, config.Goodreads.Secret)
	if values.bookID == "" {
		results, err := gr.GetSearch(values.Query)
		if err != nil {
			fmt.Println(err.Error())
			return params, err
		}
		if len(results.Search_work) == 0 {
			err = errors.New("no books found")
			return params, err
		}

		if values.Index >= len(results.Search_work) {
			values.Index = 0
		}
		values.bookID = results.Search_work[values.Index].Search_best_book.Search_id.Text
	}
	book, err := gr.GetBook(values.bookID)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	var authorBuffer bytes.Buffer
	for i, author := range book.Book_authors[0].Book_author {
		if i > 0 {
			authorBuffer.WriteString(" | ")
		}
		if author.Book_role.Text != "" {
			authorBuffer.WriteString(author.Book_role.Text)
			authorBuffer.WriteString(": ")
		}
		authorBuffer.WriteString(author.Book_name.Text)
	}

	rating := book.Book_average_rating[0].Text
	numRating, _ := strconv.ParseFloat(rating, 32)
	var stars string
	for i := 0; i < int(numRating+0.5); i++ {
		stars += ":star:"
	}

	rightValues := values.encodeValues()

	attachments := []slack.Attachment{
		slack.Attachment{
			Title:      book.Book_title[0].Text,
			TitleLink:  book.Book_url.Text,
			AuthorName: authorBuffer.String(),
			ThumbURL:   book.Book_image_url[0].Text,
			Fields: []slack.AttachmentField{
				slack.AttachmentField{
					Title: fmt.Sprintf("Avg Rating (%s)", rating),
					Value: stars,
					Short: true,
				},
				slack.AttachmentField{
					Title: "Ratings",
					Value: book.Book_ratings_count[0].Text,
					Short: true,
				},
			},
		},
		slack.Attachment{
			Text:       replaceMarkup(book.Book_description.Text),
			MarkdownIn: []string{"text", "fields"},
			Footer:     fmt.Sprintf("Posted by %s using /booky | Data from Goodreads.com | patreon.com/gobooky ", values.UserName),
		},
	}
	if wrongBookButtons {
		values.Index++
		nextValues := values.encodeValues()
		values.Index -= 2
		prevValues := values.encodeValues()

		nextBookButton := slack.AttachmentAction{
			Name:  "next book",
			Text:  "Wrong Book?",
			Type:  "button",
			Value: string(nextValues),
		}
		nvmBookButton := slack.AttachmentAction{
			Name:  "nvm",
			Text:  "nvm",
			Type:  "button",
			Style: "danger",
			Value: string(nextValues),
		}
		rightBookButton := slack.AttachmentAction{
			Name:  "right book",
			Text:  ":thumbsup:",
			Type:  "button",
			Style: "primary",
			Value: string(rightValues),
		}
		wrongBookButtons := slack.Attachment{
			CallbackID: "wrongbook",
			Fallback:   "Try using both the title and the author's name",
			Actions:    []slack.AttachmentAction{},
		}

		if values.Index >= 0 {
			prevBookButton := slack.AttachmentAction{
				Name:  "previousbook",
				Text:  "previous",
				Type:  "button",
				Value: string(prevValues),
			}
			nextBookButton.Text = "next"
			wrongBookButtons.Actions = append(wrongBookButtons.Actions, prevBookButton)
		}
		wrongBookButtons.Actions = append(wrongBookButtons.Actions, nextBookButton, nvmBookButton, rightBookButton)
		attachments = append(attachments, wrongBookButtons)
	} else {
		values := goodreadsButtonValues{
			bookID:   book.Book_id[0].Text,
			bookName: book.Book_title[0].Text,
		}
		odValue := book.Book_title[0].Text + " " + book.Book_authors[0].Book_author[0].Book_name.Text
		odValue = strings.Replace(odValue, ".", " ", -1)
		log.Printf("Book: %#v", book)
		buttons := slack.Attachment{
			Actions: []slack.AttachmentAction{

				slack.AttachmentAction{
					Name:  "addToShelf",
					Text:  "Add to Goodreads",
					Type:  "button",
					Value: values.encodeValues(),
				},
				slack.AttachmentAction{
					Name:  "checkOverdrive",
					Text:  "Check Your Library",
					Type:  "button",
					Value: odValue,
				},
			},
			CallbackID: "bookaction",
			Fallback:   "Something went wrong, try again later",
		}

		attachments = append(attachments, buttons)
	}
	params = slack.NewPostMessageParameters()
	params.Text = book.Book_title[0].Text
	params.AsUser = true
	params.Attachments = attachments
	return
}

// event responds to events from slack
func event(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	var v map[string]interface{}
	err := decoder.Decode(&v)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if v["token"].(string) != config.Slack.VerificationToken {
		writeError(w, http.StatusForbidden, "Forbidden")
		return
	}
	w.WriteHeader(http.StatusOK)
	if v["type"].(string) == "url_verification" {
		w.Write([]byte(v["challenge"].(string)))
		return
	}
	event, err := json.Marshal(v)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	switch v["event"].(map[string]interface{})["type"].(string) {
	case "message":
		var message eventMessage
		err := json.Unmarshal(event, &message)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		checkTextForBook(message)
	case "link_shared":
		var link eventLinkShared
		err := json.Unmarshal(event, &link)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		generateGoodreadsLinks(link)
	}
}

func generateGoodreadsLinks(link eventLinkShared) {

	if !strings.Contains(link.Event.Links[0].URL, "book/show/") {
		return
	}
	values := wrongBookButtonValues{bookID: strings.Split(link.Event.Links[0].URL, "book/show/")[1]}
	_, token, _, err := getSlackAuth(link.TeamID)
	if err != nil {
		log.Output(0, err.Error())
		return
	}

	post, err := createBookPost(values, false)
	if err != nil {
		log.Output(0, err.Error())
		return
	}
	post.Attachments[0].Text = post.Attachments[1].Text
	post.Attachments[0].Actions = post.Attachments[2].Actions
	post.Attachments[0].CallbackID = post.Attachments[2].CallbackID
	post.Attachments[0].Footer = post.Attachments[2].Footer

	api := slack.New(token)
	params := slack.UnfurlParameters{
		Timestamp: link.Event.MessageTs,
		Unfurls: []slack.Unfurl{
			slack.Unfurl{
				UnfurlURL:  link.Event.Links[0].URL,
				Attachment: post.Attachments[0],
			},
		},
	}
	api.Unfurl(context.Background(), link.Event.Channel, params)
}
func wrongBookButton(action action, token string) {

	var values wrongBookButtonValues
	err := values.decodeValues(action.Actions[0].Value)
	if err != nil {
		responseError(action.ResponseURL, err.Error(), token)
		return
	}

	api := slack.New(token)
	if action.User.ID != values.User {
		responseParams := slack.NewResponseMessageParameters()
		responseParams.ResponseType = "ephemeral"
		responseParams.ReplaceOriginal = false
		responseParams.Text = fmt.Sprintf("Only %s can update this book", values.UserName)
		err = api.PostResponse(action.ResponseURL, responseParams)
		if err != nil {
			responseError(action.ResponseURL, err.Error(), token)

		}
		return

	}
	var timestamp string
	var v map[string]interface{}
	err = json.Unmarshal(action.OriginalMessage, &v)
	if err != nil {
		timestamp = ""
	} else {
		timestamp = v["ts"].(string)
	}
	wrongBookButtons := true
	switch action.Actions[0].Name {
	case "right book":
		wrongBookButtons = false
	case "nvm":
		_, _, err = api.DeleteMessage(action.Channel.ID, timestamp)
		if err != nil {
			if !values.IsEphemeral {
				responseError(action.ResponseURL, err.Error(), token)
			} else {
				simpleResponse(action.ResponseURL, "Sorry you couldn't find your book. Try searching for both the author and title together", true, token)
			}

		}
		return
	}

	params, err := createBookPost(values, wrongBookButtons)
	if err != nil {
		if err.Error() == "no books found" {
			simpleResponse(action.ResponseURL, "No books found, try a broader search", true, token)
		} else {
			responseError(action.ResponseURL, err.Error(), token)
		}
		return
	}
	//If it's an ephemeral post, replace it with an in_channel post after finding the right one, otherwise just update
	if values.IsEphemeral {
		responseParams := slack.NewResponseMessageParameters()
		responseParams.Text = params.Text
		responseParams.Attachments = params.Attachments

		if action.Actions[0].Name == "right book" {
			simpleResponse(action.ResponseURL, "Posting your book", true, token)
			responseParams.ReplaceOriginal = false
			responseParams.ResponseType = "in_channel"
		} else {
			responseParams.ReplaceOriginal = true
			responseParams.ResponseType = "ephemeral"
		}

		err = api.PostResponse(action.ResponseURL, responseParams)
		if err != nil {
			responseError(action.ResponseURL, err.Error(), token)
		}
	} else {
		updateParams := slack.UpdateMessageParameters{
			Timestamp:   timestamp,
			Text:        params.Text,
			Attachments: params.Attachments,
			Parse:       params.Parse,
			LinkNames:   params.LinkNames,
			AsUser:      params.AsUser,
		}

		_, _, _, err = api.UpdateMessageWithAttachments(context.Background(), action.Channel.ID, updateParams)
		if err != nil {
			responseError(action.ResponseURL, err.Error(), token)
			return
		}
	}
}

type wrongBookButtonValues struct {
	User        string `json:"user"`
	UserName    string `json:"user_name"`
	Query       string `json:"query"`
	Index       int    `json:"index"`
	IsEphemeral bool   `json:"is_ephemeral"`
	bookID      string
}

func (values *wrongBookButtonValues) encodeValues() string {
	return fmt.Sprintf("%v|+|%v|+|%v|+|%v|+|%v", values.Index, values.IsEphemeral, values.Query, values.User, values.UserName)
}
func (values *wrongBookButtonValues) decodeValues(valueString string) (err error) {
	valueStrings := strings.Split(valueString, "|+|")
	if len(valueStrings) < 5 {
		err = errors.New("not enough values")
		return
	}
	index, err := strconv.ParseInt(valueStrings[0], 10, 32)
	if err != nil {
		return
	}
	values.Index = int(index)
	values.IsEphemeral, err = strconv.ParseBool(valueStrings[1])
	if err != nil {
		return
	}
	values.Query = valueStrings[2]
	values.User = valueStrings[3]
	values.UserName = valueStrings[4]
	return
}
