"""Colab GPU profiles + UniMERNet/OCR setup optimized for T4 vs A100.

Optimization idea: never OCR full exam pages when a .docx exists.
OOXML already has text/OMML/answers. Vision only sees:
  * MathType WMF crops → UniMERNet (tiny, batched)
  * drawing images → Unlimited-OCR (optional, one figure at a time)
Unload UniMERNet before loading Unlimited-OCR so T4 16GB does not OOM.
"""

from __future__ import annotations

import gc
import re
from pathlib import Path
from typing import Any

PROFILES: dict[str, dict[str, Any]] = {
    "cpu": {
        "unimernet": "tiny",
        "um_batch": 1,
        "ocr_figures": False,
        "ocr_pages": False,
        "ocr_mode": "base",
        "max_ocr_len": 2048,
        "dpi": 120,
        "unload_between": True,
        "fp16": False,
    },
    "t4": {
        "unimernet": "tiny",
        "um_batch": 8,
        "ocr_figures": False,  # bật tay nếu còn VRAM sau UniMERNet
        "ocr_pages": False,
        "ocr_mode": "base",
        "max_ocr_len": 4096,
        "dpi": 150,
        "unload_between": True,
        "fp16": True,
    },
    "a100": {
        "unimernet": "small",
        "um_batch": 16,
        "ocr_figures": True,
        "ocr_pages": True,
        "ocr_mode": "gundam",
        "max_ocr_len": 8192,
        "dpi": 200,
        "unload_between": False,
        "fp16": True,
    },
}

HF_REPOS = {
    "tiny": ("wanderkid/unimernet_tiny", "unimernet_tiny.pth"),
    "small": ("wanderkid/unimernet_small", "unimernet_small.pth"),
    "base": ("wanderkid/unimernet_base", "unimernet_base.pth"),
}


def detect_profile() -> tuple[str, dict[str, Any]]:
    try:
        import torch
    except ImportError:
        return "cpu", dict(PROFILES["cpu"])
    if not torch.cuda.is_available():
        return "cpu", dict(PROFILES["cpu"])
    name = torch.cuda.get_device_name(0).lower()
    vram_gb = torch.cuda.get_device_properties(0).total_memory / 1024**3
    print(f"GPU: {torch.cuda.get_device_name(0)}  VRAM={vram_gb:.1f} GB")
    if vram_gb >= 30 or any(k in name for k in ("a100", "l40", "h100", "a6000")):
        return "a100", dict(PROFILES["a100"])
    if "l4" in name and vram_gb >= 20:
        p = dict(PROFILES["a100"])
        p["unimernet"] = "tiny"
        p["ocr_pages"] = False
        return "l4", p
    return "t4", dict(PROFILES["t4"])


def free_cuda(*objs: object) -> None:
    for obj in objs:
        del obj
    gc.collect()
    try:
        import torch

        if torch.cuda.is_available():
            torch.cuda.empty_cache()
            torch.cuda.ipc_collect()
    except Exception:
        pass


def prepare_unimernet_checkpoint(
    size: str = "tiny",
    dest_root: str | Path = "/content/models",
) -> Path:
    """Download UniMERNet weights and write a Colab yaml with absolute paths."""
    from huggingface_hub import snapshot_download

    if size not in HF_REPOS:
        raise ValueError(f"size must be tiny|small|base, got {size}")
    repo, pth_name = HF_REPOS[size]
    dest_root = Path(dest_root)
    model_dir = dest_root / f"unimernet_{size}"
    snapshot_download(repo_id=repo, local_dir=str(model_dir))
    pth = model_dir / pth_name
    if not pth.exists():
        found = next(model_dir.glob("*.pth"), None)
        if found is None:
            raise FileNotFoundError(f"No .pth in {model_dir}")
        pth = found

    tmpl = Path(__file__).resolve().parent / "configs" / "unimernet_tiny.yaml"
    yaml_text = tmpl.read_text(encoding="utf-8")
    yaml_text = yaml_text.replace("unimernet_tiny.pth", pth.name)
    yaml_text = yaml_text.replace("MODEL_DIR", str(model_dir).replace("\\", "/"))
    cfg_path = dest_root / f"unimernet_{size}_colab.yaml"
    cfg_path.write_text(yaml_text, encoding="utf-8")
    print("checkpoint", pth, "cfg", cfg_path)
    return cfg_path


def vision_jobs_from_manifest(
    manifest: dict[str, Any],
    out_dir: str | Path,
    kinds: tuple[str, ...] = ("mathtype",),
) -> list[tuple[str, Path]]:
    """Only formula/drawing sidecars — never full exam pages."""
    out_dir = Path(out_dir)
    jobs: list[tuple[str, Path]] = []
    for asset in manifest.get("assets", []):
        if asset.get("kind") not in kinds:
            continue
        src = out_dir / asset["sidecar"]
        if src.is_file() and src.stat().st_size > 0:
            jobs.append((asset["id"], src))
    return jobs


def is_plausible_latex(latex: str) -> bool:
    """Skip UniMERNet fragments that would make Azota markup worse than placeholders."""
    body = latex.strip().strip("$")
    if len(body) < 2:
        return False
    compact = re.sub(r"\s+", "", body)
    if compact.startswith("^") or compact.startswith("_"):
        return False
    if "\\" not in body and len(compact) < 8:
        return False
    return True


def inject_latex_into_markup(markup: str, predictions: dict[str, str]) -> str:
    """Replace [!m:$mathtype_N$] with $latex$ so Azota renders without OLE."""
    for aid, latex in predictions.items():
        body = latex.strip().strip("$")
        if not body or not is_plausible_latex(body):
            continue
        markup = markup.replace(f"[!m:${aid}$]", f"${body}$")
    return markup
