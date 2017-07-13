package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"golang.org/x/oauth2"

	"github.com/clockworkcoding/goverdrive"
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

	oauthToken := &oauth2.Token{
		AccessToken:  auth.token,
		RefreshToken: auth.refreshToken,
		Expiry:       auth.expiry,
		TokenType:    auth.tokenType,
	}
	c := goverdrive.NewClient(config.Overdrive.Key, config.Overdrive.Secret, "", oauthToken, true)
	library, err := c.GetLibrary(auth.overdriveAccountID)
	if err != nil {
		overdriveAuthMessage(action, token, err.Error())
		return
	}
	if auth.overdriveAccountID == "" {
		overdriveAuthMessage(action, token, "Something went wrong, make sure your Library's Overdrive catalog is connected to Booky")
	}
	var odValues overdriveSearchButtonValues
	odValues.decodeValues(action.Actions[0].Value)
	searchParams := goverdrive.NewSearchParamters()
	searchParams.Query = odValues.query
	result, err := c.GetSearch(library.Links.Products.Href, searchParams)
	if err != nil {
		overdriveAuthMessage(action, token, err.Error())
		return
	}
	if result.TotalItems == 0 {
		simpleResponse(action.ResponseURL, "I couldn't find this title in your library's Overdrive catalog", false, token)
	}
	attachments := []slack.Attachment{}
	for i, book := range result.Products {
		if i >= 5 {
			break
		}
		var formatBuffer bytes.Buffer
		for j, format := range book.Formats {
			if j > 0 {
				formatBuffer.WriteString(" | ")
			}
			formatBuffer.WriteString(format.Name)
		}
		availability, err := c.GetAvailability(book.Links.Availability.Href)
		if err != nil {
			simpleResponse(action.ResponseURL, "Something went wrong, please try again", false, token)
		}
		log.Println(availability)
		actions := []slack.AttachmentAction{}
		if availability.CopiesAvailable > 0 {
			actions = append(actions, slack.AttachmentAction{
				Name:  "checkout",
				Text:  "Checkout",
				Type:  "button",
				Value: "?",
			})
		} else if availability.CopiesOwned > 0 {
			actions = append(actions, slack.AttachmentAction{
				Name:  "placehold",
				Text:  "Place Hold",
				Type:  "button",
				Value: "?",
			})
		}
		attachments = append(attachments, slack.Attachment{
			Title:      fmt.Sprintf("%s by %s", book.Title, book.PrimaryCreator.Name),
			TitleLink:  book.Links.Self.Href,
			AuthorName: formatBuffer.String(),
			ThumbURL:   book.Images.Thumbnail.Href,
			Fields: []slack.AttachmentField{
				slack.AttachmentField{
					Title: "Available Copies",
					Short: true,
					Value: fmt.Sprintf("%d/%d", availability.CopiesAvailable, availability.CopiesOwned),
				},
				slack.AttachmentField{
					Title: "Holds",
					Value: strconv.Itoa(availability.NumberOfHolds),
					Short: true,
				},
			},
			Actions:    actions,
			CallbackID: "overdrive",
			Fallback:   "Something went wrong, try again later",
		})
	}
	api := slack.New(token)
	params := slack.NewResponseMessageParameters()
	params.ResponseType = "ephemeral"
	params.ReplaceOriginal = false
	params.Text = fmt.Sprintf("Here are the results from %s", library.Name)
	params.AsUser = false
	params.Attachments = attachments
	err = api.PostResponse(action.ResponseURL, params)
	if err != nil {
		responseError(action.ResponseURL, err.Error(), token)

	}

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

type overdriveSearchButtonValues struct {
	query        string
	index        string
	collectionID string
}

func (values *overdriveSearchButtonValues) encodeValues() string {
	return fmt.Sprintf("%v|+|%v|+|%v", values.query, values.index, values.collectionID)
}
func (values *overdriveSearchButtonValues) decodeValues(valueString string) (err error) {
	valueStrings := strings.Split(valueString, "|+|")
	if len(valueStrings) < 3 {
		err = errors.New("not enough values")
		return
	}
	values.query = valueStrings[0]
	values.index = valueStrings[1]
	values.collectionID = valueStrings[2]
	return
}
