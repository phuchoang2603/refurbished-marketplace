"""Persist the Azota toolkit on the user's Google Drive folder ``markdown azota``.

Colab ``/content`` is wiped every runtime. Drive is not. Clone into a staging
dir, then copy code into Drive without deleting ``azota_out`` / ``models``.
"""

from __future__ import annotations

import shutil
import subprocess
import sys
from pathlib import Path

FOLDER_NAMES = ("markdown azota", "markdown-azota", "markdown_azota")
CANONICAL_NAME = "markdown azota"
SKIP_COPY = {
    "azota_out",
    "azota_out.zip",
    "models",
    "uploads",
    "__pycache__",
    ".pytest_cache",
    ".git",
    ".ipynb_checkpoints",
}
DEFAULT_REPO = "https://github.com/phuchoang2603/refurbished-marketplace.git"
DEFAULT_BRANCH = "cursor/docx-to-azota-pipeline-4d56"
STAGING = Path("/content/_repo_azota")
DRIVE_ROOTS = (Path("/content/drive/MyDrive"), Path("/content/drive/My Drive"))
REQUIRED_FILES = ("convert.py", "install_colab.py", "compat_transformers.py")
LOCAL_FALLBACKS = (
    Path("/content/docx-to-azota"),
    Path("/content/_repo_azota/tools/docx-to-azota"),
    Path("/content/repo/tools/docx-to-azota"),
)


def _norm(name: str) -> str:
    return " ".join(name.strip().lower().replace("_", " ").replace("-", " ").split())


def is_markdown_azota_name(name: str) -> bool:
    return _norm(name) == _norm(CANONICAL_NAME)


def resolve_markdown_azota(
    search_roots: list[Path],
    *,
    create_under: Path | None = None,
) -> Path:
    """Find an existing folder (exact name preferred) or create ``markdown azota``."""
    matches: list[Path] = []
    for root in search_roots:
        if not root.is_dir():
            continue
        try:
            children = list(root.iterdir())
        except OSError:
            continue
        for child in children:
            if child.is_dir() and is_markdown_azota_name(child.name):
                matches.append(child)
            if not child.is_dir():
                continue
            try:
                grandchildren = list(child.iterdir())
            except OSError:
                continue
            for grand in grandchildren:
                if grand.is_dir() and is_markdown_azota_name(grand.name):
                    matches.append(grand)
    if matches:
        exact = [p for p in matches if p.name == CANONICAL_NAME]
        return exact[0] if exact else matches[0]
    if create_under is None:
        raise FileNotFoundError("không thấy folder markdown azota")
    dest = create_under / CANONICAL_NAME
    dest.mkdir(parents=True, exist_ok=True)
    return dest


def toolkit_is_complete(root: Path) -> bool:
    return root.is_dir() and all((root / name).is_file() for name in REQUIRED_FILES)


def iter_named_markdown_azota(search_roots: list[Path]) -> list[Path]:
    found: list[Path] = []
    try:
        found.append(resolve_markdown_azota(search_roots))
    except FileNotFoundError:
        pass
    return found


def find_complete_toolkit(search_roots: list[Path] | None = None) -> Path | None:
    """Prefer Drive ``markdown azota`` if it has install_colab.py, else Colab copies."""
    ordered: list[Path] = []
    if search_roots:
        ordered.extend(iter_named_markdown_azota(search_roots))
        for root in search_roots:
            if root.is_dir():
                ordered.append(root)
    ordered.extend(LOCAL_FALLBACKS)
    seen: set[Path] = set()
    for path in ordered:
        try:
            path = path.resolve()
        except OSError:
            continue
        if path in seen:
            continue
        seen.add(path)
        if toolkit_is_complete(path):
            return path
    return None


def copy_toolkit(src: Path, dest: Path) -> None:
    """Copy converter code into dest. Never delete dest or user outputs."""
    dest.mkdir(parents=True, exist_ok=True)
    for item in src.iterdir():
        if item.name in SKIP_COPY:
            continue
        target = dest / item.name
        if item.is_dir():
            shutil.copytree(item, target, dirs_exist_ok=True)
        else:
            shutil.copy2(item, target)


def clone_toolkit(
    repo: str = DEFAULT_REPO,
    branch: str = DEFAULT_BRANCH,
    staging: Path = STAGING,
) -> Path:
    if staging.exists():
        shutil.rmtree(staging)
    subprocess.check_call(
        [
            "git",
            "clone",
            "-b",
            branch,
            "--depth",
            "1",
            "--single-branch",
            "--filter=blob:none",
            "--sparse",
            repo,
            str(staging),
        ]
    )
    subprocess.check_call(
        ["git", "-C", str(staging), "sparse-checkout", "set", "tools/docx-to-azota"]
    )
    src = staging / "tools" / "docx-to-azota" / "convert.py"
    if not src.exists():
        raise SystemExit("clone chưa đủ file — chạy lại ô này, đừng bấm Stop")
    return src.parent


def mount_drive() -> Path:
    try:
        from google.colab import drive
    except ImportError as exc:
        raise SystemExit("Ô này chỉ chạy trên Google Colab (cần mount Drive).") from exc
    drive.mount("/content/drive", force_remount=False)
    for root in DRIVE_ROOTS:
        if root.is_dir():
            return root
    raise SystemExit("Mount Drive xong nhưng không thấy /content/drive/MyDrive")


def use_path(root: Path) -> Path:
    inserted = str(root)
    if inserted in sys.path:
        sys.path.remove(inserted)
    sys.path.insert(0, inserted)
    return root


def sync_to_markdown_azota(
    *,
    repo: str = DEFAULT_REPO,
    branch: str = DEFAULT_BRANCH,
    refresh: bool = True,
) -> Path:
    """Mount Drive, clone (unless Drive already has convert.py and refresh=False), copy.

    Returns the Drive folder path. Safe to re-run: outputs stay.
    """
    my_drive = mount_drive()
    root = resolve_markdown_azota(list(DRIVE_ROOTS), create_under=my_drive)
    have_code = (root / "convert.py").exists()
    if refresh or not have_code:
        try:
            src = clone_toolkit(repo=repo, branch=branch)
            copy_toolkit(src, root)
        except (subprocess.CalledProcessError, SystemExit):
            if not have_code:
                raise
            print("clone lỗi — dùng code đã có trong Drive", flush=True)
    use_path(root)
    (root / "uploads").mkdir(exist_ok=True)
    (root / "azota_out").mkdir(exist_ok=True)
    (root / "models").mkdir(exist_ok=True)
    print("ROOT", root, flush=True)
    print("OK", (root / "convert.py").exists(), flush=True)
    return root


def reconnect_toolkit(*, refresh_if_incomplete: bool = True) -> Path:
    """Re-attach sys.path after a Colab Restart session (Part B import).

    Does not require ``ROOT`` from an earlier cell.
    """
    my_drive: Path | None
    try:
        my_drive = mount_drive()
        search = list(DRIVE_ROOTS)
    except SystemExit:
        my_drive = None
        search = []
    found = find_complete_toolkit(search or None)
    if found is not None:
        return use_path(found)
    if not refresh_if_incomplete:
        raise SystemExit("không thấy install_colab.py — chạy lại ô A2")
    src = clone_toolkit()
    if my_drive is not None:
        dest = resolve_markdown_azota(list(DRIVE_ROOTS), create_under=my_drive)
        copy_toolkit(src, dest)
        if toolkit_is_complete(dest):
            (dest / "uploads").mkdir(exist_ok=True)
            (dest / "azota_out").mkdir(exist_ok=True)
            (dest / "models").mkdir(exist_ok=True)
            return use_path(dest)
    if toolkit_is_complete(src):
        return use_path(src)
    raise SystemExit("clone xong vẫn thiếu install_colab.py — chạy lại ô A2, đừng bấm Stop")


if __name__ == "__main__":
    sync_to_markdown_azota()
