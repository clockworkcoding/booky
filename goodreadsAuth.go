package main

import (
	"fmt"
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
		writeError(w, 401, "Oops, you didn't authorize Booky")
		return
	}
	auth, err := getGoodreadsAuth(goodreadsAuth{token: token})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	c := goodreads.NewClient(config.Goodreads.Key, config.Goodreads.Secret)
	accessToken, err := c.Consumer.AuthorizeToken(&oauth.RequestToken{Secret: auth.secret, Token: auth.token}, token)
	if err != nil {
		writeError(w, 401, err.Error())
		return
	}

	auth.token = accessToken.Token
	auth.secret = accessToken.Secret

	w.Write([]byte(fmt.Sprintf("OAuth successful for team %s and user %s", auth.teamID, auth.userID)))
	if err = saveGoodreadsAuth(auth); err != nil {
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
	auth := goodreadsAuth{
		secret: rtoken.Secret,
		token:  rtoken.Token,
		userID: userID,
		teamID: teamID,
	}
	err = saveGoodreadsAuth(auth)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	http.Redirect(w, r, url, http.StatusFound)
}
