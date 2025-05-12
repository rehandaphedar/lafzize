package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func fetchAccessToken(clientID, clientSecret string) (string, error) {
	authString := base64.StdEncoding.EncodeToString(fmt.Appendf(nil, "%s:%s", clientID, clientSecret))

	tokenURL := "https://oauth2.quran.foundation/oauth2/token"

	values := url.Values{}
	values.Set("scope", "content")
	values.Set("grant_type", "client_credentials")

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(values.Encode()))
	if err != nil {
		return "", fmt.Errorf("Error while creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", authString))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Error while sending request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Error while reading response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Error in response with status %d: %s", resp.StatusCode, string(body))
	}

	var token struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}

	if err := json.Unmarshal(body, &token); err != nil {
		return "", fmt.Errorf("Error while parsing token response: %v", err)
	}

	return token.AccessToken, nil
}
