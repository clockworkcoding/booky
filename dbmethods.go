package main

import (
	"errors"
	"fmt"
	"net/http"

	_ "github.com/lib/pq"
	"github.com/nlopes/slack"
)

func saveSlackAuth(oAuth *slack.OAuthResponse) (err error) {

	fmt.Printf("T: %s", oAuth.AccessToken)
	if _, err = db.Exec(`CREATE TABLE IF NOT EXISTS slack_auth (
		team varchar(200),
		teamid varchar(20),
		token varchar(200),
		url varchar(200),
		configUrl varchar(200),
		channel varchar(200),
		channelid varchar(200),
		createdtime	timestamp
		)`); err != nil {
		fmt.Println("Error creating database table: " + err.Error())
		return
	}
	if _, err = db.Exec(fmt.Sprintf("INSERT INTO slack_auth VALUES ('%s','%s','%s','%s','%s','%s','%s', now())", oAuth.TeamName, oAuth.TeamID,
		oAuth.AccessToken, oAuth.IncomingWebhook.URL, oAuth.IncomingWebhook.ConfigurationURL, oAuth.IncomingWebhook.Channel, oAuth.IncomingWebhook.ChannelID)); err != nil {
		fmt.Println("Error saving auth: " + err.Error())
		return
	}

	rows, err := db.Query("SELECT 		team ,		teamid ,		token ,		url ,		configUrl ,		channel ,		channelid FROM slack_auth")
	if err != nil {
		fmt.Println("Error reading auth: " + err.Error())
		return
	}
	defer rows.Close()
	for rows.Next() {
		var a, b, c, d, e, f, g string
		if err = rows.Scan(&a, &b, &c, &d, &e, &f, &g); err != nil {
			fmt.Println("Error scanning auth:" + err.Error())
			return
		}
		fmt.Printf("Read from DB: %s - %s - %s - %s - %s - %s - %s \n", a, b, c, d, e, f, g)
	}
	return
}

func getAuth(teamID string) (token, channelid string, err error) {
	fmt.Printf("TeamId: %s\n", teamID)
	rows, err := db.Query(fmt.Sprintf("SELECT token, channelid FROM slack_auth WHERE teamid = '%s' ORDER BY createdtime DESC FETCH FIRST 1 ROWS ONLY", teamID))
	if err != nil {
		fmt.Println("Error reading auth: " + err.Error())
		return
	}
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&token, &channelid); err != nil {
			fmt.Println("Error scanning auth:" + err.Error())
			return
		}
		return
	}
	return token, channelid, errors.New("Team not found")
}

func dbFunc(w http.ResponseWriter, r *http.Request) {

	if _, err := db.Exec("DROP TABLE IF EXISTS slack_auth"); err != nil {
		fmt.Println("Error creating database table: " + err.Error())
		return
	}
	w.Write([]byte(fmt.Sprintf("Table Dropped")))
}
