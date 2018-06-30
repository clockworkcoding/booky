package slack

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type DialogPost struct {
	TriggerID string `json:"trigger_id"`
	Dialog    Dialog `json:"dialog"`
}

//Dialog object sent to Slack
type Dialog struct {
	CallbackID     string          `json:"callback_id"`
	Title          string          `json:"title"`
	SubmitLabel    string          `json:"submit_label"`
	NotifyOnCancel bool            `json:"notify_on_cancel"`
	Elements       []DialogElement `json:"elements"`
}

//Elements in dialog
type DialogElement struct {
	Type            string              `json:"type"`
	Label           string              `json:"label"`
	Name            string              `json:"name"`
	Subtype         string              `json:"subtype,omitempty"`
	MaxLength       int                 `json:"max_length,omitempty"`
	MinLength       int                 `json:"min_length,omitempty"`
	Optional        bool                `json:"optional,omitempty"`
	Hint            string              `json:"hint,omitempty"`
	Value           string              `json:"value,omitempty"`
	Placeholder     string              `json:"placeholder,omitempty"`
	Options         []DialogOption      `json:"options,omitempty"`
	OptionGroups    []DialogOptionGroup `json:"option_groups,omitempty"`
	DataSource      string              `json:"data_source,omitempty"`
	MinQueryLength  int                 `json:"min_query_length,omitempty"`
	SelectedOptions []DialogOption      `json:"selected_options,omitempty"`
}

type DialogOption struct {
	Label string `json:"label"` // Required.
	Value string `json:"value"` // Required.
}

// AttachmentActionOptionGroup is a semi-hierarchal way to list available options to appear in action menu.
type DialogOptionGroup struct {
	Label   string         `json:"label"`   // Required.
	Options []DialogOption `json:"options"` // Required.
}

// PostDialog sends a dialog in response to a Slash Command or Action.
func (api *Client) PostDialog(triggerID string, token string, dialog Dialog) error {
	post := DialogPost{TriggerID: triggerID, Dialog: dialog}
	jsonStr, err := json.Marshal(post)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", SLACK_API+"dialog.open", bytes.NewBuffer(jsonStr))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+token)
	client := &http.Client{}
	_, err = client.Do(req)
	if err != nil {
		return err
	}
	return err
}
