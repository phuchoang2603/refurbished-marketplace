"""Unit tests for Azota table rows (tables.py)."""

from __future__ import annotations

import io
import zipfile
from pathlib import Path
from xml.etree import ElementTree as ET

from docx_extract.assemble import Context
from docx_extract.math_assets import AssetStore
from docx_extract.ns import qn
from docx_extract.tables import render_table


def _cell(text: str) -> ET.Element:
    tc = ET.Element(qn("w", "tc"))
    p = ET.SubElement(tc, qn("w", "p"))
    r = ET.SubElement(p, qn("w", "r"))
    t = ET.SubElement(r, qn("w", "t"))
    t.text = text
    return tc


def test_one_row_three_cells(tmp_path: Path) -> None:
    tbl = ET.Element(qn("w", "tbl"))
    tr = ET.SubElement(tbl, qn("w", "tr"))
    tr.append(_cell("c1"))
    tr.append(_cell("c2"))
    tr.append(_cell("c3"))
    buf = io.BytesIO()
    with zipfile.ZipFile(buf, "w") as zf:
        zf.writestr("dummy", b"")
    zf = zipfile.ZipFile(io.BytesIO(buf.getvalue()), "r")
    ctx = Context(
        zip=zf,
        document_path="word/document.xml",
        rels={},
        sidecar=AssetStore(tmp_path / "sidecar"),
    )
    out = render_table(tbl, ctx)
    assert out == ["[* c1 | c2 | c3 *]"]
