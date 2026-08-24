package service

import (
	"task212-specline/internal/matching"
	"task212-specline/internal/model"
)

// Stats 系统统计信息。
type Stats struct {
	Observations    int `json:"observations"`
	Peaks           int `json:"peaks"`
	Candidates      int `json:"candidates"`
	Versions        int `json:"versions"`
	TransitionLines int `json:"transition_lines"`
	LibVersion      string `json:"transition_lib_version"`
}

// Stats 计算系统统计。
func (a *App) Stats() (*Stats, error) {
	obs, err := a.DB.CountObservations()
	if err != nil {
		return nil, err
	}
	st := &Stats{
		Observations:    obs,
		TransitionLines: matching.LibrarySize(),
		LibVersion:      model.TransitionLibVersion,
	}
	// 峰数、候选数、版本数按全库聚合。
	obsList, err := a.DB.ListObservations()
	if err != nil {
		return nil, err
	}
	for _, o := range obsList {
		if n, e := a.DB.CountPeaks(o.ID); e == nil {
			st.Peaks += n
		}
		if n, e := a.DB.CountCandidates(o.ID); e == nil {
			st.Candidates += n
		}
		if vs, e := a.DB.ListVersions(o.ID); e == nil {
			st.Versions += len(vs)
		}
	}
	return st, nil
}
