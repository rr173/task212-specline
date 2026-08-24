package store

import (
	"database/sql"
	"time"

	"task212-specline/internal/model"
)

// InsertObservation 插入新观测集，返回其 ID。
func (db *DB) InsertObservation(o *model.ObservationSet) (int64, error) {
	res, err := db.conn.Exec(
		`INSERT INTO observation_sets (name, target, wavelength_unit, status, content_hash, created_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		o.Name, o.Target, o.WavelengthUnit, o.Status, o.ContentHash, nowText(),
	)
	if err != nil {
		return 0, mapSQLError(err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, mapSQLError(err)
	}
	o.ID = id
	o.CreatedAt = time.Now().UTC()
	return id, nil
}

// GetObservation 按 ID 读取观测集。
func (db *DB) GetObservation(id int64) (*model.ObservationSet, error) {
	var o model.ObservationSet
	var createdAt, publishedAt, archivedAt sql.NullString
	err := db.conn.QueryRow(
		`SELECT id, name, target, wavelength_unit, status, content_hash, created_at, published_at, archived_at
		 FROM observation_sets WHERE id = ?`, id,
	).Scan(&o.ID, &o.Name, &o.Target, &o.WavelengthUnit, &o.Status, &o.ContentHash,
		&createdAt, &publishedAt, &archivedAt)
	if err != nil {
		return nil, mapSQLError(err)
	}
	if createdAt.Valid {
		if t, e := parseTime(createdAt.String); e == nil {
			o.CreatedAt = t
		}
	}
	if publishedAt.Valid {
		if t, e := parseTime(publishedAt.String); e == nil {
			o.PublishedAt = &t
		}
	}
	if archivedAt.Valid {
		if t, e := parseTime(archivedAt.String); e == nil {
			o.ArchivedAt = &t
		}
	}
	return &o, nil
}

// ListObservations 列出全部观测集（按创建时间倒序）。
func (db *DB) ListObservations() ([]*model.ObservationSet, error) {
	rows, err := db.conn.Query(
		`SELECT id, name, target, wavelength_unit, status, content_hash, created_at, published_at, archived_at
		 FROM observation_sets ORDER BY id DESC`)
	if err != nil {
		return nil, mapSQLError(err)
	}
	defer rows.Close()
	var out []*model.ObservationSet
	for rows.Next() {
		var o model.ObservationSet
		var createdAt, publishedAt, archivedAt sql.NullString
		if err := rows.Scan(&o.ID, &o.Name, &o.Target, &o.WavelengthUnit, &o.Status, &o.ContentHash,
			&createdAt, &publishedAt, &archivedAt); err != nil {
			return nil, mapSQLError(err)
		}
		if createdAt.Valid {
			if t, e := parseTime(createdAt.String); e == nil {
				o.CreatedAt = t
			}
		}
		if publishedAt.Valid {
			if t, e := parseTime(publishedAt.String); e == nil {
				o.PublishedAt = &t
			}
		}
		if archivedAt.Valid {
			if t, e := parseTime(archivedAt.String); e == nil {
				o.ArchivedAt = &t
			}
		}
		out = append(out, &o)
	}
	return out, rows.Err()
}

// UpdateObservationStatus 更新观测集状态（受调用方状态机约束）。
func (db *DB) UpdateObservationStatus(id int64, status string) error {
	_, err := db.conn.Exec(`UPDATE observation_sets SET status = ? WHERE id = ?`, status, id)
	return mapSQLError(err)
}

// UpdateObservationContentHash 更新观测集内容哈希（峰序列变更后）。
func (db *DB) UpdateObservationContentHash(id int64, hash string) error {
	_, err := db.conn.Exec(`UPDATE observation_sets SET content_hash = ? WHERE id = ?`, hash, id)
	return mapSQLError(err)
}

// MarkObservationPublished 标记观测集已发布。
func (db *DB) MarkObservationPublished(id int64) error {
	_, err := db.conn.Exec(`UPDATE observation_sets SET status = ?, published_at = ? WHERE id = ?`,
		model.ObsPublished, nowText(), id)
	return mapSQLError(err)
}

// MarkObservationArchived 标记观测集已封存（只读）。
func (db *DB) MarkObservationArchived(id int64) error {
	_, err := db.conn.Exec(`UPDATE observation_sets SET status = ?, archived_at = ? WHERE id = ?`,
		model.ObsArchived, nowText(), id)
	return mapSQLError(err)
}

// CountObservations 统计观测集数量。
func (db *DB) CountObservations() (int, error) {
	var n int
	err := db.conn.QueryRow(`SELECT COUNT(*) FROM observation_sets`).Scan(&n)
	return n, mapSQLError(err)
}
