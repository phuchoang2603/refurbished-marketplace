"""OOXML namespaces and tiny XML helpers."""

from __future__ import annotations

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

for prefix, uri in NS.items():
    ET.register_namespace(prefix, uri)


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
