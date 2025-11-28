# Introduction

A program to generate word level/word by word timestamps of Qurʾān recitations.

Unfortunately, there is no public demo/instance at this point. I would host it on my VPS, but it would not be able to handle running AI models.

# Usage

Send a POST request to the API endpoint with the following data:
- `audio`: The audio file of the recitation.
- `segments`: The range of verses recited in the format `[start_verse_key]:[end_verse_key]`.

A verse key is of the format `[chapter_number]:[verse_number]`.

There are some special segments that do not represent words from a verse:

- `taawwudh`: Represents Taʿawwudh.
- `basmalah`: Represents Basmalah.

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
[
  {
    "type": "phrase",
    "key": "taawwudh",
    "start": 320,
    "end": 1220
  },
  {
    "type": "phrase",
    "key": "basmalah",
    "start": 1280,
    "end": 1820
  },
  {
    "type": "word",
    "key": "85:1:1",
    "start": 1860,
    "end": 1980
  },
  {
    "type": "word",
    "key": "85:1:2",
    "start": 2060,
    "end": 2160
  },
  {
    "type": "word",
    "key": "85:1:3",
    "start": 2200,
    "end": 3220
  }
]
```

Each segment is with the following keys:
- `type`: The type of the segment (`word` or `phrase`).
- `key`: The key of the segment (`word_key` for `word`, either `taawwudh` or `basmalah` for `phrase`).
- `start`: The start time of the segment in milliseconds.
- `end`: The end time of the segment in milliseconds.

# Deployment

Firstly, make sure you have `ffmpeg` installed.

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
- whether you need CUDA/XPU/MPS specific Torch versions.
- the deployment method you want to use `fastapi run` vs `uvicorn` (with or without `uvloop`) vs `gunicorn` etc.


Obtain `qpc-hafs-word-by-word.json` from [QUL](https://qul.tarteel.ai/resources/quran-script).
Obtain `quran-metadata-misc.json` from [quranic-universal-library-extras](https://sr.ht/~rehandaphedar/quranic-universal-library-extras).

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
- `LAFZIZE_WORDS`: Path to words. Default: `"qpc-hafs-word-by-word.json"`.
- `LAFZIZE_METADATA`: Path to metadata. Default: `"quran-metadata-misc.json"`.
