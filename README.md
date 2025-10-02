# Introduction

A program to generate word level/word by word timestamps of Qurʾān recitations.

Unfortunately, there is no public demo/instance at this point. I would host it on my VPS, but it would not be able to handle running AI models.

# Usage

Send a POST request to the API endpoint with the following data:
- `audio`: The audio file of the recitation.
- `segments`: The range of verses recited in the format `[start_verse_key]:[end_verse_key]`.

A verse key is of the format `[chapter_number]:[verse_number]`.

There are some special segments that do not represent words from a verse:

- `taawwudh`: Represents Taʿawwudh
- `basmalah`: Represents Basmalah

There is also OpenAPI documentation available at `/docs` and `/redoc`.

## Examples

For an audio file containing the recitation of Sūrah Al Fātiḥah preceded by Taʿawwudh:

```sh
curl \
	-X POST \
	-F "audio=@001.mp3" \
	-F "segments=taawwudh" \
	-F "segments=1:1,1:7" \
		"http://localhost:8004"
```

If only one verse is recited, `start_verse_key` and `end_verse_key` should be the same.

```sh
curl \
	-X POST \
	-F "audio=@001001.mp3" \
	-F "segments=taawwudh" \
	-F "segments=1:1,1:1" \
		"http://localhost:8004"
```

You can pass multiple `segments` values. They will be evaluated in the order they are passed, and joined into a single list of verses. This is helpful in case of recitations with some verses missing.

```sh
curl \
	-X POST \
	-F "audio=@085.mp3" \
	-F "segments=taawwudh" \
	-F "segments=basmalah" \
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
      1,60,1200,0,0,0,1
    ],
    [
      2,1240,1820,0,0,0,2
    ],
    [
      3,1860,1980,85,1,1,0
    ],
    [
      4,2060,2160,85,1,2,0
    ],
    [
      5,2200,3220,85,1,3,0
    ],
    [
      6,3780,4340,85,2,1,0
    ],
    [
      7,4460,5760,85,2,2,0
    ],
    [
      8,6300,7540,85,3,1,0
    ],
    [
      9,7580,9180,85,3,2,0
    ]
}
```

Each segment is an array of the format:

```
[
	segmentNumber, startMs, endMs, chapterNumber, verseNumber, wordNumber, specialSegmentType
]
```

- `segmentNumber`: The word number *in the context of the given verse range*
- `startMs`: The start time of the word in milliseconds
- `endMs`: The end time of the word in milliseconds
- `chapterNumber`: The chapter number of the word
- `verseNumber`: The verse number of the word
- `wordNumber`: The word number *in the chapter* of the word being recited
- `specialSegmentType`: `0` -> No special segment (The segment is a word from a verse). `1` -> Taʿawwudh. `2` -> Basmalah.

If `specialSegmentType` is not 0, `chapterNumber`, `verseNumber`, and `wordNumber` will all be set to 0 and should be ignored.

For example, consider that the recording provided contains Taʿawwudh followed by verses `1:3-1:5`.

The first segment will have:
- `segmentNumber` = 1
- `chapterNumber` = 0
- `verseNumber` = 0
- `wordNumber` = 0

The first word of verse `1:3` will have:
- `segmentNumber` = 2
- `chapterNumber` = 1
- `verseNumber` = 3
- `wordNumber` = 1

The first word of verse `1:4` will have:
- `segmentNumber` = 4
- `chapterNumber` = 1
- `verseNumber` = 4
- `wordNumber` = 1

This format is compatible with [Quranic Universal Library's format](https://qul.tarteel.ai/docs/with-segments), also used by the Quran Foundation API.

# Deployment

Clone the repository:
```sh
git clone https://git.sr.ht/~rehandaphedar/lafzize
cd lafzize
```

Create a virtualenv if you want:
```sh
python -m venv venv
source venv/bin/activate
```

Install dependencies:
```sh
pip install -r requirements.txt
```

Note that you may need to change the requirements depending on:
- whether you need CUDA/XPU/MPS specific Torch versions
- the deployment method you want to use `fastapi run` vs `uvicorn` (with or without `uvloop`) vs `gunicorn` etc.

Obtain `data.json` and `data_extra.json` from [qf-cache](https://sr.ht/~rehandaphedar/qf-client-golang/#caching).

Then, run the server using your preferred deployment method:
```sh
fastapi run --port 8004
```

# Configuration

Configuration is done through the following environment variables:

- `LAFZIZE_MODEL`: Model to use for inference. Default: `MahmoudAshraf/mms-300m-1130-forced-aligner`.
- `LAFZIZE_DEVICE`: Device to use for inference. Default: `"cuda"` if available, else `"cpu"`.
- `LAFZIZE_DTYPE`: Torch dtype to use for inference. Default: `float16` if device is CUDA, else `float32`.
- `LAFZIZE_BATCH_SIZE`: Batch size for inference. Default: `4`.
- `LAFZIZE_WINDOW_SIZE`: Window size in seconds for audio chunking. Default: `30`.
- `LAFZIZE_CONTEXT_SIZE`: Overlap between chunks in seconds. Default: `2`.
- `LAFZIZE_ROMANIZE`: Whether to enable romanization for non latin scripts, or for multilingual models regardless of the language. Required when using the default model. Default: `True`.

Details regarding the above options can be found in the [ctc-forced-aligner](https://github.com/MahmoudAshraf97/ctc-forced-aligner) repository.

Additionally, there are some lafzize specific options:
- `LAFZIZE_MAX_UPLOAD_SIZE`: Maximum allowed upload size in MB. Default: `128`.
- `LAFZIZE_TAAWWUDH`: The segment to interpret as Taʿawwudh. Default: `"taawwudh"`.
- `LAFZIZE_BASMALAH`: The segment to interpret as Basmalah. Default: `"basmalah"`.
- `LAFZIZE_DATA`: Path to `data`. Default: `"data.json"`.
- `LAFZIZE_DATA_EXTRA`: Path to `data_extra`. Default: `"data_extra.json"`.
