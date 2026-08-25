"""Optional GPU overlays: UniMERNet (formulas) + Unlimited-OCR (pages/figures).

These helpers are imported from Colab. They degrade cleanly when torch / the
models are not installed so the structural converter still runs on CPU.
"""

from __future__ import annotations

import re
from pathlib import Path
from typing import Iterable

DET_RE = re.compile(r"<\|det\|>([^<\s]+)(?:\s*\[[^\]]*\])?\s*<\|/det\|>(.*)", re.DOTALL)


def strip_unlimited_ocr_det(raw: str) -> str:
    """Remove <|det|>type [bbox]<|/det|> markers (official Unlimited-OCR postprocess)."""
    blocks: list[list[str]] = []
    cur: list[str] | None = None
    for line in raw.splitlines():
        line = line.rstrip()
        if not line:
            continue
        m = DET_RE.match(line)
        if m:
            category, content = m.group(1).strip(), m.group(2).strip()
            if category == "image":
                continue
            if cur is not None:
                blocks.append(cur)
            cur = [content] if content else []
            continue
        if cur is None:
            cur = []
        cur.append(line)
    if cur is not None:
        blocks.append(cur)
    return "\n\n".join("\n".join(b) for b in blocks).strip()


def rasterize_formula_image(src: Path, dest_png: Path, dpi: int = 300) -> Path | None:
    """WMF/EMF/PNG/JPEG → PNG for UniMERNet.

    Tries Pillow first (PNG/JPEG), then Wand/ImageMagick, then LibreOffice.
    Returns dest path or None if conversion failed.
    """
    src = Path(src)
    dest_png = Path(dest_png)
    dest_png.parent.mkdir(parents=True, exist_ok=True)
    ext = src.suffix.lower()
    if ext in {".png", ".jpg", ".jpeg", ".webp", ".gif"}:
        try:
            from PIL import Image

            Image.open(src).convert("RGB").save(dest_png)
            return dest_png
        except Exception:
            dest_png.write_bytes(src.read_bytes())
            return dest_png

    # Wand (ImageMagick)
    try:
        from wand.image import Image as WandImage

        with WandImage(filename=str(src), resolution=dpi) as img:
            img.format = "png"
            img.save(filename=str(dest_png))
        return dest_png
    except Exception:
        pass

    # LibreOffice headless
    import shutil
    import subprocess
    import tempfile

    soffice = shutil.which("soffice") or shutil.which("libreoffice")
    if soffice:
        with tempfile.TemporaryDirectory() as td:
            try:
                subprocess.run(
                    [soffice, "--headless", "--convert-to", "png", "--outdir", td, str(src)],
                    check=True,
                    capture_output=True,
                    timeout=60,
                )
                produced = next(Path(td).glob("*.png"), None)
                if produced:
                    dest_png.write_bytes(produced.read_bytes())
                    return dest_png
            except Exception:
                pass
    return None


def load_unimernet(cfg_path: str | Path | None = None, device: str | None = None, fp16: bool = True):
    """Load UniMERNet once so Colab can infer image-by-image."""
    import argparse

    import torch

    from compat_transformers import patch_transformers_for_unimernet

    patch_transformers_for_unimernet()
    from unimernet.common.config import Config
    import unimernet.tasks as tasks
    from unimernet.processors import load_processor
    import unimernet

    pkg = Path(unimernet.__file__).resolve().parent
    cfg_candidates = []
    if cfg_path:
        cfg_candidates.append(Path(cfg_path))
    cfg_candidates.extend(
        [
            pkg / "configs" / "demo.yaml",
            Path("UniMERNet/configs/demo.yaml"),
            Path("/content/UniMERNet/configs/demo.yaml"),
            Path("/content/models/unimernet_tiny_colab.yaml"),
        ]
    )
    found = next((p for p in cfg_candidates if p.exists()), None)
    if found is None:
        raise FileNotFoundError("Cannot find UniMERNet yaml — run prepare_unimernet_checkpoint()")
    args = argparse.Namespace(cfg_path=str(found), options=None)
    cfg = Config(args)
    task = tasks.setup_task(cfg)
    if device is None:
        device = "cuda" if torch.cuda.is_available() else "cpu"
    model = task.build_model(cfg).to(device).eval()
    if fp16 and device == "cuda":
        model = model.half()
    vis_processor = load_processor(
        "formula_image_eval",
        cfg.config.datasets.formula_rec_eval.vis_processor.eval,
    )
    return model, vis_processor, device


def unimernet_one(model, vis_processor, device, image_path: str | Path) -> str:
    import torch
    from PIL import Image

    raw = Image.open(image_path).convert("RGB")
    image = vis_processor(raw).unsqueeze(0).to(device)
    if next(model.parameters()).dtype == torch.float16:
        image = image.half()
    with torch.inference_mode():
        output = model.generate({"image": image})
    return output["pred_str"][0]


def unimernet_batch(
    model,
    vis_processor,
    device,
    jobs: list[tuple[str, Path]],
    batch_size: int = 8,
) -> dict[str, str]:
    """Batched UniMERNet — main T4 speedup vs one-image-at-a-time."""
    import torch
    from PIL import Image

    preds: dict[str, str] = {}
    use_half = next(model.parameters()).dtype == torch.float16
    for i in range(0, len(jobs), batch_size):
        chunk = jobs[i : i + batch_size]
        tensors = [vis_processor(Image.open(p).convert("RGB")) for _, p in chunk]
        batch = torch.stack(tensors).to(device)
        if use_half:
            batch = batch.half()
        with torch.inference_mode():
            output = model.generate({"image": batch})
        for (aid, _), latex in zip(chunk, output["pred_str"]):
            preds[aid] = latex
    return preds


def run_unimernet(
    image_paths: Iterable[tuple[str, Path]],
    model_size: str = "tiny",
    device: str | None = None,
    cfg_path: str | Path | None = None,
) -> dict[str, str]:
    """Recognize LaTeX for (asset_id, png_path) pairs. Requires `unimernet` + GPU/CPU torch."""
    _ = model_size
    model, vis_processor, device = load_unimernet(cfg_path=cfg_path, device=device)
    preds: dict[str, str] = {}
    for asset_id, path in image_paths:
        preds[asset_id] = unimernet_one(model, vis_processor, device, path)
    return preds


def load_unlimited_ocr(model_name: str = "baidu/Unlimited-OCR"):
    import torch
    from transformers import AutoModel, AutoTokenizer

    tokenizer = AutoTokenizer.from_pretrained(model_name, trust_remote_code=True)
    dtype = torch.bfloat16 if torch.cuda.is_available() else torch.float32
    model = AutoModel.from_pretrained(
        model_name,
        trust_remote_code=True,
        use_safetensors=True,
        torch_dtype=dtype,
    )
    if torch.cuda.is_available():
        model = model.cuda()
    return model.eval(), tokenizer


def unlimited_ocr_one(
    model,
    tokenizer,
    image_file: str,
    output_path: str,
    gundam: bool = True,
    prompt: str = "<image>document parsing.",
    max_length: int = 4096,
) -> str:
    Path(output_path).mkdir(parents=True, exist_ok=True)
    model.infer(
        tokenizer,
        prompt=prompt,
        image_file=image_file,
        output_path=output_path,
        base_size=1024,
        image_size=640 if gundam else 1024,
        crop_mode=gundam,
        max_length=max_length,
        no_repeat_ngram_size=35,
        ngram_window=128,
        save_results=True,
    )
    texts = []
    for p in sorted(Path(output_path).rglob("*")):
        if p.suffix.lower() in {".md", ".txt"}:
            texts.append(p.read_text(encoding="utf-8", errors="replace"))
    return "\n\n".join(texts)


def run_unlimited_ocr_pages(
    image_files: list[str],
    output_path: str,
    prompt: str = "<image>document parsing.",
    gundam: bool = True,
) -> str:
    """Run Baidu Unlimited-OCR on one page (gundam) or many pages (base/multi)."""
    import torch
    from transformers import AutoModel, AutoTokenizer

    model_name = "baidu/Unlimited-OCR"
    tokenizer = AutoTokenizer.from_pretrained(model_name, trust_remote_code=True)
    dtype = torch.bfloat16 if torch.cuda.is_available() else torch.float32
    model = AutoModel.from_pretrained(
        model_name,
        trust_remote_code=True,
        use_safetensors=True,
        torch_dtype=dtype,
    )
    if torch.cuda.is_available():
        model = model.cuda()
    model = model.eval()

    Path(output_path).mkdir(parents=True, exist_ok=True)
    if len(image_files) == 1:
        model.infer(
            tokenizer,
            prompt=prompt,
            image_file=image_files[0],
            output_path=output_path,
            base_size=1024,
            image_size=640 if gundam else 1024,
            crop_mode=gundam,
            max_length=32768,
            no_repeat_ngram_size=35,
            ngram_window=128,
            save_results=True,
        )
    else:
        model.infer_multi(
            tokenizer,
            prompt="<image>Multi page parsing.",
            image_files=image_files,
            output_path=output_path,
            image_size=1024,
            max_length=32768,
            no_repeat_ngram_size=35,
            ngram_window=1024,
            save_results=True,
        )

    texts = []
    for p in sorted(Path(output_path).rglob("*")):
        if p.suffix.lower() in {".md", ".txt"}:
            texts.append(p.read_text(encoding="utf-8", errors="replace"))
    return "\n\n".join(texts)


def pdf_to_images(pdf_path: str, dpi: int = 200) -> list[str]:
    import os
    import tempfile

    import fitz

    doc = fitz.open(pdf_path)
    tmp_dir = tempfile.mkdtemp(prefix="pdf_ocr_")
    mat = fitz.Matrix(dpi / 72, dpi / 72)
    paths = []
    for i, page in enumerate(doc):
        out = os.path.join(tmp_dir, f"page_{i + 1:04d}.png")
        page.get_pixmap(matrix=mat).save(out)
        paths.append(out)
    doc.close()
    return paths
