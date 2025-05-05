package auth

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
)

func (i *AdobeAuthInput) Authenticate() (*AdobeAuthenticationContext, error) {
	response, err := http.PostForm(
		"https://pdf-services.adobe.io/token",
		url.Values{
			"client_id":     {i.ClientID},
			"client_secret": {i.ClientSecret},
		},
	)
	if err != nil {
		log.Fatal(err)
		return &AdobeAuthenticationContext{}, errors.New(err.Error())
	}

	defer response.Body.Close()

	// Read response
	data, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	var jsonData AdobeAuthenticationContext
	err = json.Unmarshal(data, &jsonData)
	if err != nil {
		log.Fatal(err)
	}

	// Set the session globally
	AdobeSession.setAccessToken(jsonData.AccessToken)
	AdobeSession.setExpiresIn(jsonData.ExpiresIn)
	AdobeSession.setTokenType(jsonData.TokenType)
	AdobeSession.setClientID(i.ClientID)

	return &jsonData, nil
}
