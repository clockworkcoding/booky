package main

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/clockworkcoding/slack"
	_ "github.com/lib/pq"
)

func saveGoodreadsAuth(param goodreadsAuth) (err error) {

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
	if param.id != 0 {
		if _, err = db.Exec(fmt.Sprintf(`UPDATE goodreads_auth
	SET token = '%s' ,
	secret = '%s'
	where id = %v`,
			param.token, param.secret, param.id)); err != nil {
			fmt.Printf(`UPDATE goodreads_auth
		SET token = '%s' ,
		secret = '%s'
		where id = %v\n`,
				param.token, param.secret, param.id)
			fmt.Println("Error saving goodreads auth: " + err.Error())
			return
		}

	} else {
		if _, err = db.Exec(fmt.Sprintf(`INSERT INTO goodreads_auth(
		teamid ,
		userid ,
		token ,
		secret ,
		createdtime
		) VALUES ('%s','%s','%s','%s', now())`,
			param.teamID, param.userID, param.token, param.secret)); err != nil {
			fmt.Println("Error saving goodreads auth: " + err.Error())
			return
		}
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
	rows, err := db.Query(fmt.Sprintf("SELECT id, token, channelid FROM slack_auth WHERE teamid = '%s' ORDER BY createdtime DESC FETCH FIRST 1 ROWS ONLY", teamID))
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

type goodreadsAuth struct {
	id     int
	teamID string
	userID string
	token  string
	secret string
}

func getGoodreadsAuth(param goodreadsAuth) (result goodreadsAuth, err error) {
	var query bytes.Buffer
	query.WriteString("SELECT id, teamid, userid, token, secret FROM goodreads_auth WHERE 1 = 1 ")
	if param.id != 0 {
		query.WriteString(" AND id = ")
		query.WriteString(string(param.id))
	}
	if param.teamID != "" {
		query.WriteString(" AND teamid = '")
		query.WriteString(param.teamID)
		query.WriteString("'")
	}
	if param.userID != "" {
		query.WriteString(" AND userid = '")
		query.WriteString(param.userID)
		query.WriteString("'")
	}
	if param.token != "" {
		query.WriteString(" AND token = '")
		query.WriteString(param.token)
		query.WriteString("'")
	}
	if param.secret != "" {
		query.WriteString(" AND secret = '")
		query.WriteString(param.secret)
		query.WriteString("'")
	}
	query.WriteString(" ORDER BY createdtime DESC FETCH FIRST 1 ROWS ONLY")

	rows, err := db.Query(query.String())
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&result.id, &result.teamID, &result.userID, &result.token, &result.secret); err != nil {
			fmt.Println("Error scanning auth:" + err.Error())
			return
		}
		return
	}

	return result, errors.New("User not found")
}
