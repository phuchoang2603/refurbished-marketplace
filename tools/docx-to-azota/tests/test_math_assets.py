"""Unit tests for math/image placeholders (math_assets.py)."""

from __future__ import annotations

import io
import zipfile
from pathlib import Path
from xml.etree import ElementTree as ET

from docx_extract.assemble import Context
from docx_extract.math_assets import AssetStore, emit_drawing, emit_mathml, emit_object
from docx_extract.ns import qn


def _zip_with(name: str, data: bytes) -> zipfile.ZipFile:
    buf = io.BytesIO()
    with zipfile.ZipFile(buf, "w") as zf:
        zf.writestr(name, data)
    return zipfile.ZipFile(io.BytesIO(buf.getvalue()), "r")


def _ctx(tmp_path: Path, zf: zipfile.ZipFile | None = None, rels: dict | None = None) -> Context:
    if zf is None:
        zf = _zip_with("dummy", b"")
    return Context(
        zip=zf,
        document_path="word/document.xml",
        rels=rels or {},
        sidecar=AssetStore(tmp_path / "sidecar"),
    )


def test_omath_emits_mathml_and_increments(tmp_path: Path) -> None:
    ctx = _ctx(tmp_path)
    omath = ET.Element(qn("m", "oMath"))
    ET.SubElement(omath, qn("m", "r"))
    ph = emit_mathml(omath, ctx)
    assert ph == "[!m:$mathml_1$]"
    assert ctx.sidecar.assets[0].kind == "mathml"
    assert b"<m:oMath" in ctx.sidecar.assets[0].sidecar.encode() or (
        tmp_path / "sidecar" / "mathml_1.xml"
    ).read_bytes()
    xml = (tmp_path / "sidecar" / "mathml_1.xml").read_text(encoding="utf-8")
    assert "<m:oMath" in xml


def test_drawing_emits_img(tmp_path: Path) -> None:
    zf = _zip_with("word/media/image1.png", b"\x89PNG")
    ctx = _ctx(
        tmp_path,
        zf=zf,
        rels={"rId5": {"target": "media/image1.png", "type": "", "id": "rId5"}},
    )
    drawing = ET.Element(qn("w", "drawing"))
    inline = ET.SubElement(drawing, qn("wp", "inline"))
    blip = ET.SubElement(inline, qn("a", "blip"))
    blip.set(qn("r", "embed"), "rId5")
    ph = emit_drawing(drawing, ctx)
    assert ph == "[img:$img_1$]"
    assert ctx.sidecar.assets[0].id == "img_1"
    assert (tmp_path / "sidecar" / "img_1.png").read_bytes() == b"\x89PNG"


def test_mathtype_ole_emits_placeholder(tmp_path: Path) -> None:
    zf = _zip_with("word/media/image1.wmf", b"wmf-bytes")
    ctx = _ctx(
        tmp_path,
        zf=zf,
        rels={"rId9": {"target": "media/image1.wmf", "type": "", "id": "rId9"}},
    )
    obj = ET.Element(qn("w", "object"))
    ole = ET.SubElement(obj, qn("o", "OLEObject"))
    ole.set("ProgID", "Equation.DSMT4")
    ole.set(qn("r", "id"), "rId9")
    imagedata = ET.SubElement(obj, qn("v", "imagedata"))
    imagedata.set(qn("r", "id"), "rId9")
    ph = emit_object(obj, ctx)
    assert ph == "[!m:$mathtype_1$]"
    assert (tmp_path / "sidecar" / "mathtype_1.wmf").read_bytes() == b"wmf-bytes"


def test_per_kind_counters(tmp_path: Path) -> None:
    ctx = _ctx(tmp_path)
    emit_mathml(ET.Element(qn("m", "oMath")), ctx)
    zf = _zip_with("word/media/a.png", b"PNG")
    ctx.zip = zf
    ctx.rels = {"rId1": {"target": "media/a.png", "type": "", "id": "rId1"}}
    drawing = ET.Element(qn("w", "drawing"))
    blip = ET.SubElement(drawing, qn("a", "blip"))
    blip.set(qn("r", "embed"), "rId1")
    ph = emit_drawing(drawing, ctx)
    assert ph == "[img:$img_1$]"
    assert ctx.order == 2
    assert [a.id for a in ctx.sidecar.assets] == ["mathml_1", "img_1"]
