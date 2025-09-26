package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/xuri/excelize/v2"
)

type rec map[string]any

var (
	oylar = []string{"Yanvar","Fevral","Mart","Aprel","May","Iyun","Iyul","Avgust","Sentabr","Oktabr","Noyabr","Dekabr"}
	choraklar = []string{"1-chorak","2-chorak","3-chorak","4-chorak"}
	kalitlar = []struct{ Key string; Val int }{
		{"gaz",350},{"suv",360},{"kafe",470},{"market",470},{"savdo",470},
		{"kino",590},{"aloqa",610},{"pochta",630},{"maktab",850},{"stomatolog",860},
		{"sport",930},
	}
)

func translitLower(s string) string {
	r := []rune(strings.ToLower(s))
	for i, ch := range r {
		switch ch {
		case 'ғ': r[i] = 'g'
		case 'ҳ': r[i] = 'h'
		case 'қ': r[i] = 'q'
		case 'ў': r[i] = 'o'
		}
	}
	return string(r)
}

func okedQosh(name string) int {
	l := translitLower(name)
	for _, kv := range kalitlar {
		if strings.Contains(l, kv.Key) {
			return kv.Val
		}
	}
	return 960
}

func parseFloatFlexible(v any) (float64, error) {
	switch t := v.(type) {
	case float64:
		return t, nil
	case json.Number:
		return t.Float64()
	case string:
		s := strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(t, " ", ""), ",", ""))
		if s == "" { return 0, errors.New("empty") }
		return strconv.ParseFloat(s, 64)
	default:
		return 0, fmt.Errorf("unsupported %T", v)
	}
}

func readCSV(path string) ([]rec, error) {
	f, err := os.Open(path)
	if err != nil { return nil, err }
	defer f.Close()
	r := csv.NewReader(f)
	r.FieldsPerRecord = -1
	header, err := r.Read()
	if err != nil { return nil, err }
	var out []rec
	for {
		row, err := r.Read()
		if err != nil { break }
		if len(row) == 0 { continue }
		m := rec{}
		for i, h := range header {
			if i < len(row) {
				m[strings.TrimSpace(h)] = row[i]
			}
		}
		out = append(out, m)
	}
	return out, nil
}

func readJSON(path string, jsonLines bool) ([]rec, error) {
	if jsonLines {
		f, err := os.Open(path)
		if err != nil { return nil, err }
		defer f.Close()
		sc := bufio.NewScanner(f)
		var out []rec
		for sc.Scan() {
			var m rec
			if err := json.Unmarshal(sc.Bytes(), &m); err == nil {
				out = append(out, m)
			}
		}
		return out, sc.Err()
	}
	b, err := os.ReadFile(path)
	if err != nil { return nil, err }
	var arr []rec
	dec := json.NewDecoder(strings.NewReader(string(b)))
	dec.UseNumber()
	if err := dec.Decode(&arr); err != nil { return nil, err }
	return arr, nil
}

func toFloat(m rec, key string) float64 {
	v, ok := m[key]
	if !ok { return 0 }
	f, _ := parseFloatFlexible(v)
	return f
}

func toInt(m rec, key string) int {
	v, ok := m[key]
	if !ok { return 0 }
	switch t := v.(type) {
	case float64: return int(t)
	case json.Number:
		i, _ := t.Int64(); return int(i)
	case string:
		i, _ := strconv.Atoi(strings.TrimSpace(t)); return i
	default: return 0
	}
}

func toString(m rec, key string) string {
	v, ok := m[key]
	if !ok { return "" }
	return fmt.Sprintf("%v", v)
}

func groupSum(rows []rec, groupKeys []string, valKey string) map[string]float64 {
	sums := map[string]float64{}
	for _, m := range rows {
		var parts []string
		miss := false
		for _, k := range groupKeys {
			kk := fmt.Sprintf("%v", m[k])
			if kk == "" { miss = true; break }
			parts = append(parts, kk)
		}
		if miss { continue }
		key := strings.Join(parts, "|")
		sums[key] += toFloat(m, valKey)
	}
	return sums
}

func writeExcel(outPath string, headers []string, rows [][]any) error {
	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	for j, h := range headers {
		col := string('A' + j)
		cell := fmt.Sprintf("%s1", col)
		f.SetCellValue(sheet, cell, h)
	}
	for i, r := range rows {
		for j, v := range r {
			col := string('A' + j)
			cell := fmt.Sprintf("%s%d", col, i+2)
			f.SetCellValue(sheet, cell, v)
		}
	}
	_ = f.SetColWidth(sheet, "A", "Z", 16)
	return f.SaveAs(outPath)
}

func main() {
	a := app.NewWithID("kadastr.fullgui")
	w := a.NewWindow("KadastrApp (Go/Fyne) — Full GUI")
	w.Resize(fyne.NewSize(900, 640))

	folderEntry := widget.NewEntry()
	folderBtn := widget.NewButtonWithIcon("Papkani tanlash", theme.FolderOpenIcon(), func() {
		d := dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil || uri == nil { return }
			folderEntry.SetText(uri.Path())
		}, w)
		d.Show()
	})

	filePattern := widget.NewEntry()
	filePattern.SetText("*.csv;*.json")
	jsonLinesCheck := widget.NewCheck("JSON Lines (NDJSON)", func(bool){})

	yillik := widget.NewCheck("Yillik", func(bool){})
	oylar := []string{"Yanvar","Fevral","Mart","Aprel","May","Iyun","Iyul","Avgust","Sentabr","Oktabr","Noyabr","Dekabr"}
	oySelect := widget.NewSelect(oylar, func(string){})
	now := time.Now()
	mo := int(now.Month()) - 1
	if mo <= 0 { mo = 12 }
	oySelect.SetSelected(oylar[mo-1])

	keyCol := widget.NewEntry()
	keyCol.SetPlaceHolder("Guruhlash ustuni: masalan soato_tum")
	valCol := widget.NewEntry()
	valCol.SetPlaceHolder("Yig'iladigan ustun: masalan 'qiymat_oy3' yoki 'kadastr qiymati, ming so`m'")

	status := widget.NewMultiLineEntry()
	status.SetMinRowsVisible(12)

	runBtn := widget.NewButtonWithIcon("Hisobla va saqla", theme.ConfirmIcon(), func() {
		folder := strings.TrimSpace(folderEntry.Text)
		if folder == "" {
			dialog.ShowError(errors.New("Papka tanlanmagan"), w); return
		}
		patterns := strings.Split(strings.TrimSpace(filePattern.Text), ";")
		_isJSONL := jsonLinesCheck.Checked
		kcol := strings.TrimSpace(keyCol.Text)
		vcol := strings.TrimSpace(valCol.Text)
		if kcol == "" || vcol == "" {
			dialog.ShowError(errors.New("Key/Value ustunlari to'ldirilmagan"), w); return
		}

		var all []rec
		for _, pat := range patterns {
			p := strings.TrimSpace(pat)
			if p == "" { continue }
			glob := filepath.Join(folder, p)
			files, _ := filepath.Glob(glob)
			for _, fp := range files {
				ext := strings.ToLower(filepath.Ext(fp))
				var rows []rec
				var err error
				if ext == ".csv" {
					rows, err = readCSV(fp)
				} else if ext == ".json" {
					rows, err = readJSON(fp, _isJSONL)
				} else {
					continue
				}
				if err != nil {
					status.SetText(status.Text + fmt.Sprintf("O'tkazildi: %s (%v)\n", filepath.Base(fp), err))
					continue
				}
				all = append(all, rows...)
				status.SetText(status.Text + fmt.Sprintf("OK: %s\n", filepath.Base(fp)))
			}
		}
		if len(all) == 0 {
			dialog.ShowInformation("Natija yo'q", "Fayl topilmadi yoki o'qib bo'lmadi", w)
			return
		}

		for _, m := range all {
			if _, ok := m["oked"]; !ok {
				if s := toString(m, "name_liter"); s != "" {
					m["oked"] = okedQosh(s)
				}
			}
			soatot := toInt(m, "soato_tum")
			m["okpo"] = 61500000 + (soatot % 1000)
			m["soato"] = soatot / 1000
		}

		// 1) soato_tum bo'yicha yig'indi
		sums1 := groupSum(all, []string{kcol}, vcol)

		// 2) ns = 41 yoki 42 (oked 350/360 bo'lsa 42)
		var rows2 []rec
		for _, m := range all {
			ns := 41
			ov := fmt.Sprintf("%v", m["oked"])
			if ov == "350" || ov == "360" {
				ns = 42
			}
			r := rec{}
			for k, v := range m { r[k] = v }
			r["ns"] = ns
			rows2 = append(rows2, r)
		}
		sums2 := groupSum(rows2, []string{"ns", kcol}, vcol)

		// Tayyorlash
		type kv struct{ Key string; Val float64 }
		var out1 []kv
		for k, v := range sums1 {
			out1 = append(out1, kv{k, v/1000.0})
		}
		sort.Slice(out1, func(i,j int) bool { return out1[i].Key < out1[j].Key })

		var out2 []struct{ NS, K string; V float64 }
		for k, v := range sums2 {
			parts := strings.SplitN(k, "|", 2)
			if len(parts) != 2 { continue }
			out2 = append(out2, struct{NS,K string; V float64}{parts[0], parts[1], v/1000.0})
		}
		sort.Slice(out2, func(i,j int) bool {
			if out2[i].NS == out2[j].NS { return out2[i].K < out2[j].K }
			return out2[i].NS < out2[j].NS
		})

		dialog.NewFileSave(func(uc fyne.URIWriteCloser, err error) {
			if err != nil || uc == nil { return }
			outPath := uc.URI().Path()
			_ = uc.Close()
			if strings.ToLower(filepath.Ext(outPath)) != ".xlsx" {
				outPath += ".xlsx"
			}
			var headers = []string{"mes","okpo","soato","razdel","ns","g1"}
			var rows [][]any
			// mes (oy raqami, yillik bo'lsa 12)
			mes := 12
			if !yillik.Checked {
				sel := oySelect.Selected
				for i, name := range oylar {
					if name == sel { mes = i+1; break }
				}
			}
			// Jadval-1
			for _, kv := range out1 {
				soatot, _ := strconv.Atoi(strings.Split(kv.Key, "|")[0])
				okpo := 61500000 + (soatot % 1000)
				soato := soatot / 1000
				rows = append(rows, []any{mes, okpo, soato, 1, 12, kv.Val})
			}
			// Jadval-2 (ns 41/42)
			for _, r := range out2 {
				soatot, _ := strconv.Atoi(strings.Split(r.K, "|")[0])
				okpo := 61500000 + (soatot % 1000)
				soato := soatot / 1000
				nsInt, _ := strconv.Atoi(r.NS)
				rows = append(rows, []any{mes, okpo, soato, 1, nsInt, r.V})
			}
			if err := writeExcel(outPath, headers, rows); err != nil {
				dialog.ShowError(err, w); return
			}
			dialog.ShowInformation("Tayyor", "Natija saqlandi: "+outPath, w)
		}, w).SetFileName("natija.xlsx").Show()
	})

	form := container.NewVBox(
		widget.NewLabelWithStyle("KadastrApp (Go/Fyne) — Python GUI funksionalligiga yaqin", fyne.TextAlignLeading, fyne.TextStyle{Bold:true}),
		widget.NewSeparator(),
		container.NewBorder(nil, nil, nil, folderBtn, folderEntry),
		widget.NewSeparator(),
		container.NewGridWithColumns(2,
			container.NewVBox(widget.NewLabel("Fayl shablonlari"), filePattern, jsonLinesCheck),
			container.NewVBox(widget.NewLabel("Davr"), yillik, oySelect),
		),
		widget.NewSeparator(),
		container.NewGridWithColumns(2,
			container.NewVBox(widget.NewLabel("Guruhlash ustuni (mas: soato_tum)"), keyCol),
			container.NewVBox(widget.NewLabel("Yig'indisi olinadigan ustun (mas: qiymat_oy3)"), valCol),
		),
		widget.NewSeparator(),
		runBtn,
		widget.NewLabel("Holat / log:"),
		status,
	)

	w.SetContent(container.NewPadded(form))
	w.ShowAndRun()
}
