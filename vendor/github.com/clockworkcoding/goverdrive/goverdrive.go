package goverdrive

import "net/url"

const (
	//AuthURL Overdrive's authentication endpoint
	AuthURL = "https://oauth.overdrive.com/auth"
	//TokenURL Overdrive's token endpoint
	TokenURL = "https://oauth.overdrive.com/token"
	//authURI = "https://oauth.overdrive.com/auth?client_id=%s&redirect_uri=%s&scope=accountId:%s&response_type=code&state=%s"
)

//BuildAuthURL generates the interpolated auth URI, probably won't be used
func BuildAuthURL(clientID, accountID, redirectURI, state string) string {
	var URL *url.URL
	URL, _ = url.Parse(AuthURL)
	parameters := url.Values{}
	parameters.Add("client_id", clientID)
	parameters.Add("redirect_uri", redirectURI)
	parameters.Add("scope", accountID)
	parameters.Add("state", state)
	URL.RawQuery = parameters.Encode()

	return URL.String()
}
