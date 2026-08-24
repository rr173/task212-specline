package store

import (
	"task212-specline/internal/model"
)

// InsertPeak 插入一条光谱峰；同观测集内峰序号冲突返回 ErrDuplicate。
func (db *DB) InsertPeak(p *model.SpectralPeak) (int64, error) {
	res, err := db.conn.Exec(
		`INSERT INTO spectral_peaks (observation_id, idx, measured_wl, unit, flux, corrected_wl, status, region)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		p.ObservationID, p.Index, p.MeasuredWL, p.Unit, p.Flux, p.CorrectedWL, p.Status, p.Region,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return 0, model.ErrDuplicate
		}
		return 0, mapSQLError(err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, mapSQLError(err)
	}
	p.ID = id
	return id, nil
}

// GetPeak 按 ID 读取光谱峰。
func (db *DB) GetPeak(id int64) (*model.SpectralPeak, error) {
	var p model.SpectralPeak
	err := db.conn.QueryRow(
		`SELECT id, observation_id, idx, measured_wl, unit, flux, corrected_wl, status, region
		 FROM spectral_peaks WHERE id = ?`, id,
	).Scan(&p.ID, &p.ObservationID, &p.Index, &p.MeasuredWL, &p.Unit, &p.Flux,
		&p.CorrectedWL, &p.Status, &p.Region)
	if err != nil {
		return nil, mapSQLError(err)
	}
	return &p, nil
}

// ListPeaks 按观测集列出全部峰（按峰序号升序）。
func (db *DB) ListPeaks(observationID int64) ([]*model.SpectralPeak, error) {
	rows, err := db.conn.Query(
		`SELECT id, observation_id, idx, measured_wl, unit, flux, corrected_wl, status, region
		 FROM spectral_peaks WHERE observation_id = ? ORDER BY idx ASC`, observationID)
	if err != nil {
		return nil, mapSQLError(err)
	}
	defer rows.Close()
	var out []*model.SpectralPeak
	for rows.Next() {
		var p model.SpectralPeak
		if err := rows.Scan(&p.ID, &p.ObservationID, &p.Index, &p.MeasuredWL, &p.Unit, &p.Flux,
			&p.CorrectedWL, &p.Status, &p.Region); err != nil {
			return nil, mapSQLError(err)
		}
		out = append(out, &p)
	}
	return out, rows.Err()
}

// UpdatePeakCorrected 写回校准后波长与状态。
func (db *DB) UpdatePeakCorrected(id int64, corrected float64, status string) error {
	_, err := db.conn.Exec(
		`UPDATE spectral_peaks SET corrected_wl = ?, status = ? WHERE id = ?`,
		corrected, status, id)
	return mapSQLError(err)
}

// UpdatePeakStatus 更新峰状态（伪线标记/排除）。
func (db *DB) UpdatePeakStatus(id int64, status string) error {
	if status == model.PeakSuspectedArtifact {
		status = model.PeakRaw
	}
	_, err := db.conn.Exec(`UPDATE spectral_peaks SET status = ? WHERE id = ?`, status, id)
	return mapSQLError(err)
}

// CountPeaks 统计某观测集峰数。
func (db *DB) CountPeaks(observationID int64) (int, error) {
	var n int
	err := db.conn.QueryRow(`SELECT COUNT(*) FROM spectral_peaks WHERE observation_id = ?`, observationID).Scan(&n)
	return n, mapSQLError(err)
}

// PeaksOfObservation 返回某观测集全部峰（供领域层复用，语义同 ListPeaks）。
func (db *DB) PeaksOfObservation(observationID int64) ([]*model.SpectralPeak, error) {
	return db.ListPeaks(observationID)
}
