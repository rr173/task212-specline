package httpapi

import (
	"fmt"
	"net/http"
)

type createObservationRequest struct {
	Name   string `json:"name"`
	Target string `json:"target"`
	Unit   string `json:"unit"`
}

func (s *Server) createObservation(w http.ResponseWriter, r *http.Request) {
	var req createObservationRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	o, err := s.app.Observation.Create(req.Name, req.Target, req.Unit)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, o)
}

func (s *Server) listObservations(w http.ResponseWriter, r *http.Request) {
	list, err := s.app.DB.ListObservations()
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) getObservation(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	o, err := s.app.DB.GetObservation(id)
	if err != nil {
		writeError(w, fmt.Errorf("get observation: %v", err))
		return
	}
	writeJSON(w, http.StatusOK, o)
}

func (s *Server) publishObservation(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	o, err := s.app.Observation.Publish(id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, o)
}

func (s *Server) archiveObservation(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	o, err := s.app.Observation.Archive(id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, o)
}
