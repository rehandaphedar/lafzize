from collections import defaultdict
from pydantic import BaseModel, Field

from .config import config


class Word(BaseModel):
    id: int
    surah: str
    ayah: str
    word: str
    location: str
    text: str


class MetadataPhrases(BaseModel):
    taawwudh: str
    basmalah: str
    taawwudh_code_v2: str = Field(alias="taawwudh-code_v2")
    basmalah_normal_code_v2: str = Field(alias="basmalah_normal-code_v2")
    basmalah_idgham_code_v2: str = Field(alias="basmalah_idgham-code_v2")
    makkah: str = Field(alias="makkah-code_v4")
    madinah: str = Field(alias="madinah-code_v4")


class MetadataChapter(BaseModel):
    code_v2: str
    code_v4: str | None = None


class MetadataJuz(BaseModel):
    code_v2: str
    code_v4: str | None = None


class Metadata(BaseModel):
    phrases: MetadataPhrases
    chapters: list[MetadataChapter]
    juzs: list[MetadataJuz]


Words = dict[str, Word]
Verses = dict[str, list[str]]


def generate_verses(words: Words):
    verses: Verses = defaultdict(list)
    for word in sorted(words.values(), key=sort_word):
        verse_key = f"{word.surah}:{word.ayah}"
        verses[verse_key].append(word.text)
    return dict(verses)


def get_word_keys(words: Words):
    word_keys: list[str] = []
    for word in sorted(words.values(), key=sort_word):
        word_key = f"{word.surah}:{word.ayah}:{word.word}"
        word_keys.append(word_key)
    return word_keys


def get_verse_keys(words: Words):
    verse_keys: list[str] = []
    for word in sorted(words.values(), key=sort_word):
        if int(word.word) != 1:
            continue
        verse_key = f"{word.surah}:{word.ayah}"
        verse_keys.append(verse_key)
    return verse_keys


def sort_word(word: Word):
    return int(word.surah), int(word.ayah), int(word.word)


def generate_segments(request_segments: list[str], words: Words, verses: Verses):
    verse_key_segments: list[str] = []
    word_key_segments: list[str] = []

    verse_keys = get_verse_keys(words)
    word_keys = get_word_keys(words)

    for request_segment in request_segments:
        if request_segment in [config.taawwudh, config.basmalah]:
            verse_key_segments.append(request_segment)
            word_key_segments.append(request_segment)
            continue

        verse_key_range = request_segment.split(",")
        start_verse_key = verse_key_range[0]
        end_verse_key = verse_key_range[1]

        start_word_key = f"{start_verse_key}:1"
        end_word_key = f"{end_verse_key}:{ len(verses[end_verse_key])}"

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
