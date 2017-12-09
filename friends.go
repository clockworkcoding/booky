package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/clockworkcoding/goodreads"
	"github.com/clockworkcoding/slack"
)

func friendsCommand(w http.ResponseWriter, r *http.Request) {
	queryText := r.FormValue("text")
	teamID := r.FormValue("team_id")
	responseURL := r.FormValue("response_url")
	_, token, _, err := getSlackAuth(teamID)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	w.Write([]byte(""))
	if queryText == "?" {
		simpleResponse(responseURL, "If you're having trouble or just want to leave a message go to http://booky.fyi/contact or email Max@ClockworkCoding.com", true, token)
		return
	}
	go simpleResponse(responseURL, "Getting things ready", true, token)

	attachments := []slack.Attachment{
		slack.Attachment{
			Title:      "Become friends on Goodreads",
			AuthorName: "Visit Goodreads to accept friend requests",
			AuthorLink: "https://www.goodreads.com/friend/requests",
			ThumbURL:   "https://s.gr-assets.com/assets/icons/goodreads_icon_100x100-4a7d81b31d932cfc0be621ee15a14e70.png",
			CallbackID: "friendAction",
			Fallback:   "Something went wrong, try again later",
			Footer:     "Goodreads will show the name, email, and books associated with your Goodreads account to anybody you add as a friend. | patreon.com/gobooky",
			Actions: []slack.AttachmentAction{

				slack.AttachmentAction{
					Name:  "addSelfToFriends",
					Text:  "Let others add you as a friend",
					Type:  "button",
					Style: "primary",
				},
				slack.AttachmentAction{
					Name:  "addAllFriends",
					Text:  "Add all these awesome people",
					Type:  "button",
					Style: "danger",
					Confirm: &slack.ConfirmationField{
						Title:       "Are you sure?",
						Text:        "Goodreads will show the name, email, and books associated with your Goodreads account to anybody you add as a friend",
						OkText:      "Yes, they're cool",
						DismissText: "No, thanks",
					},
				},
			},
		},
	}
	responseParams := slack.NewResponseMessageParameters()
	responseParams.ResponseType = "in_channel"
	responseParams.ReplaceOriginal = true
	responseParams.Text = "Goodreads Friend-o-rama!"
	responseParams.Attachments = attachments
	api := slack.New(token)
	err = api.PostResponse(responseURL, responseParams)
	if err != nil {
		responseError(responseURL, err.Error(), token)
	}
}

func addSelfToFriendsButton(action action, token string) {
	log.Println("adding self to the list")
	auth, err := getGoodreadsAuth(goodreadsAuth{slackUserID: action.User.ID, teamID: action.Team.ID})
	if err != nil {
		if err.Error() == "User not found" {
			goodreadsAuthMessage(action, token, "You have to connect Booky to your Goodreads account to do that")
			return
		}
		responseError(action.ResponseURL, err.Error(), token)
		return
	}
	values := friendButtonValues{
		User:        action.User.ID,
		UserName:    action.User.Name,
		goodreadsID: auth.goodreadsUserID,
	}

	newButton := slack.AttachmentAction{
		Name:  "addFriend",
		Text:  "add @" + values.UserName,
		Type:  "button",
		Value: values.encodeValues(),
	}

	var orignalMessage slack.Message
	err = json.Unmarshal(action.OriginalMessage, &orignalMessage)
	if err != nil {
		responseError(action.ResponseURL, err.Error(), token)
		return
	}
	attachmentCount := len(orignalMessage.Attachments)
	actionCount := len(orignalMessage.Attachments[attachmentCount-1].Actions)
	if attachmentCount == 20 && actionCount == 5 {
		simpleResponse(action.ResponseURL, "Oops! I can't add any more people here, start another Friend-O-Rama by typing '/booky -friends'", false, token)
		return
	}

	var oldValues friendButtonValues
	for _, attachment := range orignalMessage.Attachments {
		for _, oldAction := range attachment.Actions {
			if oldAction.Name == "addFriend" {
				oldValues.decodeValues(oldAction.Value)
				if oldValues.User == action.User.ID {
					simpleResponse(action.ResponseURL, "Don't worry, you're already on the list", false, token)
					return
				}
			}
		}
	}

	if actionCount == 5 {
		orignalMessage.Attachments = append(orignalMessage.Attachments, slack.Attachment{CallbackID: "friendAction"})
		attachmentCount++
	}
	orignalMessage.Attachments[attachmentCount-1].Actions = append(orignalMessage.Attachments[attachmentCount-1].Actions, newButton)

	api := slack.New(token)

	responseParams := slack.NewResponseMessageParameters()
	responseParams.ResponseType = "in_channel"
	responseParams.ReplaceOriginal = true
	responseParams.Text = orignalMessage.Text
	responseParams.Attachments = orignalMessage.Attachments
	err = api.PostResponse(action.ResponseURL, responseParams)
	if err != nil {
		responseError(action.ResponseURL, err.Error(), token)
		return
	}
}

func addAllFriendButton(action action, token string) {
	log.Println("add all friends button!")

	simpleResponse(action.ResponseURL, "Sending a bunch of friend requests, this might take a while", false, token)
	auth, err := getGoodreadsAuth(goodreadsAuth{slackUserID: action.User.ID, teamID: action.Team.ID})
	if err != nil {
		if err.Error() == "User not found" {
			goodreadsAuthMessage(action, token, "You have to connect Booky to your Goodreads account to do that")
			return
		}
		responseError(action.ResponseURL, err.Error(), token)
		return
	}
	var orignalMessage slack.Message
	err = json.Unmarshal(action.OriginalMessage, &orignalMessage)
	if err != nil {
		responseError(action.ResponseURL, err.Error(), token)
		return
	}
	count := 0
	c := goodreads.NewClientWithToken(config.Goodreads.Key, config.Goodreads.Secret, auth.token, auth.secret)
	var values friendButtonValues
	for _, attachment := range orignalMessage.Attachments {
		for _, oldAction := range attachment.Actions {
			if oldAction.Name == "addFriend" {
				values.decodeValues(oldAction.Value)
				if values.User != action.User.ID {
					err = c.AddFriend(values.goodreadsID)
					if err != nil && err.Error() != "409 Conflict" {
						log.Println(err.Error())
						goodreadsAuthMessage(action, token, "Something went wrong, make sure your Goodreads account is connected to Booky")
						return
					}
					time.Sleep(2000)
					count++
				}
			}
		}
	}

	simpleResponse(action.ResponseURL, fmt.Sprintf("%v friend requests sent!", count), false, token)
}

func addFriendButton(action action, token string) {
	log.Println("add friend button!")
	auth, err := getGoodreadsAuth(goodreadsAuth{slackUserID: action.User.ID, teamID: action.Team.ID})
	if err != nil {
		if err.Error() == "User not found" {
			goodreadsAuthMessage(action, token, "You have to connect Booky to your Goodreads account to do that")
			return
		}
		responseError(action.ResponseURL, err.Error(), token)
		return
	}
	var values friendButtonValues
	err = values.decodeValues(action.Actions[0].Value)
	if action.User.ID == values.User {
		simpleResponse(action.ResponseURL, "It's great that you want to be your own friend!", false, token)
		return
	}
	time.Sleep(1000)
	c := goodreads.NewClientWithToken(config.Goodreads.Key, config.Goodreads.Secret, auth.token, auth.secret)
	err = c.AddFriend(values.goodreadsID)
	if err != nil {
		if err.Error() == "409 Conflict" {
			simpleResponse(action.ResponseURL, fmt.Sprintf("A friend request to @%s already exists, or you're already friends", values.UserName), false, token)
			return
		}
		if err.Error() == "404 Not Found" {
			time.Sleep(time.Duration(1000 + rand.Intn(1000)))
			err = c.AddFriend(values.goodreadsID)
		}
		if err != nil {
			simpleResponse(action.ResponseURL, fmt.Sprintf("Please try again later, the application may have exceeded the Goodreads rate limit", values.UserName), false, token)

		}
	}
	simpleResponse(action.ResponseURL, "Sending a friend request to @"+values.UserName, false, token)

}

type friendButtonValues struct {
	User        string `json:"user"`
	UserName    string `json:"user_name"`
	goodreadsID string `json:"goodreads_id"`
}

func (values *friendButtonValues) encodeValues() string {
	return fmt.Sprintf("%v|+|%v|+|%v", values.User, values.UserName, values.goodreadsID)
}
func (values *friendButtonValues) decodeValues(valueString string) (err error) {
	valueStrings := strings.Split(valueString, "|+|")
	if len(valueStrings) < 3 {
		err = errors.New("not enough values")
		return
	}
	values.User = valueStrings[0]
	values.UserName = valueStrings[1]
	values.goodreadsID = valueStrings[2]
	return
}
