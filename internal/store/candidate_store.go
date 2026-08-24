package store

import (
	"time"

	"task212-specline/internal/model"
)

// InsertCandidate 插入归属候选，返回 ID。
func (db *DB) InsertCandidate(c *model.AttributionCandidate) (int64, error) {
	res, err := db.conn.Exec(
		`INSERT INTO attribution_candidates (observation_id, peak_id, transition_key, element, ionization, rest_wl, corrected_wl, residual, tolerance, score, status, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		c.ObservationID, c.PeakID, c.TransitionKey, c.Element, c.Ionization, c.RestWL,
		c.CorrectedWL, c.Residual, c.Tolerance, c.Score, c.Status, nowText(),
	)
	if err != nil {
		return 0, mapSQLError(err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, mapSQLError(err)
	}
	c.ID = id
	c.CreatedAt = time.Now().UTC()
	return id, nil
}

// GetCandidate 按 ID 读取候选。
func (db *DB) GetCandidate(id int64) (*model.AttributionCandidate, error) {
	var c model.AttributionCandidate
	var createdAt string
	err := db.conn.QueryRow(
		`SELECT id, observation_id, peak_id, transition_key, element, ionization, rest_wl, corrected_wl, residual, tolerance, score, status, created_at
		 FROM attribution_candidates WHERE id = ?`, id,
	).Scan(&c.ID, &c.ObservationID, &c.PeakID, &c.TransitionKey, &c.Element, &c.Ionization,
		&c.RestWL, &c.CorrectedWL, &c.Residual, &c.Tolerance, &c.Score, &c.Status, &createdAt)
	if err != nil {
		return nil, mapSQLError(err)
	}
	c.CreatedAt, _ = parseTime(createdAt)
	return &c, nil
}

// ListCandidates 按观测集列出全部候选。
func (db *DB) ListCandidates(observationID int64) ([]*model.AttributionCandidate, error) {
	rows, err := db.conn.Query(
		`SELECT id, observation_id, peak_id, transition_key, element, ionization, rest_wl, corrected_wl, residual, tolerance, score, status, created_at
		 FROM attribution_candidates WHERE observation_id = ? ORDER BY id ASC`, observationID)
	if err != nil {
		return nil, mapSQLError(err)
	}
	defer rows.Close()
	var out []*model.AttributionCandidate
	for rows.Next() {
		var c model.AttributionCandidate
		var createdAt string
		if err := rows.Scan(&c.ID, &c.ObservationID, &c.PeakID, &c.TransitionKey, &c.Element, &c.Ionization,
			&c.RestWL, &c.CorrectedWL, &c.Residual, &c.Tolerance, &c.Score, &c.Status, &createdAt); err != nil {
			return nil, mapSQLError(err)
		}
		c.CreatedAt, _ = parseTime(createdAt)
		out = append(out, &c)
	}
	return out, rows.Err()
}

// ListCandidatesByPeak 按峰列出候选（互斥判定用）。
func (db *DB) ListCandidatesByPeak(peakID int64) ([]*model.AttributionCandidate, error) {
	rows, err := db.conn.Query(
		`SELECT id, observation_id, peak_id, transition_key, element, ionization, rest_wl, corrected_wl, residual, tolerance, score, status, created_at
		 FROM attribution_candidates WHERE peak_id = ? ORDER BY id ASC`, peakID)
	if err != nil {
		return nil, mapSQLError(err)
	}
	defer rows.Close()
	var out []*model.AttributionCandidate
	for rows.Next() {
		var c model.AttributionCandidate
		var createdAt string
		if err := rows.Scan(&c.ID, &c.ObservationID, &c.PeakID, &c.TransitionKey, &c.Element, &c.Ionization,
			&c.RestWL, &c.CorrectedWL, &c.Residual, &c.Tolerance, &c.Score, &c.Status, &createdAt); err != nil {
			return nil, mapSQLError(err)
		}
		c.CreatedAt, _ = parseTime(createdAt)
		out = append(out, &c)
	}
	return out, rows.Err()
}

// UpdateCandidateStatus 更新候选状态。
func (db *DB) UpdateCandidateStatus(id int64, status string) error {
	_, err := db.conn.Exec(`UPDATE attribution_candidates SET status = ? WHERE id = ?`, status, id)
	return mapSQLError(err)
}

// UpdateCandidateScore 更新候选评分（先验调整后重算）。
func (db *DB) UpdateCandidateScore(id int64, score float64) error {
	_, err := db.conn.Exec(`UPDATE attribution_candidates SET score = ? WHERE id = ?`, score, id)
	return mapSQLError(err)
}

// DeleteCandidatesForPeak 删除某峰的全部候选（伪线排除时）。
func (db *DB) DeleteCandidatesForPeak(peakID int64) error {
	_, err := db.conn.Exec(`DELETE FROM attribution_candidates WHERE peak_id = ?`, peakID)
	return mapSQLError(err)
}

// DeleteCandidatesForObservation 删除某观测集全部候选（重建匹配前调用）。
func (db *DB) DeleteCandidatesForObservation(observationID int64) error {
	_, err := db.conn.Exec(`DELETE FROM attribution_candidates WHERE observation_id = ?`, observationID)
	return mapSQLError(err)
}

// CountCandidates 统计某观测集候选数。
func (db *DB) CountCandidates(observationID int64) (int, error) {
	var n int
	err := db.conn.QueryRow(`SELECT COUNT(*) FROM attribution_candidates WHERE observation_id = ?`, observationID).Scan(&n)
	return n, mapSQLError(err)
}
