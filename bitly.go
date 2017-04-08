package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
)

const bitlyAPIRoot = "https://api-ssl.bitly.com"

//shortenURl returns the original url if there's an error
func shortenURl(longURL string) (shortURL string) {
	if config.BitlyKey == "" || config.BitlyKey == "{your key here}" {
		return longURL
	}
	safeURL := url.QueryEscape(longURL)

	apiURL := bitlyAPIRoot + fmt.Sprintf("/v3/shorten?access_token=%s&longUrl=%s", config.BitlyKey, safeURL)

	// Build the request
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		log.Println("bit.ly error: ", err)
		return longURL
	}
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Println("Do: ", err)
		return longURL
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return longURL
	}

	var response bitly
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Println(err)
		return longURL
	}

	return response.Data.URL
}

type bitly struct {
	StatusCode int    `json:"status_code"`
	StatusTxt  string `json:"status_txt"`
	Data       struct {
		URL        string `json:"url"`
		Hash       string `json:"hash"`
		GlobalHash string `json:"global_hash"`
		LongURL    string `json:"long_url"`
		NewHash    int    `json:"new_hash"`
	} `json:"data"`
}
