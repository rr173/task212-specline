// Package httpapi 提供 HTTP/JSON API 与轻量 Web 呈现层（路由前缀 /api）。
package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"task212-specline/internal/model"
	"task212-specline/internal/service"
)

// Server HTTP 服务。
type Server struct {
	app *service.App
	mux *http.ServeMux
}

// New 构造 HTTP 服务并注册全部路由。
func New(app *service.App) *Server {
	s := &Server{app: app, mux: http.NewServeMux()}
	s.routes()
	return s
}

// Handler 返回底层 http.Handler。
func (s *Server) Handler() http.Handler { return s.mux }

// routes 注册全部 API 与页面路由。
func (s *Server) routes() {
	// 观测集
	s.mux.HandleFunc("POST /api/observations", s.createObservation)
	s.mux.HandleFunc("GET /api/observations", s.listObservations)
	s.mux.HandleFunc("GET /api/observations/{id}", s.getObservation)
	s.mux.HandleFunc("POST /api/observations/{id}/publish", s.publishObservation)
	s.mux.HandleFunc("POST /api/observations/{id}/archive", s.archiveObservation)

	// 光谱峰
	s.mux.HandleFunc("POST /api/observations/{id}/peaks", s.addPeaks)
	s.mux.HandleFunc("GET /api/observations/{id}/peaks", s.listPeaks)
	s.mux.HandleFunc("POST /api/peaks/{id}/mark-artifact", s.markArtifact)
	s.mux.HandleFunc("POST /api/peaks/{id}/exclude", s.excludePeak)
	s.mux.HandleFunc("POST /api/peaks/{id}/reopen", s.reopenPeak)

	// 校准
	s.mux.HandleFunc("POST /api/observations/{id}/calibrate", s.calibrate)
	s.mux.HandleFunc("GET /api/observations/{id}/calibration", s.getCalibration)

	// 归属匹配
	s.mux.HandleFunc("POST /api/observations/{id}/match", s.match)
	s.mux.HandleFunc("GET /api/observations/{id}/candidates", s.listCandidates)
	s.mux.HandleFunc("GET /api/candidates/{id}", s.getCandidate)
	s.mux.HandleFunc("POST /api/candidates/{id}/confirm", s.confirmCandidate)
	s.mux.HandleFunc("POST /api/candidates/{id}/reject", s.rejectCandidate)

	// 反证与先验
	s.mux.HandleFunc("POST /api/candidates/{id}/refutations", s.addRefutation)
	s.mux.HandleFunc("GET /api/candidates/{id}/refutations", s.listRefutations)
	s.mux.HandleFunc("PUT /api/observations/{id}/prior", s.adjustPrior)

	// 归属版本
	s.mux.HandleFunc("POST /api/versions", s.createVersion)
	s.mux.HandleFunc("GET /api/versions", s.listVersions)
	s.mux.HandleFunc("GET /api/versions/{id}", s.getVersion)
	s.mux.HandleFunc("POST /api/versions/{id}/share", s.shareVersion)
	s.mux.HandleFunc("POST /api/versions/{id}/freeze", s.freezeVersion)

	// 统计与健康
	s.mux.HandleFunc("GET /api/stats", s.stats)
	s.mux.HandleFunc("GET /api/health", s.health)

	// 页面（全栈 Web 呈现）
	s.mux.HandleFunc("GET /{$}", s.indexPage)
	s.mux.HandleFunc("GET /observations/{id}", s.observationPage)
}

// --- 工具函数 ---

// writeJSON 输出 JSON 响应。
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError 把业务错误映射为 HTTP 状态码并输出错误体。
func writeError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, model.ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, model.ErrInvalid):
		status = http.StatusBadRequest
	case errors.Is(err, model.ErrConflict), errors.Is(err, model.ErrDuplicate):
		status = http.StatusConflict
	case errors.Is(err, model.ErrArchived):
		status = http.StatusLocked
	}
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

// pathID 解析路径中的整数 ID。
func pathID(r *http.Request) (int64, error) {
	return strconv.ParseInt(r.PathValue("id"), 10, 64)
}

// decodeJSON 解析 JSON 请求体。
func decodeJSON(r *http.Request, dst any) error {
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}
