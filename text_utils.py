from pydantic import BaseModel
from ctc_forced_aligner import text_normalize, get_uroman_tokens
from ctc_forced_aligner.alignment_utils import Segment as Span

from .config import config


class Segment(BaseModel):
    type: str
    key: str
    start: int
    end: int


def preprocess_text(
    text: list[str], romanize: bool, language: str
) -> tuple[list[str], list[str]]:

    normalized_text: list[str] = [text_normalize(segment, language) for segment in text]

    if romanize:
        tokens: list[str] = get_uroman_tokens(normalized_text, language)
    else:
        tokens = [" ".join(list(word)) for word in normalized_text]

    return tokens, text


def postprocess_results(
    text_starred: list[str],
    spans: list[list[Span]],
    stride: float,
    word_segments: list[str],
):
    results: list[Segment] = []

    segment_idx = 0
    for idx, text in enumerate(text_starred):
        if text == "<star>":
            continue

        span = spans[idx]
        span_start = span[0].start
        span_end = span[-1].end
        start_ms = int(span_start * (stride))
        end_ms = int(span_end * (stride))

        segment_key = word_segments[segment_idx]

        segment_type = "word"
        if segment_key in [config.taawwudh, config.basmalah]:
            segment_type = "phrase"

        results.append(
            Segment(type=segment_type, key=segment_key, start=start_ms, end=end_ms)
        )

        segment_idx += 1

    return results
