"""Colab-safe UniMERNet install.

Do NOT use `pip install unimernet[full]` on Colab: it pins transformers==4.42.4
which tries to compile `tokenizers` from source (needs Rust) and fails.

Keep Colab's torch. Install tokenizers from a wheel, then unimernet --no-deps.
"""

from __future__ import annotations

import subprocess
import sys


def _run(cmd: list[str]) -> None:
    print("+", " ".join(cmd), flush=True)
    subprocess.check_call(cmd)


def install_unimernet_colab() -> None:
    py = sys.executable
    # Wheel first — never build tokenizers from source.
    _run(
        [
            py,
            "-m",
            "pip",
            "install",
            "-q",
            "tokenizers>=0.19.1,<0.20",
            "transformers==4.42.4",
        ]
    )
    _run([py, "-m", "pip", "install", "-q", "unimernet", "--no-deps"])
    _run(
        [
            py,
            "-m",
            "pip",
            "install",
            "-q",
            "omegaconf>=2.3.0",
            "timm>=0.9.16,<0.10",
            "iopath>=0.1.9,<0.2",
            "fairscale>=0.4.13,<0.5",
            "ftfy>=6.2.0",
            "albumentations>=1.4.4,<2",
            "rapidfuzz>=3.8.1,<4",
            "evaluate>=0.4.1,<0.5",
            "webdataset>=0.2.86,<0.3",
            "Wand",
        ]
    )
    import unimernet

    print("unimernet OK", unimernet.__file__)


def allow_wmf_in_imagemagick() -> None:
    from pathlib import Path

    for policy in (
        Path("/etc/ImageMagick-6/policy.xml"),
        Path("/etc/ImageMagick/policy.xml"),
    ):
        if not policy.exists():
            continue
        text = policy.read_text(encoding="utf-8")
        for name in ("WMF", "EMF", "WMZ", "EMZ"):
            text = text.replace(
                f'rights="none" pattern="{name}"',
                f'rights="read|write" pattern="{name}"',
            )
        policy.write_text(text, encoding="utf-8")
        print("patched", policy)


if __name__ == "__main__":
    allow_wmf_in_imagemagick()
    install_unimernet_colab()
