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

func createBookPost(values wrongBookButtonValues, wrongBookButtons bool, showFullDescription bool) (params slack.PostMessageParameters, err error) {
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
		if i > 4 {
			authorBuffer.WriteString("...")
			break
		}
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

	patreonText := ""
	if len(config.Patreon) > 0 {
		patreonText = " | " + config.Patreon
	}

	bookshopLink := getBookshopLink(book.Book_isbn13[0].Text, book.Book_work[0].Book_original_title.Text)
	if len(bookshopLink) > 0 {
    bookshopLink = " \n<" + bookshopLink + " | Buy this book from Bookshop.org> (<http://booky.fyi/affiliate |affiate disclosure>)"
	}

	attachments := []slack.Attachment{
		{
			Title:      book.Book_title[0].Text,
			TitleLink:  book.Book_url.Text,
			AuthorName: authorBuffer.String(),
			ThumbURL:   book.Book_image_url[0].Text,
			Fields: []slack.AttachmentField{
				{
					Title: fmt.Sprintf("Avg Rating (%s)", rating),
					Value: stars,
					Short: true,
				},
				{
					Title: "Ratings",
					Value: book.Book_ratings_count[0].Text,
					Short: true,
				},
			},
		},
		{
			Text:       replaceMarkup(book.Book_description.Text),
			MarkdownIn: []string{"text", "fields"},
			Footer:     fmt.Sprintf("Posted by @%s using /booky | Data from Goodreads.com%s%s", values.UserName, patreonText, bookshopLink),
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
		buttons := slack.Attachment{
			Actions: []slack.AttachmentAction{

				{
					Name:  "addToShelf",
					Text:  "Add to Goodreads",
					Type:  "button",
					Value: values.encodeValues(),
				},
				//slack.AttachmentAction{
				//	Name:  "checkOverdrive",
				//	Text:  "Check Your Library",
				//	Type:  "button",
				//	Value: odValue,
				//},

			},
			CallbackID: "bookaction",
			Fallback:   "Something went wrong, try again later",
		}

		attachments = append(attachments, buttons)
	}

	maxLength := 140
	if config.DescriptionLength > 0 {
		maxLength = config.DescriptionLength
	}
	if !showFullDescription && len(attachments[1].Text) > maxLength+3 {
		attachments[2].Actions = append(attachments[2].Actions,
			slack.AttachmentAction{
				Name:  "fullDescription",
				Text:  "Show Full Description",
				Type:  "button",
				Value: values.encodeValues(),
			})
		attachments[1].Text = attachments[1].Text[:maxLength] + "..."
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
	if v["type"].(string) == "url_verification" {
		w.Write([]byte(v["challenge"].(string)))
		return
	}
	if v["token"].(string) != config.Slack.VerificationToken {
		writeError(w, http.StatusForbidden, "Forbidden: verification failed")
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
		tokenized := strings.Split(message.Event.Text, "_")
		if len(tokenized) < 2 {
			return
		}
		queryText := tokenized[1]
		channel := message.Event.Channel
		teamID := message.TeamID
		user := message.Event.User
		checkTextForBook(queryText, teamID, channel, user)
	case "link_shared":
		var link eventLinkShared
		err := json.Unmarshal(event, &link)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		generateGoodreadsLinks(link)
	}
}
func menuSearch(action action) {
	_, token, _, err := getSlackAuth(action.Team.ID)
	api := slack.New(token)
	if len(strings.Split(action.Message.Text, " ")) == 1 {
		values := wrongBookButtonValues{
			Index:       0,
			User:        action.User.ID,
			Query:       action.Message.Text,
			IsEphemeral: true,
			UserName:    action.User.Name,
		}

		responseURL := action.ResponseURL

		params, err := createBookPost(values, true, true)
		if err != nil {
			if err.Error() == "no books found" {
				simpleResponse(responseURL, "No books found, try a broader search", false, token)
			} else {
				responseError(responseURL, err.Error(), token)
			}
			return
		}
		responseParams := slack.NewResponseMessageParameters()
		responseParams.ResponseType = "ephemeral"
		responseParams.ReplaceOriginal = true
		responseParams.Text = params.Text
		responseParams.Attachments = params.Attachments
		err = api.PostResponse(responseURL, responseParams)
		if err != nil {
			responseError(responseURL, err.Error(), token)
		}
	}
	var elements []slack.DialogElement

	options := findTitleOptions(action.Message.Text, "*")
	options = append(options, findTitleOptions(action.Message.Text, "_")...)
	options = append(options, findTitleOptions(action.Message.Text, "\"")...)

	if len(options) > 0 {
		elements = append(elements, slack.DialogElement{
			Label:    "Potential titles",
			Type:     "select",
			Name:     "selecttitle",
			Options:  options,
			Optional: false,
			Value:    options[0].Value,
		})
	} else {
		elements = append(elements, slack.DialogElement{
			Label:    "Custom Search",
			Type:     "text",
			Name:     "searchtext",
			Hint:     "booky can't find a title in the post, but you can search for one here",
			Optional: false,
		})
	}

	lookUpDialog := slack.Dialog{
		CallbackID:     "lookUpDialog",
		NotifyOnCancel: false,
		SubmitLabel:    "Search",
		Title:          "Look up from post",
		Elements:       elements,
	}

	err = api.PostDialog(action.TriggerID, token, lookUpDialog)
	if err != nil {
		log.Println(0, err.Error())
		return
	}
}

func findTitleOptions(text string, sep string) (options []slack.DialogOption) {
	if bold := strings.Split(text, sep); len(bold) > 1 {
		for i, title := range bold {
			if i%2 == 0 {
				continue
			}
			option := slack.DialogOption{
				Label: title,
				Value: title,
			}

			options = append(options, option)
		}
	}
	return options
}

func generateGoodreadsLinks(link eventLinkShared) {

  log.Println(link.Event.Links[0].URL)
	if !strings.Contains(link.Event.Links[0].URL, "book/show/") {
		return
	}
	values := wrongBookButtonValues{bookID: strings.Split(link.Event.Links[0].URL, "book/show/")[1]}
	_, token, _, err := getSlackAuth(link.TeamID)
	if err != nil {
		log.Output(0, err.Error())
		return
	}

	post, err := createBookPost(values, false, false)
	if err != nil {
		log.Output(0, err.Error())
		return
	}
	post.Attachments[0].Text = post.Attachments[1].Text
	post.Attachments[0].Actions = post.Attachments[2].Actions
	post.Attachments[0].CallbackID = post.Attachments[2].CallbackID
	post.Attachments[0].Footer = post.Attachments[1].Footer

	api := slack.New(token)
	params := slack.UnfurlParameters{
		Timestamp: link.Event.MessageTs,
		Unfurls: []slack.Unfurl{
			{
				UnfurlURL:  link.Event.Links[0].URL,
				Attachment: post.Attachments[0],
			},
		},
	}
  err = api.Unfurl(context.Background(), link.Event.Channel, params)
	if err != nil {
    log.Println("Error unfurling link: " + err.Error())
		return
	}
}
func wrongBookButton(action action, token string) {

	var values wrongBookButtonValues
	log.Output(0, action.Actions[0].Value)
	err := values.decodeValues(action.Actions[0].Value)
	if err != nil {
		responseError(action.ResponseURL, err.Error(), token)
		return
	}

	api := slack.New(token)
	if action.User.ID != values.User && action.Actions[0].Name != "fullDescription" {
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
	fullDescription := true
	switch action.Actions[0].Name {
	case "right book":
		wrongBookButtons = false
		fullDescription = false
	case "fullDescription":
		wrongBookButtons = false
		fullDescription = true
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

	params, err := createBookPost(values, wrongBookButtons, fullDescription)
	if err != nil {
		if err.Error() == "no books found" {
			simpleResponse(action.ResponseURL, "No books found, try a broader search", true, token)
		} else {
			responseError(action.ResponseURL, err.Error(), token)
		}
		return
	}
	//If it's an ephemeral post, replace it with an in_channel post after finding the right one, otherwise just update
	if action.Actions[0].Name == "fullDescription" {
		responseParams := slack.NewResponseMessageParameters()
		responseParams.Text = params.Text + " (Full Description)"
		responseParams.Attachments = params.Attachments
		responseParams.ReplaceOriginal = false
		responseParams.ResponseType = "ephemeral"

		err = api.PostResponse(action.ResponseURL, responseParams)
		if err != nil {
			responseError(action.ResponseURL, err.Error(), token)
		}
	} else if values.IsEphemeral {
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
	valueString := fmt.Sprintf("%v|+|%v|+|%v|+|%v|+|%v", values.Index, values.IsEphemeral, values.Query, values.User, values.UserName)
	return valueString
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
