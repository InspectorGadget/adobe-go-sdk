package auth

type AdobeAuthInput struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type AdobeAuthenticationContext struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
	ClientID    string `json:"client_id"`
}

var (
	AdobeSession *AdobeAuthenticationContext = &AdobeAuthenticationContext{}
)

func (c *AdobeAuthenticationContext) setAccessToken(accessToken string) {
	c.AccessToken = accessToken
}

func (c *AdobeAuthenticationContext) setExpiresIn(expiresIn int) {
	c.ExpiresIn = expiresIn
}

func (c *AdobeAuthenticationContext) setTokenType(tokenType string) {
	c.TokenType = tokenType
}

func (c *AdobeAuthenticationContext) setClientID(clientID string) {
	c.ClientID = clientID
}

func (c *AdobeAuthenticationContext) GetAccessToken() string {
	return c.AccessToken
}

func (c *AdobeAuthenticationContext) GetExpiresIn() int {
	return c.ExpiresIn
}

func (c *AdobeAuthenticationContext) GetTokenType() string {
	return c.TokenType
}

func (c *AdobeAuthenticationContext) GetClientID() string {
	return c.ClientID
}
