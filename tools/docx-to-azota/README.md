# DOCX → Azota markup (hybrid UniMERNet + Unlimited-OCR)

Chuyển đề thi Word (`.docx`) sang **cú pháp Azota phẳng theo dòng**, giữ nguyên công thức / ảnh / bảng dưới dạng placeholder + file sidecar.

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

## Google Colab (GPU) — overlay 2 mô hình

Mở `colab_docx_to_azota.ipynb` (runtime **T4** cho UniMERNet-tiny; **A100/L4** nếu bật Unlimited-OCR 3B).

Thứ tự cell:

1. Cài `convert.py` (clone repo hoặc upload folder này)
2. Upload `.docx` → chạy `convert_docx()` → có ngay `markup.txt` / `sidecar/` / `manifest.json`
3. Raster WMF → PNG
4. UniMERNet nhận dạng từng `mathtype_N` → ghi `sidecar/mathtype_N.tex` + `manifest.latex`
5. (Tùy chọn) LibreOffice `.docx` → PDF → Unlimited-OCR `infer` / `infer_multi`
6. Tải zip kết quả

Cờ trong notebook:

```python
RUN_UNIMERNET = True          # T4 16GB đủ tiny/small
RUN_UNLIMITED_OCR = False     # bật khi A100; T4 dễ OOM
```

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
