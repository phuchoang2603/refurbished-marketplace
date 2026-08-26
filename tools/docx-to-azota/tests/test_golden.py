"""Golden regression: sample physics exam vs checked-in Azota markup."""

from __future__ import annotations

import json
from pathlib import Path

from convert import convert_docx

SAMPLE = Path(__file__).resolve().parents[1] / "samples" / "de-vat-li-lan-3.docx"
GOLDEN = Path(__file__).resolve().parents[1] / "examples" / "de-vat-li-lan-3" / "markup.txt"


def test_golden_markup_matches_example(tmp_path: Path) -> None:
    assert SAMPLE.is_file(), f"missing sample {SAMPLE}"
    assert GOLDEN.is_file(), f"missing golden {GOLDEN}"
    out_dir = tmp_path / "out"
    convert_docx(SAMPLE, out_dir)
    got = (out_dir / "markup.txt").read_text(encoding="utf-8")
    expected = GOLDEN.read_text(encoding="utf-8")
    if got != expected:
        g_lines = got.splitlines()
        e_lines = expected.splitlines()
        diffs = []
        for i, (a, b) in enumerate(zip(g_lines, e_lines), start=1):
            if a != b:
                diffs.append(f"L{i}: got={a!r} expected={b!r}")
            if len(diffs) >= 8:
                break
        if len(g_lines) != len(e_lines):
            diffs.append(f"line count {len(g_lines)} vs {len(e_lines)}")
        raise AssertionError("golden mismatch:\n" + "\n".join(diffs))


def test_asset_counts_and_key_lines(tmp_path: Path) -> None:
    out_dir = tmp_path / "out"
    convert_docx(SAMPLE, out_dir)
    data = json.loads((out_dir / "manifest.json").read_text(encoding="utf-8"))
    kinds: dict[str, int] = {}
    for a in data["assets"]:
        kinds[a["kind"]] = kinds.get(a["kind"], 0) + 1
    assert kinds.get("mathml") == 69
    assert kinds.get("mathtype") == 16
    assert kinds.get("img") == 8
    text = (out_dir / "markup.txt").read_text(encoding="utf-8")
    assert "*D. ngưng tụ." in text
    assert "→ Đáp án: 69,6" in text
