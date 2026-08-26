# Module làm sạch DOCX → markup Azota — Design Spec

- **Ngày:** 2026-07-06
- **Trạng thái:** Đã duyệt thiết kế; code Tầng A: package `docx_extract/`
- **Phạm vi:** Tái tạo "Tầng A" của Azota — biến file `.docx` đề thi thành markup phẳng giống panel phải của Azota, kèm trích xuất asset (toán/ảnh) ra sidecar. KHÔNG bao gồm convert công thức, KHÔNG build cây câu hỏi JSON.

---

## 1. Mục tiêu & tiêu chí thành công

Đầu vào: một file `.docx` (mẫu chuẩn: `ĐỀ VẬT LÍ LẦN 3_VER 2 (2).docx`).

Đầu ra:

- `markup.txt` — văn bản phẳng theo dòng, **y hệt cú pháp Azota**: `[!b:$...$]`, `[!i:$...$]`, `[!sup:$...$]`, `[!sub:$...$]`, `[!b!i:$...$]`, `[!m:$mathml_N$]`, `[!m:$mathtype_N$]`, `[img:$img_N$]`, bảng `[* c1 | c2 *]`, và dấu `*` trước chữ cái đáp án đúng.
- `sidecar/` — asset thô: `mathml_N.xml` (OMML fragment), `mathtype_N.wmf|emf` (render MathType OLE), `img_N.png|jpeg|wmf` (ảnh hình vẽ).
- `manifest.json` — map mỗi id placeholder → nguồn (rId, tên file media/embedding, vị trí trong document order).

**Tiêu chí thành công (thước đo chính):** golden diff giữa `markup.txt` và đoạn markup Azota đã có sẵn cho đúng file này (do người dùng cung cấp). Càng ít khác biệt càng tốt; mọi khác biệt còn lại phải giải thích được.

Regression hiện tại (khi chưa có dump panel Azota): `examples/de-vat-li-lan-3/markup.txt`.

## 2. Nguyên tắc cốt lõi

Markup Azota là **văn bản phẳng theo dòng**, KHÔNG phải cây JSON lồng nhau. Các mỏ neo cấu trúc (`Câu N`, `PHẦN I/II/III`, `Nhóm I..IV`, `[GT]`, `[/]`, `Chọn X.`, `Đáp án là {a}`) **đã là text sẵn trong file docx** (do giáo viên gõ). Vì vậy module là một **pipeline tuyến tính**, không cần nhận diện ngữ nghĩa câu hỏi.

Module chỉ **thêm hoặc biến đổi** đúng 5 nhóm việc:

1. Inline-tag định dạng (đậm/nghiêng/sup/sub) theo whitelist.
2. Placeholder cho toán/ảnh + trích asset.
3. Dấu `*` trước chữ cái đáp án đúng.
4. Cú pháp bảng `[* | *]`.
5. Gộp run liền kề cùng định dạng.

**Làm sạch = VỨT ĐI, không phải SỬA.** Chỉ giữ lại whitelist (b/i/sup/sub/math/img/table); mọi thuộc tính khác (màu, cỡ chữ, font, highlight, tab thừa, đoạn rỗng, run bị Word xé vụn) bị loại bỏ.

## 3. Kiến trúc — 6 unit

| Unit | File | Nhiệm vụ |
|---|---|---|
| loader | `docx_extract/loader.py` | Unzip; `document.xml`; rels `rId → media/embedding` |
| runs | `docx_extract/runs.py` | Whitelist tag; gộp run; tab/br |
| math_assets | `docx_extract/math_assets.py` | OMML / MathType OLE / drawing → placeholder + sidecar |
| answer | `docx_extract/answer.py` | `*` trước đáp án đúng (gạch chân) |
| tables | `docx_extract/tables.py` | `w:tbl` → `[* c1 \| c2 *]` |
| assemble | `docx_extract/assemble.py` | Duyệt body, ghi markup + sidecar + manifest |

CLI: `python3 -m docx_extract đề.docx -o azota_out` hoặc `python3 convert.py` (shim Colab).

## 4–5. Quy tắc & lỗi

Như spec đã duyệt: whitelist b/i/sup/sub; một bộ đếm N theo loại; `w:u` trên `A.`–`D.` → `*D.`; element lạ bỏ qua + log; thiếu rId vẫn phát placeholder.

## 6. Kiểm thử

- Unit: `tests/test_runs.py`, `test_math_assets.py`, `test_answer.py`, `test_tables.py`
- Golden/regression: `tests/test_golden.py` vs `examples/de-vat-li-lan-3/markup.txt`

## 7. Bố cục (mapping repo)

```
tools/docx-to-azota/
  docx_extract/          # = spec src/docx_extract/
  convert.py             # shim: from docx_extract import convert_docx
  tests/
  samples/de-vat-li-lan-3.docx
  examples/de-vat-li-lan-3/
  docs/superpowers/specs/2026-07-06-docx-clean-module-design.md
```

## 8. Ngoài phạm vi (YAGNI v1)

- Convert OMML→MathML, WMF→PNG/SVG, MathType→LaTeX (UniMERNet).
- OCR trang / Unlimited-OCR.
- Build cây JSON câu hỏi.
- Suy luận đáp án từ "Chọn X." khi không có tín hiệu định dạng (code có backfill nhẹ từ Lời giải — có thể tắt).
- UI sửa tay.
- `.doc` / PDF.
