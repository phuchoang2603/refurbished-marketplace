#!/usr/bin/env python3
"""Colab / CLI shim — Tầng A lives in ``docx_extract``.

``from convert import convert_docx`` stays valid. UniMERNet / OCR helpers stay
here because they are out of spec v1 (YAGNI).
"""

from __future__ import annotations

import json
from pathlib import Path
from typing import Any

from docx_extract import convert_docx
from docx_extract.answer import star_correct_marker
from docx_extract.assemble import main as _package_main
from docx_extract.ns import NS, qn
from docx_extract.runs import RunStyle, azota_wrap, postprocess_lines, render_paragraph
from docx_extract.tables import render_table

NSMAP = NS

__all__ = [
    "NS",
    "NSMAP",
    "qn",
    "RunStyle",
    "azota_wrap",
    "star_correct_marker",
    "postprocess_lines",
    "render_paragraph",
    "render_table",
    "convert_docx",
    "apply_unimernet_latex",
    "write_ocr_sidecar",
    "main",
]


def apply_unimernet_latex(
    manifest: dict[str, Any],
    predictions: dict[str, str],
    out_dir: Path,
) -> dict[str, Any]:
    """Attach UniMERNet LaTeX to mathtype/img formula assets and write .tex sidecars."""
    out_dir = Path(out_dir)
    sidecar = out_dir / "sidecar"
    sidecar.mkdir(parents=True, exist_ok=True)
    for asset in manifest.get("assets", []):
        pred = predictions.get(asset["id"])
        if not pred:
            continue
        asset["latex"] = pred
        tex_path = sidecar / f"{asset['id']}.tex"
        tex_path.write_text(pred.strip() + "\n", encoding="utf-8")
        asset.setdefault("extras", {})["latex_sidecar"] = f"sidecar/{asset['id']}.tex"
    (out_dir / "manifest.json").write_text(
        json.dumps(manifest, ensure_ascii=False, indent=2) + "\n", encoding="utf-8"
    )
    return manifest


def write_ocr_sidecar(ocr_text: str, out_dir: Path, name: str = "unlimited_ocr.md") -> Path:
    path = Path(out_dir) / "ocr" / name
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(ocr_text, encoding="utf-8")
    return path


def main(argv: list[str] | None = None) -> int:
    return _package_main(argv)


if __name__ == "__main__":
    raise SystemExit(main())
