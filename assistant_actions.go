package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/campoy/apiai"
	"github.com/clockworkcoding/goodreads"
	"github.com/go-redis/redis"
)

func lookUpHandler(w http.ResponseWriter, r *http.Request) {

	var req apiai.Request

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "could not decode request: "+err.Error(), http.StatusBadRequest)
		return
	}

	log.Println("Entered look up handler", req)

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

	var resolvedURL = os.Getenv("REDIS_URL")
	var password = ""
	if !strings.Contains(resolvedURL, "localhost") {
		parsedURL, _ := url.Parse(resolvedURL)
		password, _ = parsedURL.User.Password()
		resolvedURL = parsedURL.Host
	}

	client := redis.NewClient(&redis.Options{
		Addr:     resolvedURL,
		Password: password, // no password set
		DB:       0,        // use default DB
	})

	pong, err := client.Ping().Result()
	fmt.Println(pong, err)
	// Output: PONG <nil>

	err = client.Set("key", "value", 0).Err()
	if err != nil {
		fmt.Println(err)
	}

	val, err := client.Get("key").Result()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("key", val)

	val2, err := client.Get("key2").Result()
	if err == redis.Nil {
		fmt.Println("key2 does not exists")
	} else if err != nil {
		panic(err)
	} else {
		fmt.Println("key2", val2)
	}
	fmt.Fprintf(w, "%s", b)
}
