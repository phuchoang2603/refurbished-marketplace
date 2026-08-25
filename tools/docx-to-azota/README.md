# DOCX → Azota markup (hybrid UniMERNet + Unlimited-OCR)

Chuyển đề thi Word (`.docx`) sang **cú pháp Azota phẳng theo dòng**, giữ nguyên công thức / ảnh / bảng dưới dạng placeholder + file sidecar.

**Quy trình đúng (có cần Colab không?): xem [`HUONG_DAN.md`](HUONG_DAN.md).** Tóm tắt: CPU `convert.py` là bắt buộc; Colab GPU chỉ khi cần UniMERNet/OCR.

Mẫu chuẩn đã chạy thành công: `samples/de-vat-li-lan-3.docx`
(`ĐỀ VẬT LÍ LẦN 3_VER 2 (2).docx`).

## Tiêu chí thành công

| Đầu ra | Ý nghĩa |
| --- | --- |
| `markup.txt` | Văn bản phẳng, **đúng cú pháp Azota** |
| `sidecar/` | Asset thô theo document order |
| `manifest.json` | Map mỗi id → nguồn (rId, media/embedding, vị trí) |

Cú pháp Azota bắt buộc:

```
[!b:$đậm$]
[!i:$nghiêng$]
[!sup:$mũ$]
[!sub:$chỉ số$]
[!b!i:$đậm nghiêng$]
[!m:$mathml_N$]
[!m:$mathtype_N$]
[img:$img_N$]
[* cột 1 | cột 2 *]
*D. đáp án đúng          (trắc nghiệm A–D)
*a) mệnh đề đúng         (đúng/sai)
→ Đáp án: 69,6           (trả lời ngắn)
```

Sidecar:

- `mathml_N.xml` — fragment OMML (`m:oMath`)
- `mathtype_N.wmf\|emf` — preview render của OLE MathType (`Equation.DSMT4`)
- `mathtype_N_ole.bin` — file OLE gốc (để debug)
- `img_N.png\|jpeg\|wmf` — hình vẽ / ảnh trong đề

## Vì sao kết hợp 2 mô hình (không OCR toàn văn làm nguồn chính)

Đề vật lý kiểu này **không phải PDF scan**. File Word chứa đồng thời:

- OMML (Equation của Word) — 69 công thức trên mẫu
- MathType OLE (`Equation.DSMT4`) — 16 công thức, **không có LaTeX/MathML trong XML**
- Ảnh hình vẽ PNG/JPEG
- Gạch chân đáp án đúng, đậm/nghiêng/mũ/chỉ số, bảng

Nếu OCR cả trang rồi mới tách công thức, sẽ mất:

- thứ tự document-order chính xác
- rId / OLE / OMML gốc
- gạch chân → dấu `*`
- phân biệt `mathml_*` vs `mathtype_*`

Vì vậy pipeline là **cấu trúc OOXML làm xương sống**, vision chỉ vá những chỗ Word không cho text:

```
                    ┌─────────────────────────────────────┐
  .docx ──extract──►│  A. OOXML walker (CPU, lossless)    │──► markup.txt
                    │     runs → [!b:] [!i:] [!sup:] …    │    sidecar/
                    │     m:oMath → mathml_N.xml          │    manifest.json
                    │     MathType OLE → mathtype_N.wmf   │
                    │     drawing → img_N.*               │
                    │     underline A–D → *D.             │
                    └──────────────┬──────────────────────┘
                                   │
              ┌────────────────────┴────────────────────┐
              ▼                                         ▼
   B. UniMERNet (MER)                        C. Unlimited-OCR (R-SWA)
   crop WMF/EMF công thức                    render từng trang đề
   → LaTeX  (sidecar/*.tex)                  → markdown đọc-theo-thứ-tự
   overlay vào manifest.latex                QA hình vẽ, bảng-ảnh, missing text
```

### Vai trò từng mô hình

**UniMERNet** ([opendatalab/UniMERNet](https://github.com/opendatalab/UniMERNet)) — Swin encoder + mBART decoder, Length-Aware Module. Đúng bài toán *formula image → LaTeX*. Dùng cho:

- preview WMF của MathType (OLE không parse được)
- công thức nằm trong ảnh hình vẽ (sau khi Unlimited-OCR / detector cắt bbox)

**Unlimited-OCR** ([baidu/Unlimited-OCR](https://github.com/baidu/Unlimited-OCR)) — DeepEncoder + decoder R-SWA, KV cache không tăng theo độ dài, parse hàng chục trang một lượt (kiến trúc trong [Unlimited-OCR.png](https://github.com/baidu/Unlimited-OCR/blob/main/assets/Unlimited-OCR.png)). Dùng cho:

- QA reading-order toàn đề (so với `markup.txt`)
- chữ trong hình thí nghiệm / sơ đồ (Rutherford, mạch điện, đồ thị)
- bảng đang là ảnh chứ không phải `w:tbl`

Không thay thế OOXML bằng output OCR.

## Chạy local (CPU, không GPU)

```bash
cd tools/docx-to-azota
python3 convert.py samples/de-vat-li-lan-3.docx -o azota_out
```

Kết quả mẫu (file này):

| Asset | Số lượng |
| --- | ---: |
| `mathml_*` | 69 |
| `mathtype_*` | 16 |
| `img_*` | 8 |
| `*A.`–`*D.` (Phần I) | 18/18 |
| `→ Đáp án:` (Phần III) | 6/6 |

## Google Colab — chạy được + tối ưu (UniMERNet + Unlimited-OCR)

Có **hai notebook**:

| File | Khi nào dùng |
| --- | --- |
| `colab_optimized.ipynb` | **Chạy thật trên T4** — auto-profile, batch UniMERNet, không OCR cả trang |
| `colab_docx_to_azota.ipynb` | Đánh giá từng cell / PDF từng trang |

### Vì sao không OCR 5 trang trên Colab T4

Log cũ ~**57s/trang × 5 = 286s**, Unlimited-OCR 3B dễ **OOM** nếu load cùng UniMERNet.

Với đề `.docx` (ĐỀ VẬT LÍ LẦN 3):

| Việc | Công cụ | T4 |
| --- | --- | --- |
| Chữ, OMML, `*D.`, bảng | OOXML walker | ~0.2s, 0 VRAM |
| 16 công thức MathType | **UniMERNet-tiny batch** | ~441MB, vài giây |
| 8 hình vẽ | `[img:$img_N$]` giữ nguyên; Unlimited-OCR chỉ khi còn VRAM | tắt mặc định trên T4 |

Colab: **không** `pip install unimernet[full]` hay `tokenizers` 0.19 (Python 3.12/3.13, thiếu wheel, compile Rust fail). Dùng `install_colab.py` (`unimernet --no-deps` + `compat_transformers` cho transformers 5.x). Nếu vừa thấy `Building wheel for tokenizers` hoặc `apply_chunking_to_forward`: xem `HUONG_DAN.md` mục phục hồi.

Mở `colab_optimized.ipynb` → Runtime GPU T4 → `Shift+Enter`:

1. `detect_profile()` chọn tiny / batch=8 / fp16
2. `prepare_unimernet_checkpoint("tiny")` tải HuggingFace
3. Raster **chỉ** WMF MathType, `unimernet_batch`
4. `free_cuda()` rồi mới (tuỳ chọn) OCR hình
5. `inject_latex_into_markup` → `$\\alpha$` thay `[!m:$mathtype_N$]`

Bật OCR hình trên T4 (nếu bước 3 còn VRAM):

```python
PROFILE["ocr_figures"] = True
```

A100: profile tự bật `ocr_pages` + UniMERNet-small.

Mở `colab_docx_to_azota.ipynb`. Runtime **T4** cho Bước 1–3a; **A100** nếu bật Unlimited-OCR.

**Cách đánh giá:** `Shift+Enter` từng cell. Đừng Run all khi đang đo thời gian.

Mỗi bước in đúng format log của bạn:

```
⏱ THỜI GIAN:
Bước 1: 0.2s (OOXML extract)     # hoặc PDF → PNG
Bước 2: 1.8s (raster WMF)
Bước 3: 49s (UniMERNet 4 ảnh)    # từng ảnh, giống "OCR 4 ảnh"
Bước 4: 0.4s (ghép+fix)
TỔNG : 51.4s
```

Với PDF, Bước 3b in thời gian từng trang (giống log Kaggle 85s / 74s / …).

Cờ cắt mẫu để đánh giá nhanh:

```python
MAX_UNIMERNET_IMAGES = 4   # None = hết MathType
MAX_OCR_PAGES = 1          # None = hết trang (~57s × N)
RUN_UNLIMITED_OCR = False  # True trên A100
```

Thứ tự cell: config → clone → pip (3 cell) → upload → Bước 1 → preview → Bước 2 → UniMERNet load / infer từng ảnh → OCR từng trang → ghép+fix → `timer.print_summary()` → zip.

## Ánh xạ đề mẫu → Azota

| Trong Word | Trong `markup.txt` |
| --- | --- |
| Gạch chân `D.` | `*D.` |
| `[D]` / `[S]` (đúng/sai) | `*a)` / `a)` |
| `[GT] … [/]` | `Lời giải` (bỏ `[/]`) |
| `Đáp án là {a}` + `A.69,6` | `→ Đáp án: 69,6` |
| `A. … B. …` cùng dòng | tách 4 dòng |
| `w:tbl` | `[* c1 \| c2 *]` |
| `m:oMath` | `[!m:$mathml_N$]` |
| MathType OLE | `[!m:$mathtype_N$]` |
| `w:drawing` | `[img:$img_N$]` |

## Kiểm thử

```bash
python3 -m pytest tests/test_convert.py -q
```
