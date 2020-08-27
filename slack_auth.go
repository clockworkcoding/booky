package main

import (
	"crypto/rand"
	"log"
	"net/http"

	"github.com/clockworkcoding/slack"

	_ "github.com/lib/pq"
)

// auth receives the callback from Slack, validates and displays the user information
func slackAuth(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	errStr := r.FormValue("error")
	if errStr != "" {
		log.Output(0, "auth error "+errStr)
		http.Redirect(w, r, config.RedirectURL+"/Error", http.StatusTemporaryRedirect)
		return
	}
	oAuthResponse, err := slack.GetV2OAuthResponse(config.Slack.ClientID, config.Slack.ClientSecret, code, "", false)
	if err != nil {
		log.Output(0, "auth error: "+err.Error())
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

	url := "https://slack.com/oauth/v2/authorize?client_id=" + config.Slack.ClientID + "&scope=commands,links:read,links:write&user_scope=links:read"
	http.Redirect(w, r, url, http.StatusFound)
}

// home displays the add-to-slack button
func home(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`<html><head><title>Slack OAuth Test</title></head><body><a href="/add"><img alt="Add to Slack" height="40" width="139" src="https://platform.slack-edge.com/img/add_to_slack.png" srcset="https://platform.slack-edge.com/img/add_to_slack.png 1x, https://platform.slack-edge.com/img/add_to_slack@2x.png 2x" /></a></body></html>`))
}
