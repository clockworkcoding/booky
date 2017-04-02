package main

import (
	"errors"
	"fmt"

	"github.com/clockworkcoding/slack"
	_ "github.com/lib/pq"
)

func saveGoodreadsAuth(teamid, userid, token, secret string) (err error) {

	if _, err = db.Exec(`CREATE TABLE IF NOT EXISTS goodreads_auth (
			id serial,
			teamid varchar(200),
			userid varchar(200),
			token varchar(200),
			secret varchar(200),
			createdtime	timestamp
			)`); err != nil {
		fmt.Println("Error creating database table: " + err.Error())
		return
	}
	if _, err = db.Exec(fmt.Sprintf(`INSERT INTO goodreads_auth(
		teamid ,
		userid ,
		token ,
		secret ,
		createdtime
		) VALUES ('%s','%s','%s','%s', now())`,
		teamid, userid, token, secret)); err != nil {
		fmt.Println("Error saving goodreads auth: " + err.Error())
		return
	}

	return
}

func saveSlackAuth(oAuth *slack.OAuthResponse) (err error) {

	if _, err = db.Exec(`CREATE TABLE IF NOT EXISTS slack_auth (
		id serial,
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
	if _, err = db.Exec(fmt.Sprintf(`INSERT INTO slack_auth (
		team ,
		teamid,
		token ,
		url ,
		configUrl ,
		channel ,
		channelid,
		createdtime	)
		VALUES ('%s','%s','%s','%s','%s','%s','%s', now())`, oAuth.TeamName, oAuth.TeamID,
		oAuth.AccessToken, oAuth.IncomingWebhook.URL, oAuth.IncomingWebhook.ConfigurationURL, oAuth.IncomingWebhook.Channel, oAuth.IncomingWebhook.ChannelID)); err != nil {
		fmt.Println("Error saving slack auth: " + err.Error())
		return
	}

	return
}

func getSlackAuth(teamID string) (id int, token, channelid string, err error) {
	rows, err := db.Query(fmt.Sprintf("SELECT token, channelid FROM slack_auth WHERE teamid = '%s' ORDER BY createdtime DESC FETCH FIRST 1 ROWS ONLY", teamID))
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&id, &token, &channelid); err != nil {
			fmt.Println("Error scanning auth:" + err.Error())
			return
		}
		return
	}

	return id, token, channelid, errors.New("Team not found")
}

func getGoodreadsAuth(teamID, userID string) (id int, token, secret string, err error) {
	rows, err := db.Query(fmt.Sprintf("SELECT id, token, secret FROM goodreads_auth WHERE teamid = '%s' and userid = '%s' ORDER BY createdtime DESC FETCH FIRST 1 ROWS ONLY", teamID, userID))
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&id, &token, &secret); err != nil {
			fmt.Println("Error scanning auth:" + err.Error())
			return
		}
		return
	}

	return id, token, secret, errors.New("User not found")
}
