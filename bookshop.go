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
  if config.BookshopID == "" {
		return
	}
  
  if isbn != ""{
    link = getIsbnLink(isbn)
  }
  if link == "" && title != "" {
    link = getTitleLink(title)
  }
 
	return
}

func getIsbnLink(isbn string) (link string){
  testIsbnURL := bookshopURL + "a/0/" + url.QueryEscape(isbn)
  isbnURL := bookshopURL + "a/" + config.BookshopID + "/" + url.QueryEscape(isbn)
  req, err := http.NewRequest("GET", testIsbnURL, nil)
  if err != nil {
    log.Println("bookshop error : ", err)
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
  link = isbnURL
  return
}

func getTitleLink(title string) (link string){
  
	titleURL := bookshopURL + "books/" + url.QueryEscape(formatTitle(title))
	req, err := http.NewRequest("GET", titleURL, nil)
	if err != nil {
		log.Println("bookshop error: ", err)
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
	link = titleURL + "?aid=" + config.BookshopID
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
