"""Colab-safe UniMERNet install.

Do NOT use `pip install unimernet[full]` on Colab: it pins transformers==4.42.4
which tries to compile `tokenizers` from source (needs Rust) and fails.

Keep Colab's torch. Install tokenizers from a wheel, then unimernet --no-deps.
"""

from __future__ import annotations

import subprocess
import sys


"""Colab-safe UniMERNet install.

Never pip-install tokenizers/transformers on Colab: those pins compile
Rust tokenizers from source and fail. Use the wheels Colab already has.
"""

from __future__ import annotations

import subprocess
import sys


def _run(cmd: list[str]) -> None:
    print("+", " ".join(cmd), flush=True)
    subprocess.check_call(cmd)


def install_unimernet_colab() -> None:
    import tokenizers
    import transformers

    print("keep Colab tokenizers", tokenizers.__version__, "transformers", transformers.__version__)
    py = sys.executable
    _run([py, "-m", "pip", "install", "-q", "unimernet", "--no-deps"])
    _run(
        [
            py,
            "-m",
            "pip",
            "install",
            "-q",
            "--only-binary=:all:",
            "omegaconf",
            "timm",
            "iopath",
            "fairscale",
            "ftfy",
            "albumentations",
            "rapidfuzz",
            "webdataset",
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
