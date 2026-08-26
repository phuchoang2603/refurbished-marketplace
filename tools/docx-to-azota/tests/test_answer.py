"""Unit tests for correct-answer starring (answer.py)."""

from __future__ import annotations

from docx_extract.answer import backfill_stars_from_giai, star_correct_marker
from docx_extract.runs import RunStyle


def test_underline_letter_gets_star() -> None:
    out = star_correct_marker("D.", RunStyle(underline=True))
    assert out.startswith("*D")


def test_plain_letter_no_star() -> None:
    out = star_correct_marker("D.", RunStyle())
    assert not out.startswith("*")


def test_backfill_from_chon() -> None:
    lines = [
        "Câu 1:",
        "A. sai",
        "B. đúng",
        "Lời giải",
        "Chọn B.",
    ]
    out = backfill_stars_from_giai(lines)
    assert any(x.startswith("*B") for x in out)
