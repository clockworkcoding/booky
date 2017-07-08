package goverdrive

import (
	"net/http"

	"golang.org/x/oauth2"
)

//Client is the overdrive API Client object
type Client struct {
	client *http.Client
	test   bool
}

// Constructor with Consumer key/secret and user token/secret
func NewClient(clientID, clientSecret, libraryAccountId string, token *oauth2.Token, test bool) *Client {

	conf := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  AuthURL,
			TokenURL: TokenURL,
		},
	}

	return &Client{client: conf.Client(oauth2.NoContext, token)}
}
