"""Load a .docx (zip) into document.xml + rId map + zip handle."""

from __future__ import annotations

import re
import zipfile
from dataclasses import dataclass, field
from pathlib import Path
from xml.etree import ElementTree as ET

from .ns import PKG_REL, qn


def load_document_target(docx_zip: zipfile.ZipFile) -> str:
    rels = ET.fromstring(docx_zip.read("_rels/.rels"))
    for rel in rels:
        if rel.get("Type") == PKG_REL:
            target = rel.get("Target", "word/document.xml")
            return target.lstrip("/")
    return "word/document.xml"


def load_rels(docx_zip: zipfile.ZipFile, document_path: str) -> dict[str, dict[str, str]]:
    rels_path = str(Path(document_path).parent / "_rels" / (Path(document_path).name + ".rels"))
    mapping: dict[str, dict[str, str]] = {}
    try:
        root = ET.fromstring(docx_zip.read(rels_path))
    except KeyError:
        return mapping
    for rel in root:
        rid = rel.get("Id")
        if not rid:
            continue
        mapping[rid] = {
            "id": rid,
            "type": rel.get("Type", ""),
            "target": rel.get("Target", ""),
            "target_mode": rel.get("TargetMode", "Internal"),
        }
    return mapping


def resolve_media(document_path: str, target: str) -> str:
    base = Path(document_path).parent
    resolved = (base / target).as_posix()
    while "/../" in resolved:
        resolved = re.sub(r"[^/]+/\.\./", "", resolved)
    return resolved


@dataclass
class Document:
    zip: zipfile.ZipFile
    document_path: str
    body: ET.Element
    rels: dict[str, dict[str, str]]
    doc_root: ET.Element = field(repr=False)


def load(path: str | Path) -> Document:
    zf = zipfile.ZipFile(Path(path))
    document_path = load_document_target(zf)
    doc_root = ET.fromstring(zf.read(document_path))
    body = doc_root.find(qn("w", "body"))
    if body is None:
        zf.close()
        raise ValueError("document.xml has no w:body")
    return Document(
        zip=zf,
        document_path=document_path,
        body=body,
        rels=load_rels(zf, document_path),
        doc_root=doc_root,
    )
