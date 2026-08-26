# DOCX → Azota

Hai việc riêng. **Không** trộn UniMERNet vào bước extract.

| Việc | Cần GPU? | File |
| --- | --- | --- |
| 1. Extract Azota | Không | `markup.txt` + `sidecar/` + `manifest.json` |
| 2. MathType → `$latex$` | Colab T4 | thay `[!m:$mathtype_N$]` |

Azota nhận placeholder. Đề mẫu: 69 mathml, 16 mathtype, 8 img **không GPU**.

**Cấm trên Colab:** `pip install unimernet[full]`, `pip install tokenizers`, `pip install transformers==4.42.4`.

---

## Máy bạn (khuyến nghị)

```bash
cd tools/docx-to-azota
python3 convert.py "đề.docx" -o azota_out
```

Mẫu: `python3 convert.py samples/de-vat-li-lan-3.docx -o azota_out`

---

## Colab — lưu hết vào Drive folder **markdown azota**

Code + đề + `azota_out` + zip nằm ở `MyDrive/markdown azota` (không phải `/content`, không mất khi tắt máy).

1. Trên Google Drive, giữ folder tên **`markdown azota`** (đã có thì dùng; chưa có notebook sẽ tạo).
2. Runtime → **Disconnect and delete runtime**.
3. Tải notebook: [colab_start_here.ipynb](https://github.com/phuchoang2603/refurbished-marketplace/blob/cursor/docx-to-azota-pipeline-4d56/tools/docx-to-azota/colab_start_here.ipynb)
4. Colab → File → Upload notebook.
5. Runtime → **T4 GPU**.
6. Chạy ô **A2** trước — cấp quyền Drive, đợi `OK True`. **Đừng bấm Stop.**
7. `Shift+Enter` từng ô. Không Run all. Không dán 3 dòng `from install_colab import …` từ chat (mất `sys.path`).
8. Nếu `ModuleNotFoundError: install_colab`: chạy **ô B2 đầy đủ** (tự tìm Drive / clone), không Restart rồi chỉ chạy 3 dòng import.

Phần A (Drive → upload → extract → zip trong folder của bạn) **không cài UniMERNet**. Xong phần A là đủ nộp Azota.

Phần B chỉ khi cần `$latex$` từ ảnh MathType. Checkpoint model cũng vào `markdown azota/models`.
