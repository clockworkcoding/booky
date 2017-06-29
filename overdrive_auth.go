package main

import (
	"net/http"

	"github.com/clockworkcoding/goodreads"
	"github.com/mrjones/oauth"

	_ "github.com/lib/pq"
)

// auth receives the callback from Overdrive, validates and displays the user information
func overdriveAuthCallback(w http.ResponseWriter, r *http.Request) {
	token := r.FormValue("oauth_token")
	authorize := r.FormValue("authorize")
	if authorize != "1" || token == "" {
		http.Redirect(w, r, config.RedirectURL+"/Error", http.StatusTemporaryRedirect)
		return
	}
	auth, err := getOverdriveAuth(overdriveAuth{token: token})
	if err != nil {
		http.Redirect(w, r, config.RedirectURL+"/Error", http.StatusTemporaryRedirect)
		return
	}

	c := goodreads.NewClient(config.Goodreads.Key, config.Goodreads.Secret)
	accessToken, err := c.Consumer.AuthorizeToken(&oauth.RequestToken{Secret: auth.refreshToken, Token: auth.token}, token)
	if err != nil {
		http.Redirect(w, r, config.RedirectURL+"/Error", http.StatusTemporaryRedirect)
		return
	}

	auth.token = accessToken.Token
	auth.refreshToken = accessToken.Secret

	c = goodreads.NewClientWithToken(config.Goodreads.Key, config.Goodreads.Secret, auth.token, auth.refreshToken)
	grUser, err := c.QueryUser()
	if err != nil {
		http.Redirect(w, r, config.RedirectURL+"/Error", http.StatusTemporaryRedirect)
		return
	}
	auth.overdriveUserID = grUser.Attr_id
	if err = saveOverdriveAuth(auth); err != nil {
		http.Redirect(w, r, config.RedirectURL+"/Error", http.StatusTemporaryRedirect)
		return
	}
	http.Redirect(w, r, config.RedirectURL+"/OverdriveSuccess", http.StatusTemporaryRedirect)

}

// addToOverdrive initializes the oauth process and redirects to Overdrive
func addToOverdrive(w http.ResponseWriter, r *http.Request) {
	teamID := r.FormValue("team")
	userID := r.FormValue("user")
	if teamID == "" || userID == "" {
		http.Redirect(w, r, config.RedirectURL+"/Error", http.StatusTemporaryRedirect)
		return
	}
	c := goodreads.NewClient(config.Goodreads.Key, config.Goodreads.Secret)
	rtoken, url, err := c.Consumer.GetRequestTokenAndUrl(config.URL + "/odauth")
	if err != nil {
		http.Redirect(w, r, config.RedirectURL+"/Error", http.StatusTemporaryRedirect)
		return
	}

	auth := overdriveAuth{
		refreshToken: rtoken.Secret,
		token:        rtoken.Token,
		slackUserID:  userID,
		teamID:       teamID,
	}
	err = saveOverdriveAuth(auth)
	if err != nil {
		http.Redirect(w, r, config.RedirectURL+"/Error", http.StatusTemporaryRedirect)
		return
	}

	http.Redirect(w, r, url, http.StatusFound)
}
