"""OMML → LaTeX (CPU, no UniMERNet)."""

from __future__ import annotations

from xml.etree import ElementTree as ET

from docx_extract.ns import qn
from docx_extract.omml_latex import omml_to_latex


def _omath(*children: ET.Element) -> ET.Element:
    om = ET.Element(qn("m", "oMath"))
    for c in children:
        om.append(c)
    return om


def _t(text: str) -> ET.Element:
    r = ET.Element(qn("m", "r"))
    t = ET.SubElement(r, qn("m", "t"))
    t.text = text
    return r


def test_plain_text() -> None:
    assert omml_to_latex(_omath(_t("E=mc"))) == "E=mc"


def test_fraction_and_subscript() -> None:
    f = ET.Element(qn("m", "f"))
    num = ET.SubElement(f, qn("m", "num"))
    num.append(_t("W"))
    den = ET.SubElement(f, qn("m", "den"))
    den.append(_t("A"))
    sub = ET.Element(qn("m", "sSub"))
    e = ET.SubElement(sub, qn("m", "e"))
    e.append(_t("W"))
    sb = ET.SubElement(sub, qn("m", "sub"))
    sb.append(_t("lkr"))
    latex = omml_to_latex(_omath(sub, _t("="), f))
    assert r"\frac{W}{A}" in latex
    assert r"{W}_{lkr}" in latex or "W_{lkr}" in latex.replace("{W}", "W")


def test_pi_symbol() -> None:
    assert r"\pi" in omml_to_latex(_omath(_t("π=3,14")))
