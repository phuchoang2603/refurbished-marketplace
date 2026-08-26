# DOCX → Azota

Hai việc riêng. **Không** trộn UniMERNet vào bước extract. Tầng A (spec đã duyệt) là `docx_extract/` — `python3 convert.py` chỉ là shim.

| Việc | Cần GPU? | File |
| --- | --- | --- |
| 1. Extract Azota | Không | `markup.txt` + `sidecar/` + `manifest.json` |
| 2. MathType → `$latex$` | Colab T4 | thay `[!m:$mathtype_N$]` |

Azota nhận placeholder. Đề mẫu: 69 mathml, 16 mathtype, 8 img **không GPU**.

**Cấm trên Colab:** `pip install unimernet[full]`, `pip install tokenizers`, `pip install transformers==4.42.4`.

**Không cần Google Drive.**

---

## Máy bạn (khuyến nghị)

```bash
cd tools/docx-to-azota
python3 convert.py "đề.docx" -o azota_out
```

Mẫu: `python3 convert.py samples/de-vat-li-lan-3.docx -o azota_out`

---

## Colab — làm lại từ đầu

1. Runtime → **Disconnect and delete runtime** (xóa runtime bẩn).
2. Tải notebook: [colab_start_here.ipynb](https://github.com/phuchoang2603/refurbished-marketplace/blob/cursor/docx-to-azota-pipeline-4d56/tools/docx-to-azota/colab_start_here.ipynb)
3. Colab → File → Upload notebook → chọn file vừa tải.
4. Runtime → **T4 GPU**.
5. `Shift+Enter` từng ô. Không Run all. Không dán cell từ chat cũ / `Untitled1.ipynb`.
6. Ô A2 clone vào `/content/docx-to-azota` — **đừng bấm Stop**, đợi `OK True`. **Không** hiện hộp Google Drive.

Phần A (clone → upload → extract → zip) **không cài UniMERNet**. Ô A5 in `EXTRACT HOÀN TẤT` (cùng khuôn log Kaggle). Xong phần A là đủ nộp Azota.

Phần B chỉ khi cần `$latex$` từ ảnh MathType. Nếu `ModuleNotFoundError: install_colab`, chạy **cả ô B2** (tự clone `/content/docx-to-azota`), không chỉ 3 dòng import.
