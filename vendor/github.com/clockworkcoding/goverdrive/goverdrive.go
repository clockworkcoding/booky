package goverdrive

import (
	"context"
	"log"
	"net/url"

	"golang.org/x/oauth2"
)

const (
	//AuthURL Overdrive's authentication endpoint
	AuthURL = "https://oauth.overdrive.com/auth"
	//TokenURL Overdrive's token endpoint
	TokenURL = "https://oauth.overdrive.com/token"
	//authURI = "https://oauth.overdrive.com/auth?client_id=%s&redirect_uri=%s&scope=accountId:%s&response_type=code&state=%s"
)

//BuildAuthURL generates the interpolated auth URI, probably won't be used
func BuildAuthURL(clientID, accountID, state, redirectURI string) string {
	var URL *url.URL
	URL, _ = url.Parse(AuthURL)
	parameters := url.Values{}
	parameters.Add("response_type", "code")
	parameters.Add("client_id", clientID)
	parameters.Add("redirect_uri", redirectURI)
	parameters.Add("scope", "accountId:"+accountID)
	parameters.Add("state", state)
	URL.RawQuery = parameters.Encode()
	log.Println(URL.String())
	return URL.String()
}

func GetToken(clientID, clientSecret, libraryAccountId, code, redirectURI string) (token *oauth2.Token, err error) {
	ctx := context.Background()
	conf := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURI,
		Scopes:       []string{libraryAccountId},
		Endpoint: oauth2.Endpoint{
			AuthURL:  AuthURL,
			TokenURL: TokenURL,
		},
	}

	token, err = conf.Exchange(ctx, code)
	if err != nil {
		log.Println(err)
		return
	}

	return
}
