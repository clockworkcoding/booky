package goverdrive

import (
	"encoding/json"
	"errors"
	"log"
)

func (c *Client) GetPatron() (patron Patron, err error) {

	resp, err := c.client.Get(c.getCirculationAPI() + "/v1/patrons/me")
	if err != nil {
		return
	}
	if resp.StatusCode != 200 {
		return patron, errors.New(resp.Status)
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&patron); err != nil {
		log.Println(err)
	}
	return
}

type Patron struct {
	PatronID        int    `json:"patronId"`
	WebsiteID       int    `json:"websiteId"`
	ExistingPatron  bool   `json:"existingPatron"`
	CollectionToken string `json:"collectionToken"`
	HoldLimit       int    `json:"holdLimit"`
	LastHoldEmail   string `json:"lastHoldEmail"`
	CheckoutLimit   int    `json:"checkoutLimit"`
	LendingPeriods  []struct {
		FormatType    string `json:"formatType"`
		LendingPeriod int    `json:"lendingPeriod"`
		Units         string `json:"units"`
	} `json:"lendingPeriods"`
	Links struct {
		Self struct {
			Href string `json:"href"`
			Type string `json:"type"`
		} `json:"self"`
		Checkouts struct {
			Href string `json:"href"`
			Type string `json:"type"`
		} `json:"checkouts"`
		Holds struct {
			Href string `json:"href"`
			Type string `json:"type"`
		} `json:"holds"`
	} `json:"links"`
	LinkTemplates struct {
		Search struct {
			Href string `json:"href"`
			Type string `json:"type"`
		} `json:"search"`
		Availability struct {
			Href string `json:"href"`
			Type string `json:"type"`
		} `json:"availability"`
	} `json:"linkTemplates"`
	Actions []struct {
		EditLendingPeriod struct {
			Href   string `json:"href"`
			Type   string `json:"type"`
			Method string `json:"method"`
			Fields []struct {
				Name     string   `json:"name"`
				Type     string   `json:"type"`
				Value    string   `json:"value,omitempty"`
				Optional bool     `json:"optional"`
				Options  []string `json:"options,omitempty"`
			} `json:"fields"`
		} `json:"editLendingPeriod"`
	} `json:"actions"`
}
