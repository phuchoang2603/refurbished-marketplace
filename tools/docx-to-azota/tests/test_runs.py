"""Unit tests for whitelist inline tags (runs.py)."""

from __future__ import annotations

import io
import zipfile
from pathlib import Path
from xml.etree import ElementTree as ET

from docx_extract.assemble import Context
from docx_extract.math_assets import AssetStore
from docx_extract.ns import qn
from docx_extract.runs import RunStyle, azota_wrap, render_paragraph


def _rpr(**kwargs) -> ET.Element:
    rpr = ET.Element(qn("w", "rPr"))
    if kwargs.get("b"):
        ET.SubElement(rpr, qn("w", "b"))
    if kwargs.get("i"):
        ET.SubElement(rpr, qn("w", "i"))
    va = kwargs.get("vert")
    if va:
        ET.SubElement(rpr, qn("w", "vertAlign")).set(qn("w", "val"), va)
    return rpr


def _r(text: str | None = None, **kwargs) -> ET.Element:
    r = ET.Element(qn("w", "r"))
    rpr = kwargs.pop("rpr", None)
    if rpr is None and kwargs:
        rpr = _rpr(**kwargs)
    if rpr is not None:
        r.append(rpr)
    if text is not None:
        t = ET.SubElement(r, qn("w", "t"))
        t.text = text
    return r


def _p(*children: ET.Element) -> ET.Element:
    p = ET.Element(qn("w", "p"))
    for c in children:
        p.append(c)
    return p


def _ctx(tmp_path: Path) -> Context:
    buf = io.BytesIO()
    with zipfile.ZipFile(buf, "w") as zf:
        zf.writestr("dummy", b"")
    zf = zipfile.ZipFile(io.BytesIO(buf.getvalue()), "r")
    return Context(
        zip=zf,
        document_path="word/document.xml",
        rels={},
        sidecar=AssetStore(tmp_path / "sidecar"),
    )


def test_bold_italic_sup_sub_wrap() -> None:
    assert azota_wrap("x", RunStyle(bold=True)) == "[!b:$x$]"
    assert azota_wrap("x", RunStyle(italic=True)) == "[!i:$x$]"
    assert azota_wrap("x", RunStyle(bold=True, italic=True)) == "[!b!i:$x$]"
    assert azota_wrap("2", RunStyle(vert="superscript")) == "[!sup:$2$]"
    assert azota_wrap("2", RunStyle(vert="subscript")) == "[!sub:$2$]"


def test_color_size_font_are_discarded(tmp_path: Path) -> None:
    rpr = ET.Element(qn("w", "rPr"))
    ET.SubElement(rpr, qn("w", "b"))
    ET.SubElement(rpr, qn("w", "color")).set(qn("w", "val"), "FF0000")
    ET.SubElement(rpr, qn("w", "sz")).set(qn("w", "val"), "28")
    ET.SubElement(rpr, qn("w", "rFonts")).set(qn("w", "ascii"), "Times New Roman")
    out = render_paragraph(_p(_r("ab", rpr=rpr)), _ctx(tmp_path))
    assert out == "[!b:$ab$]"


def test_coalesce_adjacent_same_style(tmp_path: Path) -> None:
    out = render_paragraph(_p(_r("a", b=True), _r("b", b=True)), _ctx(tmp_path))
    assert out == "[!b:$ab$]"
    assert out.count("[!b:") == 1


def test_tab_is_space_br_is_space(tmp_path: Path) -> None:
    r = _r()
    ET.SubElement(r, qn("w", "tab"))
    t = ET.SubElement(r, qn("w", "t"))
    t.text = "x"
    br_r = _r()
    ET.SubElement(br_r, qn("w", "br"))
    t2 = ET.SubElement(br_r, qn("w", "t"))
    t2.text = "y"
    out = render_paragraph(_p(r, br_r), _ctx(tmp_path))
    assert out == "x y"


def test_wrap_moves_outer_spaces_outside_tags() -> None:
    assert azota_wrap(" ab ", RunStyle(italic=True)) == " [!i:$ab$] "
