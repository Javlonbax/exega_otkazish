# KadastrApp (Go/Fyne) — to'liq GUI funksionaliga yaqin
Bu versiya Python'dagi `kadastr_prog.py` mantiqini Go'da Fyne GUI bilan takrorlaydi: papkadan CSV/JSON o'qish, ustun bo'yicha guruhlash + yig'indi, NS/OKPO/SOATO hisoblash va natijani `natija.xlsx`ga yozish.

## Nega GitHub Actions?
Sizning kompyuteringizda admin va internet cheklovlari bor; Fyne/Excelize kutubxonalarini lokal yuklab bo'lmaydi. Shu sababli **bulutda build** qilib, tayyor `KadastrApp.exe`ni Artifacts'dan yuklab olasiz.

## Qadamlar
1. GitHub'da yangi repository yarating (private bo'lishi mumkin).
2. Ushbu papkadagi hamma fayllarni repo'ga push qiling:
   - `kadastr_gui_fyne.go`
   - `go.mod`
   - `.github/workflows/build_windows.yml`

3. Actions bo'limida workflow'ni ishga tushiring (Run workflow).
4. Tugagach, **Artifacts** dan `KadastrApp.exe` ni yuklab oling.

## Lokal build (internet ochiq va Go o'rnatilgan bo'lsa)
```
go mod tidy
go build -ldflags "-s -w -H=windowsgui" -o KadastrApp.exe kadastr_gui_fyne.go
```

## Ishlatish
- Dasturda papkani tanlang, `*.csv;*.json` qoldiring (JSON Lines bo'lsa flagni belgilang).
- Guruhlash (`soato_tum`) va yig'indi ustun nomini kiriting (`qiymat_oy3` yoki `kadastr qiymati, ming so\`m`).
- "Hisobla va saqla" bosib, `natija.xlsx` faylini saqlang.

> Eslatma: Python'dagi barcha biznes qoidalar 1:1 ko'chirilmagan — soddalashtirilgan. Kerak bo'lsa, aniq qoidalarni bosqichma-bosqich qo'shamiz.
