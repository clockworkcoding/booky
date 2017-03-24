package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/clockworkcoding/goodreads"
	"github.com/demisto/slack"
	"golang.org/x/oauth2"
	"log"
	"net/http"
	"os"
	"time"
)

var (
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

// addToSlack initializes the oauth process and redirects to Slack
func addToSlack(w http.ResponseWriter, r *http.Request) {
	// Just generate random state
	b := make([]byte, 10)
	_, err := rand.Read(b)
	if err != nil {
		writeError(w, 500, err.Error())
	}
	globalState = state{auth: hex.EncodeToString(b), ts: time.Now()}
	conf := &oauth2.Config{
		ClientID:     config.Slack.ClientID,
		ClientSecret: config.Slack.ClientSecret,
		Scopes:       []string{"client"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://slack.com/oauth/authorize",
			TokenURL: "https://slack.com/api/oauth.access", // not actually used here
		},
	}
	url := conf.AuthCodeURL(globalState.auth)
	http.Redirect(w, r, url, http.StatusFound)
}

// auth receives the callback from Slack, validates and displays the user information
func auth(w http.ResponseWriter, r *http.Request) {
	state := r.FormValue("state")
	code := r.FormValue("code")
	errStr := r.FormValue("error")
	if errStr != "" {
		writeError(w, 401, errStr)
		return
	}
	if state == "" || code == "" {
		writeError(w, 400, "Missing state or code")
		return
	}
	if state != globalState.auth {
		writeError(w, 403, "State does not match")
		return
	}
	// As an example, we allow only 5 min between requests
	if time.Since(globalState.ts) > 5*time.Minute {
		writeError(w, 403, "State is too old")
		return
	}
	token, err := slack.OAuthAccess(config.Slack.ClientID, config.Slack.ClientSecret, code, "")
	if err != nil {
		writeError(w, 401, err.Error())
		return
	}
	s, err := slack.New(slack.SetToken(token.AccessToken))
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	// Get our own user id
	test, err := s.AuthTest()
	if err != nil {
		writeError(w, 500, err.Error())
		return
	}
	w.Write([]byte(fmt.Sprintf("OAuth successful for team %s and user %s", test.Team, test.User)))
}

// home displays the add-to-slack button
func home(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`<html><head><title>Slack OAuth Test</title></head><body><a href="/add">Add To Slack</a></body></html>`))
}

type Configuration struct {
	Goodreads struct {
		Key    string `json:"Key"`
		Secret string `json:"Secret"`
	} `json:"Goodreads"`
	Slack struct {
		ClientID     string `json:"ClientID"`
		ClientSecret string `json:"ClientSecret"`
	} `json:"slack"`
	Db struct {
		Host     string `json:"Host"`
		Database string `json:"Database"`
		User     string `json:"User"`
		Port     string `json:"Port"`
		Password string `json:"Password"`
		URI      string `json:"URI"`
	} `json:"db"`
}

func main() {
	gr := goodreads.NewClient(config.Goodreads.Key, config.Goodreads.Secret)
	results, err := gr.GetSearch("Collapsing Empire")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	book, err := gr.GetBook(results.Search_work[0].Search_best_book.Search_id.Text)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println(book.Book_title[0].Text)
	fmt.Println(book.Book_description.Text)

	http.HandleFunc("/add", addToSlack)
	http.HandleFunc("/auth", auth)
	http.HandleFunc("/", home)
	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), nil))

}

func init() {
	config = readConfig()
}

func readConfig() Configuration {
	configuration := Configuration{}

	if configuration.Slack.ClientID = os.Getenv("slackClientID"); configuration.Slack.ClientID != "" {
		configuration.Slack.ClientSecret = os.Getenv("slackClientSecret")
		configuration.Goodreads.Secret = os.Getenv("goodReadsSecret")
		configuration.Goodreads.Key = os.Getenv("goodReadsKey")
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
