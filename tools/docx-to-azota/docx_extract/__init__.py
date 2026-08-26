"""Tầng A: DOCX → Azota markup.txt + sidecar + manifest.json."""

from .assemble import clean_docx, convert_docx, main
from .answer import star_correct_marker
from .runs import RunStyle, azota_wrap, postprocess_lines, render_paragraph

__all__ = [
    "clean_docx",
    "convert_docx",
    "main",
    "RunStyle",
    "azota_wrap",
    "postprocess_lines",
    "render_paragraph",
    "star_correct_marker",
]
