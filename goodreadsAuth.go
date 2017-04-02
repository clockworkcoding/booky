package main

import (
	"fmt"
	"net/http"

	"github.com/clockworkcoding/goodreads"
	"github.com/clockworkcoding/slack"

	_ "github.com/lib/pq"
)

// auth receives the callback from Slack, validates and displays the user information
func goodReadsAuth(w http.ResponseWriter, r *http.Request) {
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

// addToGoodreads initializes the oauth process and redirects to Goodreads
func addToGoodreads(w http.ResponseWriter, r *http.Request) {
	teamID := r.FormValue("team")
	userID := r.FormValue("user")
	if teamID == "" || userID == "" {
		writeError(w, http.StatusBadRequest, "Bad Request")
		return
	}
	c := goodreads.NewClient(config.Goodreads.Key, config.Goodreads.Secret)
	rtoken, url, err := c.Consumer.GetRequestTokenAndUrl(config.URL + "/grauth")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	fmt.Printf("%s got token %s and secret %s\n", userID, rtoken.Token, rtoken.Secret)
	err = saveGoodreadsAuth(teamID, userID, rtoken.Token, rtoken.Secret)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	http.Redirect(w, r, url, http.StatusFound)
}
