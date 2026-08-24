package httpapi

import (
	"html/template"
	"net/http"

	"task212-specline/internal/model"
)

// indexTemplate 观测集列表页。
var indexTemplate = template.Must(template.New("index").Parse(`<!DOCTYPE html>
<html lang="zh-CN"><head><meta charset="utf-8"><title>天文光谱线归属复核台</title>
<style>body{font-family:system-ui,sans-serif;max-width:900px;margin:2rem auto;padding:0 1rem;color:#222}
table{border-collapse:collapse;width:100%}td,th{border:1px solid #ddd;padding:.5rem;text-align:left}
th{background:#f5f5f5}a{color:#0366d6;text-decoration:none}</style></head>
<body><h1>天文光谱线归属复核台</h1>
<p>导入光谱峰 → 波长漂移校准 → 元素跃迁匹配 → 证据复核 → 冻结归属版本。</p>
<table><thead><tr><th>ID</th><th>名称</th><th>目标</th><th>状态</th><th>操作</th></tr></thead>
<tbody>{{range .}}<tr><td>{{.ID}}</td><td>{{.Name}}</td><td>{{.Target}}</td>
<td>{{.Status}}</td><td><a href="/observations/{{.ID}}">查看</a></td></tr>{{else}}
<tr><td colspan="5">暂无观测集</td></tr>{{end}}</tbody></table>
</body></html>`))

// indexPage 渲染观测集列表页。
func (s *Server) indexPage(w http.ResponseWriter, r *http.Request) {
	list, err := s.app.DB.ListObservations()
	if err != nil {
		writeError(w, err)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = indexTemplate.Execute(w, list)
}

// observationPageData 观测详情页数据。
type observationPageData struct {
	Observation *model.ObservationSet
	Peaks       []*model.SpectralPeak
	Candidates  []*model.AttributionCandidate
	Versions    []*model.AttributionVersion
}

var observationTemplate = template.Must(template.New("obs").Parse(`<!DOCTYPE html>
<html lang="zh-CN"><head><meta charset="utf-8"><title>观测 {{.Observation.ID}} - 归属复核台</title>
<style>body{font-family:system-ui,sans-serif;max-width:1000px;margin:2rem auto;padding:0 1rem;color:#222}
table{border-collapse:collapse;width:100%;margin:1rem 0}td,th{border:1px solid #ddd;padding:.4rem;text-align:left;font-size:.9rem}
th{background:#f5f5f5}.artifact{color:#c00}.frozen{color:#060}h2{border-bottom:2px solid #eee;padding-bottom:.3rem}
a{color:#0366d6;text-decoration:none}</style></head>
<body><p><a href="/">← 返回列表</a></p>
<h1>观测 {{.Observation.Name}}</h1>
<p>目标：{{.Observation.Target}} ｜ 状态：<b>{{.Observation.Status}}</b> ｜ 单位：{{.Observation.WavelengthUnit}}</p>

<h2>光谱峰（波长区间）</h2>
<table><thead><tr><th>序号</th><th>测量波长</th><th>校准波长</th><th>状态</th></tr></thead>
<tbody>{{range .Peaks}}<tr><td>{{.Index}}</td><td>{{printf "%.3f" .MeasuredWL}} Å</td>
<td>{{printf "%.3f" .CorrectedWL}} Å</td><td>{{if eq .Status "excluded"}}<span class="artifact">{{.Status}}</span>{{else}}{{.Status}}{{end}}</td></tr>
{{end}}</tbody></table>

<h2>归属候选与证据</h2>
<table><thead><tr><th>ID</th><th>跃迁</th><th>元素</th><th>静止波长</th><th>残差</th><th>评分</th><th>状态</th></tr></thead>
<tbody>{{range .Candidates}}<tr><td>{{.ID}}</td><td>{{.TransitionKey}}</td><td>{{.Element}} {{.Ionization}}</td>
<td>{{printf "%.3f" .RestWL}} Å</td><td>{{printf "%.3f" .Residual}}</td><td>{{printf "%.3f" .Score}}</td><td>{{.Status}}</td></tr>
{{else}}<tr><td colspan="7">尚未生成候选</td></tr>{{end}}</tbody></table>

<h2>归属版本</h2>
<table><thead><tr><th>ID</th><th>标签</th><th>状态</th><th>跃迁库</th></tr></thead>
<tbody>{{range .Versions}}<tr><td>{{.ID}}</td><td>{{.Label}}</td><td>{{if eq .Status "frozen"}}<span class="frozen">{{.Status}}</span>{{else}}{{.Status}}{{end}}</td><td>{{.TransitionLibVersion}}</td></tr>
{{else}}<tr><td colspan="4">暂无版本</td></tr>{{end}}</tbody></table>
</body></html>`))

// observationPage 渲染观测详情页（波长区间 + 候选证据）。
func (s *Server) observationPage(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	o, err := s.app.DB.GetObservation(id)
	if err != nil {
		writeError(w, err)
		return
	}
	peaks, _ := s.app.DB.ListPeaks(id)
	cands, _ := s.app.DB.ListCandidates(id)
	vers, _ := s.app.DB.ListVersions(id)
	data := observationPageData{Observation: o, Peaks: peaks, Candidates: cands, Versions: vers}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = observationTemplate.Execute(w, data)
}
