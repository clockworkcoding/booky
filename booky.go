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

func bookyCommand(w http.ResponseWriter, r *http.Request) {
	queryText := r.FormValue("text")
	w.Write([]byte("Looking up your book using " + queryText + "..."))

}

// event responds to events from slack
func event(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	var v map[string]interface{}
	err := decoder.Decode(&v)
	if err != nil {
		fmt.Println("ERR: " + err.Error())
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if v["token"].(string) != config.Slack.VerificationToken {
		writeError(w, http.StatusForbidden, "Forbidden")
		return
	}
	w.WriteHeader(http.StatusOK)
	if v["type"] == "challenge" {
		w.Write([]byte(v["challenge"].(string)))
		return
	}
	event, err := json.Marshal(v)
	if err != nil {
		fmt.Println("ERR: " + err.Error())
		writeError(w, http.StatusInternalServerError, err.Error())
	}
	fmt.Println(v["event"].(map[string]interface{})["type"].(string))
	switch v["event"].(map[string]interface{})["type"].(string) {
	case "message":
		var message EventMessage
		err := json.Unmarshal(event, &message)
		if err != nil {
			fmt.Println("ERR: " + err.Error())
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		fmt.Println(message.Event.Text)
	case "link_shared":
		fmt.Println("It's a link!")
		var link EventLinkShared
		err := json.Unmarshal(event, &link)
		if err != nil {
			fmt.Println("ERR: " + err.Error())
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		fmt.Println(link.Event.Links[0].URL)
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
	http.HandleFunc("/booky", bookyCommand)
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
