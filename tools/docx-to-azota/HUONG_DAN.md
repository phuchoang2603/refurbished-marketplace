# Quy trình đúng: có cần Colab không?

**Không bắt buộc.** Đề `.docx` (ĐỀ VẬT LÍ LẦN 3) chạy **máy bạn, CPU, không GPU** đã ra đủ 3 file Azota.

Colab chỉ cần khi bạn muốn thêm:

- UniMERNet: ảnh MathType → `$latex$`
- Unlimited-OCR: chữ trong hình vẽ / file PDF scan

```
Đề .docx ──► convert.py (CPU) ──► markup.txt + sidecar/ + manifest.json
                                      │
                                      │  TÙY CHỌN, GPU Colab T4
                                      ▼
                              UniMERNet (công thức WMF)
                              Unlimited-OCR (chỉ hình, nếu còn VRAM)
```

---

## Phần A — làm xong đề (không Colab)

Trên máy có Python 3.10+:

```bash
cd tools/docx-to-azota
python3 convert.py "ĐỀ VẬT LÍ LẦN 3_VER 2 (2).docx" -o azota_out
```

Hoặc dùng file mẫu trong repo:

```bash
python3 convert.py samples/de-vat-li-lan-3.docx -o azota_out
```

Mở thư mục `azota_out/`:

| File | Dùng để |
| --- | --- |
| `markup.txt` | dán/import Azota |
| `sidecar/` | `mathml_*.xml`, `mathtype_*.wmf`, `img_*.png` |
| `manifest.json` | map id → file nguồn |

**Dừng ở đây** nếu Azota nhận `[!m:$mathml_1$]` / `[!m:$mathtype_1$]` / `[img:$img_1$]`. Không cần GPU.

---

## Phần B — Colab (chỉ khi cần LaTeX / OCR hình)

### B.0 Chuẩn bị (click, chưa code)

1. Vào [Google Colab](https://colab.research.google.com/).
2. File → New notebook.
3. Runtime → Change runtime type → **T4 GPU** → Save.
4. Mỗi ô dưới đây: dán vào **một cell** → `Shift+Enter`. Đợi xong mới sang ô tiếp. **Không** Run all.

---

### Ô 1 — kiểm tra GPU

```python
!nvidia-smi -L
```

Phải thấy `Tesla T4` (hoặc GPU khác). Nếu in `CPU only` thì quay lại B.0.

---

### Ô 2 — lấy converter

```python
REPO = "https://github.com/phuchoang2603/refurbished-marketplace.git"
BRANCH = "cursor/docx-to-azota-pipeline-4d56"
!git clone -b {BRANCH} --depth 1 {REPO} /content/repo
```

---

### Ô 3 — đưa code vào PYTHONPATH

```python
import sys, shutil
from pathlib import Path
shutil.copytree("/content/repo/tools/docx-to-azota", "/content/docx-to-azota", dirs_exist_ok=True)
sys.path.insert(0, "/content/docx-to-azota")
print("OK", Path("/content/docx-to-azota/convert.py").exists())
```

Phải in `OK True`.

---

### Ô 4 — import

```python
from convert import convert_docx, apply_unimernet_latex
from eval_timer import StepTimer
from colab_opt import detect_profile, prepare_unimernet_checkpoint, free_cuda
from colab_opt import vision_jobs_from_manifest, inject_latex_into_markup
from vision import rasterize_formula_image, load_unimernet, unimernet_batch
print("import OK")
```

---

### Ô 5 — nhận GPU profile

```python
NAME, PROFILE = detect_profile()
timer = StepTimer()
OUT = "/content/azota_out"
print(NAME, PROFILE)
```

T4 sẽ in `t4` và `unimernet: tiny`.

---

### Ô 6 — cài thư viện CPU + ImageMagick (WMF → PNG)

```python
!pip -q install pillow pymupdf huggingface_hub
!apt-get -qq install -y imagemagick libmagickwand-dev
!pip -q install Wand
```

---

### Ô 7 — cài UniMERNet (ô này lâu ~1–3 phút)

```python
!pip -q install -U "unimernet[full]"
```

---

### Ô 8 — upload đề `.docx`

```python
from google.colab import files
uploaded = files.upload()
DOCX = "/content/" + next(iter(uploaded))
print(DOCX)
```

Cửa sổ chọn file hiện ra → chọn `ĐỀ VẬT LÍ LẦN 3_VER 2 (2).docx`.

---

### Ô 9 — Bước 1: extract Azota (CPU, ~0.2s) — **bắt buộc**

```python
with timer.step("Bước 1", "OOXML"):
    man = convert_docx(DOCX, OUT)
print(man["counts"])
```

Kỳ vọng khoảng: `mathml 69`, `mathtype 16`, `img 8`.

**Đã có ` /content/azota_out/markup.txt `.** Có thể tải về và dừng, không cần Ô 10–16.

---

### Ô 10 — xem markup

```python
print("\n".join(open(OUT + "/markup.txt", encoding="utf-8").read().splitlines()[:40]))
```

---

### Ô 11 — Bước 2: WMF → PNG (chỉ công thức MathType)

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

---

### Ô 12 — tải checkpoint UniMERNet-tiny (ô này lâu)

```python
with timer.step("Bước 3-load", "download tiny"):
    cfg = prepare_unimernet_checkpoint("tiny", "/content/models")
    model, vis, device = load_unimernet(cfg_path=cfg, fp16=True)
print("device =", device)
```

---

### Ô 13 — Bước 3: nhận dạng LaTeX (batch)

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

In ra dạng:

```
⏱ THỜI GIAN:
Bước 1: 0.2s (OOXML)
Bước 2: …s (raster WMF)
Bước 3: …s (UniMERNet batch)
Bước 4: 0.1s (inject LaTeX)
TỔNG : …
```

---

### Ô 15 — giải phóng GPU (bắt buộc trước khi OCR)

```python
free_cuda(model, vis)
model = vis = None
print("đã unload UniMERNet")
```

---

### Ô 16 — Unlimited-OCR hình vẽ (TÙY CHỌN, dễ OOM trên T4)

Mặc định **bỏ qua**. Chỉ chạy nếu Ô 13 xong mà `nvidia-smi` còn >8GB trống.

```python
PROFILE = dict(PROFILE) if "PROFILE" in dir() else {}
# Bật tay:
RUN_OCR_HINH = False
```

```python
if RUN_OCR_HINH:
    !pip -q install -U transformers==4.57.1 einops addict easydict
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

### Ô 17 — tải kết quả về máy

```python
from google.colab import files
!cd /content && zip -qr azota_out.zip azota_out
files.download("/content/azota_out.zip")
```

Giải nén: `markup.txt` + `sidecar/` + `manifest.json`.

---

## Không làm

- Không OCR cả 5 trang PDF khi đã có `.docx` (chậm ~57s/trang, dễ OOM).
- Không `Run all` khi đang đo thời gian.
- Không load UniMERNet và Unlimited-OCR cùng lúc trên T4.

## Notebook sẵn

Cùng nội dung trên, đã tách cell sẵn:

- `colab_optimized.ipynb` — chạy thật trên T4  
- `colab_docx_to_azota.ipynb` — đo từng bước / PDF
