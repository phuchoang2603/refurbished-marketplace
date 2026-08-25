"""Convert OCR Markdown + LaTeX (pandoc / Unlimited-OCR) into Azota markup.

Intermediate MD from the page-OCR path looks like:

    **Câu 1:** ... $\\log_{a} b = 1$
    ![ ](/kaggle/working/.../image1.png){width="2in"}

This mapper keeps `$latex$`, turns images into `[img:$img_N$]`,
markdown tables into `[* c1 | c2 *]`, and `**bold**` into `[!b:$…$]`.
"""

from __future__ import annotations

import re
import shutil
from pathlib import Path

from convert import postprocess_lines

IMG_RE = re.compile(r"!\[.*?\]\(([^)]+)\)(?:\{[^}]*\})?", re.DOTALL)
BOLD_RE = re.compile(r"\*\*(.+?)\*\*")
# Italic *text* but not ** and not *A. / *a)
ITALIC_RE = re.compile(r"(?<!\*)\*(?![A-Da-d][.)])([^*]+?)\*(?!\*)")
TABLE_SEP_RE = re.compile(r"^\s*\|?(?:\s*:?-{3,}:?\s*\|)+\s*:?-{3,}:?\s*\|?\s*$")


def markdown_to_azota(
    md: str,
    sidecar_dir: Path | None = None,
    media_root: Path | None = None,
    img_start: int = 1,
) -> tuple[str, list[dict]]:
    sidecar_dir = Path(sidecar_dir) if sidecar_dir else None
    if sidecar_dir:
        sidecar_dir.mkdir(parents=True, exist_ok=True)
    media_root = Path(media_root) if media_root else None
    assets: list[dict] = []
    img_n = img_start

    def replace_img(match: re.Match[str]) -> str:
        nonlocal img_n
        src = match.group(1).strip().strip('"').strip("'")
        asset_id = f"img_{img_n}"
        img_n += 1
        dest_name = None
        if sidecar_dir:
            src_path = Path(src)
            if not src_path.is_file() and media_root is not None:
                src_path = media_root / src.lstrip("/")
            if src_path.is_file():
                dest_name = f"{asset_id}{src_path.suffix or '.png'}"
                shutil.copy2(src_path, sidecar_dir / dest_name)
            else:
                dest_name = f"{asset_id}.png"
        assets.append(
            {
                "id": asset_id,
                "kind": "img",
                "placeholder": f"[img:${asset_id}$]",
                "source": src,
                "sidecar": f"sidecar/{dest_name}" if dest_name else None,
            }
        )
        return f"[img:${asset_id}$]"

    text = IMG_RE.sub(replace_img, md)
    lines = [_convert_line(l) for l in text.splitlines()]
    lines = _tables_to_azota(lines)
    lines = postprocess_lines(lines)
    return "\n".join(lines).rstrip() + "\n", assets


def _convert_line(line: str) -> str:
    # Protect latex $...$ so bold/italic regex does not eat it.
    holes: list[str] = []

    def stash(m: re.Match[str]) -> str:
        holes.append(m.group(0))
        return f"@@LATEX{len(holes) - 1}@@"

    protected = re.sub(r"\$[^$]+\$", stash, line)
    protected = BOLD_RE.sub(lambda m: f"[!b:${m.group(1)}$]", protected)
    protected = ITALIC_RE.sub(lambda m: f"[!i:${m.group(1)}$]", protected)
    for i, raw in enumerate(holes):
        protected = protected.replace(f"@@LATEX{i}@@", raw)
    return protected


def _tables_to_azota(lines: list[str]) -> list[str]:
    out: list[str] = []
    i = 0
    while i < len(lines):
        if "|" in lines[i] and i + 1 < len(lines) and TABLE_SEP_RE.match(lines[i + 1]):
            block = [lines[i]]
            i += 2
            while i < len(lines) and "|" in lines[i] and lines[i].strip():
                block.append(lines[i])
                i += 1
            for row in block:
                cells = [c.strip() for c in row.strip().strip("|").split("|")]
                out.append("[* " + " | ".join(cells) + " *]")
            continue
        out.append(lines[i])
        i += 1
    return out
