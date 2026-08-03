// Package webui, s3intel için tarayıcıda çalışan basit bir arayüz sunar.
// Kendi başına ne AWS SDK ne de grayhatwarfare/HTTP-dış-çağrı mantığı içerir —
// tüm iş mantığını internal/scanjobs üzerinden çağırır, sadece sonucu HTML
// olarak render eder. SADECE 127.0.0.1'e bağlanır; dışarıya açık bir servis
// değildir (bkz. cmd/serve.go).
package webui

import (
	"fmt"
	"html/template"
	"net/http"

	"s3intel/internal/riskengine"
	"s3intel/internal/scanjobs"
)

const pageTemplate = `<!doctype html>
<html lang="tr">
<head>
<meta charset="utf-8">
<title>s3intel</title>
<style>
  body { font-family: system-ui, -apple-system, "Segoe UI", sans-serif; max-width: 920px; margin: 40px auto; padding: 0 16px; color: #1e293b; }
  h1 { margin-bottom: 4px; }
  .sub { color: #64748b; margin-top: 0; margin-bottom: 24px; font-size: 14px; }
  form { display: flex; gap: 10px; align-items: center; margin-bottom: 20px; flex-wrap: wrap; }
  select, input[type=text] { padding: 9px 11px; border: 1px solid #cbd5e1; border-radius: 6px; font-size: 14px; }
  input[type=text] { flex: 1 1 240px; }
  button { padding: 9px 18px; border: none; border-radius: 6px; background: #1d4ed8; color: white; font-size: 14px; font-weight: bold; cursor: pointer; }
  button:hover { background: #1e40af; }
  table { width: 100%; border-collapse: collapse; margin-top: 6px; }
  th { background: #1d4ed8; color: white; text-align: left; padding: 8px 10px; font-size: 12.5px; }
  td { padding: 7px 10px; border-bottom: 1px solid #e2e8f0; font-size: 13px; }
  tr:nth-child(even) td { background: #f8fafc; }
  .score { font-weight: bold; padding: 2px 9px; border-radius: 4px; color: white; display: inline-block; min-width: 24px; text-align: center; }
  .error { background: #fef2f2; border-left: 4px solid #dc2626; padding: 10px 14px; border-radius: 0 6px 6px 0; color: #b91c1c; margin-bottom: 16px; font-size: 13.5px; }
  .empty { color: #94a3b8; font-size: 13.5px; }
  .hint { color: #94a3b8; font-size: 11.5px; margin-top: 28px; }
  .badge { display: inline-block; font-size: 10.5px; font-weight: bold; padding: 2px 8px; border-radius: 999px; margin-left: 8px; vertical-align: middle; }
  .badge.active { background: #fee2e2; color: #b91c1c; }
  .badge.passive { background: #ccfbf1; color: #0f766e; }
</style>
</head>
<body>
  <h1>s3intel
    {{if eq .Mode "active"}}<span class="badge active">AKTİF — gerçek AWS</span>{{else}}<span class="badge passive">PASİF — grayhatwarfare/mock</span>{{end}}
  </h1>
  <p class="sub">Aktif mod SADECE whitelist'teki bucket'lara gerçek AWS çağrısı yapar. Pasif mod hiçbir zaman gerçek bir bucket'a bağlanmaz.</p>

  <form method="get" action="/">
    <select name="mode">
      <option value="passive" {{if eq .Mode "passive"}}selected{{end}}>Pasif (grayhatwarfare)</option>
      <option value="active" {{if eq .Mode "active"}}selected{{end}}>Aktif (whitelist bucket'ın)</option>
    </select>
    <input type="text" name="target" placeholder="bucket adı (aktif) ya da anahtar kelime (pasif)" value="{{.Target}}" required>
    <button type="submit">Tara</button>
  </form>

  {{if .Error}}<div class="error"><b>Hata:</b> {{.Error}}</div>{{end}}

  {{if .Submitted}}
    {{if .Results}}
    <table>
      <tr><th>Skor</th><th>Kategori</th><th>Kaynak</th><th>Secret</th><th>Dosya / Anahtar</th></tr>
      {{range .Results}}
      <tr>
        <td><span class="score" style="background:{{.Color}}">{{.Score}}</span></td>
        <td>{{.Category}}</td>
        <td>{{.Source}}</td>
        <td>{{.SecretCount}}</td>
        <td>{{.Key}}</td>
      </tr>
      {{end}}
    </table>
    {{else if not .Error}}
    <p class="empty">Sonuç bulunamadı.</p>
    {{end}}
  {{end}}

  <p class="hint">Bu sayfa sadece 127.0.0.1 (localhost) üzerinden erişilebilir. Durdurmak için terminalde Ctrl+C.</p>
</body>
</html>`

var tmpl = template.Must(template.New("page").Parse(pageTemplate))

type row struct {
	Score       int
	Color       string
	Category    string
	Source      string
	SecretCount int
	Key         string
}

type pageData struct {
	Mode      string
	Target    string
	Submitted bool
	Error     string
	Results   []row
}

func riskColor(score int) string {
	switch {
	case score >= 70:
		return "#dc2626"
	case score >= 40:
		return "#d97706"
	default:
		return "#16a34a"
	}
}

func toRows(results []riskengine.Result) []row {
	rows := make([]row, 0, len(results))
	for _, r := range results {
		rows = append(rows, row{
			Score:       r.Score,
			Color:       riskColor(r.Score),
			Category:    string(r.Category),
			Source:      r.Source,
			SecretCount: len(r.Secrets),
			Key:         r.Key,
		})
	}
	return rows
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	mode := r.URL.Query().Get("mode")
	if mode != "active" {
		mode = "passive"
	}
	target := r.URL.Query().Get("target")

	data := pageData{Mode: mode, Target: target}

	if target != "" {
		data.Submitted = true

		var results []riskengine.Result
		var err error

		switch mode {
		case "active":
			results, err = scanjobs.RunActive(r.Context(), target)
		case "passive":
			results, err = scanjobs.RunPassive(r.Context(), target)
		}

		if err != nil {
			data.Error = err.Error()
		} else {
			data.Results = toRows(results)
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = tmpl.Execute(w, data)
}

// Serve, s3intel web arayüzünü verilen adreste başlatır ve terminali bloklar.
// addr, cmd/serve.go tarafından her zaman "127.0.0.1:<port>" biçiminde verilir
// — bu paket kendi başına dış ağa açık bir adrese bağlanmaz.
func Serve(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleIndex)

	fmt.Printf("s3intel web arayüzü çalışıyor: http://%s\n", addr)
	fmt.Println("Tarayıcında bu adresi aç. Durdurmak için Ctrl+C.")

	return http.ListenAndServe(addr, mux)
}
