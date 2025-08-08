package api

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"golang.org/x/oauth2/clientcredentials"
)

type API struct {
	Chapters []Chapter        `json:"chapters"`
	Verses   map[string]Verse `json:"verses"`
}

type Chapter struct {
	ID              int            `json:"id"`
	RevelationPlace string         `json:"revelation_place"`
	RevelationOrder int            `json:"revelation_order"`
	BismillahPre    bool           `json:"bismillah_pre"`
	NameSimple      string         `json:"name_simple"`
	NameComplex     string         `json:"name_complex"`
	NameArabic      string         `json:"name_arabic"`
	VersesCount     int            `json:"verses_count"`
	Pages           []int          `json:"pages"`
	TranslatedName  TranslatedName `json:"translated_name"`
	CodeV2          string         `json:"code_v2,omitempty"` // Extra
}

type Verse struct {
	HizbNumber      int                `json:"hizb_number"`
	ID              int                `json:"id"`
	JuzNumber       int                `json:"juz_number"`
	ManzilNumber    int                `json:"manzil_number"`
	PageNumber      int                `json:"page_number"`
	RubElHizbNumber int                `json:"rub_el_hizb_number"`
	RukuNumber      int                `json:"ruku_number"`
	SajdahNumber    *int               `json:"sajdah_number"` // Pointer to handle null
	VerseKey        string             `json:"verse_key"`
	VerseNumber     int                `json:"verse_number"`
	Words           []Word             `json:"words"`
	Translations    []VerseTranslation `json:"translations"`
}

type TranslatedName struct {
	LanguageName string `json:"language_name"`
	Name         string `json:"name"`
}

type Word struct {
	AudioURL        *string             `json:"audio_url"` // Pointer to handle null
	CharTypeName    string              `json:"char_type_name"`
	ID              int                 `json:"id"`
	LineNumber      int                 `json:"line_number"`
	PageNumber      int                 `json:"page_number"`
	Position        int                 `json:"position"`
	Text            string              `json:"text"`
	TextUthmani     string              `json:"text_uthmani"`
	Translation     WordTranslation     `json:"translation"`
	Transliteration WordTransliteration `json:"transliteration"`
}

type VerseTranslation struct {
	ResourceID   int    `json:"resource_id"`
	ResourceName string `json:"resource_name"`
	ID           int    `json:"id"`
	Text         string `json:"text"`
	VerseID      int    `json:"verse_id"`
	LanguageID   int    `json:"language_id"`
	LanguageName string `json:"language_name"`
	VerseKey     string `json:"verse_key"`
	ChapterID    int    `json:"chapter_id"`
	VerseNumber  int    `json:"verse_number"`
	JuzNumber    int    `json:"juz_number"`
	HizbNumber   int    `json:"hizb_number"`
	RubNumber    int    `json:"rub_number"`
	PageNumber   int    `json:"page_number"`
}

type WordTranslation struct {
	LanguageName string `json:"language_name"`
	Text         string `json:"text"`
}

type WordTransliteration struct {
	LanguageName string  `json:"language_name"`
	Text         *string `json:"text"` // Pointer to handle null
}

type APIResponseVerses struct {
	Verses     []Verse               `json:"verses"`
	Pagination APIResponsePagination `json:"pagination"`
}

type APIResponseChapters struct {
	Chapters []Chapter `json:"chapters"`
}

type APIResponsePagination struct {
	PerPage      int `json:"per_page"`
	CurrentPage  int `json:"current_page"`
	NextPage     int `json:"next_page"`
	TotalPages   int `json:"total_pages"`
	TotalRecords int `json:"total_records"`
}

func RunApiCommand(args []string) {
	apiFlagSet := flag.NewFlagSet("api", flag.ExitOnError)

	clientID := apiFlagSet.String("client_id", "", "Quran API Client ID")
	clientSecret := apiFlagSet.String("client_secret", "", "Quran API Client Secret")
	output := apiFlagSet.String("output", "data.json", "Path to JSON file to write output to")

	err := apiFlagSet.Parse(args)
	if err != nil {
		apiFlagSet.Usage()
		os.Exit(1)
	}

	config := clientcredentials.Config{
		ClientID:     *clientID,
		ClientSecret: *clientSecret,
		TokenURL:     "https://oauth2.quran.foundation/oauth2/token",
		Scopes:       []string{"content"},
	}
	tokenSource := config.TokenSource(context.Background())

	log.Println("Fetching Chapter data")

	url := "https://apis.quran.foundation/content/api/v4/chapters"
	body, err := FetchWithCredentials(url, *clientID, tokenSource)
	if err != nil {
		log.Fatalf("error fetching data from Quran.com API: %v\n", err)
	}

	var chapters APIResponseChapters
	err = json.Unmarshal(body, &chapters)
	if err != nil {
		log.Fatalf("error unmarshaling JSON: %v\n", err)
	}

	var verses []Verse
	for idx := range 30 {
		juzNumber := idx + 1
		nextPage := 1

		log.Printf("Processing Juz %d\n", juzNumber)

		for nextPage > 0 {
			log.Printf("Processing Page %d\n", nextPage)

			url := fmt.Sprintf("https://apis.quran.foundation/content/api/v4/verses/by_juz/%d?words=true&per_page=50&page=%d&word_fields=code_v2,text_uthmani&translations=20", juzNumber, nextPage)
			body, err := FetchWithCredentials(url, *clientID, tokenSource)
			if err != nil {
				log.Fatalf("error fetching data from Quran.com API: %v\n", err)
			}

			var data APIResponseVerses
			err = json.Unmarshal(body, &data)
			if err != nil {
				log.Fatalf("error unmarshaling JSON: %v\n", err)
			}

			nextPage = data.Pagination.NextPage
			verses = append(verses, data.Verses...)
		}
	}

	var data API
	data.Chapters = chapters.Chapters
	data.Verses = make(map[string]Verse)
	for _, verse := range verses {
		data.Verses[verse.VerseKey] = verse
	}

	dataJSON, err := json.Marshal(data)
	if err != nil {
		log.Fatalf("error marshaling data to JSON: %v\n", err)
	}

	err = os.WriteFile(*output, dataJSON, 0644)
	if err != nil {
		log.Fatalf("error writing to JSON file: %v\n", err)
	}
}
