package store

import (
	"database/sql"
	"time"

	"task212-specline/internal/model"
)

// InsertVersion 创建归属版本，返回 ID。
func (db *DB) InsertVersion(v *model.AttributionVersion) (int64, error) {
	res, err := db.conn.Exec(
		`INSERT INTO attribution_versions (observation_id, label, status, content_hash, prior_hash, transition_lib_version, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		v.ObservationID, v.Label, v.Status, v.ContentHash, v.PriorHash, v.TransitionLibVersion, nowText(),
	)
	if err != nil {
		return 0, mapSQLError(err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, mapSQLError(err)
	}
	v.ID = id
	v.CreatedAt = time.Now().UTC()
	return id, nil
}

// GetVersion 按 ID 读取归属版本。
func (db *DB) GetVersion(id int64) (*model.AttributionVersion, error) {
	var v model.AttributionVersion
	var createdAt, frozenAt, supersededAt sql.NullString
	err := db.conn.QueryRow(
		`SELECT id, observation_id, label, status, content_hash, prior_hash, transition_lib_version, created_at, frozen_at, superseded_at, superseded_by
		 FROM attribution_versions WHERE id = ?`, id,
	).Scan(&v.ID, &v.ObservationID, &v.Label, &v.Status, &v.ContentHash, &v.PriorHash,
		&v.TransitionLibVersion, &createdAt, &frozenAt, &supersededAt, &v.SupersededBy)
	if err != nil {
		return nil, mapSQLError(err)
	}
	if createdAt.Valid {
		if t, e := parseTime(createdAt.String); e == nil {
			v.CreatedAt = t
		}
	}
	if frozenAt.Valid {
		if t, e := parseTime(frozenAt.String); e == nil {
			v.FrozenAt = &t
		}
	}
	if supersededAt.Valid {
		if t, e := parseTime(supersededAt.String); e == nil {
			v.SupersededAt = &t
		}
	}
	return &v, nil
}

// ListVersions 按观测集列出全部版本（按创建时间倒序）。
func (db *DB) ListVersions(observationID int64) ([]*model.AttributionVersion, error) {
	rows, err := db.conn.Query(
		`SELECT id, observation_id, label, status, content_hash, prior_hash, transition_lib_version, created_at, frozen_at, superseded_at, superseded_by
		 FROM attribution_versions WHERE observation_id = ? ORDER BY id DESC`, observationID)
	if err != nil {
		return nil, mapSQLError(err)
	}
	defer rows.Close()
	var out []*model.AttributionVersion
	for rows.Next() {
		var v model.AttributionVersion
		var createdAt, frozenAt, supersededAt sql.NullString
		if err := rows.Scan(&v.ID, &v.ObservationID, &v.Label, &v.Status, &v.ContentHash, &v.PriorHash,
			&v.TransitionLibVersion, &createdAt, &frozenAt, &supersededAt, &v.SupersededBy); err != nil {
			return nil, mapSQLError(err)
		}
		if createdAt.Valid {
			if t, e := parseTime(createdAt.String); e == nil {
				v.CreatedAt = t
			}
		}
		if frozenAt.Valid {
			if t, e := parseTime(frozenAt.String); e == nil {
				v.FrozenAt = &t
			}
		}
		if supersededAt.Valid {
			if t, e := parseTime(supersededAt.String); e == nil {
				v.SupersededAt = &t
			}
		}
		out = append(out, &v)
	}
	return out, rows.Err()
}

// UpdateVersionStatus 更新版本状态。
func (db *DB) UpdateVersionStatus(id int64, status string) error {
	_, err := db.conn.Exec(`UPDATE attribution_versions SET status = ? WHERE id = ?`, status, id)
	return mapSQLError(err)
}

// MarkVersionFrozen 冻结版本（写入冻结时间）。
func (db *DB) MarkVersionFrozen(id int64) error {
	_, err := db.conn.Exec(`UPDATE attribution_versions SET status = ?, frozen_at = ? WHERE id = ?`,
		model.VersionFrozen, nowText(), id)
	return mapSQLError(err)
}

// MarkVersionSuperseded 标记版本被替代（写入替代时间与被替代者 ID）。
func (db *DB) MarkVersionSuperseded(id int64, byID int64) error {
	_, err := db.conn.Exec(`UPDATE attribution_versions SET status = ?, superseded_at = ?, superseded_by = ? WHERE id = ?`,
		model.VersionSuperseded, nowText(), byID, id)
	return mapSQLError(err)
}

// ActiveVersionOf 返回观测集当前活动版本（非 superseded 的最新一条）。
func (db *DB) ActiveVersionOf(observationID int64) (*model.AttributionVersion, error) {
	var v model.AttributionVersion
	var createdAt, frozenAt, supersededAt sql.NullString
	err := db.conn.QueryRow(
		`SELECT id, observation_id, label, status, content_hash, prior_hash, transition_lib_version, created_at, frozen_at, superseded_at, superseded_by
		 FROM attribution_versions WHERE observation_id = ? AND status != ? ORDER BY id DESC LIMIT 1`,
		observationID, model.VersionSuperseded,
	).Scan(&v.ID, &v.ObservationID, &v.Label, &v.Status, &v.ContentHash, &v.PriorHash,
		&v.TransitionLibVersion, &createdAt, &frozenAt, &supersededAt, &v.SupersededBy)
	if err != nil {
		return nil, mapSQLError(err)
	}
	if createdAt.Valid {
		if t, e := parseTime(createdAt.String); e == nil {
			v.CreatedAt = t
		}
	}
	if frozenAt.Valid {
		if t, e := parseTime(frozenAt.String); e == nil {
			v.FrozenAt = &t
		}
	}
	if supersededAt.Valid {
		if t, e := parseTime(supersededAt.String); e == nil {
			v.SupersededAt = &t
		}
	}
	return &v, nil
}
