package main

import (
	"net/http"

	"github.com/clockworkcoding/goodreads"
	"github.com/mrjones/oauth"

	_ "github.com/lib/pq"
)

// auth receives the callback from Goodreads, validates and displays the user information
func goodreadsAuthCallback(w http.ResponseWriter, r *http.Request) {
	token := r.FormValue("oauth_token")
	authorize := r.FormValue("authorize")
	if authorize != "1" || token == "" {
		http.Redirect(w, r, config.RedirectURL+"/Error", http.StatusTemporaryRedirect)
		return
	}
	auth, err := getGoodreadsAuth(goodreadsAuth{token: token})
	if err != nil {
		http.Redirect(w, r, config.RedirectURL+"/Error", http.StatusTemporaryRedirect)
		return
	}

	c := goodreads.NewClient(config.Goodreads.Key, config.Goodreads.Secret)
	accessToken, err := c.Consumer.AuthorizeToken(&oauth.RequestToken{Secret: auth.secret, Token: auth.token}, token)
	if err != nil {
		http.Redirect(w, r, config.RedirectURL+"/Error", http.StatusTemporaryRedirect)
		return
	}

	auth.token = accessToken.Token
	auth.secret = accessToken.Secret

	c = goodreads.NewClientWithToken(config.Goodreads.Key, config.Goodreads.Secret, auth.token, auth.secret)
	grUser, err := c.QueryUser()
	if err != nil {
		http.Redirect(w, r, config.RedirectURL+"/Error", http.StatusTemporaryRedirect)
		return
	}
	auth.goodreadsUserID = grUser.Attr_id
	if err = saveGoodreadsAuth(auth); err != nil {
		http.Redirect(w, r, config.RedirectURL+"/Error", http.StatusTemporaryRedirect)
		return
	}
	http.Redirect(w, r, config.RedirectURL+"/GoodreadsSuccess", http.StatusTemporaryRedirect)

}

// addToGoodreads initializes the oauth process and redirects to Goodreads
func addToGoodreads(w http.ResponseWriter, r *http.Request) {
	teamID := r.FormValue("team")
	userID := r.FormValue("user")
	if teamID == "" || userID == "" {
		http.Redirect(w, r, config.RedirectURL+"/Error", http.StatusTemporaryRedirect)
		return
	}
	c := goodreads.NewClient(config.Goodreads.Key, config.Goodreads.Secret)
	rtoken, url, err := c.Consumer.GetRequestTokenAndUrl(config.URL + "/grauth")
	if err != nil {
		http.Redirect(w, r, config.RedirectURL+"/Error", http.StatusTemporaryRedirect)
		return
	}

	auth := goodreadsAuth{
		secret:      rtoken.Secret,
		token:       rtoken.Token,
		slackUserID: userID,
		teamID:      teamID,
	}
	err = saveGoodreadsAuth(auth)
	if err != nil {
		http.Redirect(w, r, config.RedirectURL+"/Error", http.StatusTemporaryRedirect)
		return
	}

	http.Redirect(w, r, url, http.StatusFound)
}
