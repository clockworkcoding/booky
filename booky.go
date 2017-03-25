package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/clockworkcoding/goodreads"
	_ "github.com/lib/pq"
)

var (
	db     *sql.DB
	config Configuration
)

type state struct {
	auth string
	ts   time.Time
}

// globalState is an example of how to store a state between calls
var globalState state

// writeError writes an error to the reply - example only
func writeError(w http.ResponseWriter, status int, err string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(err))
}

// event responds to events from slack
func event(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	var eventMeta EventMeta
	err := decoder.Decode(&eventMeta)
	if err != nil {
		fmt.Println("ERR: " + err.Error())
		var challenge Challenge
		err = decoder.Decode(&challenge)
		if err != nil {
			fmt.Println("ERR: " + err.Error())
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return

	}
	fmt.Println(eventMeta.Event.Type)

	w.WriteHeader(http.StatusOK)

	decoder = json.NewDecoder(r.Body)
	switch eventMeta.Event.Type {
	case "message":
		var message EventMessage
		err = decoder.Decode(&message)
		if err != nil {
			fmt.Println("ERR: " + err.Error())
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		fmt.Println(message.Event.User, message.Event.Text)
	case "link_shared":
		var linkShared EventLinkShared
		err = decoder.Decode(&linkShared)
		if err != nil {
			fmt.Println("ERR: " + err.Error())
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		fmt.Println(linkShared.Event.Links[0].URL)
	}
}

type Configuration struct {
	Goodreads struct {
		Key    string `json:"Key"`
		Secret string `json:"Secret"`
	} `json:"Goodreads"`
	Slack struct {
		ClientID          string `json:"ClientID"`
		ClientSecret      string `json:"ClientSecret"`
		VerificationToken string `json:"VerificationToken"`
	} `json:"slack"`
	Db struct {
		URI string `json:"URI"`
	} `json:"db"`
}

func main() {
	gr := goodreads.NewClient(config.Goodreads.Key, config.Goodreads.Secret)

	results, err := gr.GetSearch("Collapsing Empire")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	_, err = gr.GetBook(results.Search_work[0].Search_best_book.Search_id.Text)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	db, err = sql.Open("postgres", config.Db.URI)
	if err != nil {
		log.Fatalf("Error opening database: %q", err)
	}

	//fmt.Println(book.Book_title[0].Text)
	//fmt.Println(book.Book_description.Text)

	routing()
	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), nil))
}

func routing() {
	http.HandleFunc("/dbfunc", dbFunc)
	http.HandleFunc("/add", addToSlack)
	http.HandleFunc("/auth", auth)
	http.HandleFunc("/event", event)
	http.HandleFunc("/", home)

}

func init() {
	config = readConfig()
}

func readConfig() Configuration {
	configuration := Configuration{}

	if configuration.Slack.ClientID = os.Getenv("SLACK_CLIENT_ID"); configuration.Slack.ClientID != "" {
		configuration.Slack.ClientSecret = os.Getenv("SLACK_CLIENT_SECRET")
		configuration.Goodreads.Secret = os.Getenv("GOODREADS_SECRET")
		configuration.Goodreads.Key = os.Getenv("GOODREADS_KEY")
		configuration.Db.URI = os.Getenv("DATABASE_URL")
		configuration.Slack.VerificationToken = os.Getenv("SLACK_VERIFICATION_TOKEN")
	} else {
		file, _ := os.Open("conf.json")
		decoder := json.NewDecoder(file)
		err := decoder.Decode(&configuration)
		if err != nil {
			fmt.Println("error:", err)
		}
	}
	return configuration
}
