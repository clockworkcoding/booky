package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/clockworkcoding/goodreads"
	_ "github.com/lib/pq"
	"github.com/nlopes/slack"
	"golang.org/x/oauth2"
	"log"
	"net/http"
	"os"
	"time"
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
		Scopes:       []string{"channels:history", "incoming-webhook"},
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
	oAuthResponse, err := slack.GetOAuthResponse(config.Slack.ClientID, config.Slack.ClientSecret, code, "", false)
	if err != nil {
		writeError(w, 401, err.Error())
		return
	}
	fmt.Println(oAuthResponse.IncomingWebhook.Channel)

	w.Write([]byte(fmt.Sprintf("OAuth successful for team %s and user %s", oAuthResponse.TeamName, oAuthResponse.UserID)))
	saveSlackAuth(oAuthResponse)

}

// home displays the add-to-slack button
func home(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`<html><head><title>Slack OAuth Test</title></head><body><a href="/add">Add To Slack</a></body></html>`))
}

type Challenge struct {
	Token     string `json:"token"`
	Challenge string `json:"challenge"`
	Type      string `json:"type"`
}

// event responds to events from slack
func event(w http.ResponseWriter, r *http.Request) {
	fmt.Println("event reached")
	decoder := json.NewDecoder(r.Body)

	var challenge Challenge
	err := decoder.Decode(&challenge)
	if err != nil {
		fmt.Println("ERR: " + err.Error())
	}
	defer r.Body.Close()
	fmt.Println(challenge.Challenge)

	w.Write([]byte(challenge.Challenge))
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

	http.HandleFunc("/dbfunc", dbFunc)
	http.HandleFunc("/add", addToSlack)
	http.HandleFunc("/auth", auth)
	http.HandleFunc("/event", event)
	http.HandleFunc("/", home)
	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), nil))

}

func saveSlackAuth(oAuth *slack.OAuthResponse) {

	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS slack_auth (
		team varchar(200),
		teamid varchar(20),
		token varchar(200),
		url varchar(200),
		configUrl varchar(200),
		channel varchar(200),
		channelid varchar(200),
		createdtime	timestamp
		)`); err != nil {
		fmt.Println("Error creating database table: " + err.Error())
		return
	}

	if _, err := db.Exec(fmt.Sprintf("INSERT INTO slack_auth VALUES ('%s','%s','%s','%s','%s','%s','%s, now()')", oAuth.TeamName, oAuth.TeamID,
		oAuth.AccessToken, oAuth.IncomingWebhook.URL, oAuth.IncomingWebhook.ConfigurationURL, oAuth.IncomingWebhook.Channel, oAuth.IncomingWebhook.ChannelID)); err != nil {
		fmt.Println("Error incrementing tick: " + err.Error())
		return
	}

	rows, err := db.Query("SELECT * FROM slack_auth")
	if err != nil {
		fmt.Println("Error reading ticks: " + err.Error())
		return
	}
	defer rows.Close()
	for rows.Next() {
		var a, b, c, d, e, f, g string
		if err := rows.Scan(&a, &b, &c, &d, &e, &f, &g); err != nil {
			fmt.Println("Error scanning ticks:" + err.Error())
			return
		}
		fmt.Printf("Read from DB: %s - %s - %s - %s - %s - %s - %s\n", a, b, c, d, e, f, g)
	}
}
func dbFunc(w http.ResponseWriter, r *http.Request) {

	if _, err := db.Exec("DROP TABLE IF EXISTS slack_auth"); err != nil {
		fmt.Println("Error creating database table: " + err.Error())
		return
	}
	w.Write([]byte(fmt.Sprintf("Table Dropped")))
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
