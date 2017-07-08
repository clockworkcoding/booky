package main

import (
	"log"
	"net/http"

	"github.com/clockworkcoding/goverdrive"

	_ "github.com/lib/pq"
)

// auth receives the callback from Overdrive, validates and displays the user information
func overdriveAuthCallback(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	userid := r.FormValue("state")
	log.Printf("code: %s, user: %s\n", code, userid)
	if userid == "" || code == "" {
		http.Redirect(w, r, config.RedirectURL+"/Error", http.StatusTemporaryRedirect)
		return
	}
	auth, err := getOverdriveAuth(overdriveAuth{slackUserID: userid})
	if err != nil {
		http.Redirect(w, r, config.RedirectURL+"/Error", http.StatusTemporaryRedirect)
		return
	}
	log.Println("loaded auth from db")
	accessToken, err := goverdrive.GetToken(config.Overdrive.Key, config.Overdrive.Secret, auth.overdriveAccountID, code, config.URL+"/odauth")
	if err != nil {
		http.Redirect(w, r, config.RedirectURL+"/Error", http.StatusTemporaryRedirect)
		return
	}
	auth.token = accessToken.AccessToken
	auth.refreshToken = accessToken.RefreshToken
	auth.tokenType = accessToken.TokenType
	auth.expiry = accessToken.Expiry
	log.Println("retrieved accesstoken")
	err = saveOverdriveAuth(auth)
	if err != nil {
		log.Println(err.Error())
		http.Redirect(w, r, config.RedirectURL+"/Error", http.StatusTemporaryRedirect)
		return
	}
	log.Println("saved the token!")
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

	url := goverdrive.BuildAuthURL(config.Overdrive.Key, "4425", userID, config.URL+"/odauth")

	auth := overdriveAuth{
		slackUserID:        userID,
		teamID:             teamID,
		overdriveAccountID: "4425",
	}
	err := saveOverdriveAuth(auth)
	if err != nil {
		http.Redirect(w, r, config.RedirectURL+"/Error", http.StatusTemporaryRedirect)
		return
	}

	http.Redirect(w, r, url, http.StatusFound)
}
