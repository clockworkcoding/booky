package goverdrive

import (
	"encoding/json"
	"errors"
	"log"
	"net/url"
	"strconv"
	"time"
)

type SearchParameters struct {
	Avilability         bool
	Formats             string
	Identifier          string
	CrossRefId          int
	LastTitleUpdateTime time.Time
	LastUpdateTime      time.Time
	Limit               int
	Minimum             bool
	Offset              int
	Query               string
	Series              string
	Sort                string
	SortAscending       bool
}

func NewSearchParamters() SearchParameters {
	return SearchParameters{
		SortAscending: true,
	}
}

func (c *Client) GetSearch(libraryProductHref string, params SearchParameters) (result SearchResult, err error) {
	var URL *url.URL
	URL, err = url.Parse(libraryProductHref)
	urlParameters := URL.Query()
	if params.Avilability {
		urlParameters.Add("availability", "true")
	}
	if params.Formats != "" {
		urlParameters.Add("formats", params.Formats)
	}
	if params.Identifier != "" {
		urlParameters.Add("identifier", params.Identifier)
	}
	if params.CrossRefId != 0 {
		urlParameters.Add("crossRefId", strconv.Itoa(params.CrossRefId))
	}
	var newTime time.Time
	if params.LastTitleUpdateTime != newTime {
		urlParameters.Add("lastTitleUpdateTime", params.LastTitleUpdateTime.Format(time.RFC3339))
	}
	if params.LastUpdateTime != newTime {
		urlParameters.Add("lastUpdateTime", params.LastUpdateTime.Format(time.RFC3339))
	}
	if params.Limit != 0 {
		urlParameters.Add("limit", strconv.Itoa(params.Limit))
	}
	if params.Minimum {
		urlParameters.Add("minimum", "true")
	}
	if params.Offset != 0 {
		urlParameters.Add("offset", strconv.Itoa(params.Offset))
	}
	if params.Query != "" {
		urlParameters.Add("q", params.Query)
	}
	if params.Series != "" {
		urlParameters.Add("series", params.Series)
	}
	if params.Sort != "" {
		if params.SortAscending {
			params.Sort += ":asc"
		} else {
			params.Sort += ":desc"
		}
	}
	URL.RawQuery = urlParameters.Encode()
	resp, err := c.client.Get(URL.String())
	if err != nil {
		return
	}
	if resp.StatusCode != 200 {
		return result, errors.New(resp.Status)
	}

	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Println(err)
	}
	return
}

type SearchResult struct {
	Limit      int    `json:"limit"`
	Offset     int    `json:"offset"`
	TotalItems int    `json:"totalItems"`
	ID         string `json:"id"`
	Products   []struct {
		ID             string `json:"id"`
		CrossRefID     int    `json:"crossRefId"`
		MediaType      string `json:"mediaType"`
		Title          string `json:"title"`
		Subtitle       string `json:"subtitle,omitempty"`
		SortTitle      string `json:"sortTitle"`
		PrimaryCreator struct {
			Role string `json:"role"`
			Name string `json:"name"`
		} `json:"primaryCreator,omitempty"`
		StarRating float64 `json:"starRating"`
		DateAdded  string  `json:"dateAdded"`
		Formats    []struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			Identifiers []struct {
				Type  string `json:"type"`
				Value string `json:"value"`
			} `json:"identifiers"`
		} `json:"formats"`
		Images struct {
			Thumbnail struct {
				Href string `json:"href"`
				Type string `json:"type"`
			} `json:"thumbnail"`
			Cover150Wide struct {
				Href string `json:"href"`
				Type string `json:"type"`
			} `json:"cover150Wide"`
			Cover struct {
				Href string `json:"href"`
				Type string `json:"type"`
			} `json:"cover"`
			Cover300Wide struct {
				Href string `json:"href"`
				Type string `json:"type"`
			} `json:"cover300Wide"`
		} `json:"images"`
		ContentDetails []struct {
			Href    string `json:"href"`
			Type    string `json:"type"`
			Account struct {
				ID   int    `json:"id"`
				Name string `json:"name"`
			} `json:"account"`
		} `json:"contentDetails"`
		Links struct {
			Self struct {
				Href string `json:"href"`
				Type string `json:"type"`
			} `json:"self"`
			Metadata struct {
				Href string `json:"href"`
				Type string `json:"type"`
			} `json:"metadata"`
			Availability struct {
				Href string `json:"href"`
				Type string `json:"type"`
			} `json:"availability"`
		} `json:"links"`
		Series                 string `json:"series,omitempty"`
		OtherFormatIdentifiers []struct {
			Type  string `json:"type"`
			Value string `json:"value"`
		} `json:"otherFormatIdentifiers,omitempty"`
	} `json:"products"`
	Links struct {
		Self struct {
			Href string `json:"href"`
			Type string `json:"type"`
		} `json:"self"`
		First struct {
			Href string `json:"href"`
			Type string `json:"type"`
		} `json:"first"`
		Prev struct {
			Href string `json:"href"`
			Type string `json:"type"`
		} `json:"prev"`
		Next struct {
			Href string `json:"href"`
			Type string `json:"type"`
		} `json:"next"`
		Last struct {
			Href string `json:"href"`
			Type string `json:"type"`
		} `json:"last"`
	} `json:"links"`
}
