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

	"strconv"

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
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte("Something went wrong, please try again or contact Max@ClockworkCoding.com if the problem persists."))
	log.Output(1, fmt.Sprintf("Err: %s", err))
}

func responseError(responseURL, message, token string) {
	log.Output(1, fmt.Sprintf("Err: %s", message))
	simpleResponse(responseURL, "Something went wrong, please try again or contact Max@ClockworkCoding.com if the problem persists.", false, token)
}

func simpleResponse(responseURL, message string, replace bool, token string) {
	params := slack.NewResponseMessageParameters()
	params.ResponseType = "ephemeral"
	params.ReplaceOriginal = replace
	params.Text = message

	api := slack.New(token)
	err := api.PostResponse(responseURL, params)
	if err != nil {
		log.Output(0, fmt.Sprintf("Err: %s", err.Error()))
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
		writeError(w, http.StatusUnauthorized, "Unauthorized. Expected: "+config.Slack.VerificationToken+" Received: "+action.Token)
		return
	}
	_, token, _, err := getSlackAuth(action.Team.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Write([]byte(""))
	go simpleResponse(action.ResponseURL, "", false, token)

	switch action.CallbackID {
	case "wrongbook":
		wrongBookButton(action, token)
	case "goodreads":
		goodreadsButton(action, token)
	case "overdrive":
		overdriveButton(action, token)
	case "friendAction":
		switch action.Actions[0].Name {
		case "addSelfToFriends":
			addSelfToFriendsButton(action, token)
		case "addFriend":
			addFriendButton(action, token)
		case "addAllFriends":
			addAllFriendButton(action, token)
		default:
			log.Println(action.Actions[0].Name)
		}
	case "bookaction":
		switch action.Actions[0].Name {
		case "checkOverdrive":
			overdriveButton(action, token)
		case "addToShelf":
			goodreadsButton(action, token)
		case "fullDescription":
			wrongBookButton(action, token)
		}
	case "menuSearch":
		menuSearch(action)
	case "lookUpDialog":
		dialogSearch(action, w)
	default:
		log.Println(action.CallbackID)
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
	if queryText == "-friends" {
		friendsCommand(w, r)
		return
	}
	w.Write([]byte(""))
	if queryText == "?" || queryText == "help" {
		simpleResponse(responseURL, "If you're having trouble or just want to leave a message go to http://booky.fyi/contact or email Max@ClockworkCoding.com", true, token)
		return
	}
	lookUpBook(responseURL, token, userID, userName, queryText)
}

func dialogSearch(action action, w http.ResponseWriter) {
	_, token, _, err := getSlackAuth(action.Team.ID)
	if err != nil {
		responseError(action.ResponseURL, err.Error(), token)
		return
	}

	//err = json.Unmarshal(action.Submission, &submission)
	if err != nil {
		responseError(action.ResponseURL, err.Error(), token)
		return
	}

	w.WriteHeader(200)

	var queryText string
	if len(strings.TrimSpace(action.Submission["searchtext"])) > 0 {
		queryText = action.Submission["searchtext"]
	} else {
		queryText = action.Submission["selecttitle"]
	}

	lookUpBook(action.ResponseURL, token, action.User.ID, action.User.Name, queryText)
}

func lookUpBook(responseURL, token, userID, userName, queryText string) {
	go simpleResponse(responseURL, "Looking up your book", true, token)

	values := wrongBookButtonValues{
		Index:       0,
		User:        userID,
		Query:       queryText,
		IsEphemeral: true,
		UserName:    userName,
	}

	params, err := createBookPost(values, true, true)
	if err != nil {
		if err.Error() == "no books found" {
			simpleResponse(responseURL, "No books found, try a broader search", false, token)
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
	err = api.PostResponse(responseURL, responseParams)
	if err != nil {
		responseError(responseURL, err.Error(), token)
	}
}

func checkTextForBook(queryText string, teamID string, channel string, user string) {
	_, token, authedChannel, err := getSlackAuth(teamID)
	if err != nil || channel != authedChannel {
		if err != nil {
			log.Output(0, err.Error())
		} else {
			fmt.Printf("Found: %s, Expected %s", channel, authedChannel)
		}
		return
	}
	values := wrongBookButtonValues{
		User:        user,
		Query:       queryText,
		Index:       0,
		IsEphemeral: false,
		UserName:    "booky user",
	}
	params, err := createBookPost(values, true, true)
	if err != nil {
		log.Output(0, err.Error())
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

//Configuration booky's config struct
type Configuration struct {
	Goodreads struct {
		Key    string `json:"Key"`
		Secret string `json:"Secret"`
	} `json:"Goodreads"`
	Overdrive struct {
		Key    string `json:"Key"`
		Secret string `json:"Secret"`
	} `json:"Overdrive"`
	Slack struct {
		ClientID          string `json:"ClientID"`
		ClientSecret      string `json:"ClientSecret"`
		VerificationToken string `json:"VerificationToken"`
	} `json:"slack"`
	Db struct {
		URI string `json:"URI"`
	} `json:"db"`
	Keys struct {
		Key1 string `json:"Key1"`
		Key2 string `json:"Key2"`
	} `json:"Keys"`
	RedisURL          string `json:"RedisURL"`
	URL               string `json:"URL"`
	BitlyKey          string `json:"BitlyKey"`
	BookshopID        string `json:"BookshopId"`
  BookshopFallback  string `json:"BookshopFallback"`
	RedirectURL       string `json:"RedirectURL"`
	Patreon           string `json:"patreon"`
	DescriptionLength int    `json:"description_length"`
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
	mux.Handle("/odadd", http.HandlerFunc(addToOverdrive))
	mux.Handle("/odauth", http.HandlerFunc(overdriveAuthCallback))
	mux.Handle("/event", http.HandlerFunc(event))
	mux.Handle("/booky", http.HandlerFunc(bookyCommand))
	mux.Handle("/friends", http.HandlerFunc(friendsCommand))
	mux.Handle("/button", http.HandlerFunc(buttonPressed))
	mux.Handle("/gaction", http.HandlerFunc(lookUpHandler))
	mux.Handle("/", http.HandlerFunc(redirect))
	err := http.ListenAndServe(":"+os.Getenv("PORT"), mux)
	if err != nil {
		log.Fatal("ListenAndServe error: ", err)
	}

}

func redirect(w http.ResponseWriter, r *http.Request) {
	if url := os.Getenv(strings.Replace(r.URL.Path, "/", "URL_", 1)); url != "" {
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	} else {
		http.Redirect(w, r, config.RedirectURL+r.URL.Path, http.StatusTemporaryRedirect)
	}
}

func init() {
	config = readConfig()
}

func readConfig() Configuration {
	configuration := Configuration{}

	if configuration.Slack.ClientID = os.Getenv("SLACK_CLIENT_ID"); configuration.Slack.ClientID != "" {
		var err error
		configuration.Slack.ClientSecret = os.Getenv("SLACK_CLIENT_SECRET")
		configuration.Goodreads.Secret = os.Getenv("GOODREADS_SECRET")
		configuration.Goodreads.Key = os.Getenv("GOODREADS_KEY")
		configuration.Overdrive.Secret = os.Getenv("OVERDRIVE_SECRET")
		configuration.Overdrive.Key = os.Getenv("OVERDRIVE_KEY")
		configuration.Db.URI = os.Getenv("DATABASE_URL")
		configuration.Slack.VerificationToken = os.Getenv("SLACK_VERIFICATION_TOKEN")
		configuration.URL = os.Getenv("URL")
		configuration.BitlyKey = os.Getenv("BITLY_KEY")
		configuration.BookshopID = os.Getenv("BOOKSHOP_ID")
    configuration.BookshopFallback = os.Getenv("BOOKSHOP_FALLBACK")
		configuration.RedirectURL = os.Getenv("REDIRECT_URL")
		configuration.Keys.Key1 = os.Getenv("KEY_1")
		configuration.Keys.Key2 = os.Getenv("KEY_2")
		configuration.RedisURL = os.Getenv("REDIS_URL")
		configuration.Patreon = os.Getenv("patreon")
		configuration.DescriptionLength, err = strconv.Atoi(os.Getenv("description_length"))
		if err != nil {
			configuration.DescriptionLength = 140
		}
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
