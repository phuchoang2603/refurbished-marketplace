from __future__ import annotations

import io
import json
import zipfile
from pathlib import Path
from xml.etree import ElementTree as ET

import pytest

from convert import convert_docx, postprocess_lines, star_correct_marker, RunStyle, azota_wrap

W = "http://schemas.openxmlformats.org/wordprocessingml/2006/main"
M = "http://schemas.openxmlformats.org/officeDocument/2006/math"
R = "http://schemas.openxmlformats.org/officeDocument/2006/relationships"
REL = "http://schemas.openxmlformats.org/package/2006/relationships"
OD = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument"
IMG = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/image"


def _el(tag: str, text: str | None = None, **attrs) -> ET.Element:
    el = ET.Element(tag, {k: v for k, v in attrs.items() if v is not None})
    if text is not None:
        el.text = text
    return el


def _r(text: str, *, b=False, i=False, u=False, vert: str | None = None) -> ET.Element:
    run = _el(f"{{{W}}}r")
    rpr = _el(f"{{{W}}}rPr")
    if b:
        rpr.append(_el(f"{{{W}}}b"))
    if i:
        rpr.append(_el(f"{{{W}}}i"))
    if u:
        rpr.append(_el(f"{{{W}}}u", **{f"{{{W}}}val": "single"}))
    if vert:
        rpr.append(_el(f"{{{W}}}vertAlign", **{f"{{{W}}}val": vert}))
    run.append(rpr)
    t = _el(f"{{{W}}}t", text)
    t.set("{http://www.w3.org/XML/1998/namespace}space", "preserve")
    run.append(t)
    return run


def _p(*runs: ET.Element) -> ET.Element:
    p = _el(f"{{{W}}}p")
    for run in runs:
        p.append(run)
    return p


def _omath(text: str) -> ET.Element:
    om = _el(f"{{{M}}}oMath")
    r = _el(f"{{{M}}}r")
    t = _el(f"{{{M}}}t", text)
    r.append(t)
    om.append(r)
    return om


def build_docx(body_children: list[ET.Element], media: dict[str, bytes] | None = None) -> bytes:
    media = media or {}
    document = _el(f"{{{W}}}document")
    body = _el(f"{{{W}}}body")
    for child in body_children:
        body.append(child)
    document.append(body)
    doc_xml = ET.tostring(document, encoding="utf-8", xml_declaration=True)

    rels_pkg = _el(f"{{{REL}}}Relationships")
    rels_pkg.append(
        _el(
            f"{{{REL}}}Relationship",
            Id="rId1",
            Type=OD,
            Target="word/document.xml",
        )
    )

    doc_rels = _el(f"{{{REL}}}Relationships")
    rid_n = 1
    media_entries: list[tuple[str, str, bytes]] = []
    for name, blob in media.items():
        rid_n += 1
        rid = f"rId{rid_n}"
        doc_rels.append(_el(f"{{{REL}}}Relationship", Id=rid, Type=IMG, Target=f"media/{name}"))
        media_entries.append((rid, name, blob))

    buf = io.BytesIO()
    with zipfile.ZipFile(buf, "w") as zf:
        zf.writestr("[Content_Types].xml", """<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Default Extension="png" ContentType="image/png"/>
  <Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
</Types>
""")
        zf.writestr("_rels/.rels", ET.tostring(rels_pkg, encoding="utf-8", xml_declaration=True))
        zf.writestr("word/document.xml", doc_xml)
        zf.writestr("word/_rels/document.xml.rels", ET.tostring(doc_rels, encoding="utf-8", xml_declaration=True))
        for _rid, name, blob in media_entries:
            zf.writestr(f"word/media/{name}", blob)
    return buf.getvalue()


def test_azota_wrap_and_star():
    assert azota_wrap("hello", RunStyle(bold=True)) == "[!b:$hello$]"
    assert azota_wrap("x", RunStyle(italic=True)) == "[!i:$x$]"
    assert azota_wrap("ab", RunStyle(bold=True, italic=True)) == "[!b!i:$ab$]"
    assert azota_wrap("2", RunStyle(vert="superscript")) == "[!sup:$2$]"
    assert azota_wrap("i", RunStyle(vert="subscript")) == "[!sub:$i$]"
    assert azota_wrap("Câu 1.", RunStyle(bold=True)) == "Câu 1."
    starred = star_correct_marker("D.", RunStyle(bold=True, underline=True))
    assert starred == "*D."


def test_postprocess_mcq_tf_short_answer():
    lines = [
        "Câu 1. Hoi",
        "A. mot.B. hai.C. ba.*D. bon.",
        "[GT]",
        "Chọn D.",
        "[/]",
        "Nhóm I. stem",
        "[S] sai",
        "[D] dung",
        "[S] sai 2",
        "[D] dung 2",
        "Câu 2. short",
        "Đáp án là {a}",
        "A.69,6",
    ]
    out = postprocess_lines(lines)
    assert "A. mot." in out
    assert "B. hai." in out
    assert "*D. bon." in out
    assert "Lời giải" in out
    assert "[/]" not in out
    assert "a) sai" in out
    assert "*b) dung" in out
    assert "c) sai 2" in out
    assert "*d) dung 2" in out
    assert "→ Đáp án: 69,6" in out
    unwrapped = postprocess_lines(
        ["Câu 6. x", "A. Áp kế.B. Pit. [!b:$C$][!b:$.$] Giá đỡ.*D. Cân."]
    )
    assert any(l.startswith("C.") for l in unwrapped)


def test_convert_mini_docx(tmp_path: Path):
    p1 = _p(_r("Câu 1. Van de ", b=True), _r("nang luong ", i=True))
    p1.append(_omath("E=mc^2"))
    p2 = _p(_r("A. sai"), _r("B. ", b=True), _r("C. "), _r("D.", b=True, u=True), _r(" dung"))
    blob = build_docx([p1, p2])
    src = tmp_path / "mini.docx"
    src.write_bytes(blob)
    out = tmp_path / "out"
    manifest = convert_docx(src, out)
    markup = (out / "markup.txt").read_text(encoding="utf-8")
    assert "[!m:$mathml_1$]" in markup
    assert "[!i:$nang luong$]" in markup or "[!i:$nang luong $]" in markup
    assert "*D." in markup
    assert manifest["counts"]["mathml"] == 1
    xml = (out / "sidecar" / "mathml_1.xml").read_text(encoding="utf-8")
    assert "E=mc^2" in xml
    data = json.loads((out / "manifest.json").read_text(encoding="utf-8"))
    assert data["assets"][0]["id"] == "mathml_1"
    assert data["assets"][0]["document_order"] == 1


def test_strip_unlimited_ocr_det():
    from vision import strip_unlimited_ocr_det

    raw = (
        "<|det|>text [0,0,10,10]<|/det|>Câu 1. Hoi\n"
        "A. mot\n"
        "<|det|>image [0,0,1,1]<|/det|>\n"
        "<|det|>text<|/det|>B. hai"
    )
    out = strip_unlimited_ocr_det(raw)
    assert "Câu 1. Hoi" in out
    assert "B. hai" in out


def test_markdown_to_azota(tmp_path: Path):
    from markdown_to_azota import markdown_to_azota

    png = tmp_path / "image1.png"
    png.write_bytes(b"\x89PNG\r\n\x1a\n")
    md = """
**Câu 1:** Cho $\\log_{a} b = 1$.

| Doanh thu | Số ngày |
|-----------|---------|
| 10 | 2 |
| 20 | 3 |

![fig](image1.png){width="2in"}

A. 1.
*B. 2.
C. 3.
D. 4.
""".strip()
    sidecar = tmp_path / "sidecar"
    text, assets = markdown_to_azota(md, sidecar_dir=sidecar, media_root=tmp_path)
    assert "[!b:$Câu 1:$]" in text
    assert r"$\log_{a} b = 1$" in text
    assert "[* Doanh thu | Số ngày *]" in text
    assert "[img:$img_1$]" in text
    assert "*B. 2." in text
    assert assets[0]["id"] == "img_1"
    assert (sidecar / "img_1.png").exists()


def test_colab_opt_helpers(tmp_path: Path):
    from colab_opt import inject_latex_into_markup, vision_jobs_from_manifest, detect_profile

    name, profile = detect_profile()
    assert name in {"cpu", "t4", "a100", "l4"}
    assert "unimernet" in profile
    markup = "He [!m:$mathtype_1$] va [!m:$mathml_1$]"
    out = inject_latex_into_markup(markup, {"mathtype_1": r"\alpha + \beta"})
    assert r"$\alpha + \beta$" in out
    assert "[!m:$mathml_1$]" in out
    skipped = inject_latex_into_markup(markup, {"mathtype_1": r"^ { 2 } H"})
    assert "[!m:$mathtype_1$]" in skipped
    img = tmp_path / "sidecar"
    img.mkdir()
    f = img / "mathtype_1.wmf"
    f.write_bytes(b"x")
    jobs = vision_jobs_from_manifest(
        {"assets": [{"id": "mathtype_1", "kind": "mathtype", "sidecar": "sidecar/mathtype_1.wmf"}]},
        tmp_path,
    )
    assert jobs[0][0] == "mathtype_1"


def test_patch_targets_pytorch_utils():
    src = (Path(__file__).resolve().parent.parent / "compat_transformers.py").read_text(encoding="utf-8")
    assert "import transformers.pytorch_utils as pytorch_utils" in src
    assert 'setattr(mod, "find_pruneable_heads_and_indices", heads)' in src
    from compat_transformers import rewrite_qformer_imports

    src = """
from transformers.modeling_utils import (
    PreTrainedModel,
    apply_chunking_to_forward,
    find_pruneable_heads_and_indices,
    prune_linear_layer,
)
"""
    out = rewrite_qformer_imports(src)
    assert "from transformers.pytorch_utils import apply_chunking_to_forward" in out
    assert "from transformers.modeling_utils import PreTrainedModel" in out
    assert rewrite_qformer_imports(out) == out


def test_install_colab_never_pins_tokenizers():
    src = (Path(__file__).resolve().parent.parent / "install_colab.py").read_text(encoding="utf-8")
    assert '[py, "-m", "pip", "install", "-q", "unimernet", "--no-deps"]' in src
    install_fn = src.split("def install_unimernet_colab")[1].split("def allow_wmf")[0]
    assert "tokenizers>=" not in install_fn
    assert "unimernet[full]" not in install_fn
    from install_colab import FORBIDDEN_PIP, REQUIRED_EXTRAS

    assert "unimernet[full]" in FORBIDDEN_PIP
    assert "transformers==4.42.4" in FORBIDDEN_PIP
    assert "evaluate" in REQUIRED_EXTRAS
    assert "tokenizers" not in REQUIRED_EXTRAS


def test_rewrite_decoder_onnx_imports():
    from compat_transformers import rewrite_decoder_onnx_imports

    src = (
        "from transformers import PreTrainedTokenizer\n"
        "from transformers.onnx import OnnxConfig, OnnxConfigWithPast, OnnxSeq2SeqConfigWithPast\n"
        "from transformers.onnx.utils import compute_effective_axis_dimension\n"
        "from transformers.utils import TensorType\n"
    )
    out = rewrite_decoder_onnx_imports(src)
    assert "azota_onnx_stub" in out
    assert "from transformers.onnx import" in out
    assert rewrite_decoder_onnx_imports(out) == out
    src = (Path(__file__).resolve().parent.parent / "compat_transformers.py").read_text(encoding="utf-8")
    assert "def install_transformers_onnx_stub" in src
    assert "transformers.onnx" in src
    assert "OnnxSeq2SeqConfigWithPast" in src
    assert "install_transformers_onnx_stub()" in src


def test_force_eager_attention_is_wired() -> None:
    src = (Path(__file__).resolve().parent.parent / "compat_transformers.py").read_text(encoding="utf-8")
    assert "def force_eager_attention" in src
    assert "force_eager_attention()" in src
    assert 'requested_attention = "eager"' in src
    vis = (Path(__file__).resolve().parent.parent / "vision.py").read_text(encoding="utf-8")
    assert "set_attn_implementation" in vis


def test_get_head_mask_none_layers() -> None:
    from compat_transformers import get_head_mask

    class Dummy:
        dtype = "float32"

    assert get_head_mask(Dummy(), None, 4) == [None, None, None, None]


def test_rewrite_encoder_decoder_kv_cache() -> None:
    from compat_transformers import rewrite_encoder_decoder_kv_cache, to_legacy_kv_cache

    src = "        past_key_values_length = past_key_values[0][0].shape[2] if past_key_values is not None else 0\n"
    out = rewrite_encoder_decoder_kv_cache(src)
    assert "azota_kv_cache" in out
    assert rewrite_encoder_decoder_kv_cache(out) == out
    assert to_legacy_kv_cache(None) is None
    assert to_legacy_kv_cache(()) is None
    assert to_legacy_kv_cache(((1,),)) == ((1,),)


def test_step_timer():
    from eval_timer import StepTimer

    t = StepTimer()
    with t.step("Bước 1", "OOXML"):
        pass
    s = t.summary()
    assert "Bước 1:" in s
    assert "TỔNG :" in s


def test_sample_exam_counts(tmp_path: Path):
    sample = Path(__file__).resolve().parents[1] / "samples" / "de-vat-li-lan-3.docx"
    if not sample.exists():
        pytest.skip("sample docx missing")
    manifest = convert_docx(sample, tmp_path)
    assert manifest["counts"]["mathml"] == 69
    assert manifest["counts"]["mathtype"] == 16
    assert manifest["counts"]["img"] == 8
    text = (tmp_path / "markup.txt").read_text(encoding="utf-8")
    assert text.count("[!m:$mathml_") == 69
    assert text.count("[!m:$mathtype_") == 16
    assert text.count("[img:$img_") == 8
    assert "*D. ngưng tụ." in text
    assert "→ Đáp án: 69,6" in text
    assert "[* Số lần bơm" in text
