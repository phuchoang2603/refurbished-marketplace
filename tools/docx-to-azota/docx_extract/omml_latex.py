"""OMML (Word equation) → LaTeX. CPU only — no UniMERNet."""

from __future__ import annotations

import re
from xml.etree import ElementTree as ET

from .ns import local

SKIP = {"rPr", "ctrlPr", "fPr", "dPr", "sSubPr", "sSupPr", "sSubSupPr", "sPrePr", "funcPr", "barPr", "accPr", "mPr", "mcPr", "mcs", "mc", "sty", "nor", "pos", "chr", "begChr", "endChr"}

FUNCS = {
    "sin", "cos", "tan", "cot", "sec", "csc",
    "log", "ln", "exp", "lim", "max", "min", "det", "gcd", "arg",
    "sinh", "cosh", "tanh", "arcsin", "arccos", "arctan",
}

CHAR = {
    "π": r"\pi ",
    "α": r"\alpha ",
    "β": r"\beta ",
    "γ": r"\gamma ",
    "δ": r"\delta ",
    "Δ": r"\Delta ",
    "θ": r"\theta ",
    "λ": r"\lambda ",
    "μ": r"\mu ",
    "ρ": r"\rho ",
    "σ": r"\sigma ",
    "τ": r"\tau ",
    "φ": r"\phi ",
    "ω": r"\omega ",
    "Ω": r"\Omega ",
    "ε": r"\varepsilon ",
    "η": r"\eta ",
    "ξ": r"\xi ",
    "ψ": r"\psi ",
    "∑": r"\sum ",
    "∫": r"\int ",
    "∂": r"\partial ",
    "∇": r"\nabla ",
    "∞": r"\infty ",
    "⋅": r"\cdot ",
    "·": r"\cdot ",
    "×": r"\times ",
    "÷": r"\div ",
    "±": r"\pm ",
    "∓": r"\mp ",
    "≈": r"\approx ",
    "≠": r"\neq ",
    "≤": r"\le ",
    "≥": r"\ge ",
    "→": r"\rightarrow ",
    "←": r"\leftarrow ",
    "⇒": r"\Rightarrow ",
    "⇔": r"\Leftrightarrow ",
    "↔": r"\leftrightarrow ",
    "∈": r"\in ",
    "∘": r"\circ ",
    "°": r"\circ ",
    "−": "-",
    "–": "-",
    "—": "-",
    "⋅": r"\cdot ",
    "…": r"\ldots ",
    "∝": r"\propto ",
    "∼": r"\sim ",
    "≃": r"\simeq ",
    "≡": r"\equiv ",
    "≪": r"\ll ",
    "≫": r"\gg ",
    "⊥": r"\perp ",
    "∠": r"\angle ",
    "△": r"\triangle ",
    "⃗": r"\vec{}",
    "\u00a0": " ",
    "\u2009": " ",
    "\u200a": " ",
    "\u202f": " ",
    "\u2003": " ",
    "\u2212": "-",
}


def omml_to_latex(el: ET.Element | None) -> str:
    if el is None:
        return ""
    raw = _node(el)
    raw = re.sub(r"[ \t]+", " ", raw)
    return raw.strip(" $")


def inject_mathml_latex(markup: str, assets: list) -> str:
    """Replace ``[!m:$mathml_N$]`` with ``$latex$`` when conversion succeeded."""
    for asset in assets:
        latex = getattr(asset, "latex", None)
        kind = getattr(asset, "kind", None)
        placeholder = getattr(asset, "placeholder", None)
        if kind != "mathml" or not latex or not placeholder:
            continue
        body = latex.strip().strip("$")
        if not body:
            continue
        markup = markup.replace(placeholder, f"${body}$")
    return markup


def _val(el: ET.Element | None) -> str:
    if el is None:
        return ""
    return el.get("{http://schemas.openxmlformats.org/officeDocument/2006/math}val") or el.get("val") or ""


def _math_children(el: ET.Element) -> list[ET.Element]:
    out = []
    for c in el:
        name = local(c.tag)
        if name in SKIP or name.endswith("Pr"):
            continue
        out.append(c)
    return out


def _join(el: ET.Element) -> str:
    return "".join(_node(c) for c in _math_children(el))


def _escape_delim(ch: str) -> str:
    if not ch:
        return "."
    if ch in "{}":
        return "\\" + ch
    if ch == "\\":
        return r"\backslash"
    return ch


def _text_to_latex(text: str) -> str:
    text = text.replace("\u00a0", " ").replace("\ufeff", "")
    out = []
    for ch in text:
        out.append(CHAR.get(ch, ch))
    s = "".join(out)
    s = s.replace("%", r"\%").replace("&", r"\&")
    return s


def _node(el: ET.Element) -> str:
    name = local(el.tag)
    if name in SKIP or name.endswith("Pr"):
        return ""
    if name in {"oMath", "oMathPara", "e", "box", "num", "den", "deg", "sub", "sup", "lim", "fName"}:
        return _join(el)
    if name == "r":
        parts = []
        for c in el:
            if local(c.tag) == "t":
                parts.append(_text_to_latex(c.text or ""))
            elif local(c.tag) not in SKIP and not local(c.tag).endswith("Pr"):
                parts.append(_node(c))
        return "".join(parts)
    if name == "t":
        return _text_to_latex(el.text or "")
    if name == "f":
        num = den = ""
        for c in el:
            n = local(c.tag)
            if n == "num":
                num = _join(c)
            elif n == "den":
                den = _join(c)
        return rf"\frac{{{num}}}{{{den}}}"
    if name == "sSub":
        return _script(el, sub=True, sup=False)
    if name == "sSup":
        return _script(el, sub=False, sup=True)
    if name == "sSubSup":
        return _script(el, sub=True, sup=True)
    if name == "sPre":
        base = sub = sup = ""
        for c in el:
            n = local(c.tag)
            if n == "e":
                base = _join(c)
            elif n == "sub":
                sub = _join(c)
            elif n == "sup":
                sup = _join(c)
        left = ""
        if sub:
            left += f"_{{{sub}}}"
        if sup:
            left += f"^{{{sup}}}"
        return f"{{}}{left}{base}" if (sub or sup) else base
    if name == "d":
        return _delim(el)
    if name == "func":
        fname = arg = ""
        for c in el:
            n = local(c.tag)
            if n == "fName":
                fname = _join(c).strip()
            elif n == "e":
                arg = _join(c)
        key = re.sub(r"[^A-Za-z]", "", fname).lower()
        if key in FUNCS:
            return rf"\{key} {arg}"
        return rf"\operatorname{{{fname}}}{arg}"
    if name == "bar":
        inner = ""
        pos = "top"
        for c in el:
            if local(c.tag) == "e":
                inner = _join(c)
            elif local(c.tag) == "barPr":
                p = c.find("{http://schemas.openxmlformats.org/officeDocument/2006/math}pos")
                if p is not None and _val(p) == "bot":
                    pos = "bot"
        if pos == "bot":
            return rf"\underline{{{inner}}}"
        return rf"\overline{{{inner}}}"
    if name == "acc":
        inner = ""
        mark = "^"
        for c in el:
            if local(c.tag) == "e":
                inner = _join(c)
            elif local(c.tag) == "accPr":
                ch = c.find("{http://schemas.openxmlformats.org/officeDocument/2006/math}chr")
                if ch is not None:
                    mark = _val(ch) or mark
        acc_map = {
            "^": r"\hat{{{}}}",
            "\u0302": r"\hat{{{}}}",
            "~": r"\tilde{{{}}}",
            "\u0303": r"\tilde{{{}}}",
            "\u2192": r"\vec{{{}}}",
            "\u20d7": r"\vec{{{}}}",
            ".": r"\dot{{{}}}",
            "\u0307": r"\dot{{{}}}",
            "\u00a8": r"\ddot{{{}}}",
        }
        tmpl = acc_map.get(mark, r"\hat{{{}}}")
        return tmpl.format(inner)
    if name == "m":
        rows = []
        for c in el:
            if local(c.tag) != "mr":
                continue
            cells = [_join(e) for e in c if local(e.tag) == "e"]
            rows.append(" & ".join(cells))
        body = r" \\ ".join(rows)
        return rf"\begin{{matrix}}{body}\end{{matrix}}"
    if name == "mr":
        return " & ".join(_join(e) for e in el if local(e.tag) == "e")
    if name == "rad":
        deg = inner = ""
        hide_deg = False
        for c in el:
            n = local(c.tag)
            if n == "deg":
                deg = _join(c)
            elif n == "e":
                inner = _join(c)
            elif n == "radPr":
                hide = c.find("{http://schemas.openxmlformats.org/officeDocument/2006/math}degHide")
                if hide is not None and str(_val(hide) or "1") not in {"0", "false", "off"}:
                    hide_deg = True
        if deg and not hide_deg:
            return rf"\sqrt[{deg}]{{{inner}}}"
        return rf"\sqrt{{{inner}}}"
    if name == "eqArr":
        rows = [_join(c) for c in el if local(c.tag) == "e"]
        return r"\begin{aligned}" + r" \\ ".join(rows) + r"\end{aligned}"
    if name == "groupChr":
        return _join(el)
    if name == "phant":
        return _join(el)
    if name == "nary":
        return _nary(el)
    return _join(el)


def _script(el: ET.Element, *, sub: bool, sup: bool) -> str:
    base = s = p = ""
    for c in el:
        n = local(c.tag)
        if n == "e":
            base = _join(c)
        elif n == "sub":
            s = _join(c)
        elif n == "sup":
            p = _join(c)
    out = f"{{{base}}}" if base else "{}"
    if sub and s:
        out += f"_{{{s}}}"
    if sup and p:
        out += f"^{{{p}}}"
    return out


def _delim(el: ET.Element) -> str:
    left, right = "(", ")"
    inner_parts = []
    for c in el:
        n = local(c.tag)
        if n == "dPr":
            b = c.find("{http://schemas.openxmlformats.org/officeDocument/2006/math}begChr")
            e = c.find("{http://schemas.openxmlformats.org/officeDocument/2006/math}endChr")
            if b is not None:
                left = _val(b) or left
            if e is not None:
                right = _val(e)
        elif n == "e":
            inner_parts.append(_join(c))
    inner = "".join(inner_parts)
    l = _escape_delim(left) if left else "."
    r = _escape_delim(right) if right else "."
    return rf"\left{l} {inner}\right{r}"


def _nary(el: ET.Element) -> str:
    op = r"\sum"
    sub = sup = inner = ""
    for c in el:
        n = local(c.tag)
        if n == "naryPr":
            ch = c.find("{http://schemas.openxmlformats.org/officeDocument/2006/math}chr")
            raw = _val(ch) if ch is not None else "∑"
            op_map = {"∑": r"\sum", "∫": r"\int", "∏": r"\prod", "⋃": r"\bigcup", "⋂": r"\bigcap"}
            op = op_map.get(raw, CHAR.get(raw, raw).strip() or r"\sum")
        elif n == "sub":
            sub = _join(c)
        elif n == "sup":
            sup = _join(c)
        elif n == "e":
            inner = _join(c)
    out = op
    if sub:
        out += f"_{{{sub}}}"
    if sup:
        out += f"^{{{sup}}}"
    return out + inner
