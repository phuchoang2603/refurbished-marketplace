"""Render w:tbl as Azota [* c1 | c2 *] rows."""

from __future__ import annotations

from xml.etree import ElementTree as ET

from .ns import qn
from .runs import render_paragraph


def render_table(tbl: ET.Element, ctx, convert_paragraph=None) -> list[str]:
    para = convert_paragraph or render_paragraph
    rows: list[str] = []
    for ri, tr in enumerate(tbl.findall(qn("w", "tr")), start=1):
        cells: list[str] = []
        for ci, tc in enumerate(tr.findall(qn("w", "tc")), start=1):
            ctx.xpath_stack.append(f"tr[{ri}]/tc[{ci}]")
            cell_parts: list[str] = []
            for pi, p in enumerate(tc.findall(qn("w", "p")), start=1):
                ctx.xpath_stack.append(f"p[{pi}]")
                text = para(p, ctx)
                if text:
                    cell_parts.append(text)
                ctx.xpath_stack.pop()
            for ti, nested in enumerate(tc.findall(qn("w", "tbl")), start=1):
                ctx.xpath_stack.append(f"tbl[{ti}]")
                cell_parts.extend(render_table(nested, ctx, para))
                ctx.xpath_stack.pop()
            cells.append(" ".join(cell_parts).strip())
            ctx.xpath_stack.pop()
        rows.append("[* " + " | ".join(cells) + " *]")
    return rows
