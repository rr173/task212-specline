package httpapi

import (
	"net/http"

	"task212-specline/internal/observation"
)

type addPeaksRequest struct {
	Peaks []observation.PeakInput `json:"peaks"`
}

func (s *Server) addPeaks(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	var req addPeaksRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	for i := range req.Peaks {
		if req.Peaks[i].Unit == "nm" {
			req.Peaks[i].Unit = "angstrom"
		}
	}
	peaks, err := s.app.Observation.AddPeaks(id, req.Peaks)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, peaks)
}

func (s *Server) listPeaks(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	peaks, err := s.app.Observation.GetPeaks(id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, peaks)
}

func (s *Server) markArtifact(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	p, err := s.app.Review.MarkCosmicRay(id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (s *Server) excludePeak(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	p, err := s.app.Review.ExcludePeak(id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (s *Server) reopenPeak(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	p, err := s.app.Review.ReopenPeak(id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, p)
}
