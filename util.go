package main

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"git.sr.ht/~rehandaphedar/lafzize/v3/pkg/api"
)

type AlignerOutput struct {
	Segments []AlignerSegment
}

type AlignerSegment struct {
	Start float64 `json:"start"`
	End   float64 `json:"end"`
	Score float64 `json:"score"`
}

func generateSegments(requestSegments []string) ([]string, []string) {
	verseKeySegments := []string{}
	wordKeySegments := []string{}
	verseKeys := api.GetVerseKeys(data)
	wordKeys := api.GetWordKeys(data)

	for _, requestSegment := range requestSegments {
		verseKeyRange := strings.Split(requestSegment, ",")
		startVerseKey := verseKeyRange[0]
		endVerseKey := verseKeyRange[1]

		startWordKey := fmt.Sprintf("%s:%d", startVerseKey, 1)
		endWordKey := fmt.Sprintf("%s:%d", endVerseKey, len(data.Verses[endVerseKey].Words)-1)

		verseKeySegments = append(verseKeySegments, sliceBetween(verseKeys, startVerseKey, endVerseKey)...)
		wordKeySegments = append(wordKeySegments, sliceBetween(wordKeys, startWordKey, endWordKey)...)
	}

	return verseKeySegments, wordKeySegments
}

func sliceBetween[T comparable](slice []T, start, end T) []T {
	startIdx := slices.Index(slice, start)
	endIdx := slices.Index(slice, end)

	if startIdx == -1 || endIdx == -1 {
		return []T{}
	}
	if startIdx > endIdx {
		return []T{}
	}

	return slice[startIdx : endIdx+1]
}

func convertAlignerOutput(alignerOutput AlignerOutput, wordSegments []string) ([][]int, error) {
	outputSegments := [][]int{}
	for idx, alignerOutputSegment := range alignerOutput.Segments {
		segmentNumber := idx + 1

		wordKey := wordSegments[idx]
		wordKeySegments := strings.Split(wordKey, ":")

		chapterNumber, err := strconv.Atoi(wordKeySegments[0])
		if err != nil {
			return nil, err
		}
		verseNumber, err := strconv.Atoi(wordKeySegments[1])
		if err != nil {
			return nil, err
		}
		wordNumber, err := strconv.Atoi(wordKeySegments[2])
		if err != nil {
			return nil, err
		}

		startMs := int(alignerOutputSegment.Start * 1000)
		endMs := int(alignerOutputSegment.End * 1000)

		outputSegments = append(outputSegments,
			[]int{segmentNumber,
				startMs,
				endMs,
				chapterNumber,
				verseNumber,
				wordNumber})
	}

	return outputSegments, nil
}
