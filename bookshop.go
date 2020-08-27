package main

import (
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode"
)

const bookshopURL = "https://bookshop.org/"

func getBookshopLink(isbn string, title string) (link string) {
  if config.BookshopID == "" || (isbn== "" && title == "") {
		return
	}
  

  log.Println(title)
  log.Println(isbn)
  log.Println(bookURL)
	// Build the request
  isbnURL := bookshopURL + "a/" + config.BookshopID + "/" + url.QueryEscape(isbn)
  eq, err := http.NewRequest("GET", bookURL, nil)
	if err != nil {
		log.Println("bit.ly error: ", err)
		return
	}
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("bookshop query: ", err)
		return
	}

	if resp.StatusCode != 200 {
    log.Println(resp.StatusCode)
		return
	}
	link = bookURL + "?aid=" + config.BookshopID

	titleURL := bookshopURL + "books/" + url.QueryEscape(formatTitle(title))
	req, err := http.NewRequest("GET", bookURL, nil)
	if err != nil {
		log.Println("bit.ly error: ", err)
		return
	}
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("bookshop query: ", err)
		return
	}

	if resp.StatusCode != 200 {
    log.Println(resp.StatusCode)
		return
	}
	link = bookURL + "?aid=" + config.BookshopID
	return
}

func formatTitle(title string) (linkTitle string) {
	var sb strings.Builder
	for _, r := range title {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			sb.WriteRune(r)
		} else {
			sb.WriteRune('-')
		}
    
	}
	return sb.String()
}
