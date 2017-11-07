package main

import (
	"bytes"
	"fmt"
	"log"
	"strconv"

	"github.com/clockworkcoding/goverdrive"
	"github.com/clockworkcoding/slack"
	"golang.org/x/oauth2"
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

	switch action.Actions[0].Name {
	case "checkOverdrive":
		checkOverdrive(action, token, c, auth.overdriveAccountID)
	case "checkout":
		overdriveCheckout(action, token, c)
	}

}

func checkOverdrive(action action, token string, c *goverdrive.Client, accountID string) {

	library, err := c.GetLibrary(accountID)
	if err != nil {
		overdriveAuthMessage(action, token, err.Error())
		return
	}

	searchParams := goverdrive.NewSearchParamters()
	searchParams.Query = action.Actions[0].Value
	searchParams.Limit = 20
	result, err := c.GetSearch(library.Links.Products.Href, searchParams)
	if err != nil {
		overdriveAuthMessage(action, token, err.Error())
		return
	}
	if result.TotalItems == 0 {
		simpleResponse(action.ResponseURL, "I couldn't find this title in your library's Overdrive catalog", false, token)
	}
	attachments := []slack.Attachment{}
	for _, book := range result.Products {

		var formatBuffer bytes.Buffer
		for i, format := range book.Formats {
			if i > 0 {
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
				Value: availability.ReserveID,
			})
		} else if availability.CopiesOwned > 0 {
			actions = append(actions, slack.AttachmentAction{
				Name:  "placeHold",
				Text:  "Place Hold",
				Type:  "button",
				Value: availability.ReserveID,
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

func overdriveCheckout(action action, token string, c *goverdrive.Client) {
	log.Println("reserveId: " + action.Actions[0].Value)
	result, err := c.CheckoutTitle(action.Actions[0].Value, "")
	if err != nil {
		log.Println(err.Error())
		overdriveAuthMessage(action, token, "Something went wrong, make sure your Library's Overdrive catalog is connected to Booky")
		return
	}

	simpleResponse(action.ResponseURL, fmt.Sprintf("The title has been checked out until %s", result.Expires), true, token)
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
