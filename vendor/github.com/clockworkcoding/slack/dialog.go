package slack

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type DialogPost struct {
	Token     string `json:"token"`
	TriggerID string `json:"trigger_id"`
	Dialog    Dialog `json:"dialog"`
}

//Dialog object sent to Slack
type Dialog struct {
	CallbackID     string           `json:"callback_id"`
	Title          string           `json:"title"`
	SubmitLabel    string           `json:"submit_label"`
	NotifyOnCancel bool             `json:"notify_on_cancel"`
	Elements       []DialogElements `json:"elements"`
}

//Elements in dialog
type DialogElements struct {
	Type            string                        `json:"type"`
	Label           string                        `json:"label"`
	Name            string                        `json:"name"`
	Subtype         string                        `json:"subtype,omitempty"`
	MaxLength       int                           `json:"max_length,omitempty"`
	MinLength       int                           `json:"min_length,omitempty"`
	Optional        bool                          `json:"optional,omitempty"`
	Hint            string                        `json:"hint,omitempty"`
	Value           string                        `json:"value,omitempty"`
	Placeholder     string                        `json:"placeholder,omitempty"`
	Options         []AttachmentActionOption      `json:"options,omitempty"`
	OptionGroups    []AttachmentActionOptionGroup `json:"option_groups,omitempty"`
	DataSource      string                        `json:"data_source,omitempty"`
	MinQueryLength  int                           `json:"min_query_length,omitempty"`
	SelectedOptions []AttachmentActionOption      `json:"selected_options,omitempty"`
}

// PostDialog sends a dialog in response to a Slash Command or Action.
func (api *Client) PostDialog(triggerID string, token string, dialog Dialog) error {
	post := DialogPost{TriggerID: triggerID, Token: token, Dialog: dialog}
	jsonStr, err := json.Marshal(post)
	req, err := http.NewRequest("POST", SLACK_API+"dialog.open", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	bodyString := string(bodyBytes)
	logger.Println(bodyString)
	return err
}
