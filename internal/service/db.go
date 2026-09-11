package service

import (
	"context"
	"os"
	"path/filepath"
	"time"

	_ "github.com/gogf/gf/contrib/drivers/sqlite/v2"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"

	"ops-dump/internal/model"
)

// Cfg holds the process configuration (loaded at startup).
var Cfg *model.AppConfig

// Env holds resolved runtime variables from the .env file.
var Env *EnvMap

var schemaStatements = []string{
	`CREATE TABLE IF NOT EXISTS od_users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		display_name TEXT,
		enabled INTEGER NOT NULL DEFAULT 1,
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL
	)`,
	`CREATE TABLE IF NOT EXISTS od_sessions (
		id TEXT PRIMARY KEY,
		username TEXT NOT NULL,
		expires TEXT NOT NULL,
		created_at TEXT NOT NULL
	)`,
	`CREATE TABLE IF NOT EXISTS od_jobs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		kind TEXT NOT NULL,
		enabled INTEGER NOT NULL DEFAULT 0,
		sort_order INTEGER NOT NULL DEFAULT 0,
		cron TEXT NOT NULL,
		depends_on TEXT,
		description TEXT,
		container TEXT,
		database TEXT,
		command TEXT,
		verify_command TEXT,
		min_free_mb INTEGER NOT NULL DEFAULT 2048,
		out_dir TEXT,
		retention_days INTEGER NOT NULL DEFAULT 7,
		kv TEXT,
		notify_on_success INTEGER NOT NULL DEFAULT 0,
		notify_on_failure INTEGER NOT NULL DEFAULT 1,
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL
	)`,
	`CREATE TABLE IF NOT EXISTS od_job_runs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		job_id INTEGER NOT NULL,
		status TEXT NOT NULL,
		started_at TEXT NOT NULL,
		finished_at TEXT,
		elapsed_ms INTEGER,
		exit_code INTEGER,
		error TEXT,
		command TEXT,
		output TEXT,
		verify_status TEXT,
		verify_output TEXT,
		size_mb REAL,
		disk_free_mb REAL,
		created_at TEXT NOT NULL
	)`,
	`CREATE INDEX IF NOT EXISTS idx_runs_job ON od_job_runs(job_id, started_at DESC)`,
	`CREATE TABLE IF NOT EXISTS od_notify_config (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		channel TEXT NOT NULL DEFAULT 'feishu',
		enabled INTEGER NOT NULL DEFAULT 1,
		config TEXT NOT NULL,
		updated_at TEXT NOT NULL
	)`,
}

// dataNoId converts a struct to a g.Map suitable for insert/update, stripping the
// primary key so the database can autoincrement it and so updates don't overwrite
// the id column.
func dataNoId(v any) g.Map {
	m := gconv.Map(v)
	delete(m, "id")
	return m
}

// BootService initializes the database, migrations and seeds.
type BootService struct{}

// NewBoot creates a BootService instance.
func NewBoot() *BootService { return &BootService{} }

// Init initializes the database, migrations and seeds. It must be called before
// any service is used.
func (s *BootService) Init(ctx context.Context, cfg *model.AppConfig) error {
	Cfg = cfg
	Env = NewEnv(cfg.Credentials.EnvFile)

	if dir := filepath.Dir(cfg.DbPath); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	g.Log().Info(ctx, "dbPath=", cfg.DbPath, " envFile=", cfg.Credentials.EnvFile)
	gdb.AddConfigNode(gdb.DefaultGroupName, gdb.ConfigNode{
		Type: "sqlite",
		Name: cfg.DbPath,
	})
	// Ensure WAL for concurrent read/write safety.
	if _, err := g.DB().Exec(ctx, "PRAGMA journal_mode=WAL;"); err != nil {
		return err
	}
	for _, stmt := range schemaStatements {
		if _, err := g.DB().Exec(ctx, stmt); err != nil {
			return err
		}
	}
	// Migration: rename legacy run_after → depends_on on pre-existing databases
	// (CREATE TABLE IF NOT EXISTS keeps old columns untouched).
	if hasColumn(ctx, "od_jobs", "run_after") {
		if _, err := g.DB().Exec(ctx, "ALTER TABLE od_jobs RENAME COLUMN run_after TO depends_on"); err != nil {
			return err
		}
	}
	if !hasColumn(ctx, "od_jobs", "sort_order") {
		if _, err := g.DB().Exec(ctx, "ALTER TABLE od_jobs ADD COLUMN sort_order INTEGER NOT NULL DEFAULT 0"); err != nil {
			return err
		}
	}
	if !hasColumn(ctx, "od_jobs", "notify_on_success") {
		if _, err := g.DB().Exec(ctx, "ALTER TABLE od_jobs ADD COLUMN notify_on_success INTEGER NOT NULL DEFAULT 0"); err != nil {
			return err
		}
	}
	if !hasColumn(ctx, "od_jobs", "notify_on_failure") {
		if _, err := g.DB().Exec(ctx, "ALTER TABLE od_jobs ADD COLUMN notify_on_failure INTEGER NOT NULL DEFAULT 1"); err != nil {
			return err
		}
	}
	if !hasColumn(ctx, "od_job_runs", "command") {
		if _, err := g.DB().Exec(ctx, "ALTER TABLE od_job_runs ADD COLUMN command TEXT"); err != nil {
			return err
		}
	}
	if err := NewAuth().SeedAdmin(ctx); err != nil {
		return err
	}
	if err := NewTemplate().SeedTemplates(ctx); err != nil {
		return err
	}
	return nil
}

// hasColumn reports whether the given table has a column with the given name.
func hasColumn(ctx context.Context, table, column string) bool {
	rows, err := g.DB().GetAll(ctx, "PRAGMA table_info('"+table+"')")
	if err != nil {
		return false
	}
	for _, r := range rows {
		if r["name"].String() == column {
			return true
		}
	}
	return false
}

// now returns the current time as a display-friendly local string.
// Note: gtime.Time.Format() returns the Go layout string verbatim when given
// "2006-01-02 15:04:05" (gtime uses PHP-style format tokens), so the standard
// library is used here.
func now() string {
	return time.Now().Format("2006-01-02 15:04:05")
}
