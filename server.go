package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/google/uuid"
)

func runServer() {
	port := os.Args[2]
	address := fmt.Sprintf(":%s", port)

	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(address, nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}

	// Parse multipart form data
	if err := r.ParseMultipartForm(128 << 20); err != nil { // 128 MB max file size
		http.Error(w, fmt.Sprintf("Error parsing form data: %s", err), http.StatusBadRequest)
		return
	}

	requestUuid := uuid.NewString()
	requestDir := filepath.Join("data", "requests", requestUuid)
	uploadedFilePath := filepath.Join(requestDir, "uploaded")
	transcodedFilePath := filepath.Join(requestDir, "audio.mp3")
	wordTimestampsFilePath := filepath.Join(requestDir, "audio.json")
	verseKey := r.FormValue("verse_key")
	verseTextFilepath := filepath.Join("data", "verse-text", fmt.Sprintf("%s.txt", verseKey))

	if err := os.MkdirAll(requestDir, 0755); err != nil {
		http.Error(w, fmt.Sprintf("Error creating directory: %s", err), http.StatusInternalServerError)
		return
	}

	// Get uploaded file
	uploadedAudioFile, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving uploaded file: %s", err), http.StatusBadRequest)
		return
	}
	defer uploadedAudioFile.Close()

	// Save uploaded file
	if err := saveFile(uploadedFilePath, uploadedAudioFile); err != nil {
		http.Error(w, fmt.Sprintf("Error saving uploaded file: %s", err), http.StatusInternalServerError)
		return
	}

	// Transcode
	if err := transcode(uploadedFilePath, transcodedFilePath); err != nil {
		http.Error(w, fmt.Sprintf("Error processing uploaded file: %s", err), http.StatusInternalServerError)
		return
	}

	// Generate timestamps
	generateWordTimestamps(transcodedFilePath, verseTextFilepath)

	// Return audio JSON
	jsonData, err := readJSONData(wordTimestampsFilePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading JSON data: %s", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(jsonData)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error returning JSON data: %s", err), http.StatusInternalServerError)
		return
	}

	err = os.RemoveAll(requestDir)
	if err != nil {
		log.Printf("Error while deleting directory: %v", err)
	}
}

func saveFile(filePath string, file io.Reader) error {
	localFile, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer localFile.Close()

	// Copy the file content to the destination
	_, err = io.Copy(localFile, file)
	if err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}
	return nil
}

func transcode(inputPath, outputPath string) error {
	cmd := exec.Command("ffmpeg",
		"-i", inputPath,
		outputPath,
	)

	return cmd.Run()
}

func generateWordTimestamps(audioPath string, textPath string) error {
	cmd := exec.Command("ctc-forced-aligner", "--audio_path", audioPath, "--text_path", textPath, "--language", "\"ara\"", "--romanize", "--preserve_split", "True")
	_, err := cmd.Output()
	if err != nil {
		return err
	}

	return nil
}

func readJSONData(filePath string) ([]byte, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return data, nil
}
