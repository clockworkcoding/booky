package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type eventMessage struct {
	Token    string `json:"token"`
	TeamID   string `json:"team_id"`
	APIAppID string `json:"api_app_id"`
	Event    struct {
		Type    string `json:"type"`
		Channel string `json:"channel"`
		User    string `json:"user"`
		Text    string `json:"text"`
		Ts      string `json:"ts"`
	} `json:"event"`
	Type        string   `json:"type"`
	AuthedUsers []string `json:"authed_users"`
	EventID     string   `json:"event_id"`
	EventTime   int      `json:"event_time"`
}

type eventLinkShared struct {
	Token    string `json:"token"`
	TeamID   string `json:"team_id"`
	APIAppID string `json:"api_app_id"`
	Event    struct {
		Type      string `json:"type"`
		Channel   string `json:"channel"`
		User      string `json:"user"`
		MessageTs string `json:"message_ts"`
		Links     []struct {
			Domain string `json:"domain"`
			URL    string `json:"url"`
		} `json:"links"`
	} `json:"event"`
	Type        string   `json:"type"`
	AuthedUsers []string `json:"authed_users"`
	EventID     string   `json:"event_id"`
	EventTime   int      `json:"event_time"`
}

type challenge struct {
	Token     string `json:"token"`
	Challenge string `json:"challenge"`
	Type      string `json:"type"`
}

type buttonValues struct {
	User        string `json:"user"`
	UserName    string `json:"user_name"`
	Query       string `json:"query"`
	Index       int    `json:"index"`
	IsEphemeral bool   `json:"is_ephemeral"`
}

func (values *buttonValues) encodeValues() string {
	return fmt.Sprintf("%v|+|%v|+|%v|+|%v|+|%v", values.Index, values.IsEphemeral, values.Query, values.User, values.UserName)
}
func (values *buttonValues) decodeValues(valueString string) (err error) {
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

type action struct {
	Actions []struct {
		Name  string `json:"name"`
		Type  string `json:"type"`
		Value string `json:"value"`
	} `json:"actions"`
	CallbackID string `json:"callback_id"`
	Team       struct {
		ID     string `json:"id"`
		Domain string `json:"domain"`
	} `json:"team"`
	Channel struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"channel"`
	User struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"user"`
	ActionTs        string          `json:"action_ts"`
	MessageTs       string          `json:"message_ts"`
	AttachmentID    string          `json:"attachment_id"`
	Token           string          `json:"token"`
	IsAppUnfurl     bool            `json:"is_app_unfurl"`
	OriginalMessage json.RawMessage `json:"original_message"`
	ResponseURL     string          `json:"response_url"`
}
