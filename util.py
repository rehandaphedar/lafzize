from fastapi import HTTPException

from .config import config
from .api import API, get_verse_keys, get_word_keys


def generate_segments(request_segments: list[str], data: API):
    verse_key_segments: list[str] = []
    word_key_segments: list[str] = []

    verse_keys = get_verse_keys(data)
    word_keys = get_word_keys(data)

    for request_segment in request_segments:
        if request_segment in [config.taawwudh, config.basmalah]:
            verse_key_segments.append(request_segment)
            word_key_segments.append(request_segment)
            continue

        verse_key_range = request_segment.split(",")
        start_verse_key = verse_key_range[0]
        end_verse_key = verse_key_range[1]

        words = data.verses[end_verse_key].words
        if words is None:
            raise HTTPException(
                status_code=500,
                detail=f"Error while processing words data.",
            )

        start_word_key = f"{start_verse_key}:1"
        end_word_key = f"{end_verse_key}:{ len(words) - 1}"

        verse_key_segments.extend(
            extract_between(verse_keys, start_verse_key, end_verse_key)
        )
        word_key_segments.extend(
            extract_between(word_keys, start_word_key, end_word_key)
        )

    return verse_key_segments, word_key_segments


def extract_between(input_list: list[str], start: str, end: str) -> list[str]:
    start_idx = input_list.index(start)
    end_idx = input_list.index(end)
    return input_list[start_idx : end_idx + 1]
