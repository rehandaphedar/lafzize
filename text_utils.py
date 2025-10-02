from ctc_forced_aligner import text_normalize, get_uroman_tokens
from ctc_forced_aligner.alignment_utils import Segment

from .config import config


def preprocess_text(
    text: list[str], romanize: bool, language: str
) -> tuple[list[str], list[str]]:

    normalized_text: list[str] = [text_normalize(segment, language) for segment in text]

    if romanize:
        tokens = get_uroman_tokens(normalized_text, language)
    else:
        tokens = [" ".join(list(word)) for word in normalized_text]

    tokens_starred: list[str] = []
    text_starred: list[str] = []

    for token, chunk in zip(tokens, text):
        tokens_starred.extend(["<star>", token])
        text_starred.extend(["<star>", chunk])

    return tokens_starred, text_starred


def postprocess_results(
    text_starred: list[str],
    spans: list[list[Segment]],
    stride: float,
    word_segments: list[str],
):
    results: dict[str, list[list[int]]] = {"segments": []}

    segment_idx = 0
    segment_number = 1
    for idx, text in enumerate(text_starred):
        if text == "<star>":
            continue

        span = spans[idx]
        span_start = span[0].start
        span_end = span[-1].end
        start_ms = int(span_start * (stride))
        end_ms = int(span_end * (stride))

        chapter_number = 0
        verse_number = 0
        word_number = 0
        phrase = 0

        word_key = word_segments[segment_idx]

        if word_key == config.taawwudh:
            phrase = 1
        elif word_key == config.basmalah:
            phrase = 2
        else:
            word_key_segments = word_key.split(":")

            chapter_number = int(word_key_segments[0])
            verse_number = int(word_key_segments[1])
            word_number = int(word_key_segments[2])

        results["segments"].append(
            [
                segment_number,
                start_ms,
                end_ms,
                chapter_number,
                verse_number,
                word_number,
                phrase,
            ]
        )

        segment_idx += 1
        segment_number += 1

    return results
