package slack

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strings"
)

const (
	DEFAULT_MESSAGE_USERNAME         = ""
	DEFAULT_MESSAGE_THREAD_TIMESTAMP = ""
	DEFAULT_MESSAGE_ASUSER           = false
	DEFAULT_MESSAGE_PARSE            = ""
	DEFAULT_MESSAGE_LINK_NAMES       = 0
	DEFAULT_MESSAGE_UNFURL_LINKS     = false
	DEFAULT_MESSAGE_UNFURL_MEDIA     = true
	DEFAULT_MESSAGE_ICON_URL         = ""
	DEFAULT_MESSAGE_ICON_EMOJI       = ""
	DEFAULT_MESSAGE_MARKDOWN         = true
	DEFAULT_MESSAGE_ESCAPE_TEXT      = true
	DEFAULT_MESSAGE_RESPONSE_TYPE    = "in_channel"
)

type chatResponseFull struct {
	Channel   string `json:"channel"`
	Timestamp string `json:"ts"`
	Text      string `json:"text"`
	SlackResponse
}

// PostMessageParameters contains all the parameters necessary (including the optional ones) for a PostMessage() request
type PostMessageParameters struct {
	Text            string       `json:"text"`
	Username        string       `json:"user_name"`
	AsUser          bool         `json:"as_user"`
	Parse           string       `json:"parse"`
	ThreadTimestamp string       `json:"thread_ts"`
	LinkNames       int          `json:"link_names"`
	Attachments     []Attachment `json:"attachments"`
	UnfurlLinks     bool         `json:"unfurl_links"`
	UnfurlMedia     bool         `json:"unfurl_media"`
	IconURL         string       `json:"icon_url"`
	IconEmoji       string       `json:"icon_emoji"`
	Markdown        bool         `json:"mrkdwn,omitempty"`
	EscapeText      bool         `json:"escape_text"`
}

// UpdateMessageParameters contains all the parameters necessary (including the optional ones) for a UpdateMessage() request
type UpdateMessageParameters struct {
	Timestamp   string       `json:"ts"`
	Text        string       `json:"text"`
	Attachments []Attachment `json:"attachments"`
	Parse       string       `json:"parse"`
	LinkNames   int          `json:"link_names"`
	AsUser      bool         `json:"as_user"`
}

//UnfurlParmaeters contains all the parameters necessary (including the optional ones) for a ChatUnfurl() request
type UnfurlParameters struct {
	Timestamp        string   `json:"ts"`
	Unfurls          []Unfurl `json:"-"`
	UserAuthRequired bool     `json:"user_auth_required"`
}

type Unfurl struct {
	Attachment Attachment `json:"Attachment"`
	UnfurlURL  string     `json:"-"`
}

// NewPostMessageParameters provides an instance of PostMessageParameters with all the sane default values set
func NewPostMessageParameters() PostMessageParameters {
	return PostMessageParameters{
		Username:    DEFAULT_MESSAGE_USERNAME,
		AsUser:      DEFAULT_MESSAGE_ASUSER,
		Parse:       DEFAULT_MESSAGE_PARSE,
		LinkNames:   DEFAULT_MESSAGE_LINK_NAMES,
		Attachments: nil,
		UnfurlLinks: DEFAULT_MESSAGE_UNFURL_LINKS,
		UnfurlMedia: DEFAULT_MESSAGE_UNFURL_MEDIA,
		IconURL:     DEFAULT_MESSAGE_ICON_URL,
		IconEmoji:   DEFAULT_MESSAGE_ICON_EMOJI,
		Markdown:    DEFAULT_MESSAGE_MARKDOWN,
		EscapeText:  DEFAULT_MESSAGE_ESCAPE_TEXT,
	}
}

func (b UnfurlParameters) PrintFields() {
	val := reflect.ValueOf(b)
	for i := 0; i < val.Type().NumField(); i++ {
		fmt.Println(val.Type().Field(i).Tag.Get("json"))
	}
}

func chatRequest(path string, values url.Values, debug bool) (*chatResponseFull, error) {
	response := &chatResponseFull{}
	err := post(path, values, response, debug)
	if err != nil {
		return nil, err
	}
	if !response.Ok {
		return nil, errors.New(response.Error)
	}
	return response, nil
}

// DeleteMessage deletes a message in a channel
func (api *Client) DeleteMessage(channel, messageTimestamp string) (string, string, error) {
	values := url.Values{
		"token":   {api.config.token},
		"channel": {channel},
		"ts":      {messageTimestamp},
	}
	response, err := chatRequest("chat.delete", values, api.debug)
	if err != nil {
		return "", "", err
	}
	return response.Channel, response.Timestamp, nil
}

func escapeMessage(message string) string {
	replacer := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")
	return replacer.Replace(message)
}

// PostMessage sends a message to a channel.
// Message is escaped by default according to https://api.slack.com/docs/formatting
// Use http://davestevens.github.io/slack-message-builder/ to help crafting your message.
func (api *Client) PostMessage(channel, text string, params PostMessageParameters) (string, string, error) {
	if params.EscapeText {
		text = escapeMessage(text)
	}
	values := url.Values{
		"token":   {api.config.token},
		"channel": {channel},
		"text":    {text},
	}
	if params.Username != DEFAULT_MESSAGE_USERNAME {
		values.Set("username", string(params.Username))
	}
	if params.AsUser != DEFAULT_MESSAGE_ASUSER {
		values.Set("as_user", "true")
	}
	if params.Parse != DEFAULT_MESSAGE_PARSE {
		values.Set("parse", string(params.Parse))
	}
	if params.LinkNames != DEFAULT_MESSAGE_LINK_NAMES {
		values.Set("link_names", "1")
	}
	if params.Attachments != nil {
		attachments, err := json.Marshal(params.Attachments)
		if err != nil {
			return "", "", err
		}
		values.Set("attachments", string(attachments))
	}
	if params.UnfurlLinks != DEFAULT_MESSAGE_UNFURL_LINKS {
		values.Set("unfurl_links", "true")
	}
	// I want to send a message with explicit `as_user` `true` and `unfurl_links` `false` in request.
	// Because setting `as_user` to `true` will change the default value for `unfurl_links` to `true` on Slack API side.
	if params.AsUser != DEFAULT_MESSAGE_ASUSER && params.UnfurlLinks == DEFAULT_MESSAGE_UNFURL_LINKS {
		values.Set("unfurl_links", "false")
	}
	if params.UnfurlMedia != DEFAULT_MESSAGE_UNFURL_MEDIA {
		values.Set("unfurl_media", "false")
	}
	if params.IconURL != DEFAULT_MESSAGE_ICON_URL {
		values.Set("icon_url", params.IconURL)
	}
	if params.IconEmoji != DEFAULT_MESSAGE_ICON_EMOJI {
		values.Set("icon_emoji", params.IconEmoji)
	}
	if params.Markdown != DEFAULT_MESSAGE_MARKDOWN {
		values.Set("mrkdwn", "false")
	}
	if params.ThreadTimestamp != DEFAULT_MESSAGE_THREAD_TIMESTAMP {
		values.Set("thread_ts", params.ThreadTimestamp)
	}

	response, err := chatRequest("chat.postMessage", values, api.debug)
	if err != nil {
		return "", "", err
	}
	return response.Channel, response.Timestamp, nil
}

// UpdateMessage updates a message in a channel
func (api *Client) UpdateMessage(channel, timestamp, text string) (string, string, string, error) {
	values := url.Values{
		"token":   {api.config.token},
		"channel": {channel},
		"text":    {escapeMessage(text)},
		"ts":      {timestamp},
	}
	response, err := chatRequest("chat.update", values, api.debug)
	if err != nil {
		return "", "", "", err
	}
	return response.Channel, response.Timestamp, response.Text, nil
}

// UpdateMessageWithAttachments updates a message in a channel with attachments
func (api *Client) UpdateMessageWithAttachments(channel string, params UpdateMessageParameters) (string, string, string, error) {
	values := url.Values{
		"token":   {api.config.token},
		"channel": {channel},
		"text":    {escapeMessage(params.Text)},
		"ts":      {params.Timestamp},
	}
	if params.AsUser != DEFAULT_MESSAGE_ASUSER {
		values.Set("as_user", "true")
	}
	if params.Parse != DEFAULT_MESSAGE_PARSE {
		values.Set("parse", string(params.Parse))
	}
	if params.LinkNames != DEFAULT_MESSAGE_LINK_NAMES {
		values.Set("link_names", "1")
	}
	if params.Attachments != nil {
		attachments, err := json.Marshal(params.Attachments)
		if err != nil {
			return "", "", "", err
		}
		values.Set("attachments", string(attachments))
	}
	response, err := chatRequest("chat.update", values, api.debug)
	if err != nil {
		return "", "", "", err
	}
	return response.Channel, response.Timestamp, response.Text, nil
}

// Unfurl unfurls links posted in channels
func (api *Client) Unfurl(channel string, params UnfurlParameters) (string, error) {
	values := url.Values{
		"token":   {api.config.token},
		"channel": {channel},
		"ts":      {params.Timestamp},
	}

	if params.UserAuthRequired != false {
		values.Set("user_auth_required ", "true")
	}
	exportUnfurls := make([]json.RawMessage, len(params.Unfurls))
	for i, unfurl := range params.Unfurls {

		buf, _ := json.Marshal(unfurl)
		str := strings.Replace(string(buf), "Attachment", unfurl.UnfurlURL, 1)

		exportUnfurls[i] = json.RawMessage(str)
		fmt.Printf("exportunfurl: %v\n", exportUnfurls[i])
	}

	if len(params.Unfurls) > 1 {
		unfurls, err := json.Marshal(exportUnfurls)
		if err != nil {
			fmt.Printf("Error!: %s", err.Error())
			return "", err
		}
		fmt.Printf("Many unfurls %s\n", string(unfurls))
		values.Set("unfurls", string(unfurls))
	} else {
		values.Set("unfurls", string(exportUnfurls[0]))
	}
	response, err := chatRequest("chat.unfurl", values, api.debug)
	if err != nil {
		fmt.Printf("Something went wrong: %s\n", err.Error())
		return "", err
	}
	fmt.Printf("Response: %v\n", response.Ok)
	return "ok", nil
}
