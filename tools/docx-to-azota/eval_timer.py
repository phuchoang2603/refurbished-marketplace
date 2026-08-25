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
