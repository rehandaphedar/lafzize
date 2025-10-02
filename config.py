import os
import msgspec

import torch

TORCH_DTYPES = {
    "bfloat16": torch.bfloat16,
    "float16": torch.float16,
    "float32": torch.float32,
}


class Config(msgspec.Struct):
    star_frequency: str
    model: str
    dtype: torch.dtype
    batch_size: int
    window_size: int
    context_size: int
    romanize: bool
    device: str
    max_upload_size: int
    taawwudh: str
    basmalah: str
    data: str
    data_extra: str


config = Config(
    star_frequency=os.getenv("LAFZIZE_STAR_FREQUENCY", "segment"),
    model=os.getenv("LAFZIZE_MODEL", "MahmoudAshraf/mms-300m-1130-forced-aligner"),
    dtype=TORCH_DTYPES[
        os.getenv(
            "LAFZIZE_DTYPE",
            "float16" if torch.cuda.is_available() else "float32",
        )
    ],
    batch_size=int(os.getenv("LAFZIZE_BATCH_SIZE", "4")),
    window_size=int(os.getenv("LAFZIZE_WINDOW_SIZE", "30")),
    context_size=int(os.getenv("LAFZIZE_CONTEXT_SIZE", "2")),
    romanize=os.getenv("LAFZIZE_ROMANIZE", "True") == "True",
    device=os.getenv("LAFZIZE_DEVICE", "cuda" if torch.cuda.is_available() else "cpu"),
    max_upload_size=int(os.getenv("LAFZIZE_MAX_UPLOAD_SIZE", "128")),
    taawwudh=os.getenv("LAFZIZE_TAAWWUDH", "taawwudh"),
    basmalah=os.getenv("LAFZIZE_BASMALAH", "basmalah"),
    data=os.getenv("LAFZIZE_DATA", "data.json"),
    data_extra=os.getenv("LAFZIZE_DATA_EXTRA", "data_extra.json"),
)
