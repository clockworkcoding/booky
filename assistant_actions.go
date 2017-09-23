package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/antonlindstrom/pgstore"
	"github.com/campoy/apiai"
	"github.com/clockworkcoding/goodreads"
)

func lookUpHandler(w http.ResponseWriter, r *http.Request) {

	var req apiai.Request

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "could not decode request: "+err.Error(), http.StatusBadRequest)
		return
	}

	log.Println("Entered look up handler", req)

	store, err := pgstore.NewPGStore(config.Db.URI, []byte(config.Keys.Key2))
	if err != nil {
		log.Fatalf(err.Error())
	}
	defer store.Close()

	// Run a background goroutine to clean up expired sessions from the database.
	defer store.StopCleanup(store.Cleanup(time.Minute * 5))

	// Get a session.
	session, err := store.Get(r, "whatever")
	if err != nil {
		log.Fatalf(err.Error())
	}
	log.Println("id: ", req.ID, " session id: ", req.SessionID, "data", req.OriginalRequest.Data)
	log.Println("foo: ", session.Values["foo"])
	// Add a value.
	session.Values["foo"] = "barbie"

	gr := goodreads.NewClient(config.Goodreads.Key, config.Goodreads.Secret)
	var bookId string

	if req.Param("title") != "" {
		results, err := gr.GetSearch(req.Param("title") + " " + req.Param("author-last") + " " + req.Param("author-given"))
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		if len(results.Search_work) == 0 {
			err = errors.New("no books found")
			return
		}

		bookId = results.Search_work[0].Search_best_book.Search_id.Text
	} else {
		fmt.Fprintf(w, "%s", "I'm sorry, what book?")
	}

	book, err := gr.GetBook(bookId)
	if err != nil {
		fmt.Println(err.Error())
		return
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
	responseBuffer.WriteString(" reviews. The description is ")
	responseBuffer.WriteString(removeMarkup(book.Book_description.Text))
	res := apiai.Response{
		Speech: responseBuffer.String(),
	}
	b, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		log.Println("could not encode response: " + err.Error())
		return
	}

	session.Options.MaxAge = 600
	if err = session.Save(r, w); err != nil {
		log.Fatalf("Error saving session: %v", err)
	}

	fmt.Fprintf(w, "%s", b)
}
