package goverdrive

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
)

func (c *Client) GetAvailability(availabilityURL string) (result Availability, err error) {

	resp, err := c.client.Get(availabilityURL)
	if err != nil {
		return
	}
	if resp.StatusCode != 200 {
		return result, errors.New(resp.Status)
	}

	defer resp.Body.Close()
	response, _ := ioutil.ReadAll(resp.Body)
	resp.Body = ioutil.NopCloser(bytes.NewBuffer(response))
	log.Println(string(response))

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Println(err)
	}
	return
}

type Availability struct {
	ReserveID string `json:"reserveId"`
	Links     struct {
		Self struct {
			Href string `json:"href"`
			Type string `json:"type"`
		} `json:"self"`
	} `json:"links"`
	Accounts []struct {
		ID              int `json:"id"`
		CopiesOwned     int `json:"copiesOwned"`
		CopiesAvailable int `json:"copiesAvailable"`
	} `json:"accounts"`
	Available        bool   `json:"available"`
	AvailabilityType string `json:"availabilityType"`
	CopiesOwned      int    `json:"copiesOwned"`
	CopiesAvailable  int    `json:"copiesAvailable"`
	NumberOfHolds    int    `json:"numberOfHolds"`
}
