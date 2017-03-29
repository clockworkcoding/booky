package slack

import (
	"bytes"
	"encoding/json"
	"net/http"
)

const DEFAULT_RESPONSE_REPLACE_ORIGINAL = false

// ResponseMessageParameters contains all the parameters necessary (including the optional ones) for a PostResponse() request
type ResponseMessageParameters struct {
	Text            string       `json:"text"`
	ReplaceOriginal bool         `json:"replace_original"`
	Attachments     []Attachment `json:"attachments"`
	ResponseType    string       `json:"response_type"`
}

// NewResponseMessageParameters provides an instance of ResponseMessageParameters with all the sane default values set
func NewResponseMessageParameters() ResponseMessageParameters {
	return ResponseMessageParameters{
		Attachments:     nil,
		ResponseType:    DEFAULT_MESSAGE_RESPONSE_TYPE,
		ReplaceOriginal: DEFAULT_RESPONSE_REPLACE_ORIGINAL,
	}
}

// PostResponse sends a message in response to a Slash Command or Action.
func (api *Client) PostResponse(responseUrl, text string, params ResponseMessageParameters) error {
	jsonStr, err := json.Marshal(params)
	req, err := http.NewRequest("POST", responseUrl, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	_, err = client.Do(req)
	if err != nil {
		panic(err)
	}
	return err
}
