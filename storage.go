package main

import (
	"bytes"
	"errors"
	"fmt"
	"time"

	"github.com/clockworkcoding/slack"
	_ "github.com/lib/pq"
)

func saveOverdriveAuth(param overdriveAuth) (err error) {

	if _, err = db.Exec(`CREATE TABLE IF NOT EXISTS overdrive_auth (
	id serial primary key,
	teamid varchar(200),
	slackuserid varchar(200),
	overdriveaccountid varchar(200),
	token varchar,
	refreshtoken varchar,
	tokenType varchar,
	expiry timestamp,
	createdtime timestamp
	)`); err != nil {
		fmt.Println("Error creating database table: " + err.Error())
		return
	}
	if param.id != 0 {
		query := fmt.Sprintf(`UPDATE overdrive_auth
	SET token = '%s' ,
	refreshtoken = '%s',
	tokenType = '%s',
	expiry = '%v',
	overdriveaccountid = '%s'
	where id = %v`,
			param.token, param.refreshToken, param.tokenType, param.expiry.Format(time.RFC3339Nano), param.overdriveAccountID, param.id)
		if _, err = db.Exec(query); err != nil {
			fmt.Println("Error saving overdrive auth: " + err.Error())
			return
		}

	} else {
		if _, err = db.Exec(fmt.Sprintf(`INSERT INTO overdrive_auth(
		teamid ,
		slackuserid ,
		overdriveaccountid,
		token ,
		refreshtoken ,
		tokenType,
		expiry,
		createdtime
		) VALUES ('%s','%s','%s','%s','%s','%s', '%v', now())`,
			param.teamID, param.slackUserID, param.overdriveAccountID, param.token, param.refreshToken, param.tokenType, param.expiry.Format(time.RFC3339Nano))); err != nil {
			fmt.Println("Error saving overdrive auth: " + err.Error())
			return
		}
	}
	return
}

func saveGoodreadsAuth(param goodreadsAuth) (err error) {

	if _, err = db.Exec(`CREATE TABLE IF NOT EXISTS goodreads_auth (
			id serial,
			teamid varchar(200),
			slackuserid varchar(200),
			goodreadsuserid varchar(200),
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
	secret = '%s',
	goodreadsuserid = '%s'
	where id = %v`,
			param.token, param.secret, param.goodreadsUserID, param.id)); err != nil {
			fmt.Println("Error saving goodreads auth: " + err.Error())
			return
		}

	} else {
		if _, err = db.Exec(fmt.Sprintf(`INSERT INTO goodreads_auth(
		teamid ,
		slackuserid ,
		goodreadsuserid,
		token ,
		secret ,
		createdtime
		) VALUES ('%s','%s','%s','%s','%s', now())`,
			param.teamID, param.slackUserID, param.goodreadsUserID, param.token, param.secret)); err != nil {
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
	id              int
	teamID          string
	slackUserID     string
	goodreadsUserID string
	token           string
	secret          string
}

type overdriveAuth struct {
	id                 int
	teamID             string
	slackUserID        string
	overdriveAccountID string
	token              string
	refreshToken       string
	tokenType          string
	expiry             time.Time
}

func getGoodreadsAuth(param goodreadsAuth) (result goodreadsAuth, err error) {
	var query bytes.Buffer
	query.WriteString("SELECT id, teamid, slackuserid, goodreadsuserid, token, secret FROM goodreads_auth WHERE 1 = 1 ")
	if param.id != 0 {
		query.WriteString(" AND id = ")
		query.WriteString(string(param.id))
	}
	if param.teamID != "" {
		query.WriteString(" AND teamid = '")
		query.WriteString(param.teamID)
		query.WriteString("'")
	}
	if param.slackUserID != "" {
		query.WriteString(" AND slackuserid = '")
		query.WriteString(param.slackUserID)
		query.WriteString("'")
	}
	if param.goodreadsUserID != "" {
		query.WriteString(" AND goodreadsuserid = '")
		query.WriteString(param.goodreadsUserID)
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
		if err = rows.Scan(&result.id, &result.teamID, &result.slackUserID, &result.goodreadsUserID, &result.token, &result.secret); err != nil {
			fmt.Println("Error scanning auth:" + err.Error())
			return
		}
		return
	}

	return result, errors.New("User not found")
}

func getOverdriveAuth(param overdriveAuth) (result overdriveAuth, err error) {
	var query bytes.Buffer
	query.WriteString("SELECT id, teamid, slackuserid, overdriveaccountid, token, refreshtoken, tokenType, expiry FROM overdrive_auth WHERE 1 = 1 ")
	if param.id != 0 {
		query.WriteString(" AND id = ")
		query.WriteString(string(param.id))
	}
	if param.teamID != "" {
		query.WriteString(" AND teamid = '")
		query.WriteString(param.teamID)
		query.WriteString("'")
	}
	if param.slackUserID != "" {
		query.WriteString(" AND slackuserid = '")
		query.WriteString(param.slackUserID)
		query.WriteString("'")
	}
	if param.overdriveAccountID != "" {
		query.WriteString(" AND overdriveAccountID = '")
		query.WriteString(param.overdriveAccountID)
		query.WriteString("'")
	}
	if param.token != "" {
		query.WriteString(" AND token = '")
		query.WriteString(param.token)
		query.WriteString("'")
	}
	if param.refreshToken != "" {
		query.WriteString(" AND refreshtoken = '")
		query.WriteString(param.refreshToken)
		query.WriteString("'")
	}
	if param.tokenType != "" {
		query.WriteString(" AND tokenType = '")
		query.WriteString(param.tokenType)
		query.WriteString("'")
	}
	query.WriteString(" ORDER BY createdtime DESC FETCH FIRST 1 ROWS ONLY")

	rows, err := db.Query(query.String())
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(&result.id, &result.teamID, &result.slackUserID, &result.overdriveAccountID, &result.token, &result.refreshToken, &result.tokenType, &result.expiry); err != nil {
			fmt.Println("Error scanning auth:" + err.Error())
			return
		}
		return
	}

	return result, errors.New("User not found")
}
