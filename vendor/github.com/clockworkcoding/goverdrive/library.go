package goverdrive

import (
	"encoding/json"
	"errors"
	"log"
)

func (c *Client) GetLibrary(accountID string) (library Library, err error) {

	resp, err := c.client.Get(c.getDiscoveryAPI() + "/v1/libraries/" + accountID)
	if err != nil {
		return
	}
	if resp.StatusCode != 200 {
		return library, errors.New(resp.Status)
	}

	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&library); err != nil {
		log.Println(err)
	}
	return
}

type Library struct {
	ID              int    `json:"id"`
	Name            string `json:"name"`
	Type            string `json:"type"`
	CollectionToken string `json:"collectionToken"`
	Links           struct {
		Self struct {
			Href string `json:"href"`
			Type string `json:"type"`
		} `json:"self"`
		Products struct {
			Href string `json:"href"`
			Type string `json:"type"`
		} `json:"products"`
		AdvantageAccounts struct {
			Href string `json:"href"`
			Type string `json:"type"`
		} `json:"advantageAccounts"`
		DlrHomepage struct {
			Href string `json:"href"`
			Type string `json:"type"`
		} `json:"dlrHomepage"`
	} `json:"links"`
	Formats []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"formats"`
}
