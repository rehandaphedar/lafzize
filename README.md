# Introduction

A program to generate word level/word by word timestamps of Qurʾān recitations.

Unfortunately, there is no public demo/instance at this point. I would host it on my VPS, but it would not be able to handle running AI models.

# Dependencies

- `ffmpeg` should be in `$PATH`.
- [ctc-forced-aligner](https://github.com/MahmoudAshraf97/ctc-forced-aligner) should be in `$PATH`.

# Usage

Send a POST request to the API endpoint with the following data:
- `file`: The audio file of the recitation.
- `segments`: The range of verses recited in the format `[start_verse_key]:[end_verse_key]`.

A verse key is of the format `[chapter_number]:[verse_number]`.

## Examples

For an audio file containing the recitation of Sūrah Al Fātiḥah:

```sh
curl \
	-X POST \
	-F "file=@001.mp3" \
	-F "segments=1:1,1:7" \
		"http://localhost:8004"
```

If only one verse is recited, `start_verse_key` and `end_verse_key` should be the same.

```sh
curl \
	-X POST \
	-F "file=@001001.mp3" \
	-F "segments=1:1,1:1" \
		"http://localhost:8004"
```

You can pass multiple `segments` values. They will be evaluated in the order they are passed, and joined into a single list of verses. This is helpful in case of recitations with some verses missing.

```sh
curl \
	-X POST \
	-F "file=@085.mp3" \
	-F "segments=85:1,85:10" \
	-F "segments=85:13,85:18" \
	-F "segments=85:20,85:22" \
		"http://localhost:8004"
```

Example response:

```json
{
  "segments": [
    [
      1,280,840,1,1,1
    ],
    [
      2,840,1800,1,1,2
    ],
    [
      3,1800,3340,1,1,3
    ],
    [
      4,3340,6060,1,1,4
    ],
    [
      5,6060,7280,1,2,1
    ],
    [
      6,7280,8200,1,2,2
    ],
    [
      7,8200,8740,1,2,3
    ],
    [
      8,8740,11320,1,2,4
    ],
  ]
}
```

Each segment is an array of the format:

```
[
	segmentNumber, startMs, endMs, chapterNumber, verseNumber, wordNumber
]
```

- `segmentNumber`: The word number *in the context of the given verse range*
- `startMs`: The start time of the word in milliseconds
- `endMs`: The end time of the word in milliseconds
- `chapterNumber`: The chapter number of the word
- `verseNumber`: The verse number of the word
- `wordNumber`: The word number *in the chapter* of the word being recited

For example, consider that the recording provided contains verses `1:3-1:5`.

The first word of verse `1:3` will have:
- `segmentNumber` = 1
- `wordNumber` = 1

The first word of verse `1:4` will have:
- `segmentNumber` = 3
- `wordNumber` = 1

This format is compatible with [Quranic Universal Library's format](https://qul.tarteel.ai/docs/with-segments), also used by the Quran Foundation API.

# Deploying

## Installation

```sh
go install git.sr.ht/~rehandaphedar/lafzize/v3@latest
```

## Fetching Verse Text Data

Verse text data is fetched from [the Quran Foundation API](https://api-docs.quran.foundation). It requires `client_id` and `client_secret` tokens. To obtain these, visit [the Request Access page](https://api-docs.quran.foundation/request-access) and fill out the form. It takes around 48-72 hours to get approved.

Before running the program for the first time, run:

```sh
lafzize api -client_id [client_id] -client_secret [client_secret]
```

This will save the data in `data.json`. You can run `lafzize api -h` to see more options on customizing the output location.

## Running

Run `lafzize server`. See `lafzize server -h` for more options.

Possible values of `-device` option:
```
cpu, cuda, ipu, xpu, mkldnn, opengl, opencl, ideep, hip, ve, fpga, maia, xla, lazy, vulkan, mps, meta, hpu, mtia, privateuseone
```
