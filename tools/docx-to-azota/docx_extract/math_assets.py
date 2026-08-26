"""Emit Azota math/image placeholders and write sidecar assets."""

from __future__ import annotations

import shutil
from dataclasses import dataclass, field
from pathlib import Path
from typing import Any
from xml.etree import ElementTree as ET

from .loader import resolve_media
from .ns import local, qn, xml_fragment

MATHTYPE_PROGIDS = {
    "equation.dsmt4",
    "equation.3",
    "equation.dsmt6",
    "wiris.equation",
    "mathtype.equation",
}


@dataclass
class Asset:
    id: str
    kind: str
    placeholder: str
    sidecar: str
    document_order: int
    rId: str | None = None
    source_file: str | None = None
    embedding: str | None = None
    prog_id: str | None = None
    ole_rId: str | None = None
    xpath_hint: str | None = None
    latex: str | None = None
    extras: dict[str, Any] = field(default_factory=dict)

    def to_json(self) -> dict[str, Any]:
        data = {
            "id": self.id,
            "kind": self.kind,
            "placeholder": self.placeholder,
            "sidecar": self.sidecar,
            "document_order": self.document_order,
            "rId": self.rId,
            "source_file": self.source_file,
            "embedding": self.embedding,
            "prog_id": self.prog_id,
            "ole_rId": self.ole_rId,
            "xpath_hint": self.xpath_hint,
        }
        if self.latex is not None:
            data["latex"] = self.latex
        if self.extras:
            data["extras"] = self.extras
        return data


class AssetStore:
    def __init__(self, sidecar_dir: Path) -> None:
        self.sidecar_dir = sidecar_dir
        self.sidecar_dir.mkdir(parents=True, exist_ok=True)
        self.assets: list[Asset] = []
        self._counts = {"mathml": 0, "mathtype": 0, "img": 0}

    def next_id(self, kind: str) -> str:
        # Per-kind counters (mathml_1 …, mathtype_1 …, img_1 …). Spec §4.2 also
        # describes one global N; Azota panel dumps may skip ids. Until a user
        # dump exists, per-kind matches the sample golden and keeps prefixes typed.
        self._counts[kind] += 1
        return f"{kind}_{self._counts[kind]}"

    def add(self, asset: Asset) -> Asset:
        self.assets.append(asset)
        return asset


def emit_mathml(om: ET.Element, ctx) -> str:
    ctx.order += 1
    asset_id = ctx.sidecar.next_id("mathml")
    filename = f"{asset_id}.xml"
    (ctx.sidecar.sidecar_dir / filename).write_text(xml_fragment(om), encoding="utf-8")
    placeholder = f"[!m:${asset_id}$]"
    ctx.sidecar.add(
        Asset(
            id=asset_id,
            kind="mathml",
            placeholder=placeholder,
            sidecar=f"sidecar/{filename}",
            document_order=ctx.order,
            source_file=ctx.document_path,
            xpath_hint="/".join(ctx.xpath_stack + ["m:oMath"]),
        )
    )
    return placeholder


def emit_object(obj: ET.Element, ctx) -> str:
    ole = None
    for el in obj.iter():
        if local(el.tag) == "OLEObject":
            ole = el
            break
    preview_rid = None
    for im in obj.iter(qn("v", "imagedata")):
        preview_rid = im.get(qn("r", "id")) or im.get("id")
        if preview_rid:
            break
    prog = (ole.get("ProgID") if ole is not None else "") or ""
    ole_rid = None
    if ole is not None:
        ole_rid = ole.get(qn("r", "id")) or ole.get("id")
    is_math = (
        prog.lower() in MATHTYPE_PROGIDS
        or "equation" in prog.lower()
        or "mathtype" in prog.lower()
    )
    kind = "mathtype" if is_math else "img"
    return copy_binary(ctx, kind=kind, preview_rid=preview_rid, ole_rid=ole_rid, prog_id=prog)


def emit_drawing(drawing: ET.Element, ctx) -> str:
    rid = None
    for blip in drawing.iter(qn("a", "blip")):
        rid = blip.get(qn("r", "embed")) or blip.get("embed")
        if rid:
            break
    if not rid:
        for im in drawing.iter(qn("v", "imagedata")):
            rid = im.get(qn("r", "id")) or im.get("id")
            if rid:
                break
    return copy_binary(ctx, kind="img", preview_rid=rid, ole_rid=None, prog_id=None)


def emit_vml(pict: ET.Element, ctx) -> str:
    rid = None
    for im in pict.iter(qn("v", "imagedata")):
        rid = im.get(qn("r", "id")) or im.get("id")
        if rid:
            break
    return copy_binary(ctx, kind="img", preview_rid=rid, ole_rid=None, prog_id=None)


def copy_binary(
    ctx,
    kind: str,
    preview_rid: str | None,
    ole_rid: str | None,
    prog_id: str | None,
) -> str:
    ctx.order += 1
    asset_id = ctx.sidecar.next_id(kind)
    source_file = None
    dest_name = None
    if preview_rid and preview_rid in ctx.rels:
        target = ctx.rels[preview_rid]["target"]
        source_file = resolve_media(ctx.document_path, target)
        ext = Path(source_file).suffix.lstrip(".") or "bin"
        dest_name = f"{asset_id}.{ext}"
        dest = ctx.sidecar.sidecar_dir / dest_name
        try:
            with ctx.zip.open(source_file) as src, dest.open("wb") as out:
                shutil.copyfileobj(src, out)
        except KeyError:
            dest.write_bytes(b"")
            ctx.warnings.append(f"missing media {source_file} for {asset_id}")
    else:
        dest_name = f"{asset_id}.bin"
        (ctx.sidecar.sidecar_dir / dest_name).write_bytes(b"")
        if preview_rid:
            ctx.warnings.append(f"missing rId {preview_rid} for {asset_id}")

    embedding = None
    if ole_rid and ole_rid in ctx.rels:
        embedding = resolve_media(ctx.document_path, ctx.rels[ole_rid]["target"])
        ole_name = f"{asset_id}_ole.bin"
        try:
            with ctx.zip.open(embedding) as src, (ctx.sidecar.sidecar_dir / ole_name).open("wb") as out:
                shutil.copyfileobj(src, out)
        except KeyError:
            ctx.warnings.append(f"missing OLE {embedding} for {asset_id}")

    if kind == "img":
        placeholder = f"[img:${asset_id}$]"
    else:
        placeholder = f"[!m:${asset_id}$]"

    ctx.sidecar.add(
        Asset(
            id=asset_id,
            kind=kind,
            placeholder=placeholder,
            sidecar=f"sidecar/{dest_name}",
            document_order=ctx.order,
            rId=preview_rid,
            source_file=source_file,
            embedding=embedding,
            prog_id=prog_id,
            ole_rId=ole_rid,
            xpath_hint="/".join(ctx.xpath_stack),
        )
    )
    return placeholder
