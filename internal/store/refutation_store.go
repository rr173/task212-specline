package store

import (
	"time"

	"task212-specline/internal/model"
)

// InsertRefutation 保存一条反证，返回 ID。
func (db *DB) InsertRefutation(r *model.Refutation) (int64, error) {
	res, err := db.conn.Exec(
		`INSERT INTO refutations (candidate_id, kind, note, created_at) VALUES (?, ?, ?, ?)`,
		r.CandidateID, r.Kind, r.Note, nowText(),
	)
	if err != nil {
		return 0, mapSQLError(err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, mapSQLError(err)
	}
	r.ID = id
	r.CreatedAt = time.Now().UTC()
	return id, nil
}

// ListRefutations 列出某候选的全部反证。
func (db *DB) ListRefutations(candidateID int64) ([]*model.Refutation, error) {
	rows, err := db.conn.Query(
		`SELECT id, candidate_id, kind, note, created_at FROM refutations WHERE candidate_id = ? ORDER BY id ASC`,
		candidateID)
	if err != nil {
		return nil, mapSQLError(err)
	}
	defer rows.Close()
	var out []*model.Refutation
	for rows.Next() {
		var r model.Refutation
		var createdAt string
		if err := rows.Scan(&r.ID, &r.CandidateID, &r.Kind, &r.Note, &createdAt); err != nil {
			return nil, mapSQLError(err)
		}
		r.CreatedAt, _ = parseTime(createdAt)
		out = append(out, &r)
	}
	return out, rows.Err()
}
