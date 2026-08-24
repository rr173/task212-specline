// Package store 提供基于 SQLite 的持久化层（modernc.org/sqlite，纯 Go 无 CGO）。
package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// DB 封装数据库连接与建表迁移。
type DB struct {
	conn *sql.DB
	path string
}

// Open 打开（必要时创建）SQLite 数据库并执行建表迁移。
func Open(path string) (*DB, error) {
	if path == "" {
		path = filepath.Join(".", "specline.db")
	}
	if dir := filepath.Dir(path); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("create db dir: %w", err)
		}
	}
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	conn.SetMaxOpenConns(1) // 单写者，规避 SQLite 写锁竞争
	db := &DB{conn: conn, path: path}
	if err := db.migrate(); err != nil {
		conn.Close()
		return nil, err
	}
	return db, nil
}

// Close 关闭数据库连接。
func (db *DB) Close() error { return db.conn.Close() }

// Path 返回数据库文件路径。
func (db *DB) Path() string { return db.path }

// Conn 暴露底层连接，供需要事务或裸查询的场景使用。
func (db *DB) Conn() *sql.DB { return db.conn }

// migrate 执行全部建表语句（幂等）。
func (db *DB) migrate() error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS observation_sets (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			target TEXT NOT NULL DEFAULT '',
			wavelength_unit TEXT NOT NULL DEFAULT 'angstrom',
			status TEXT NOT NULL,
			content_hash TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL,
			published_at TEXT,
			archived_at TEXT
		)`,
		`CREATE INDEX IF NOT EXISTS idx_obs_status ON observation_sets(status)`,
		`CREATE TABLE IF NOT EXISTS spectral_peaks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			observation_id INTEGER NOT NULL,
			idx INTEGER NOT NULL,
			measured_wl REAL NOT NULL,
			unit TEXT NOT NULL DEFAULT 'angstrom',
			flux REAL NOT NULL DEFAULT 0,
			corrected_wl REAL NOT NULL DEFAULT 0,
			status TEXT NOT NULL,
			region TEXT NOT NULL DEFAULT '',
			UNIQUE(observation_id, idx),
			FOREIGN KEY (observation_id) REFERENCES observation_sets(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_peaks_obs ON spectral_peaks(observation_id)`,
		`CREATE TABLE IF NOT EXISTS calibrations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			observation_id INTEGER NOT NULL,
			model TEXT NOT NULL,
			offset REAL NOT NULL DEFAULT 0,
			slope REAL NOT NULL DEFAULT 0,
			reference_wl REAL NOT NULL DEFAULT 0,
			residual_rms REAL NOT NULL DEFAULT 0,
			reference_count INTEGER NOT NULL DEFAULT 0,
			status TEXT NOT NULL,
			created_at TEXT NOT NULL,
			FOREIGN KEY (observation_id) REFERENCES observation_sets(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_calib_obs ON calibrations(observation_id)`,
		`CREATE TABLE IF NOT EXISTS attribution_candidates (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			observation_id INTEGER NOT NULL,
			peak_id INTEGER NOT NULL,
			transition_key TEXT NOT NULL,
			element TEXT NOT NULL,
			ionization TEXT NOT NULL,
			rest_wl REAL NOT NULL,
			corrected_wl REAL NOT NULL,
			residual REAL NOT NULL,
			tolerance REAL NOT NULL,
			score REAL NOT NULL DEFAULT 0,
			status TEXT NOT NULL,
			created_at TEXT NOT NULL,
			FOREIGN KEY (observation_id) REFERENCES observation_sets(id),
			FOREIGN KEY (peak_id) REFERENCES spectral_peaks(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_candidates_obs ON attribution_candidates(observation_id)`,
		`CREATE INDEX IF NOT EXISTS idx_candidates_peak ON attribution_candidates(peak_id)`,
		`CREATE TABLE IF NOT EXISTS refutations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			candidate_id INTEGER NOT NULL,
			kind TEXT NOT NULL,
			note TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL,
			FOREIGN KEY (candidate_id) REFERENCES attribution_candidates(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_refutations_cand ON refutations(candidate_id)`,
		`CREATE TABLE IF NOT EXISTS attribution_versions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			observation_id INTEGER NOT NULL,
			label TEXT NOT NULL,
			status TEXT NOT NULL,
			content_hash TEXT NOT NULL,
			prior_hash TEXT NOT NULL,
			transition_lib_version TEXT NOT NULL,
			created_at TEXT NOT NULL,
			frozen_at TEXT,
			superseded_at TEXT,
			superseded_by INTEGER NOT NULL DEFAULT 0,
			FOREIGN KEY (observation_id) REFERENCES observation_sets(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_versions_obs ON attribution_versions(observation_id)`,
		`CREATE TABLE IF NOT EXISTS prior_overrides (
			observation_id INTEGER NOT NULL,
			transition_key TEXT NOT NULL,
			prior REAL NOT NULL,
			PRIMARY KEY (observation_id, transition_key),
			FOREIGN KEY (observation_id) REFERENCES observation_sets(id)
		)`,
	}
	for _, s := range stmts {
		if _, err := db.conn.Exec(s); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
	}
	return nil
}

// formatTime 将时间序列化为可存储的 RFC3339Nano 文本。
func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339Nano)
}

// parseTime 解析存储的时间文本；空串返回零值。
func parseTime(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	return time.Parse(time.RFC3339Nano, s)
}

// nowText 返回当前时间的 RFC3339Nano 文本。
func nowText() string { return time.Now().UTC().Format(time.RFC3339Nano) }
