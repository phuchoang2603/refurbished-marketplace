# Lộ trình đúng

Ba lớp, làm lần lượt. **Không** cài theo README UniMERNet trên Colab (`transformers==4.42.4` không chạy Python 3.13).

```
1. BẮT BUỘC  .docx → convert.py (CPU) → markup.txt + sidecar/ + manifest.json
2. TÙY CHỌN  MathType WMF → UniMERNet → $latex$     (Colab GPU)
3. TÙY CHỌN  hình vẽ → Unlimited-OCR                 (T4 dễ OOM, mặc định tắt)
```

| Bước | Khi nào | Kết quả Azota |
| --- | --- | --- |
| 1 | luôn | `[!m:$mathml_N$]`, `[!m:$mathtype_N$]`, `[img:$img_N$]`, `*D.` |
| 2 | chỉ khi cần `$latex$` thay MathType | `[!m:$mathtype_N$]` → `$latex$` |
| 3 | chỉ khi còn VRAM sau bước 2 | OCR chữ trong hình |

**Dừng ở bước 1** nếu Azota nhận placeholder. Đề mẫu: 69 mathml + 16 mathtype + 8 img, **không GPU**.

---

## Phần A — máy bạn, không Colab

```bash
cd tools/docx-to-azota
python3 convert.py "ĐỀ VẬT LÍ LẦN 3_VER 2 (2).docx" -o azota_out
```

Hoặc file mẫu: `python3 convert.py samples/de-vat-li-lan-3.docx -o azota_out`

| File | Dùng để |
| --- | --- |
| `markup.txt` | dán/import Azota |
| `sidecar/` | `mathml_*.xml`, `mathtype_*.wmf`, `img_*.png` |
| `manifest.json` | map id → file nguồn |

---

## Phần B — Colab, làm lại từ đầu

### B.0 Xóa runtime cũ (bắt buộc nếu đã fail pip)

1. Runtime → **Disconnect and delete runtime**.
2. File → **New notebook** (notebook trắng).
3. Runtime → Change runtime type → **T4 GPU** → Save.
4. Dán **từng ô** → `Shift+Enter`. Đợi xong mới sang ô tiếp. **Không** Run all.

**Cấm** (sẽ hỏng runtime):

```text
pip install unimernet[full]
pip install tokenizers
pip install transformers==4.42.4
```

---

### Ô 1 — GPU

```python
!nvidia-smi -L
```

Phải thấy `Tesla T4` (hoặc GPU khác). Nếu `CPU only` → quay B.0.

---

### Ô 2 — clone code (branch có patch transformers 5)

```python
import shutil
from pathlib import Path
for p in ("/content/repo", "/content/docx-to-azota", "/content/refurbished-marketplace"):
    shutil.rmtree(p, ignore_errors=True)
REPO = "https://github.com/phuchoang2603/refurbished-marketplace.git"
BRANCH = "cursor/docx-to-azota-pipeline-4d56"
!git clone -b {BRANCH} --depth 1 {REPO} /content/repo
print("cloned")
```

---

### Ô 3 — PYTHONPATH

```python
import sys, shutil
from pathlib import Path
shutil.copytree("/content/repo/tools/docx-to-azota", "/content/docx-to-azota", dirs_exist_ok=True)
sys.path.insert(0, "/content/docx-to-azota")
print("OK", Path("/content/docx-to-azota/convert.py").exists())
```

Phải in `OK True`.

---

### Ô 4 — import converter

```python
from convert import convert_docx, apply_unimernet_latex
from eval_timer import StepTimer
from colab_opt import detect_profile, prepare_unimernet_checkpoint, free_cuda
from colab_opt import vision_jobs_from_manifest, inject_latex_into_markup
from vision import rasterize_formula_image, load_unimernet, unimernet_batch
print("import OK")
```

---

### Ô 5 — profile GPU

```python
NAME, PROFILE = detect_profile()
timer = StepTimer()
OUT = "/content/azota_out"
print(NAME, PROFILE)
```

T4 in `t4` và `unimernet: tiny`.

---

### Ô 6 — ImageMagick (WMF → PNG)

```python
!pip -q install pillow pymupdf huggingface_hub
!apt-get -qq install -y imagemagick libmagickwand-dev
!pip -q install Wand
```

Đợi đến `Setting up imagemagick`. **Không** cài UniMERNet trong ô này.

---

### Ô 7 — UniMERNet (`--no-deps` + vá file decoder, không dùng `[full]`)

```python
from install_colab import allow_wmf_in_imagemagick, install_unimernet_colab
allow_wmf_in_imagemagick()
install_unimernet_colab()
```

Kỳ vọng `unimernet OK /usr/local/...`. Hàm này vá `configuration_unimernet_decoder.py` (bỏ `transformers.onnx`) rồi mới import.

Nếu clone trên Colab cũ hơn commit vá ONNX: **đừng** gọi lại `install_unimernet_colab()`. Dán ô trong chat (patch file decoder + `import unimernet`).

---

### Ô 8 — upload `.docx`

```python
from google.colab import files
uploaded = files.upload()
DOCX = "/content/" + next(iter(uploaded))
print(DOCX)
```

Chọn `ĐỀ VẬT LÍ LẦN 3_VER 2 (2).docx`.

---

### Ô 9 — Bước 1: extract Azota (CPU) — bắt buộc

```python
with timer.step("Bước 1", "OOXML"):
    man = convert_docx(DOCX, OUT)
print(man["counts"])
```

Kỳ vọng khoảng: `mathml 69`, `mathtype 16`, `img 8`.

Đã có `/content/azota_out/markup.txt`. **Có thể nhảy Ô 17 tải về và dừng.** Ô 10–16 chỉ khi cần `$latex$`.

---

### Ô 10 — xem markup

```python
print("\n".join(open(OUT + "/markup.txt", encoding="utf-8").read().splitlines()[:40]))
```

---

### Ô 11 — Bước 2: WMF → PNG (chỉ MathType)

```python
from pathlib import Path
png_dir = Path(OUT) / "sidecar_png"
png_dir.mkdir(exist_ok=True)
jobs = []
with timer.step("Bước 2", "raster WMF"):
    for aid, src in vision_jobs_from_manifest(man, OUT, kinds=("mathtype",)):
        dest = png_dir / f"{aid}.png"
        got = rasterize_formula_image(src, dest, dpi=200)
        if got:
            jobs.append((aid, got))
print(len(jobs), "ảnh công thức")
```

Kỳ vọng `16 ảnh công thức`.

---

### Ô 12 — tải UniMERNet-tiny (ô lâu)

```python
with timer.step("Bước 3-load", "download tiny"):
    cfg = prepare_unimernet_checkpoint("tiny", "/content/models")
    model, vis, device = load_unimernet(cfg_path=cfg, fp16=True)
print("device =", device)
```

---

### Ô 13 — Bước 3: nhận dạng LaTeX

```python
with timer.step("Bước 3", "UniMERNet batch"):
    preds = unimernet_batch(model, vis, device, jobs, batch_size=8)
apply_unimernet_latex(man, preds, Path(OUT))
for k, v in list(preds.items())[:5]:
    print(k, "→", v[:100])
```

---

### Ô 14 — Bước 4: gắn `$latex$` vào markup

```python
p = Path(OUT) / "markup.txt"
text = p.read_text(encoding="utf-8")
with timer.step("Bước 4", "inject LaTeX"):
    text = inject_latex_into_markup(text, preds)
    p.write_text(text, encoding="utf-8")
timer.print_summary()
```

---

### Ô 15 — unload GPU

```python
free_cuda(model, vis)
model = vis = None
print("đã unload UniMERNet")
```

---

### Ô 16 — Unlimited-OCR hình (mặc định tắt)

```python
RUN_OCR_HINH = False
if RUN_OCR_HINH:
    !pip -q install -U einops addict easydict
    from vision import load_unlimited_ocr, unlimited_ocr_one, strip_unlimited_ocr_det
    figs = vision_jobs_from_manifest(man, OUT, kinds=("img",))
    ocr_m, ocr_t = load_unlimited_ocr()
    for aid, src in figs:
        png = png_dir / f"{aid}.png"
        got = rasterize_formula_image(src, png, dpi=150) or src
        raw = unlimited_ocr_one(ocr_m, ocr_t, str(got), f"{OUT}/ocr_raw/{aid}", gundam=False, max_length=4096)
        print(aid, strip_unlimited_ocr_det(raw)[:200])
    free_cuda(ocr_m, ocr_t)
else:
    print("SKIP OCR hình — giữ [img:$img_N$]")
```

---

### Ô 17 — tải zip về máy

```python
from google.colab import files
!cd /content && zip -qr azota_out.zip azota_out
files.download("/content/azota_out.zip")
```

Giải nén: `markup.txt` + `sidecar/` + `manifest.json`.

---

## Không làm

- Không `pip install unimernet[full]`, `tokenizers`, `transformers==4.42.4`.
- Không OCR cả trang PDF khi đã có `.docx`.
- Không `Run all`.
- Không load UniMERNet và Unlimited-OCR cùng lúc trên T4.

## Notebook sẵn

- `colab_optimized.ipynb` — T4
- `colab_docx_to_azota.ipynb` — đo từng bước / PDF
