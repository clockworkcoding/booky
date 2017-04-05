package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/clockworkcoding/slack"
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
	fmt.Printf("Err: %s", err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(err))
	log.Fatal(fmt.Sprintf("Err: %s", err))
}

func responseError(error, responseURL, token string) {
	log.Output(1, fmt.Sprintf("Err: %s", error))
	params := slack.NewResponseMessageParameters()
	params.ResponseType = "ephemeral"
	params.ReplaceOriginal = false
	params.Text = fmt.Sprintf("An error occured: %s", error)

	api := slack.New(token)
	err := api.PostResponse(responseURL, params.Text, params)
	if err != nil {
		//something is very wrong if we ever get here

		log.Fatal(fmt.Sprintf("Err: %s", err.Error))
	}

}

func buttonPressed(w http.ResponseWriter, r *http.Request) {
	if r.FormValue("ssl_check") == "1" {
		w.Write([]byte("OK"))
		return
	}
	var action action
	payload := r.FormValue("payload")

	err := json.Unmarshal([]byte(payload), &action)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if action.Token == config.Slack.VerificationToken {
		w.WriteHeader(http.StatusOK)
	} else {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	_, token, _, err := getSlackAuth(action.Team.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	switch action.CallbackID {
	case "wrongbook":
		wrongBookButton(w, action, token)
	case "goodreads":
		goodreadsButton(w, action, token)
	}

}

func bookyCommand(w http.ResponseWriter, r *http.Request) {
	queryText := r.FormValue("text")
	teamID := r.FormValue("team_id")
	userID := r.FormValue("user_id")
	userName := r.FormValue("user_name")
	responseURL := r.FormValue("response_url")
	_, token, _, err := getSlackAuth(teamID)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	w.Write([]byte("Looking up your book"))

	values := wrongBookButtonValues{
		Index:       0,
		User:        userID,
		Query:       queryText,
		IsEphemeral: true,
		UserName:    userName,
	}

	params, err := createBookPost(values, true)
	if err != nil {
		if err.Error() == "no books found" {
			w.Write([]byte("No books found, try a broader search"))
		} else {
			responseError(responseURL, err.Error(), token)
		}
		return
	}
	responseParams := slack.NewResponseMessageParameters()
	responseParams.ResponseType = "ephemeral"
	responseParams.ReplaceOriginal = true
	responseParams.Text = params.Text
	responseParams.Attachments = params.Attachments
	api := slack.New(token)
	err = api.PostResponse(responseURL, responseParams.Text, responseParams)
	if err != nil {
		responseError(responseURL, err.Error(), token)
	}
}

func checkTextForBook(message eventMessage) {
	tokenized := strings.Split(message.Event.Text, "_")
	if len(tokenized) < 2 {
		return
	}
	queryText := tokenized[1]
	channel := message.Event.Channel
	teamID := message.TeamID
	_, token, authedChannel, err := getSlackAuth(teamID)
	if err != nil || channel != authedChannel {
		return
	}
	values := wrongBookButtonValues{
		User:        message.Event.User,
		Query:       queryText,
		Index:       0,
		IsEphemeral: false,
		UserName:    "booky user",
	}
	params, err := createBookPost(values, true)
	if err != nil {
		return
	}
	api := slack.New(token)
	params.AsUser = false
	_, _, err = api.PostMessage(channel, params.Text, params)
	if err != nil {
		fmt.Printf("Error posting: %s\n", err.Error())
		return
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
	URL string `json:"URL"`
}

func main() {
	var err error
	db, err = sql.Open("postgres", config.Db.URI)
	if err != nil {
		log.Fatalf("Error opening database: %q", err)
	}

	routing()
}

func routing() {

	mux := http.NewServeMux()

	mux.Handle("/add", http.HandlerFunc(addToSlack))
	mux.Handle("/auth", http.HandlerFunc(slackAuth))
	mux.Handle("/gradd", http.HandlerFunc(addToGoodreads))
	mux.Handle("/grauth", http.HandlerFunc(goodreadsAuthCallback))
	mux.Handle("/event", http.HandlerFunc(event))
	mux.Handle("/booky", http.HandlerFunc(bookyCommand))
	mux.Handle("/button", http.HandlerFunc(buttonPressed))
	mux.Handle("/", http.HandlerFunc(home))
	err := http.ListenAndServe(":"+os.Getenv("PORT"), mux)
	if err != nil {
		log.Fatal("ListenAndServe error: ", err)
	}

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
		configuration.URL = os.Getenv("URL")
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
