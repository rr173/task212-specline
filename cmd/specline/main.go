// Command specline 是天文光谱线归属复核台的服务入口。
//
// 用法：
//
//	specline --addr :8080 --db specline.db      # 启动 HTTP 服务
//	specline --smoke-test [--db smoke.db]       # 端到端自检（Docker CMD 判据）
package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"task212-specline/internal/httpapi"
	"task212-specline/internal/service"
	"task212-specline/internal/smoke"
	"task212-specline/internal/store"
)

func main() {
	addr := flag.String("addr", ":8080", "HTTP listen address")
	dbPath := flag.String("db", "specline.db", "SQLite database path")
	smokeMode := flag.Bool("smoke-test", false, "run end-to-end smoke test and exit")
	flag.Parse()

	if *smokeMode {
		smoke.Main([]string{*dbPath})
		return
	}

	db, err := store.Open(*dbPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	app := service.New(db)
	srv := httpapi.New(app)

	log.Printf("spectral line attribution review listening on %s (db=%s)", *addr, *dbPath)
	if err := http.ListenAndServe(*addr, srv.Handler()); err != nil {
		log.Fatalf("serve: %v", err)
		os.Exit(1)
	}
}
