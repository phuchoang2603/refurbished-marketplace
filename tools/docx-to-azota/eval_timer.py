"""Step timer for Colab evaluation (Bước 1 / 2 / 3 / 4)."""

from __future__ import annotations

import time
from contextlib import contextmanager
from dataclasses import dataclass, field


@dataclass
class StepTimer:
    steps: list[tuple[str, str, float]] = field(default_factory=list)

    @contextmanager
    def step(self, code: str, label: str):
        t0 = time.perf_counter()
        try:
            yield
        finally:
            elapsed = time.perf_counter() - t0
            self.steps.append((code, label, elapsed))
            print(f"{code}: {fmt_seconds(elapsed)} ({label})")

    def summary(self) -> str:
        lines = ["⏱ THỜI GIAN:"]
        total = 0.0
        for code, label, elapsed in self.steps:
            lines.append(f"{code}: {fmt_seconds(elapsed)} ({label})")
            total += elapsed
        lines.append(f"TỔNG : {fmt_seconds(total)}")
        return "\n".join(lines)

    def print_summary(self) -> None:
        print(self.summary())


def fmt_seconds(seconds: float) -> str:
    if seconds < 10:
        return f"{seconds:.1f}s"
    if seconds < 120:
        return f"{seconds:.0f}s"
    return f"{seconds:.0f}s ({seconds / 60:.2f} phút)"


def fmt_giay_phut(seconds: float) -> str:
    """Kaggle-style: ``85.16 giây (1.42 phút)``."""
    return f"{seconds:.2f} giây ({seconds / 60:.2f} phút)"


def print_ocr_complete(
    input_path: str,
    page_seconds: list[float],
    pdf_to_png_seconds: float | None = None,
) -> str:
    """Print the Kaggle OCR completion banner (per-page + totals)."""
    n = len(page_seconds)
    total = sum(page_seconds)
    avg = (total / n) if n else 0.0
    lines = [
        "OCR HOÀN TẤT",
        "=" * 80,
        f"Input: {input_path}",
        f"Số trang/ảnh: {n}",
        "===== THỜI GIAN TỪNG TRANG =====",
    ]
    for i, sec in enumerate(page_seconds, start=1):
        lines.append(f"  Page {i}: {fmt_giay_phut(sec)}")
    lines.extend(
        [
            "===== THỐNG KÊ OCR =====",
            f"  Tổng thời gian OCR: {total:.2f} giây",
            f"                      {total / 60:.2f} phút",
            f"  Trung bình mỗi trang: {avg:.2f} giây",
            f"                       {avg / 60:.2f} phút",
            "===== THỜI GIAN CONVERSION =====",
        ]
    )
    if pdf_to_png_seconds is None:
        lines.append("  PDF -> PNG: (không convert PDF)")
    else:
        lines.append(f"  PDF -> PNG: {pdf_to_png_seconds:.2f} giây")
    text = "\n".join(lines)
    print(text)
    return text


def print_extract_complete(
    input_path: str,
    convert_seconds: float,
    counts: dict | None = None,
) -> str:
    """Same banner layout as the Kaggle OCR log, for Tầng A (DOCX → Azota)."""
    counts = counts or {}
    lines = [
        "EXTRACT HOÀN TẤT",
        "=" * 80,
        f"Input: {input_path}",
        "Số trang/ảnh: — (DOCX / OOXML, không OCR từng trang)",
        "===== THỜI GIAN TỪNG BƯỚC =====",
        f"  Bước 1 (OOXML): {fmt_giay_phut(convert_seconds)}",
        "===== THỐNG KÊ =====",
        f"  mathml: {counts.get('mathml', 0)}",
        f"  mathtype: {counts.get('mathtype', 0)}",
        f"  img: {counts.get('img', 0)}",
        f"  dòng markup: {counts.get('lines', 0)}",
        "===== THỜI GIAN CONVERSION =====",
        f"  DOCX -> Azota: {convert_seconds:.2f} giây",
    ]
    text = "\n".join(lines)
    print(text)
    return text
