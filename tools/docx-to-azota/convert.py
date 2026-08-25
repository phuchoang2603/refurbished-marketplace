#!/usr/bin/env python3
"""Convert a Vietnamese exam .docx into Azota markup + sidecar assets.

Structural OOXML is the source of truth (text, bold/italic/sup/sub, OMML,
MathType OLE previews, drawings, tables, underlined correct answers).

Vision models are optional overlays:
  * UniMERNet  — formula image → LaTeX (MathType WMF/EMF)
  * Unlimited-OCR — full-page parse for figures / reading-order QA
"""

from __future__ import annotations

import argparse
import json
import re
import shutil
import zipfile
from dataclasses import dataclass, field
from pathlib import Path
from typing import Any
from xml.etree import ElementTree as ET

W_NS = "http://schemas.openxmlformats.org/wordprocessingml/2006/main"
M_NS = "http://schemas.openxmlformats.org/officeDocument/2006/math"
R_NS = "http://schemas.openxmlformats.org/officeDocument/2006/relationships"
A_NS = "http://schemas.openxmlformats.org/drawingml/2006/main"
WP_NS = "http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing"
V_NS = "urn:schemas-microsoft-com:vml"
O_NS = "urn:schemas-microsoft-com:office:office"
REL_NS = "http://schemas.openxmlformats.org/package/2006/relationships"
PKG_REL = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument"

NS = {"w": W_NS, "m": M_NS, "r": R_NS, "a": A_NS, "wp": WP_NS, "v": V_NS, "o": O_NS}

ET.register_namespace("w", W_NS)
ET.register_namespace("m", M_NS)
ET.register_namespace("r", R_NS)
ET.register_namespace("a", A_NS)
ET.register_namespace("wp", WP_NS)
ET.register_namespace("v", V_NS)
ET.register_namespace("o", O_NS)

MATHTYPE_PROGIDS = {
    "equation.dsmt4",
    "equation.3",
    "equation.dsmt6",
    "wiris.equation",
    "mathtype.equation",
}

LABEL_NO_WRAP = re.compile(
    r"^(Câu|Bài|Question|PHẦN|Phần|Nhóm|Nhom|Part|I{1,3}V?|VI{0,3})\b",
    re.IGNORECASE,
)
OPTION_MARK = re.compile(r"^\*?[A-H]\.$")
TRUEFALSE_MARK = re.compile(r"^\*?[a-h]\)$")
DS_PREFIX = re.compile(r"^\[([DS])\]\s*(.*)$", re.DOTALL)
OPTION_ITEM = re.compile(r"(\*?[A-D]\.(?:(?!\*?[A-D]\.).)+)", re.DOTALL)
SHORT_ANSWER_KEY = re.compile(r"^Đáp án là\b", re.IGNORECASE)
SHORT_ANSWER_VAL = re.compile(r"^[A-D]\.\s*(.+)$")
CHOOSE_LETTER = re.compile(r"^Chọn\s+([A-D])\.\s*$")
SECTION_RESET = re.compile(
    r"^(Câu|Bài|Question|PHẦN|Phần|Nhóm|Nhom|Part)\b", re.IGNORECASE
)


def qn(ns: str, tag: str) -> str:
    return f"{{{NS[ns]}}}{tag}"


def local(tag: str) -> str:
    return tag.rsplit("}", 1)[-1] if "}" in tag else tag


def on_off(el: ET.Element | None) -> bool:
    if el is None:
        return False
    val = el.get(qn("w", "val"))
    if val is None:
        val = el.get("val")
    if val is None:
        return True
    return str(val).lower() not in {"0", "false", "off", "none"}


def xml_fragment(el: ET.Element) -> str:
    return ET.tostring(el, encoding="unicode")


def iter_text(el: ET.Element) -> str:
    parts: list[str] = []
    if el.text:
        parts.append(el.text)
    for child in el:
        parts.append(iter_text(child))
        if child.tail:
            parts.append(child.tail)
    return "".join(parts)


@dataclass
class RunStyle:
    bold: bool = False
    italic: bool = False
    underline: bool = False
    vert: str | None = None  # superscript | subscript

    def wrap_key(self) -> tuple[bool, bool, str | None]:
        return (self.bold, self.italic, self.vert)


@dataclass
class Asset:
    id: str
    kind: str  # mathml | mathtype | img
    placeholder: str
    sidecar: str
    document_order: int
    rId: str | None = None
    source_file: str | None = None
    embedding: str | None = None
    prog_id: str | None = None
    ole_rId: str | None = None
    xpath_hint: str | None = None
    latex: str | None = None
    extras: dict[str, Any] = field(default_factory=dict)

    def to_json(self) -> dict[str, Any]:
        data = {
            "id": self.id,
            "kind": self.kind,
            "placeholder": self.placeholder,
            "sidecar": self.sidecar,
            "document_order": self.document_order,
            "rId": self.rId,
            "source_file": self.source_file,
            "embedding": self.embedding,
            "prog_id": self.prog_id,
            "ole_rId": self.ole_rId,
            "xpath_hint": self.xpath_hint,
        }
        if self.latex is not None:
            data["latex"] = self.latex
        if self.extras:
            data["extras"] = self.extras
        return data


class AssetStore:
    def __init__(self, sidecar_dir: Path) -> None:
        self.sidecar_dir = sidecar_dir
        self.sidecar_dir.mkdir(parents=True, exist_ok=True)
        self.assets: list[Asset] = []
        self._counts = {"mathml": 0, "mathtype": 0, "img": 0}

    def next_id(self, kind: str) -> str:
        self._counts[kind] += 1
        return f"{kind}_{self._counts[kind]}"

    def add(self, asset: Asset) -> Asset:
        self.assets.append(asset)
        return asset


def parse_run_style(r_pr: ET.Element | None) -> RunStyle:
    style = RunStyle()
    if r_pr is None:
        return style
    style.bold = on_off(r_pr.find(qn("w", "b"))) or on_off(r_pr.find(qn("w", "bCs")))
    style.italic = on_off(r_pr.find(qn("w", "i"))) or on_off(r_pr.find(qn("w", "iCs")))
    u = r_pr.find(qn("w", "u"))
    if u is not None:
        val = u.get(qn("w", "val")) or u.get("val") or "single"
        style.underline = str(val).lower() not in {"none", "0", "false"}
    va = r_pr.find(qn("w", "vertAlign"))
    if va is not None:
        style.vert = va.get(qn("w", "val")) or va.get("val")
    return style


def azota_wrap(text: str, style: RunStyle) -> str:
    if not text or not text.strip():
        return text
    stripped = text.strip()
    # Don't wrap punctuation-only runs (common after "Câu 10" / "Nhóm II").
    if stripped in {":", ".", ",", ";", "-", "–", "—"}:
        return text
    skip_bi = (
        LABEL_NO_WRAP.match(stripped) is not None
        or OPTION_MARK.match(stripped) is not None
        or TRUEFALSE_MARK.match(stripped) is not None
        or stripped in {"[GT]", "[GT]:", "[/]", "[D]", "[S]"}
        or re.fullmatch(r"\[([DS])\]", stripped) is not None
        or re.fullmatch(r"[A-D]", stripped) is not None
    )
    out = text
    if style.vert == "superscript":
        out = f"[!sup:${out}$]"
    elif style.vert == "subscript":
        out = f"[!sub:${out}$]"
    if skip_bi:
        return out
    if style.bold and style.italic:
        out = f"[!b!i:${out}$]"
    elif style.bold:
        out = f"[!b:${out}$]"
    elif style.italic:
        out = f"[!i:${out}$]"
    return out


def load_document_target(docx_zip: zipfile.ZipFile) -> str:
    rels = ET.fromstring(docx_zip.read("_rels/.rels"))
    for rel in rels:
        if rel.get("Type") == PKG_REL:
            target = rel.get("Target", "word/document.xml")
            return target.lstrip("/")
    return "word/document.xml"


def load_rels(docx_zip: zipfile.ZipFile, document_path: str) -> dict[str, dict[str, str]]:
    rels_path = str(Path(document_path).parent / "_rels" / (Path(document_path).name + ".rels"))
    mapping: dict[str, dict[str, str]] = {}
    try:
        root = ET.fromstring(docx_zip.read(rels_path))
    except KeyError:
        return mapping
    for rel in root:
        rid = rel.get("Id")
        if not rid:
            continue
        mapping[rid] = {
            "id": rid,
            "type": rel.get("Type", ""),
            "target": rel.get("Target", ""),
            "target_mode": rel.get("TargetMode", "Internal"),
        }
    return mapping


def resolve_media(document_path: str, target: str) -> str:
    base = Path(document_path).parent
    # Targets are relative to the document part directory (usually word/)
    resolved = (base / target).as_posix()
    while "/../" in resolved:
        resolved = re.sub(r"[^/]+/\.\./", "", resolved)
    return resolved


class DocxToAzota:
    def __init__(self, docx_path: Path, out_dir: Path) -> None:
        self.docx_path = Path(docx_path)
        self.out_dir = Path(out_dir)
        self.sidecar = AssetStore(self.out_dir / "sidecar")
        self.zip = zipfile.ZipFile(self.docx_path)
        self.document_path = load_document_target(self.zip)
        self.rels = load_rels(self.zip, self.document_path)
        self.doc_root = ET.fromstring(self.zip.read(self.document_path))
        self._order = 0
        self._xpath_stack: list[str] = []
        self.lines: list[str] = []

    def close(self) -> None:
        self.zip.close()

    def convert(self) -> dict[str, Any]:
        body = self.doc_root.find(qn("w", "body"))
        if body is None:
            raise ValueError("document.xml has no w:body")
        self._xpath_stack = ["body"]
        for i, child in enumerate(body, start=1):
            name = local(child.tag)
            if name == "sectPr":
                continue
            self._xpath_stack.append(f"{name}[{i}]")
            if name == "tbl":
                self.lines.extend(self._convert_table(child))
            elif name == "p":
                line = self._convert_paragraph(child)
                if line is not None:
                    self.lines.append(line)
            elif name == "sdt":
                self._convert_sdt(child)
            self._xpath_stack.pop()
        self.lines = postprocess_lines(self.lines)
        markup_path = self.out_dir / "markup.txt"
        markup_path.write_text("\n".join(self.lines).rstrip() + "\n", encoding="utf-8")
        manifest = {
            "source": str(self.docx_path.name),
            "document_part": self.document_path,
            "counts": {
                "mathml": sum(1 for a in self.sidecar.assets if a.kind == "mathml"),
                "mathtype": sum(1 for a in self.sidecar.assets if a.kind == "mathtype"),
                "img": sum(1 for a in self.sidecar.assets if a.kind == "img"),
                "lines": len(self.lines),
            },
            "assets": [a.to_json() for a in self.sidecar.assets],
        }
        (self.out_dir / "manifest.json").write_text(
            json.dumps(manifest, ensure_ascii=False, indent=2) + "\n",
            encoding="utf-8",
        )
        return manifest

    def _convert_sdt(self, sdt: ET.Element) -> None:
        content = sdt.find(qn("w", "sdtContent"))
        if content is None:
            return
        for i, child in enumerate(content, start=1):
            name = local(child.tag)
            self._xpath_stack.append(f"sdtContent/{name}[{i}]")
            if name == "tbl":
                self.lines.extend(self._convert_table(child))
            elif name == "p":
                line = self._convert_paragraph(child)
                if line is not None:
                    self.lines.append(line)
            self._xpath_stack.pop()

    def _convert_table(self, tbl: ET.Element) -> list[str]:
        rows: list[str] = []
        for ri, tr in enumerate(tbl.findall(qn("w", "tr")), start=1):
            cells: list[str] = []
            for ci, tc in enumerate(tr.findall(qn("w", "tc")), start=1):
                self._xpath_stack.append(f"tr[{ri}]/tc[{ci}]")
                cell_parts: list[str] = []
                for pi, p in enumerate(tc.findall(qn("w", "p")), start=1):
                    self._xpath_stack.append(f"p[{pi}]")
                    text = self._convert_paragraph(p)
                    if text:
                        cell_parts.append(text)
                    self._xpath_stack.pop()
                # nested tables
                for ti, nested in enumerate(tc.findall(qn("w", "tbl")), start=1):
                    self._xpath_stack.append(f"tbl[{ti}]")
                    cell_parts.extend(self._convert_table(nested))
                    self._xpath_stack.pop()
                cells.append(" ".join(cell_parts).strip())
                self._xpath_stack.pop()
            rows.append("[* " + " | ".join(cells) + " *]")
        return rows

    def _convert_paragraph(self, p: ET.Element) -> str | None:
        chunks: list[str] = []
        self._walk_inline(p, chunks)
        return collapse_ws_keep_newlines("".join(chunks))

    def _walk_inline(self, el: ET.Element, chunks: list[str]) -> None:
        for child in el:
            name = local(child.tag)
            if name in {"pPr", "rPr", "tblPr", "tblGrid", "sectPr", "commentRangeStart", "commentRangeEnd"}:
                continue
            if name == "oMath":
                chunks.append(self._emit_mathml(child))
                continue
            if name == "oMathPara":
                for om in child.findall(qn("m", "oMath")):
                    chunks.append(self._emit_mathml(om))
                continue
            if name == "r":
                self._walk_run(child, chunks)
                continue
            if name in {"hyperlink", "ins", "del", "smartTag", "sdt", "sdtContent", "fldSimple"}:
                self._walk_inline(child, chunks)
                continue
            if name == "object":
                chunks.append(self._emit_object(child))
                continue
            if name == "drawing":
                chunks.append(self._emit_drawing(child))
                continue
            if name == "pict":
                chunks.append(self._emit_vml(child))
                continue
            if name in {"bookmarkStart", "bookmarkEnd", "proofErr", "lastRenderedPageBreak"}:
                continue
            if list(child):
                self._walk_inline(child, chunks)

    def _walk_run(self, run: ET.Element, chunks: list[str]) -> None:
        style = parse_run_style(run.find(qn("w", "rPr")))
        buf: list[str] = []

        def flush() -> None:
            if not buf:
                return
            raw = "".join(buf)
            buf.clear()
            chunks.append(star_correct_marker(raw, style))

        for child in run:
            name = local(child.tag)
            if name == "t":
                buf.append(child.text or "")
            elif name == "tab":
                buf.append(" ")
            elif name == "br":
                buf.append(" ")
            elif name == "cr":
                buf.append(" ")
            elif name == "noBreakHyphen":
                buf.append("-")
            elif name == "softHyphen":
                continue
            elif name == "sym":
                char = child.get(qn("w", "char")) or child.get("char")
                if char:
                    try:
                        buf.append(chr(int(char, 16)))
                    except ValueError:
                        buf.append(char)
            elif name == "drawing":
                flush()
                chunks.append(self._emit_drawing(child))
            elif name == "object":
                flush()
                chunks.append(self._emit_object(child))
            elif name == "pict":
                flush()
                chunks.append(self._emit_vml(child))
            elif name == "footnoteReference":
                continue
            elif name in {"lastRenderedPageBreak", "rPr"}:
                continue
        flush()

    def _emit_mathml(self, om: ET.Element) -> str:
        self._order += 1
        asset_id = self.sidecar.next_id("mathml")
        filename = f"{asset_id}.xml"
        path = self.sidecar.sidecar_dir / filename
        path.write_text(xml_fragment(om), encoding="utf-8")
        placeholder = f"[!m:${asset_id}$]"
        self.sidecar.add(
            Asset(
                id=asset_id,
                kind="mathml",
                placeholder=placeholder,
                sidecar=f"sidecar/{filename}",
                document_order=self._order,
                source_file=self.document_path,
                xpath_hint="/".join(self._xpath_stack + ["m:oMath"]),
            )
        )
        return placeholder

    def _emit_object(self, obj: ET.Element) -> str:
        ole = None
        for el in obj.iter():
            if local(el.tag) == "OLEObject":
                ole = el
                break
        preview_rid = None
        for im in obj.iter(qn("v", "imagedata")):
            preview_rid = im.get(qn("r", "id")) or im.get("id")
            if preview_rid:
                break
        prog = (ole.get("ProgID") if ole is not None else "") or ""
        ole_rid = None
        if ole is not None:
            ole_rid = ole.get(qn("r", "id")) or ole.get("id")
        is_math = (
            prog.lower() in MATHTYPE_PROGIDS
            or "equation" in prog.lower()
            or "mathtype" in prog.lower()
        )
        kind = "mathtype" if is_math else "img"
        return self._copy_binary(
            kind=kind,
            preview_rid=preview_rid,
            ole_rid=ole_rid,
            prog_id=prog,
        )

    def _emit_drawing(self, drawing: ET.Element) -> str:
        rid = None
        for blip in drawing.iter(qn("a", "blip")):
            rid = blip.get(qn("r", "embed")) or blip.get("embed")
            if rid:
                break
        if not rid:
            for im in drawing.iter(qn("v", "imagedata")):
                rid = im.get(qn("r", "id")) or im.get("id")
                if rid:
                    break
        return self._copy_binary(kind="img", preview_rid=rid, ole_rid=None, prog_id=None)

    def _emit_vml(self, pict: ET.Element) -> str:
        rid = None
        for im in pict.iter(qn("v", "imagedata")):
            rid = im.get(qn("r", "id")) or im.get("id")
            if rid:
                break
        return self._copy_binary(kind="img", preview_rid=rid, ole_rid=None, prog_id=None)

    def _copy_binary(
        self,
        kind: str,
        preview_rid: str | None,
        ole_rid: str | None,
        prog_id: str | None,
    ) -> str:
        self._order += 1
        asset_id = self.sidecar.next_id(kind)
        source_file = None
        dest_name = None
        ext = "bin"
        if preview_rid and preview_rid in self.rels:
            target = self.rels[preview_rid]["target"]
            source_file = resolve_media(self.document_path, target)
            ext = Path(source_file).suffix.lstrip(".") or "bin"
            dest_name = f"{asset_id}.{ext}"
            dest = self.sidecar.sidecar_dir / dest_name
            try:
                with self.zip.open(source_file) as src, dest.open("wb") as out:
                    shutil.copyfileobj(src, out)
            except KeyError:
                dest.write_bytes(b"")
        else:
            dest_name = f"{asset_id}.bin"
            (self.sidecar.sidecar_dir / dest_name).write_bytes(b"")

        embedding = None
        if ole_rid and ole_rid in self.rels:
            embedding = resolve_media(self.document_path, self.rels[ole_rid]["target"])
            # Keep the OLE compound file next to the preview for UniMERNet/debug.
            ole_name = f"{asset_id}_ole.bin"
            try:
                with self.zip.open(embedding) as src, (self.sidecar.sidecar_dir / ole_name).open("wb") as out:
                    shutil.copyfileobj(src, out)
            except KeyError:
                pass

        if kind == "mathtype":
            placeholder = f"[!m:${asset_id}$]"
        elif kind == "img":
            placeholder = f"[img:${asset_id}$]"
        else:
            placeholder = f"[!m:${asset_id}$]"

        self.sidecar.add(
            Asset(
                id=asset_id,
                kind=kind,
                placeholder=placeholder,
                sidecar=f"sidecar/{dest_name}",
                document_order=self._order,
                rId=preview_rid,
                source_file=source_file,
                embedding=embedding,
                prog_id=prog_id,
                ole_rId=ole_rid,
                xpath_hint="/".join(self._xpath_stack),
            )
        )
        return placeholder


def collapse_ws_keep_newlines(text: str) -> str:
    text = text.replace("\u00a0", " ")
    text = re.sub(r"[ \t]+", " ", text)
    return text.strip()


def star_correct_marker(raw: str, style: RunStyle) -> str:
    """Prefix an underlined A./B. or a) marker with '*' then apply Azota wraps."""
    if style.underline:
        m = re.match(r"^(\s*)([A-H])\.(\s*)$", raw)
        if m:
            return f"{m.group(1)}*{m.group(2)}.{m.group(3)}"
        m = re.match(r"^(\s*)([a-h])\)(\s*)$", raw)
        if m:
            return f"{m.group(1)}*{m.group(2)}){m.group(3)}"
    return azota_wrap(raw, style)


def _split_mcq_line(text: str) -> list[str]:
    items = [m.group(1).strip() for m in OPTION_ITEM.finditer(text)]
    if len(items) >= 2:
        leftover = OPTION_ITEM.sub("", text).strip()
        if leftover:
            return [leftover] + items
        return items
    return [text]


def unwrap_structural_punctuation(text: str) -> str:
    """Join option letters / question dots split across adjacent wrapped runs."""
    text = re.sub(
        r"\[!(?:b|i|b!i):\$([A-D])\$\]\s*\[!(?:b|i|b!i):\$\.\$\]",
        r"\1.",
        text,
    )
    text = re.sub(r"\[!(?:b|i|b!i):\$([A-D]\.)\$\]", r"\1", text)
    text = re.sub(r"(Câu\s+\d+)\s*\[!(?:b|i|b!i):\$([.:])\$\]", r"\1\2", text)
    text = re.sub(r"(Nhóm\s+[IVXLC]+)\s*\[!(?:b|i|b!i):\$([.:])\$\]", r"\1\2", text)
    return text


def postprocess_lines(lines: list[str]) -> list[str]:
    # 1) Split inline A. B. C. D. into one option per line.
    expanded: list[str] = []
    for line in lines:
        line = unwrap_structural_punctuation(line)
        if SHORT_ANSWER_KEY.match(line):
            expanded.append(line)
            continue
        if re.search(r"[A-D]\.", line) and line.count("A.") + line.count("B.") + line.count("C.") + line.count("D.") >= 2:
            expanded.extend(_split_mcq_line(line))
        else:
            expanded.append(line)

    # 2) Convert [D]/[S] groups into *a) / b) Azota true-false options.
    out: list[str] = []
    tf_index = 0
    i = 0
    while i < len(expanded):
        line = expanded[i]
        if SECTION_RESET.match(line.strip()) or line.strip().startswith("PHẦN"):
            tf_index = 0
        # [GT] → Lời giải ; skip closing [/]
        stripped = line.strip()
        if stripped in {"[GT]", "[GT]:"}:
            out.append("Lời giải")
            i += 1
            continue
        if stripped == "[/]":
            i += 1
            continue
        # short answer pair
        if SHORT_ANSWER_KEY.match(stripped) and i + 1 < len(expanded):
            nxt = expanded[i + 1].strip()
            m = SHORT_ANSWER_VAL.match(nxt)
            if m:
                out.append(f"→ Đáp án: {m.group(1).strip()}")
                i += 2
                continue
        m = DS_PREFIX.match(stripped)
        if m:
            letter = "abcd"[tf_index % 4]
            tf_index += 1
            star = "*" if m.group(1) == "D" else ""
            body = m.group(2).strip()
            out.append(f"{star}{letter}) {body}".rstrip())
            i += 1
            continue
        out.append(line)
        i += 1

    # 3) If a solution says "Chọn X." and the preceding MCQ has no star, add it.
    out = _backfill_stars_from_giai(out)
    # Drop consecutive empty lines beyond 1, keep single blanks between questions.
    cleaned: list[str] = []
    blank = 0
    for line in out:
        if line.strip() == "":
            blank += 1
            if blank <= 1:
                cleaned.append("")
            continue
        blank = 0
        cleaned.append(line)
    return cleaned


def _backfill_stars_from_giai(lines: list[str]) -> list[str]:
    last_q_options: list[int] = []
    result = list(lines)
    in_giai = False
    for idx, line in enumerate(result):
        s = line.strip()
        if s == "Lời giải":
            in_giai = True
            continue
        if SECTION_RESET.match(s) or s.startswith("PHẦN") or s.startswith("Nhóm"):
            in_giai = False
            last_q_options = []
        if re.match(r"^\*?[A-D]\.", s) and not in_giai:
            last_q_options.append(idx)
            if len(last_q_options) > 8:
                last_q_options = last_q_options[-8:]
        m = CHOOSE_LETTER.match(s)
        if m and last_q_options:
            letter = m.group(1)
            if not any(result[j].lstrip().startswith("*") for j in last_q_options):
                for j in last_q_options:
                    if re.match(rf"^{letter}\.", result[j].lstrip()):
                        result[j] = "*" + result[j].lstrip()
                        break
    return result


def convert_docx(docx_path: str | Path, out_dir: str | Path) -> dict[str, Any]:
    out = Path(out_dir)
    out.mkdir(parents=True, exist_ok=True)
    conv = DocxToAzota(Path(docx_path), out)
    try:
        return conv.convert()
    finally:
        conv.close()


# ---------------------------------------------------------------------------
# Optional vision overlays (imported from Colab / GPU hosts)
# ---------------------------------------------------------------------------

def apply_unimernet_latex(manifest: dict[str, Any], predictions: dict[str, str], out_dir: Path) -> dict[str, Any]:
    """Attach UniMERNet LaTeX to mathtype/img formula assets and write .tex sidecars."""
    out_dir = Path(out_dir)
    for asset in manifest.get("assets", []):
        pred = predictions.get(asset["id"])
        if not pred:
            continue
        asset["latex"] = pred
        tex_path = out_dir / "sidecar" / f"{asset['id']}.tex"
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
    parser = argparse.ArgumentParser(description="DOCX → Azota markup.txt / sidecar / manifest.json")
    parser.add_argument("docx", type=Path, help="Input .docx")
    parser.add_argument("-o", "--out", type=Path, default=Path("azota_out"), help="Output directory")
    args = parser.parse_args(argv)
    manifest = convert_docx(args.docx, args.out)
    print(json.dumps(manifest["counts"], ensure_ascii=False, indent=2))
    print(f"Wrote {args.out / 'markup.txt'}")
    print(f"Wrote {args.out / 'manifest.json'}")
    print(f"Wrote {args.out / 'sidecar'} ({len(manifest['assets'])} assets)")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
