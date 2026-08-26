# markdown-azota

Dữ liệu **markdown Azota** lấy từ pipeline “Định dạng azota từ docx” (`tools/docx-to-azota`) trong repo `refurbished-marketplace`.

## Dán vào Azota

File chính: **`azota.md`**

Kèm theo:

| File | Việc |
| --- | --- |
| `azota.md` | Markdown Azota (cùng nội dung `markup.txt`) |
| `markup.txt` | Bản sao `.txt` |
| `sidecar/` | 69 OMML + `.tex`, 16 MathType WMF, 8 ảnh |
| `manifest.json` | Map id → file nguồn |
| `samples/de-vat-li-lan-3.docx` | Đề gốc (`ĐỀ VẬT LÍ LẦN 3_VER 2 (2).docx`) |
| `docs/` | Hướng dẫn + spec Tầng A |
| `pipeline/` | `convert.py` + `docx_extract` để chạy lại |

Đề mẫu: **69** công thức Word → `$latex$`, **16** MathType còn `[!m:$mathtype_N$]`, **8** ảnh.

## Chạy lại từ docx

```bash
cd markdown-azota/pipeline
python3 convert.py ../samples/de-vat-li-lan-3.docx -o ..
```
