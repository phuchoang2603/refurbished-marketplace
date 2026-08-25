"""Make UniMERNet importable on Colab's transformers 5.x.

UniMERNet still imports helpers that moved or were removed:

* Qformer.py: ``from transformers.modeling_utils import apply_chunking_to_forward, ...``
* modeling_unimernet_encoder.py: ``from transformers.pytorch_utils import find_pruneable_heads_and_indices``

On transformers >= 4.56 chunking/prune live in ``pytorch_utils``.
On transformers 5.x ``find_pruneable_heads_and_indices`` was removed entirely.
We re-attach a copy onto both modules before ``import unimernet``.
"""

from __future__ import annotations

import re
import sys
from typing import Any

QFORMER_IMPORT_RE = re.compile(
    r"from transformers\.modeling_utils import \("
    r"[^)]*apply_chunking_to_forward[^)]*\)",
    re.DOTALL,
)

QFORMER_IMPORT_REPLACEMENT = """from transformers.modeling_utils import PreTrainedModel
try:
    from transformers.pytorch_utils import apply_chunking_to_forward, prune_linear_layer
except ImportError:  # transformers < 4.56
    from transformers.modeling_utils import apply_chunking_to_forward, prune_linear_layer
try:
    from transformers.pytorch_utils import find_pruneable_heads_and_indices
except ImportError:
    try:
        from transformers.modeling_utils import find_pruneable_heads_and_indices
    except ImportError:
        def find_pruneable_heads_and_indices(heads, n_heads, head_size, already_pruned_heads):
            import torch
            mask = torch.ones(n_heads, head_size)
            heads = set(heads) - already_pruned_heads
            for head in heads:
                head = head - sum(1 if h < head else 0 for h in already_pruned_heads)
                mask[head] = 0
            mask = mask.view(-1).contiguous().eq(1)
            index = torch.arange(len(mask))[mask].long()
            return heads, index"""


def find_pruneable_heads_and_indices(
    heads: list[int],
    n_heads: int,
    head_size: int,
    already_pruned_heads: set[int],
) -> tuple[set[int], Any]:
    """Copy of transformers 4.42 ``pytorch_utils.find_pruneable_heads_and_indices``."""
    import torch

    mask = torch.ones(n_heads, head_size)
    heads_set = set(heads) - already_pruned_heads
    for head in heads_set:
        head = head - sum(1 if h < head else 0 for h in already_pruned_heads)
        mask[head] = 0
    mask = mask.view(-1).contiguous().eq(1)
    index = torch.arange(len(mask))[mask].long()
    return heads_set, index


def rewrite_qformer_imports(source: str) -> str:
    if "from transformers.pytorch_utils import apply_chunking_to_forward" in source:
        return source
    return QFORMER_IMPORT_RE.sub(QFORMER_IMPORT_REPLACEMENT, source, count=1)


def purge_unimernet_modules() -> None:
    for name in list(sys.modules):
        if name == "unimernet" or name.startswith("unimernet."):
            del sys.modules[name]


def _first_available(*getters):
    for getter in getters:
        try:
            value = getter()
        except Exception:
            value = None
        if value is not None:
            return value
    return None


def patch_transformers_for_unimernet() -> None:
    """Re-attach missing helpers on both modeling_utils and pytorch_utils."""
    import transformers.modeling_utils as modeling_utils
    import transformers.pytorch_utils as pytorch_utils
    import transformers.utils as tf_utils

    chunk = _first_available(
        lambda: getattr(pytorch_utils, "apply_chunking_to_forward", None),
        lambda: getattr(modeling_utils, "apply_chunking_to_forward", None),
    )
    prune = _first_available(
        lambda: getattr(pytorch_utils, "prune_linear_layer", None),
        lambda: getattr(modeling_utils, "prune_linear_layer", None),
    )
    heads = _first_available(
        lambda: getattr(pytorch_utils, "find_pruneable_heads_and_indices", None),
        lambda: getattr(modeling_utils, "find_pruneable_heads_and_indices", None),
        lambda: find_pruneable_heads_and_indices,
    )

    for mod in (modeling_utils, pytorch_utils):
        if chunk is not None:
            setattr(mod, "apply_chunking_to_forward", chunk)
        if prune is not None:
            setattr(mod, "prune_linear_layer", prune)
        if heads is not None:
            setattr(mod, "find_pruneable_heads_and_indices", heads)

    if not hasattr(tf_utils, "torch_int"):
        try:
            from transformers.utils.generic import torch_int

            tf_utils.torch_int = torch_int
        except Exception:
            pass


def rewrite_installed_qformer() -> str | None:
    """Rewrite site-packages UniMERNet Qformer.py so a later Restart still imports."""
    import importlib.util
    from pathlib import Path

    spec = importlib.util.find_spec("unimernet")
    if spec is None or not spec.origin:
        return None
    qformer = Path(spec.origin).resolve().parent / "models" / "blip2_models" / "Qformer.py"
    if not qformer.exists():
        return None
    original = qformer.read_text(encoding="utf-8")
    updated = rewrite_qformer_imports(original)
    if updated != original:
        qformer.write_text(updated, encoding="utf-8")
        print("patched", qformer)
    return str(qformer)


def import_unimernet():
    """Patch transformers, drop a failed half-import, then import UniMERNet."""
    patch_transformers_for_unimernet()
    rewrite_installed_qformer()
    purge_unimernet_modules()
    import unimernet

    return unimernet
