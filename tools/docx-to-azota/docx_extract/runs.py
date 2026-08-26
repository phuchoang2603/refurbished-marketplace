"""Whitelist inline tags: bold / italic / sup / sub; coalesce same-style text."""

from __future__ import annotations

import re
from dataclasses import dataclass
from xml.etree import ElementTree as ET

from .math_assets import emit_drawing, emit_mathml, emit_object, emit_vml
from .ns import local, on_off, qn

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


@dataclass
class RunStyle:
    bold: bool = False
    italic: bool = False
    underline: bool = False
    vert: str | None = None

    def wrap_key(self) -> tuple[bool, bool, str | None]:
        return (self.bold, self.italic, self.vert)

    def coalesce_key(self) -> tuple[bool, bool, bool, str | None]:
        return (self.bold, self.italic, self.underline, self.vert)


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
    lead_len = len(text) - len(text.lstrip())
    trail_len = len(text) - len(text.rstrip())
    lead = text[:lead_len]
    trail = text[len(text) - trail_len :] if trail_len else ""
    core = text[lead_len : len(text) - trail_len if trail_len else None]
    stripped = core.strip()
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
    out = core
    if style.vert == "superscript":
        out = f"[!sup:${out}$]"
    elif style.vert == "subscript":
        out = f"[!sub:${out}$]"
    if skip_bi:
        return lead + out + trail
    if style.bold and style.italic:
        out = f"[!b!i:${out}$]"
    elif style.bold:
        out = f"[!b:${out}$]"
    elif style.italic:
        out = f"[!i:${out}$]"
    return lead + out + trail


def collapse_ws_keep_newlines(text: str) -> str:
    text = text.replace("\u00a0", " ")
    text = re.sub(r"[ \t]+", " ", text)
    return text.strip()


def _coalesce_and_wrap(pieces: list[tuple]) -> list[str]:
    """Merge adjacent same-style text, then wrap / star."""
    from .answer import star_correct_marker

    merged: list[tuple] = []
    for piece in pieces:
        if piece[0] == "raw":
            merged.append(piece)
            continue
        _, text, style = piece
        if (
            merged
            and merged[-1][0] == "text"
            and merged[-1][2].coalesce_key() == style.coalesce_key()
        ):
            merged[-1] = ("text", merged[-1][1] + text, style)
        else:
            merged.append(("text", text, style))
    chunks: list[str] = []
    for item in merged:
        if item[0] == "raw":
            chunks.append(item[1])
        else:
            chunks.append(star_correct_marker(item[1], item[2]))
    return chunks


def render_paragraph(p: ET.Element, ctx) -> str:
    """``w:p`` → one markup string (empty if the paragraph has no whitelist content)."""
    pieces: list[tuple] = []
    _walk_inline(p, ctx, pieces)
    return collapse_ws_keep_newlines("".join(_coalesce_and_wrap(pieces)))


def _walk_inline(el: ET.Element, ctx, pieces: list[tuple]) -> None:
    for child in el:
        name = local(child.tag)
        if name in {"pPr", "rPr", "tblPr", "tblGrid", "sectPr", "commentRangeStart", "commentRangeEnd"}:
            continue
        if name == "oMath":
            pieces.append(("raw", emit_mathml(child, ctx)))
            continue
        if name == "oMathPara":
            for om in child.findall(qn("m", "oMath")):
                pieces.append(("raw", emit_mathml(om, ctx)))
            continue
        if name == "r":
            _walk_run(child, ctx, pieces)
            continue
        if name in {"hyperlink", "ins", "del", "smartTag", "sdt", "sdtContent", "fldSimple"}:
            _walk_inline(child, ctx, pieces)
            continue
        if name == "object":
            pieces.append(("raw", emit_object(child, ctx)))
            continue
        if name == "drawing":
            pieces.append(("raw", emit_drawing(child, ctx)))
            continue
        if name == "pict":
            pieces.append(("raw", emit_vml(child, ctx)))
            continue
        if name in {"bookmarkStart", "bookmarkEnd", "proofErr", "lastRenderedPageBreak"}:
            continue
        if list(child):
            _walk_inline(child, ctx, pieces)


def _walk_run(run: ET.Element, ctx, pieces: list[tuple]) -> None:
    style = parse_run_style(run.find(qn("w", "rPr")))
    buf: list[str] = []

    def flush() -> None:
        if not buf:
            return
        pieces.append(("text", "".join(buf), style))
        buf.clear()

    for child in run:
        name = local(child.tag)
        if name == "t":
            buf.append(child.text or "")
        elif name == "tab":
            buf.append(" ")
        elif name in {"br", "cr"}:
            # Paragraphs map to Azota lines; a soft break stays a space.
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
            pieces.append(("raw", emit_drawing(child, ctx)))
        elif name == "object":
            flush()
            pieces.append(("raw", emit_object(child, ctx)))
        elif name == "pict":
            flush()
            pieces.append(("raw", emit_vml(child, ctx)))
        elif name == "footnoteReference":
            continue
        elif name in {"lastRenderedPageBreak", "rPr"}:
            continue
    flush()


def unwrap_structural_punctuation(text: str) -> str:
    text = re.sub(
        r"\[!(?:b|i|b!i):\$([A-D])\$\]\s*\[!(?:b|i|b!i):\$\.\$\]",
        r"\1.",
        text,
    )
    text = re.sub(r"\[!(?:b|i|b!i):\$([A-D]\.)\$\]", r"\1", text)
    text = re.sub(r"(Câu\s+\d+)\s*\[!(?:b|i|b!i):\$([.:])\$\]", r"\1\2", text)
    text = re.sub(r"(Nhóm\s+[IVXLC]+)\s*\[!(?:b|i|b!i):\$([.:])\$\]", r"\1\2", text)
    return text


def _split_mcq_line(text: str) -> list[str]:
    items = [m.group(1).strip() for m in OPTION_ITEM.finditer(text)]
    if len(items) >= 2:
        leftover = OPTION_ITEM.sub("", text).strip()
        if leftover:
            return [leftover] + items
        return items
    return [text]


def postprocess_lines(lines: list[str]) -> list[str]:
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

    out: list[str] = []
    tf_index = 0
    i = 0
    while i < len(expanded):
        line = expanded[i]
        if SECTION_RESET.match(line.strip()) or line.strip().startswith("PHẦN"):
            tf_index = 0
        stripped = line.strip()
        if stripped in {"[GT]", "[GT]:"}:
            out.append("Lời giải")
            i += 1
            continue
        if stripped == "[/]":
            i += 1
            continue
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

    from .answer import backfill_stars_from_giai

    out = backfill_stars_from_giai(out)
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
