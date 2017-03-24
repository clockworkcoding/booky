package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"flag"
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
	config       Configuration
	address      = flag.String("address", ":8080", "Which address should I listen on")
	clientID     = flag.String("client_id", "", "The client ID from https://api.slack.com/applications")
	clientSecret = flag.String("client_secret", "", "The client secret from https://api.slack.com/applications")
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
		ClientID:     *clientID,
		ClientSecret: *clientSecret,
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
	token, err := slack.OAuthAccess(*clientID, *clientSecret, code, "")
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
	GoodReadsKey    string `json:"goodReadsKey"`
	GoodReadsSecret string `json:"goodReadsSecret"`
	SlackToken      string `json:"slackToken"`
	GoodReadsHost   string `json:"goodReadsHost"`
	GoodReadsPort   string `json:"goodReadsPort"`
	SlackHost       string `json:"slackHost"`
	SlackHostHTTP   string `json:"slackHostHttp"`
	IsHTTPS         string `json:"isHTTPS"`
}

func main() {
	gr := goodreads.NewClient(config.GoodReadsKey, config.GoodReadsSecret)
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

	flag.Parse()
	if *clientID == "" || *clientSecret == "" {
		fmt.Println("You must specify the client ID and client secret from https://api.slack.com/applications")
		os.Exit(1)
	}
	http.HandleFunc("/add", addToSlack)
	http.HandleFunc("/auth", auth)
	http.HandleFunc("/", home)
	log.Fatal(http.ListenAndServe(*address, nil))

}

func init() {
	config = readConfig()
}

func readConfig() Configuration {

	file, _ := os.Open("conf.json")
	decoder := json.NewDecoder(file)
	configuration := Configuration{}

	if configuration.SlackToken = os.Getenv("slackToken"); configuration.SlackToken != "" {
		configuration.GoodReadsHost = os.Getenv("goodReadsHost")
		configuration.GoodReadsPort = os.Getenv("goodReadsPort")
		configuration.GoodReadsSecret = os.Getenv("goodReadsSecret")
		configuration.IsHTTPS = os.Getenv("isHTTPS")
		configuration.SlackHostHTTP = os.Getenv("slackHostHTTP")
		configuration.SlackHost = os.Getenv("slackHost")
		configuration.SlackToken = os.Getenv("slackToken")
	} else {
		err := decoder.Decode(&configuration)
		if err != nil {
			fmt.Println("error:", err)
		}
	}

	return configuration
}
