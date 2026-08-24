package store

// UpsertPriorOverride 写入或更新某观测集对某跃迁的先验覆盖值。
func (db *DB) UpsertPriorOverride(observationID int64, transitionKey string, prior float64) error {
	_, err := db.conn.Exec(
		`INSERT INTO prior_overrides (observation_id, transition_key, prior) VALUES (?, ?, ?)
		 ON CONFLICT(observation_id, transition_key) DO UPDATE SET prior = excluded.prior`,
		observationID, transitionKey, prior)
	return mapSQLError(err)
}

// PriorOverrides 返回观测集的全部先验覆盖（键 → 先验值）。
func (db *DB) PriorOverrides(observationID int64) (map[string]float64, error) {
	rows, err := db.conn.Query(
		`SELECT transition_key, prior FROM prior_overrides WHERE observation_id = ?`, observationID)
	if err != nil {
		return nil, mapSQLError(err)
	}
	defer rows.Close()
	out := map[string]float64{}
	for rows.Next() {
		var key string
		var prior float64
		if err := rows.Scan(&key, &prior); err != nil {
			return nil, mapSQLError(err)
		}
		out[key] = prior
	}
	return out, rows.Err()
}
