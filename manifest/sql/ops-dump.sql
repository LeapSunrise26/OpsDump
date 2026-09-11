-- ============================================================
-- ops-dump schema (GoFrame + SQLite / glebarez pure-Go driver)
-- ============================================================
-- 所有表名以 od_ 前缀开头，方便后续使用 `gf gen dao` 生成代码时
-- 通过 gfcli 的 database.prefix: "od_" 去掉前缀得到干净的实体名（Jobs 等）。
--
-- 生成数据库：
--   go run ./hack/gen-db            （读取本文件，生成 ./ops-dump.db）
--
-- 使用 gf gen dao 生成代码（需先安装 gfcli 并配置 hack/config.yaml）：
--   gf gen dao -cfg ./hack/config.yaml
-- 或手动指定链接：
--   gf gen dao -l "sqlite:./ops-dump.db" -p "od_" -path ./internal/dao
-- ============================================================

CREATE TABLE IF NOT EXISTS od_users (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    username      TEXT    NOT NULL UNIQUE,
    password_hash TEXT    NOT NULL,
    display_name  TEXT,
    enabled       INTEGER NOT NULL DEFAULT 1,
    created_at    TEXT    NOT NULL,
    updated_at    TEXT    NOT NULL
);

CREATE TABLE IF NOT EXISTS od_sessions (
    id         TEXT NOT NULL,
    username   TEXT NOT NULL,
    expires    TEXT NOT NULL,
    created_at TEXT NOT NULL,
    PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS od_jobs (
    id                 INTEGER PRIMARY KEY AUTOINCREMENT,
    name               TEXT    NOT NULL UNIQUE,
    kind               TEXT    NOT NULL,
    enabled            INTEGER NOT NULL DEFAULT 0,
    sort_order         INTEGER NOT NULL DEFAULT 0,
    cron               TEXT    NOT NULL,
    depends_on         TEXT,
    description        TEXT,
    container          TEXT,
    database           TEXT,
    command            TEXT,
    verify_command     TEXT,
    min_free_mb        INTEGER NOT NULL DEFAULT 2048,
    out_dir            TEXT,
    retention_days     INTEGER NOT NULL DEFAULT 7,
    kv                 TEXT,
    notify_on_success  INTEGER NOT NULL DEFAULT 0,
    notify_on_failure  INTEGER NOT NULL DEFAULT 1,
    created_at         TEXT    NOT NULL,
    updated_at         TEXT    NOT NULL
);

CREATE TABLE IF NOT EXISTS od_job_runs (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    job_id        INTEGER NOT NULL,
    status        TEXT    NOT NULL,
    started_at    TEXT    NOT NULL,
    finished_at   TEXT,
    elapsed_ms    INTEGER,
    exit_code     INTEGER,
    error         TEXT,
    command       TEXT,
    output        TEXT,
    verify_status TEXT,
    verify_output TEXT,
    size_mb       REAL,
    disk_free_mb  REAL,
    created_at    TEXT    NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_runs_job ON od_job_runs(job_id, started_at DESC);

CREATE TABLE IF NOT EXISTS od_notify_config (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    channel    TEXT    NOT NULL DEFAULT 'feishu',
    enabled    INTEGER NOT NULL DEFAULT 1,
    config     TEXT    NOT NULL,
    updated_at TEXT    NOT NULL
);
