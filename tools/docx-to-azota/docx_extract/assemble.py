"""Walk w:body in document order and write markup.txt + sidecar + manifest.json."""

from __future__ import annotations

import argparse
import json
import logging
from dataclasses import dataclass, field
from pathlib import Path
from typing import Any
from xml.etree import ElementTree as ET
from zipfile import ZipFile

from .loader import Document, load
from .math_assets import AssetStore
from .ns import local, qn
from .runs import postprocess_lines, render_paragraph
from .tables import render_table

log = logging.getLogger("docx_extract")


@dataclass
class Context:
    zip: ZipFile
    document_path: str
    rels: dict[str, dict[str, str]]
    sidecar: AssetStore
    order: int = 0
    xpath_stack: list[str] = field(default_factory=list)
    lines: list[str] = field(default_factory=list)
    warnings: list[str] = field(default_factory=list)


def convert_paragraph(p: ET.Element, ctx: Context) -> str:
    return render_paragraph(p, ctx)


def _convert_sdt(sdt: ET.Element, ctx: Context) -> None:
    content = sdt.find(qn("w", "sdtContent"))
    if content is None:
        return
    for i, child in enumerate(content, start=1):
        name = local(child.tag)
        ctx.xpath_stack.append(f"sdtContent/{name}[{i}]")
        if name == "tbl":
            ctx.lines.extend(render_table(child, ctx))
        elif name == "p":
            line = convert_paragraph(child, ctx)
            if line is not None:
                ctx.lines.append(line)
        ctx.xpath_stack.pop()


def clean_docx(path: str | Path, out_dir: str | Path) -> dict[str, Any]:
    out = Path(out_dir)
    out.mkdir(parents=True, exist_ok=True)
    doc: Document = load(path)
    ctx = Context(
        zip=doc.zip,
        document_path=doc.document_path,
        rels=doc.rels,
        sidecar=AssetStore(out / "sidecar"),
        xpath_stack=["body"],
    )
    try:
        for i, child in enumerate(doc.body, start=1):
            name = local(child.tag)
            if name == "sectPr":
                continue
            ctx.xpath_stack.append(f"{name}[{i}]")
            try:
                if name == "tbl":
                    ctx.lines.extend(render_table(child, ctx))
                elif name == "p":
                    line = convert_paragraph(child, ctx)
                    if line is not None:
                        ctx.lines.append(line)
                elif name == "sdt":
                    _convert_sdt(child, ctx)
                else:
                    log.debug("skip element %s", name)
            except Exception as exc:
                ctx.warnings.append(f"{name}[{i}]: {exc}")
                log.debug("skip %s[%s]: %s", name, i, exc)
            ctx.xpath_stack.pop()
        ctx.lines = postprocess_lines(ctx.lines)
        (out / "markup.txt").write_text("\n".join(ctx.lines).rstrip() + "\n", encoding="utf-8")
        manifest = {
            "source": Path(path).name,
            "document_part": ctx.document_path,
            "counts": {
                "mathml": sum(1 for a in ctx.sidecar.assets if a.kind == "mathml"),
                "mathtype": sum(1 for a in ctx.sidecar.assets if a.kind == "mathtype"),
                "img": sum(1 for a in ctx.sidecar.assets if a.kind == "img"),
                "lines": len(ctx.lines),
            },
            "warnings": ctx.warnings,
            "assets": [a.to_json() for a in ctx.sidecar.assets],
        }
        (out / "manifest.json").write_text(
            json.dumps(manifest, ensure_ascii=False, indent=2) + "\n",
            encoding="utf-8",
        )
        if ctx.warnings:
            log.warning("%s warnings", len(ctx.warnings))
        return manifest
    finally:
        doc.zip.close()


def convert_docx(docx_path: str | Path, out_dir: str | Path) -> dict[str, Any]:
    """Colab / CLI alias for clean_docx."""
    return clean_docx(docx_path, out_dir)


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description="DOCX → Azota markup.txt / sidecar / manifest.json")
    parser.add_argument("docx", type=Path, help="Input .docx")
    parser.add_argument("-o", "--out", type=Path, default=Path("azota_out"), help="Output directory")
    args = parser.parse_args(argv)
    manifest = clean_docx(args.docx, args.out)
    print(json.dumps(manifest["counts"], ensure_ascii=False, indent=2))
    print(f"Wrote {args.out / 'markup.txt'}")
    print(f"Wrote {args.out / 'manifest.json'}")
    print(f"Wrote {args.out / 'sidecar'} ({len(manifest['assets'])} assets)")
    return 0
