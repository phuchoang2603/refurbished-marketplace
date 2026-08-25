"""Colab-safe UniMERNet install.

Do NOT pip-install any of these on Colab — they fail on Python 3.12/3.13:

    pip install unimernet[full]
    pip install "tokenizers>=0.19.1,<0.20" "transformers==4.42.4"

UniMERNet pins transformers==4.42.4 → tokenizers 0.19.x. Colab's Python 3.12/3.13
has no prebuilt tokenizers 0.19 wheels, so pip compiles from source (needs
Rust) and exits with "Building wheel for tokenizers ... did not run
successfully".

Keep Colab's existing torch / numpy / tokenizers / transformers wheels.
Install unimernet with --no-deps, then extras (not [full], not tokenizers).
Do not use ``--only-binary=:all:`` for extras: fairscale/iopath/Wand often
have no cp313 wheel and that flag aborts the whole install.

Then patch transformers 5.x (pytorch_utils helpers + decoder ONNX stub)
before ``import unimernet``.
"""

from __future__ import annotations

import subprocess
import sys


FORBIDDEN_PIP = (
    "unimernet[full]",
    "tokenizers>=",
    "tokenizers==0.19",
    "transformers==4.42.4",
)


def _run(cmd: list[str]) -> None:
    print("+", " ".join(cmd), flush=True)
    subprocess.check_call(cmd)


def assert_torch_healthy() -> None:
    """Fail fast if a previous source-build of numpy/tokenizers broke the runtime."""
    try:
        import numpy as np
        import torch

        _ = (torch.zeros(1) + float(np.array([1.0]))).item()
    except Exception as exc:
        raise SystemExit(
            "numpy/torch is broken after a failed pip build. In Colab: "
            "Runtime → Restart session (keep ImageMagick), then run "
            "install_unimernet_colab() only — never pip-install tokenizers."
        ) from exc


def install_unimernet_colab() -> None:
    import tokenizers
    import transformers

    assert_torch_healthy()
    print(
        "keep Colab wheels:",
        "python",
        f"{sys.version_info.major}.{sys.version_info.minor}",
        "tokenizers",
        tokenizers.__version__,
        "transformers",
        transformers.__version__,
        flush=True,
    )
    py = sys.executable
    _run([py, "-m", "pip", "install", "-q", "unimernet", "--no-deps"])
    _run(
        [
            py,
            "-m",
            "pip",
            "install",
            "-q",
            "omegaconf",
            "timm",
            "ftfy",
            "albumentations",
            "rapidfuzz",
            "webdataset",
        ]
    )
    for pkg in ("iopath", "fairscale", "Wand"):
        try:
            _run([py, "-m", "pip", "install", "-q", pkg])
        except subprocess.CalledProcessError:
            print("skip extra", pkg, flush=True)
    from compat_transformers import import_unimernet

    unimernet = import_unimernet()
    print("unimernet OK", unimernet.__file__, flush=True)


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
