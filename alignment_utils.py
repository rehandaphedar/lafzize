import io
import logging

from fastapi import UploadFile

import torch
import torchaudio
from ctc_forced_aligner.alignment_utils import SAMPLING_FREQ

logger = logging.getLogger("uvicorn.error")


async def load_audio(file: UploadFile, dtype: torch.dtype, device: str):
    audio_bytes = await file.read()
    with io.BytesIO(audio_bytes) as audio_buffer:
        waveform, sample_rate = torchaudio.load(audio_buffer)

    if waveform.shape[0] > 1:
        waveform = torch.mean(waveform, dim=0)
    else:
        waveform = waveform.squeeze(0)

    if sample_rate != SAMPLING_FREQ:
        try:
            waveform = torchaudio.functional.resample(
                waveform.to(device), orig_freq=sample_rate, new_freq=SAMPLING_FREQ
            )
        except Exception as error:
            logger.warning(f"Resampling on device {device} failed: {error}")
            if device != "cpu":
                logger.info("Resampling on device cpu...")
                waveform = torchaudio.functional.resample(
                    waveform.to("cpu"), orig_freq=sample_rate, new_freq=SAMPLING_FREQ
                )

    waveform = waveform.to(device)

    if not waveform.is_contiguous():
        waveform = waveform.contiguous()

    return waveform.to(dtype=dtype)
