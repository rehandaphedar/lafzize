package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type APIResponse struct {
	Verses     []Verse    `json:"verses"`
	Pagination Pagination `json:"pagination"`
}

type Verse struct {
	VerseKey string `json:"verse_key"`
	Words    []Word `json:"words"`
}

type Word struct {
	TextUthmani string `json:"text_uthmani"`
}

type Pagination struct {
	NextPage int `json:"next_page"`
}

func fetchVerseText() {
	clientID := os.Getenv("LAFZIZE_CLIENT_ID")
	clientSecret := os.Getenv("LAFZIZE_CLIENT_SECRET")

	accessToken, err := fetchAccessToken(clientID, clientSecret)

	if err != nil {
		log.Fatalf("Error fetching access token: %v", err)
	}

	rootDirectory := filepath.Join("data", "verse-text")
	os.MkdirAll(rootDirectory, 0755)

	for juzNumber := 1; juzNumber <= 30; juzNumber++ {
		log.Printf("Processing Juz %d\n", juzNumber)

		var page int = 1
		for page != 0 {
			log.Printf("Processing Page %d\n", page)

			url := fmt.Sprintf("https://api.quran.com/api/v4/verses/by_juz/%d?words=true&word_fields=text_uthmani&per_page=50&page=%d", juzNumber, page)
			method := "GET"

			client := &http.Client{}
			req, err := http.NewRequest(method, url, nil)

			if err != nil {
				fmt.Println(err)
				return
			}
			req.Header.Add("Accept", "application/json")
			req.Header.Add("x-client-id", clientID)
			req.Header.Add("x-auth-token", accessToken)

			res, err := client.Do(req)
			if err != nil {
				fmt.Println(err)
				return
			}
			defer res.Body.Close()

			body, err := io.ReadAll(res.Body)
			if err != nil {
				fmt.Println(err)
				return
			}

			var apiResponse APIResponse
			json.Unmarshal(body, &apiResponse)

			for _, verse := range apiResponse.Verses {
				log.Printf("Processing Verse %s", verse.VerseKey)

				verseText := ""
				words := verse.Words
				words = words[:len(words)-1]
				for _, word := range words {
					verseText += word.TextUthmani + "\n"
				}

				filename := filepath.Join(rootDirectory, fmt.Sprintf("%s.txt", verse.VerseKey))

				file, err := os.Create(filename)
				if err != nil {
					log.Fatalf("Error creating file: %v", err)
					return
				}
				defer file.Close()

				_, err = file.WriteString(verseText)
				if err != nil {
					log.Fatalf("Error creating file: %v", err)
					return
				}
			}
			page = apiResponse.Pagination.NextPage
		}

	}

}
