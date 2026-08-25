package httpapi

import (
	"net/http"

	"task212-specline/internal/model"
)

type calibrateRequest struct {
	Model string `json:"model"`
}

func (s *Server) calibrate(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	var req calibrateRequest
	_ = decodeJSON(r, &req) // 允许空体，默认 offset
	if req.Model == "" {
		req.Model = "offset"
	}
	c, err := s.app.Calibration.Calibrate(id, req.Model)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func (s *Server) getCalibration(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	c, err := s.app.Calibration.Result(id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, c)
}

type matchRequest struct {
	Tolerance float64 `json:"tolerance"`
}

func (s *Server) match(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	var req matchRequest
	_ = decodeJSON(r, &req)
	cands, err := s.app.Matching.Match(id, req.Tolerance)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, cands)
}

func (s *Server) listCandidates(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	// 仅查询已持久化的归属候选：GET 不应产生隐式匹配副作用。
	// 尚未生成候选（含空观测集）时返回 HTTP 200 与空列表。
	cands, err := s.app.Matching.Candidates(id)
	if err != nil {
		writeError(w, err)
		return
	}
	if cands == nil {
		cands = []*model.AttributionCandidate{}
	}
	writeJSON(w, http.StatusOK, cands)
}

func (s *Server) getCandidate(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	c, err := s.app.Matching.Candidate(id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func (s *Server) confirmCandidate(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	c, err := s.app.Matching.Confirm(id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func (s *Server) rejectCandidate(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	c, err := s.app.Matching.Reject(id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, c)
}

type refutationRequest struct {
	Kind string `json:"kind"`
	Note string `json:"note"`
}

func (s *Server) addRefutation(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	var req refutationRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	ref, err := s.app.Review.AddRefutation(id, req.Kind, req.Note)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, ref)
}

func (s *Server) listRefutations(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	refs, err := s.app.Review.Refutations(id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, refs)
}

type adjustPriorRequest struct {
	TransitionKey string  `json:"transition_key"`
	Prior         float64 `json:"prior"`
}

func (s *Server) adjustPrior(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	var req adjustPriorRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	if err := s.app.Review.AdjustPrior(id, req.TransitionKey, req.Prior); err != nil {
		writeError(w, err)
		return
	}
	hash, err := s.app.Review.PriorHash(id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"prior_hash": hash})
}
