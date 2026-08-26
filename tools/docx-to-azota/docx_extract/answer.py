"""Mark the correct MCQ letter with '*' (underline first, then fallbacks)."""

from __future__ import annotations

import re

from .runs import CHOOSE_LETTER, SECTION_RESET, RunStyle, azota_wrap


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


def backfill_stars_from_giai(lines: list[str]) -> list[str]:
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
