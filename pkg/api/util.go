package api

import (
	"fmt"
	"io"
	"net/http"

	"golang.org/x/oauth2"
)

func FetchWithCredentials(url, clientID string, tokenSource oauth2.TokenSource) ([]byte, error) {
	token, err := tokenSource.Token()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("x-auth-token", token.AccessToken)
	req.Header.Set("x-client-id", clientID)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func GetVerseKeys(data API) []string {
	verseKeys := []string{}
	for _, chapter := range data.Chapters {
		for idx := range chapter.VersesCount {
			verseNumber := idx + 1
			verseKeys = append(verseKeys, fmt.Sprintf("%d:%d", chapter.ID, verseNumber))
		}
	}
	return verseKeys
}

func GetWordKeys(data API) []string {
	wordKeys := []string{}
	for _, chapter := range data.Chapters {
		for verseIdx := range chapter.VersesCount {
			verseNumber := verseIdx + 1
			words := data.Verses[fmt.Sprintf("%d:%d", chapter.ID, verseNumber)].Words
			for wordIdx := range words[:len(words)-1] {
				wordNumber := wordIdx + 1
				wordKeys = append(wordKeys, fmt.Sprintf("%d:%d:%d", chapter.ID, verseNumber, wordNumber))
			}
		}
	}
	return wordKeys
}
