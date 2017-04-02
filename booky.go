package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/clockworkcoding/goodreads"
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
	fmt.Printf("Err: %s", err)
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

	var values buttonValues
	err = values.decodeValues(action.Actions[0].Value)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	_, token, _, err := getSlackAuth(action.Team.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	api := slack.New(token)
	if action.User.ID != values.User {
		responseParams := slack.NewResponseMessageParameters()
		responseParams.ResponseType = "ephemeral"
		responseParams.ReplaceOriginal = false
		responseParams.Text = fmt.Sprintf("Only %s can update this book", values.UserName)
		err = api.PostResponse(action.ResponseURL, responseParams.Text, responseParams)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())

		}
		return

	}
	var timestamp string
	var v map[string]interface{}
	err = json.Unmarshal(action.OriginalMessage, &v)
	if err != nil {
		timestamp = ""
	} else {
		timestamp = v["ts"].(string)
	}
	wrongBookButtons := true
	switch action.Actions[0].Name {
	case "right book":
		wrongBookButtons = false
	case "nvm":
		_, _, err = api.DeleteMessage(action.Channel.ID, timestamp)
		if err != nil {
			if !values.IsEphemeral {
				writeError(w, http.StatusInternalServerError, err.Error())
			} else {
				w.Write([]byte("Sorry you couldn't find your book. Try searching for both the author and title together"))
			}

		}
		return
	}

	params, err := createBookPost(values, wrongBookButtons)
	if err != nil {
		if err.Error() == "no books found" {
			w.Write([]byte("No books found, try a broader search"))
		} else {
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	//If it's an ephemeral post, replace it with an in_channel post after finding the right one, otherwise just update
	if values.IsEphemeral && action.Actions[0].Name != "right book" {
		responseParams := slack.NewResponseMessageParameters()
		responseParams.ResponseType = "ephemeral"
		responseParams.ReplaceOriginal = true
		responseParams.Text = params.Text
		responseParams.Attachments = params.Attachments

		err = api.PostResponse(action.ResponseURL, responseParams.Text, responseParams)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
		}
	} else if values.IsEphemeral {
		params.AsUser = false
		defer w.Write([]byte("Posting your book"))
		_, _, err = api.PostMessage(action.Channel.ID, params.Text, params)
		if err != nil {
			fmt.Printf("Error posting: %s\n", err.Error())
			return
		}
	} else {
		updateParams := slack.UpdateMessageParameters{
			Timestamp:   timestamp,
			Text:        params.Text,
			Attachments: params.Attachments,
			Parse:       params.Parse,
			LinkNames:   params.LinkNames,
			AsUser:      params.AsUser,
		}

		_, _, _, err = api.UpdateMessageWithAttachments(action.Channel.ID, updateParams)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
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

	values := buttonValues{
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
			writeError(w, http.StatusInternalServerError, err.Error())
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
		writeError(w, http.StatusInternalServerError, err.Error())
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
	values := buttonValues{
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

func createBookPost(values buttonValues, wrongBookButtons bool) (params slack.PostMessageParameters, err error) {
	gr := goodreads.NewClient(config.Goodreads.Key, config.Goodreads.Secret)

	results, err := gr.GetSearch(values.Query)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	if len(results.Search_work) == 0 {
		err = errors.New("no books found")
		return
	}

	if values.Index >= len(results.Search_work) {
		values.Index = 0
	}

	book, err := gr.GetBook(results.Search_work[values.Index].Search_best_book.Search_id.Text)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	var authorBuffer bytes.Buffer
	for i, author := range book.Book_authors[0].Book_author {
		if i > 0 {
			authorBuffer.WriteString(" | ")
		}
		if author.Book_role.Text != "" {
			authorBuffer.WriteString(author.Book_role.Text)
			authorBuffer.WriteString(": ")
		}
		authorBuffer.WriteString(author.Book_name.Text)
	}

	rating := book.Book_average_rating[0].Text
	numRating, _ := strconv.ParseFloat(rating, 32)
	var stars string
	for i := 0; i < int(numRating+0.5); i++ {
		stars += ":star:"
	}

	rightValues := values.encodeValues()

	attachments := []slack.Attachment{
		slack.Attachment{
			Title:      book.Book_title[0].Text,
			TitleLink:  book.Book_url.Text,
			AuthorName: authorBuffer.String(),
			ThumbURL:   book.Book_image_url[0].Text,
			Fields: []slack.AttachmentField{
				slack.AttachmentField{
					Title: fmt.Sprintf("Avg Rating (%s)", rating),
					Value: stars,
					Short: true,
				},
				slack.AttachmentField{
					Title: "Ratings",
					Value: book.Book_ratings_count[0].Text,
					Short: true,
				},
			},
		},
		slack.Attachment{
			Text:       replaceMarkup(book.Book_description.Text),
			MarkdownIn: []string{"text", "fields"},
			Footer:     fmt.Sprintf("Posted by %s using /booky | Data from Goodreads.com", values.UserName),
		},
	}
	if wrongBookButtons {
		values.Index += 1
		nextValues := values.encodeValues()
		values.Index -= 2
		prevValues := values.encodeValues()

		nextBookButton := slack.AttachmentAction{
			Name:  "next book",
			Text:  "Wrong Book?",
			Type:  "button",
			Value: string(nextValues),
		}
		nvmBookButton := slack.AttachmentAction{
			Name:  "nvm",
			Text:  "nvm",
			Type:  "button",
			Style: "danger",
			Value: string(nextValues),
		}
		rightBookButton := slack.AttachmentAction{
			Name:  "right book",
			Text:  ":thumbsup:",
			Type:  "button",
			Style: "primary",
			Value: string(rightValues),
		}
		wrongBookButtons := slack.Attachment{
			CallbackID: "wrongbook",
			Fallback:   "Try using both the title and the author's name",
			Actions:    []slack.AttachmentAction{},
		}

		if values.Index >= 0 {
			prevBookButton := slack.AttachmentAction{
				Name:  "previousbook",
				Text:  "previous",
				Type:  "button",
				Value: string(prevValues),
			}
			nextBookButton.Text = "next"
			wrongBookButtons.Actions = append(wrongBookButtons.Actions, prevBookButton)
		}
		wrongBookButtons.Actions = append(wrongBookButtons.Actions, nextBookButton, nvmBookButton, rightBookButton)
		attachments = append(attachments, wrongBookButtons)
	}
	params = slack.NewPostMessageParameters()
	params.Text = book.Book_title[0].Text
	params.AsUser = true
	params.Attachments = attachments
	return
}

// event responds to events from slack
func event(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	var v map[string]interface{}
	err := decoder.Decode(&v)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if v["token"].(string) != config.Slack.VerificationToken {
		writeError(w, http.StatusForbidden, "Forbidden")
		return
	}
	w.WriteHeader(http.StatusOK)
	if v["type"].(string) == "url_verification" {
		w.Write([]byte(v["challenge"].(string)))
		return
	}
	event, err := json.Marshal(v)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	switch v["event"].(map[string]interface{})["type"].(string) {
	case "message":
		var message eventMessage
		err := json.Unmarshal(event, &message)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		checkTextForBook(message)
	case "link_shared":
		var link eventLinkShared
		err := json.Unmarshal(event, &link)
		if err != nil {
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
	mux.Handle("/grauth", http.HandlerFunc(goodReadsAuth))
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
