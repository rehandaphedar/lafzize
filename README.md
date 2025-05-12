# Introduction

A program to generate word level/word by word timestamps of Qurʾān recitations.

Unfortunately, there is no public demo/instance at this point. I would host it on my VPS, but it would not be able to handle running AI models.

# Dependencies

- `ffmpeg` should be in `$PATH`.
- [ctc-forced-aligner](https://github.com/MahmoudAshraf97/ctc-forced-aligner) should be in `$PATH`.

# Usage

Send a POST request to the API endpoint with the following data:
- `file`: The audio file of the recitation.
- `verse_key`: The verse key (`[chapter_number]:[verse_number]`) of the recitation.

Example using bash:

```sh
curl \
	-X POST \
	-F "file=@001001.mp3" \
	-F "verse_key=1:1" \
		"http://localhost:8004"
```

Example response:

```json
{
  "text": "بِسْمِ ٱللَّهِ ٱلرَّحْمَـٰنِ ٱلرَّحِيمِ",
  "segments": [
    {
      "start": 0.28,
      "end": 0.86,
      "text": "بِسْمِ",
      "score": -1.6092441082000732
    },
    {
      "start": 0.86,
      "end": 1.28,
      "text": "ٱللَّهِ",
      "score": -14.186538696289062
    },
    {
      "start": 1.28,
      "end": 3.32,
      "text": "ٱلرَّحْمَـٰنِ",
      "score": -10.867626190185547
    },
    {
      "start": 3.32,
      "end": 5.52,
      "text": "ٱلرَّحِيمِ",
      "score": -17.079059600830078
    }
  ]
}
```

# Deploying

## Installation

```sh
go install git.sr.ht/~rehandaphedar/lafzize/v2@latest
```

## Fetching Verse Text Data

Verse text data is fetched from [the Quran Foundation API](https://api-docs.quran.foundation). Thus, it requires `LAFZIZE_CLIENT_ID` and `LAFZIZE_CLIENT_SECRET` environment variables to be set. To obtain these, visit [the Request Access page](https://api-docs.quran.foundation/request-access) and fill out the form. It takes around 48-72 hours to get approved.

Before running the program for the first time, run:

```sh
lafzize fetch
```


This will create a `data` folder in the current directory, and fetch verse text data from [the Quran.com API](https://api-docs.quran.com/) into `data/verse-text`.

## Running

To run the program afterwards:
```sh
lafzize server 8004
```

This will run the server on port 8004 (which can be changed).
