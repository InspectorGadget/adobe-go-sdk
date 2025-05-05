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

func (c AdobeAuthenticationContext) setAccessToken(accessToken string) {
	AdobeSession.AccessToken = accessToken
}

func (c AdobeAuthenticationContext) setExpiresIn(expiresIn int) {
	AdobeSession.ExpiresIn = expiresIn
}

func (c AdobeAuthenticationContext) setTokenType(tokenType string) {
	AdobeSession.TokenType = tokenType
}

func (c AdobeAuthenticationContext) setClientID(clientID string) {
	AdobeSession.ClientID = clientID
}

func (c AdobeAuthenticationContext) GetAccessToken() string {
	return AdobeSession.AccessToken
}

func (c AdobeAuthenticationContext) GetExpiresIn() int {
	return AdobeSession.ExpiresIn
}

func (c AdobeAuthenticationContext) GetTokenType() string {
	return AdobeSession.TokenType
}

func (c AdobeAuthenticationContext) GetClientID() string {
	return AdobeSession.ClientID
}
