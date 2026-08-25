package store

import (
	"time"

	"task212-specline/internal/model"
)

// InsertCalibration 保存一次校准结果，返回 ID。
func (db *DB) InsertCalibration(c *model.Calibration) (int64, error) {
	res, err := db.conn.Exec(
		`INSERT INTO calibrations (observation_id, model, offset, slope, reference_wl, residual_rms, reference_count, status, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		c.ObservationID, c.Model, c.Offset, c.Slope, c.ReferenceWL, c.ResidualRMS, c.ReferenceCount, c.Status, nowText(),
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

// GetCalibration 读取观测集最近一次校准结果。
func (db *DB) GetCalibration(observationID int64) (*model.Calibration, error) {
	var c model.Calibration
	var createdAt string
	err := db.conn.QueryRow(
		`SELECT id, observation_id, model, offset, slope, reference_wl, residual_rms, reference_count, status, created_at
		 FROM calibrations WHERE observation_id = ? ORDER BY id DESC LIMIT 1`, observationID,
	).Scan(&c.ID, &c.ObservationID, &c.Model, &c.Offset, &c.Slope, &c.ReferenceWL,
		&c.ResidualRMS, &c.ReferenceCount, &c.Status, &createdAt)
	if err != nil {
		return nil, mapSQLError(err)
	}
	c.CreatedAt, _ = parseTime(createdAt)
	return &c, nil
}

// SupersedeCalibrations 把观测集历史校准标记为已替代（迟到校准场景）。
func (db *DB) SupersedeCalibrations(observationID int64, exceptID int64) error {
	_, err := db.conn.Exec(
		`UPDATE calibrations SET status = ? WHERE observation_id = ? AND id != ?`,
		model.CalibSuperseded, observationID, exceptID)
	return mapSQLError(err)
}
