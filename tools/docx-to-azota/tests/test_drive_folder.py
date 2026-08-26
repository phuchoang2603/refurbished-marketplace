from __future__ import annotations

from pathlib import Path

from drive_folder import (
    copy_toolkit,
    is_markdown_azota_name,
    resolve_markdown_azota,
)


def test_folder_name_variants() -> None:
    assert is_markdown_azota_name("markdown azota")
    assert is_markdown_azota_name("markdown-azota")
    assert is_markdown_azota_name("markdown_azota")
    assert is_markdown_azota_name("  Markdown Azota  ")
    assert not is_markdown_azota_name("docx-to-azota")


def test_resolve_prefers_canonical_name(tmp_path: Path) -> None:
    (tmp_path / "markdown-azota").mkdir()
    canonical = tmp_path / "markdown azota"
    canonical.mkdir()
    found = resolve_markdown_azota([tmp_path])
    assert found == canonical


def test_resolve_nested_and_create(tmp_path: Path) -> None:
    nested = tmp_path / "Colab" / "markdown azota"
    nested.mkdir(parents=True)
    assert resolve_markdown_azota([tmp_path]) == nested

    empty = tmp_path / "empty"
    empty.mkdir()
    created = resolve_markdown_azota([empty], create_under=empty)
    assert created == empty / "markdown azota"
    assert created.is_dir()


def test_copy_toolkit_keeps_outputs(tmp_path: Path) -> None:
    src = tmp_path / "src"
    dest = tmp_path / "markdown azota"
    src.mkdir()
    (src / "convert.py").write_text("ok\n", encoding="utf-8")
    (src / "azota_out").mkdir()
    (src / "azota_out" / "drop_me.txt").write_text("no", encoding="utf-8")
    dest.mkdir()
    (dest / "azota_out").mkdir()
    keep = dest / "azota_out" / "markup.txt"
    keep.write_text("keep\n", encoding="utf-8")
    copy_toolkit(src, dest)
    assert (dest / "convert.py").read_text(encoding="utf-8") == "ok\n"
    assert keep.read_text(encoding="utf-8") == "keep\n"
    assert not (dest / "azota_out" / "drop_me.txt").exists()
