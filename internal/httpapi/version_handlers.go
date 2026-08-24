package httpapi

import (
	"net/http"
	"strconv"
)

type createVersionRequest struct {
	ObservationID int64  `json:"observation_id"`
	Label         string `json:"label"`
}

func (s *Server) createVersion(w http.ResponseWriter, r *http.Request) {
	var req createVersionRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	v, err := s.app.Versioning.Create(req.ObservationID, req.Label)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, v)
}

func (s *Server) listVersions(w http.ResponseWriter, r *http.Request) {
	obsIDStr := r.URL.Query().Get("observation_id")
	if obsIDStr == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "observation_id query required"})
		return
	}
	obsID, err := strconv.ParseInt(obsIDStr, 10, 64)
	if err != nil {
		writeError(w, err)
		return
	}
	vs, err := s.app.Versioning.List(obsID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, vs)
}

func (s *Server) getVersion(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	v, err := s.app.Versioning.Get(id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, v)
}

func (s *Server) shareVersion(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	v, err := s.app.Versioning.Share(id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, v)
}

func (s *Server) freezeVersion(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	v, err := s.app.Versioning.Freeze(id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, v)
}
