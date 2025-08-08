package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"git.sr.ht/~rehandaphedar/lafzize/v3/pkg/api"
	"github.com/google/uuid"
)

type APIRequestSegment struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type APIResponse struct {
	Segments [][]int `json:"segments"`
}

var data api.API
var maxUploadSize *int
var device *string

func runServerCommand(args []string) {
	serverFlagSet := flag.NewFlagSet("server", flag.ExitOnError)

	port := serverFlagSet.Int("port", 8004, "Port to listen on")
	maxUploadSize = serverFlagSet.Int("max_upload_size", 128, "Maximum size in MB of uploaded file")
	dataPath := serverFlagSet.String("data", "data.json", "Path to file containing API data")
	device = serverFlagSet.String("device", "cpu", "Device to use")

	if err := serverFlagSet.Parse(args); err != nil {
		serverFlagSet.Usage()
		os.Exit(1)
	}

	dataFile, err := os.ReadFile(*dataPath)
	if err != nil {
		log.Fatalf("error reading file: %v\n", err)
	}

	if err := json.Unmarshal(dataFile, &data); err != nil {
		log.Fatalf("error unmarshaling JSON: %v\n", err)
	}

	address := fmt.Sprintf(":%d", *port)

	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(address, nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}

	if err := r.ParseMultipartForm(int64(*maxUploadSize) << 20); err != nil {
		http.Error(w, fmt.Sprintf("Error parsing form data: %s", err), http.StatusBadRequest)
		return
	}

	requestUuid := uuid.NewString()
	requestDir := filepath.Join("data", "requests", requestUuid)
	uploadedFilePath := filepath.Join(requestDir, "uploaded")
	transcodedFilePath := filepath.Join(requestDir, "audio.wav")
	timingsFilePath := filepath.Join(requestDir, "audio.json")
	wordsFilePath := filepath.Join(requestDir, "words.txt")

	if err := os.MkdirAll(requestDir, 0755); err != nil {
		http.Error(w, fmt.Sprintf("Error creating directory: %s", err), http.StatusInternalServerError)
		return
	}

	verseSegments, wordSegments := generateSegments(r.Form["segments"])

	if err := writeWordsFile(wordsFilePath, verseSegments); err != nil {
		http.Error(w, fmt.Sprintf("Error writing text file: %s", err), http.StatusBadRequest)
		return
	}

	uploadedFileMultipart, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving uploaded file: %s", err), http.StatusBadRequest)
		return
	}
	defer uploadedFileMultipart.Close()

	uploadedFile, err := os.Create(uploadedFilePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error saving uploaded file: %s", err), http.StatusInternalServerError)
		return
	}
	defer uploadedFile.Close()

	if _, err := io.Copy(uploadedFile, uploadedFileMultipart); err != nil {
		http.Error(w, fmt.Sprintf("Error copying uploaded file: %s", err), http.StatusInternalServerError)
		return
	}

	if err := transcode(uploadedFilePath, transcodedFilePath); err != nil {
		http.Error(w, fmt.Sprintf("Error processing uploaded file: %s", err), http.StatusInternalServerError)
		return
	}

	if err := generateWordTimestamps(transcodedFilePath, wordsFilePath); err != nil {
		http.Error(w, fmt.Sprintf("Error generating timestamps for uploaded file: %s", err), http.StatusInternalServerError)
		return
	}

	timingsBytes, err := os.ReadFile(timingsFilePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading aligner output: %s", err), http.StatusInternalServerError)
	}

	var timings AlignerOutput
	if err = json.Unmarshal(timingsBytes, &timings); err != nil {
		http.Error(w, fmt.Sprintf("Error unmarshaling aligner output: %s", err), http.StatusInternalServerError)
	}

	var response APIResponse
	response.Segments, err = convertAlignerOutput(timings, wordSegments)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error converting aligner output to response format: %s", err), http.StatusInternalServerError)
		return
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error marshaling response to JSON: %s", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err = w.Write(responseJSON); err != nil {
		http.Error(w, fmt.Sprintf("Error returning JSON data: %s", err), http.StatusInternalServerError)
		return
	}

	if err = os.RemoveAll(requestDir); err != nil {
		log.Printf("Error while deleting directory: %v\n", err)
	}
}

func writeWordsFile(textFilePath string, verseSegments []string) error {
	textFile, err := os.Create(textFilePath)
	if err != nil {
		return err
	}
	defer textFile.Close()

	writer := bufio.NewWriter(textFile)
	defer writer.Flush()

	for _, verseKey := range verseSegments {
		words := data.Verses[verseKey].Words
		for _, word := range words[:len(words)-1] {
			if _, err := writer.WriteString(word.TextUthmani + "\n"); err != nil {
				return err
			}
		}
	}

	return nil
}

func transcode(inputPath, outputPath string) error {
	cmd := exec.Command("ffmpeg",
		"-i", inputPath,
		"-ar", "16000",
		"-ac", "1",
		outputPath,
	)

	return cmd.Run()
}

func generateWordTimestamps(audioPath string, wordsFilePath string) error {
	cmd := exec.Command("ctc-forced-aligner",
		"--audio_path", audioPath,
		"--text_path", wordsFilePath,
		"--language", "\"ara\"",
		"--romanize",
		"--preserve_split", "True",
		"--device", *device)

	return cmd.Run()
}
