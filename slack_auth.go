package main

import (
	"crypto/rand"
	"net/http"
  "log"
	"golang.org/x/oauth2"

	"github.com/clockworkcoding/slack"

	_ "github.com/lib/pq"
)

// auth receives the callback from Slack, validates and displays the user information
func slackAuth(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	errStr := r.FormValue("error")
	if errStr != "" {
    log.Output(0, "auth error " + errStr)
		http.Redirect(w, r, config.RedirectURL+"/Error", http.StatusTemporaryRedirect)
		return
	}
	oAuthResponse, err := slack.GetOAuthResponse(config.Slack.ClientID, config.Slack.ClientSecret, code, "https://clockworkcoding-booky.glitch.me/auth", false)
	if err != nil {
    log.Output(0, "auth error: " + err.Error())
		http.Redirect(w, r, config.RedirectURL+"/Error", http.StatusTemporaryRedirect)
		return
	}

	http.Redirect(w, r, config.RedirectURL+"/SlackSuccess", http.StatusTemporaryRedirect)
	if err = saveSlackAuth(oAuthResponse); err != nil {
		http.Redirect(w, r, config.RedirectURL+"/Error", http.StatusTemporaryRedirect)
		return
	}

}

// addToSlack initializes the oauth process and redirects to Slack
func addToSlack(w http.ResponseWriter, r *http.Request) {
	// Just generate random state
	b := make([]byte, 10)
	_, err := rand.Read(b)
	if err != nil {
		http.Redirect(w, r, config.RedirectURL+"/Error", http.StatusTemporaryRedirect)
	}

	conf := &oauth2.Config{
		ClientID:     config.Slack.ClientID,
		ClientSecret: config.Slack.ClientSecret,
		Scopes:       []string{"links:read", "links:write", "commands"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://slack.com/oauth/v2/authorize",
			TokenURL: "https://slack.com/api/oauth.access", // not actually used here
		},
	}
	url := conf.AuthCodeURL(globalState.auth)
	http.Redirect(w, r, url, http.StatusFound)
}

// home displays the add-to-slack button
func home(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`<html><head><title>Slack OAuth Test</title></head><body><a href="/add"><img alt="Add to Slack" height="40" width="139" src="https://platform.slack-edge.com/img/add_to_slack.png" srcset="https://platform.slack-edge.com/img/add_to_slack.png 1x, https://platform.slack-edge.com/img/add_to_slack@2x.png 2x" /></a></body></html>`))
}
