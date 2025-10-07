import logging

from fastapi import UploadFile

import torch
from torchcodec.decoders import AudioDecoder
from ctc_forced_aligner.alignment_utils import SAMPLING_FREQ

logger = logging.getLogger("uvicorn.error")


async def load_audio(file: UploadFile, dtype: torch.dtype, device: str):
    decoder = AudioDecoder(await file.read(), sample_rate=SAMPLING_FREQ)
    samples = decoder.get_all_samples()
    waveform = samples.data

    waveform = waveform.to(device, dtype)

    if waveform.shape[0] > 1:
        waveform = torch.mean(waveform, dim=0)
    else:
        waveform = waveform.squeeze(0)

    if not waveform.is_contiguous():
        waveform = waveform.contiguous()

    return waveform
