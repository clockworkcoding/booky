package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/campoy/apiai"
	"github.com/clockworkcoding/goodreads"
)

func actionError(w http.ResponseWriter, err error) {
	fmt.Println(err)
	writeSpeech(w, "Something went wrong, please try again")
}

func writeSpeech(w http.ResponseWriter, speech string) {

	res := apiai.Response{
		Speech: speech,
	}
	b, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Fprintf(w, "%s", b)
}

func lookUpHandler(w http.ResponseWriter, r *http.Request) {

	var req apiai.Request

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "could not decode request: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.Result.Action == "lookup.lookup-more" {
		descriptionHandler(w, r, req)
		return
	}
	client, err := getRedisClient()
	if err != nil {
		fmt.Println("Redis connection", err)
	}
	defer client.Close()

	gr := goodreads.NewClient(config.Goodreads.Key, config.Goodreads.Secret)
	var bookID string

	if req.Param("title") != "" {
		results, err := gr.GetSearch(req.Param("title") + " " + req.Param("author-last") + " " + req.Param("author-given"))
		if err != nil {
			actionError(w, err)
			return
		}
		if len(results.Search_work) == 0 {
			writeSpeech(w, "Sorry, I couldn't find that book")
			return
		}

		bookID = results.Search_work[0].Search_best_book.Search_id.Text
	} else {
		writeSpeech(w, "I'm sorry, what book?")
		return
	}

	book, err := gr.GetBook(bookID)
	if err != nil {
		actionError(w, err)
		return
	}

	shortDescription := removeMarkup(book.Book_description.Text)
	if len(shortDescription) > 350 {
		shortDescription = shortDescription[0:strings.LastIndex(shortDescription[0:200], " ")]
	}

	var responseBuffer bytes.Buffer
	responseBuffer.WriteString(book.Book_title[0].Text)

	responseBuffer.WriteString(" by ")

	for i, author := range book.Book_authors[0].Book_author {
		if i == len(book.Book_authors)-1 && i > 0 {
			responseBuffer.WriteString(" and ")
		}
		if author.Book_role.Text != "" {
			responseBuffer.WriteString(author.Book_role.Text)
			responseBuffer.WriteString(" ")
		}
		responseBuffer.WriteString(author.Book_name.Text)
	}

	responseBuffer.WriteString(" has an average rating of ")

	responseBuffer.WriteString(book.Book_average_rating[0].Text)
	responseBuffer.WriteString(" and ")
	responseBuffer.WriteString(book.Book_text_reviews_count.Text)
	responseBuffer.WriteString(" reviews. ")
	responseBuffer.WriteString(shortDescription)
	responseBuffer.WriteString("\n Would you like to hear the full description, a review, or add this book to a shelf?")

	writeSpeech(w, responseBuffer.String())

	if err != nil {
		actionError(w, err)
		return
	}
	var serialBook, _ = json.Marshal(book)
	err = client.Set(req.SessionID, serialBook, time.Minute*10).Err()
	if err != nil {
		actionError(w, err)
		return
	}
}

func descriptionHandler(w http.ResponseWriter, r *http.Request, req apiai.Request) {
	client, err := getRedisClient()
	if err != nil {
		fmt.Println("Redis connection", err)
	}
	defer client.Close()

	val, err := client.Get(req.SessionID).Result()
	if err != nil {
		fmt.Println(err)
	}
	var book goodreads.Book_book
	err = json.Unmarshal([]byte(val), &book)
	if err != nil {
		fmt.Println(err)
	}
	writeSpeech(w, removeMarkup(book.Book_description.Text))

}
