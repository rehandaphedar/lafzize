import msgspec
from .models import Chapter, Verse


class API(msgspec.Struct):
    chapters: list[Chapter]
    verses: dict[str, Verse]


class APIExtraChapter(msgspec.Struct):
    id: int
    code_v2: str


class APIExtraJuz(msgspec.Struct):
    id: int
    code_v2: str


class APIExtraMisc(msgspec.Struct):
    taawwudh: str
    basmalah: str
    taawwudh_code_v2: str
    basmalah_normal_code_v2: str
    basmalah_idgham_code_v2: str


class APIExtra(msgspec.Struct):
    misc: APIExtraMisc
    chapters: list[APIExtraChapter]
    juzs: list[APIExtraJuz]


def get_verse_keys(data: API):
    verse_keys: list[str] = []
    for chapter in data.chapters:
        chapter.verses_count

        assert chapter.verses_count is not None
        for idx in range(chapter.verses_count):
            verse_number = idx + 1
            verse_keys.append(f"{chapter.id}:{verse_number}")
    return verse_keys


def get_word_keys(data: API):
    word_keys: list[str] = []
    for chapter in data.chapters:

        assert chapter.verses_count is not None
        for verse_idx in range(chapter.verses_count):
            verse_number = verse_idx + 1
            words = data.verses[f"{chapter.id}:{verse_number}"].words

            assert words is not None
            for word_idx in range(len(words) - 1):
                word_number = word_idx + 1
                word_keys.append(f"{chapter.id}:{verse_number}:{word_number}")
    return word_keys
