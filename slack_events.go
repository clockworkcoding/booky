package main

import (
	"encoding/json"
)

type eventMessage struct {
	Token    string `json:"token"`
	TeamID   string `json:"team_id"`
	APIAppID string `json:"api_app_id"`
	Event    struct {
		SubType string `json:"subtype"`
		Type    string `json:"type"`
		Channel string `json:"channel"`
		User    string `json:"user"`
		Text    string `json:"text"`
		Ts      string `json:"ts"`
		EventTs string `json:"event_ts"`
	} `json:"event"`
	Type        string          `json:"type"`
	AuthedUsers []string        `json:"authed_users"`
	EventID     string          `json:"event_id"`
	EventTime   json.RawMessage `json:"event_time"`
	TriggerID   string          `json:"trigger_id"`
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
	Type        string          `json:"type"`
	AuthedUsers []string        `json:"authed_users"`
	EventID     string          `json:"event_id"`
	EventTime   json.RawMessage `json:"event_time"`
}

type challenge struct {
	Token     string `json:"token"`
	Challenge string `json:"challenge"`
	Type      string `json:"type"`
}

type action struct {
	Actions []struct {
		Name            string `json:"name"`
		Type            string `json:"type"`
		Value           string `json:"value"`
		SelectedOptions []struct {
			Value string `json:"value"`
		} `json:"selected_options"`
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
	Message struct {
		Type string `json:"type"`
		User string `json:"user"`
		Ts   string `json:"ts"`
		Text string `json:"text"`
	} `json:"message"`

	ActionTs        string          `json:"action_ts"`
	MessageTs       string          `json:"message_ts"`
	AttachmentID    string          `json:"attachment_id"`
	Token           string          `json:"token"`
	IsAppUnfurl     bool            `json:"is_app_unfurl"`
	OriginalMessage json.RawMessage `json:"original_message"`
	ResponseURL     string          `json:"response_url"`
	TriggerID       string          `json:"trigger_id"`
}
