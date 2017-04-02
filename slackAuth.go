package main

import (
	"crypto/rand"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"

	"github.com/clockworkcoding/slack"

	_ "github.com/lib/pq"
)

// auth receives the callback from Slack, validates and displays the user information
func slackAuth(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	errStr := r.FormValue("error")
	if errStr != "" {
		writeError(w, 401, errStr)
		return
	}
	oAuthResponse, err := slack.GetOAuthResponse(config.Slack.ClientID, config.Slack.ClientSecret, code, "", false)
	if err != nil {
		writeError(w, 401, err.Error())
		return
	}

	w.Write([]byte(fmt.Sprintf("OAuth successful for team %s and user %s", oAuthResponse.TeamName, oAuthResponse.UserID)))
	if err = saveSlackAuth(oAuthResponse); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

}

// addToSlack initializes the oauth process and redirects to Slack
func addToSlack(w http.ResponseWriter, r *http.Request) {
	// Just generate random state
	b := make([]byte, 10)
	_, err := rand.Read(b)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
	}

	conf := &oauth2.Config{
		ClientID:     config.Slack.ClientID,
		ClientSecret: config.Slack.ClientSecret,
		Scopes:       []string{"channels:history", "incoming-webhook", "links:read", "links:write", "chat:write:bot", "commands"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://slack.com/oauth/authorize",
			TokenURL: "https://slack.com/api/oauth.access", // not actually used here
		},
	}
	url := conf.AuthCodeURL(globalState.auth)
	http.Redirect(w, r, url, http.StatusFound)
}

// home displays the add-to-slack button
func home(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`<html><head><title>Slack OAuth Test</title></head><body><a href="https://slack.com/oauth/authorize?&client_id=87777690085.158899563392&scope=links:read,channels:history"><img alt="Add to Slack" height="40" width="139" src="https://platform.slack-edge.com/img/add_to_slack.png" srcset="https://platform.slack-edge.com/img/add_to_slack.png 1x, https://platform.slack-edge.com/img/add_to_slack@2x.png 2x" /></a></body></html>`))
}
