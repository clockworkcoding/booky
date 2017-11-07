package goverdrive

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

type CheckoutParam struct {
	Fields []Field `json:"fields"`
}

type Field struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func (c *Client) CheckoutTitle(reserveID, format string) (result CheckoutResponse, err error) {
	log.Println("entered checkout")
	if reserveID == "" {
		err = errors.New("reserveID is required for checkout")
		return
	}

	log.Println("checked reserveID")
	body := CheckoutParam{
		Fields: []Field{
			Field{
				Name:  "reserveId",
				Value: reserveID,
			},
		},
	}
	log.Println("created body")
	if format != "" {
		body.Fields = append(body.Fields, Field{Name: "format", Value: format})
	}

	log.Println("checked format")
	jsonBody, err := json.Marshal(body)
	if err != nil {
		log.Println("Json marshal error: " + err.Error())
	}
	log.Println("Marshalled JSON: " + string(jsonBody))
	// Build the request
	req, err := http.NewRequest("POST", c.getCirculationAPI()+"/v1/patrons/me/checkouts", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		log.Println("NewRequest: " + err.Error())
		return
	}

	log.Println("Built Request")
	resp, err := c.client.Do(req)
	if err != nil {
		log.Println("Do: " + err.Error())
		return
	}
	log.Println("statuscode: " + resp.Status)
	if resp.StatusCode != 201 {
		err = errors.New("Error code: " + strconv.Itoa(resp.StatusCode) + " " + resp.Status)
		return
	}

	defer resp.Body.Close()
	response, _ := ioutil.ReadAll(resp.Body)
	resp.Body = ioutil.NopCloser(bytes.NewBuffer(response))
	log.Println("response: ", string(response))

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Println("decode: " + err.Error())
	}
	return
}

type CheckoutResponse struct {
	ReserveID        string `json:"reserveId"`
	Expires          string `json:"expires"`
	IsFormatLockedIn bool   `json:"isFormatLockedIn"`
	Formats          []struct {
		ReserveID  string `json:"reserveId"`
		FormatType string `json:"formatType"`
		Links      struct {
			Self struct {
				Href string `json:"href"`
				Type string `json:"type"`
			} `json:"self"`
		} `json:"links"`
		LinkTemplates struct {
			DownloadLink struct {
				Href string `json:"href"`
				Type string `json:"type"`
			} `json:"downloadLink"`
			DownloadLinkV2 struct {
				Href string `json:"href"`
				Type string `json:"type"`
			} `json:"downloadLinkV2"`
		} `json:"linkTemplates"`
	} `json:"formats"`
	Links struct {
		Self struct {
			Href string `json:"href"`
			Type string `json:"type"`
		} `json:"self"`
		Metadata struct {
			Href string `json:"href"`
			Type string `json:"type"`
		} `json:"metadata"`
	} `json:"links"`
	Actions struct {
		Format struct {
			Href   string `json:"href"`
			Type   string `json:"type"`
			Method string `json:"method"`
			Fields []struct {
				Name     string   `json:"name"`
				Value    string   `json:"value"`
				Optional bool     `json:"optional"`
				Options  []string `json:"options,omitempty"`
			} `json:"fields"`
		} `json:"format"`
		EarlyReturn struct {
			Href   string `json:"href"`
			Method string `json:"method"`
		} `json:"earlyReturn"`
	} `json:"actions"`
}
