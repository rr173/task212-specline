// Package service 编排各业务域，提供统一的应用入口。
package service

import (
	"task212-specline/internal/calibration"
	"task212-specline/internal/matching"
	"task212-specline/internal/observation"
	"task212-specline/internal/review"
	"task212-specline/internal/store"
	"task212-specline/internal/versioning"
)

// App 应用编排根：持有各业务域服务与持久化连接。
type App struct {
	DB          *store.DB
	Observation *observation.Service
	Calibration *calibration.Service
	Matching    *matching.Service
	Review      *review.Service
	Versioning  *versioning.Service
}

// New 组装应用（各域服务共享同一 SQLite 连接）。
func New(db *store.DB) *App {
	obs := observation.New(db)
	cal := calibration.New(db)
	mat := matching.New(db)
	rev := review.New(db)
	ver := versioning.New(db, rev)
	return &App{
		DB:          db,
		Observation: obs,
		Calibration: cal,
		Matching:    mat,
		Review:      rev,
		Versioning:  ver,
	}
}
