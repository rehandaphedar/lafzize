from contextvars import ContextVar
from datetime import datetime
import logging
import uuid
import msgspec
from typing_extensions import override

from fastapi import FastAPI, Request, UploadFile, HTTPException
from fastapi_msgspec.responses import MsgSpecJSONResponse
from contextlib import asynccontextmanager

from ctc_forced_aligner import (
    load_alignment_model,
    generate_emissions,
    get_alignments,
    get_spans,
)

from .config import config
from .alignment_utils import load_audio
from .text_utils import preprocess_text, postprocess_results
from .util import Words, Metadata, generate_verses, generate_segments


request_id_var = ContextVar("request_id", default=None)


class ContextInjector(logging.Filter):
    @override
    def filter(self, record: logging.LogRecord):
        timestamp = datetime.fromtimestamp(record.created).strftime(
            "%Y-%m-%d %H:%M:%S.%f"
        )
        request_id = request_id_var.get()
        message = record.getMessage()

        if not request_id is None:
            record.msg = f"{timestamp} | [{request_id}] {message}"
        else:
            record.msg = f"{timestamp} | {message}"

        record.args = ()

        return True


logger = logging.getLogger("uvicorn.error")
logger.addFilter(ContextInjector())


@asynccontextmanager
async def lifespan(_: FastAPI):
    global words, metadata, verses, model, tokenizer

    access_logger = logging.getLogger("uvicorn.access")
    access_logger.handlers.clear()
    access_logger.propagate = True

    logger.info(f"Starting server with config: {config}")

    logger.info("Loading data...")
    try:
        with open(config.words) as f:
            words = msgspec.json.decode(f.read(), type=Words)
            verses = generate_verses(words)
        with open(config.metadata) as f:
            metadata = msgspec.json.decode(f.read(), type=Metadata)
    except Exception as e:
        print(f"Error while loading data: {e}")
        raise
    logger.info("Loaded data.")

    logger.info(f"Loading model...")
    try:
        model, tokenizer = load_alignment_model(
            device=config.device,
            model_path=config.model,
            attn_implementation=None,
            dtype=config.dtype,
        )
    except Exception as e:
        print(f"Error while loading model: {e}")
        raise
    logger.info("Loaded model.")

    yield


app = FastAPI(lifespan=lifespan)


@app.post("/")
async def handler(audio: UploadFile, segments: list[str]):
    logger.info(
        f"Started processing request with audio {audio} and segments {segments}..."
    )

    if audio.size is None or audio.size > (config.max_upload_size << 20):
        raise HTTPException(
            status_code=400,
            detail=f"File size exceeds limit of {config.max_upload_size} MB.",
        )

    logger.info("Loading audio...")
    audio_waveform = await load_audio(audio, model.dtype, model.device)
    logger.info("Loaded audio.")

    logger.info("Generating emissions...")
    emissions, stride = generate_emissions(
        model=model,
        audio_waveform=audio_waveform,
        window_length=config.window_size,
        context_length=config.context_size,
        batch_size=config.batch_size,
    )
    logger.info("Generated emissions.")

    logger.info("Generating segments...")
    verse_segments, word_segments = generate_segments(segments, words, verses)
    logger.info("Generated segments.")

    logger.info("Generating text...")
    text: list[str] = []
    for verse_segment in verse_segments:
        if verse_segment == config.taawwudh:
            text.append(metadata.phrases.taawwudh)
            continue
        if verse_segment == config.basmalah:
            text.append(metadata.phrases.basmalah)
            continue

        for word in verses[verse_segment]:
            text.append(word)
    logger.info("Generated text.")

    logger.info("Preprocessing text...")
    tokens_starred, text_starred = preprocess_text(text, config.romanize, "ara")
    logger.info("Preprocessed text.")

    logger.info("Generating alignments...")
    segments, _scores, blank_token = get_alignments(
        emissions,
        tokens_starred,
        tokenizer,
    )
    logger.info("Generated alignments.")

    logger.info("Generating spans...")
    spans = get_spans(tokens_starred, segments, blank_token)
    logger.info("Generated spans.")

    logger.info("Postprocessing results...")
    results = postprocess_results(text_starred, spans, stride, word_segments)
    logger.info("Postprocessed results.")

    logger.info(f"Processed request.")
    return MsgSpecJSONResponse(results)


@app.middleware("http")
async def add_request_id(request: Request, call_next):
    _ = request_id_var.set(str(uuid.uuid4()))
    response = await call_next(request)
    return response
