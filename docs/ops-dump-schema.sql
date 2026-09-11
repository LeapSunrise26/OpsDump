-- ============================================================
-- ops-dump 数据库 Schema (SQLite)
-- ============================================================
-- 版本：v0.1
-- 说明：本文件为 ops-dump 的完整数据库建表脚本，包含所有业务表。
--       所有表名以 od_ 前缀开头。
-- 
-- 使用方式：
--   1. 直接执行：sqlite3 ops-dump.db < ops-dump-schema.sql
--   2. 代码生成：配合 hack/config.yaml 使用 gf gen dao 生成 Go 代码
--   3. 程序自动建表：程序首次启动会自动执行建表迁移
-- ============================================================

-- 用户表：存储系统用户信息
CREATE TABLE IF NOT EXISTS od_users (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    username      TEXT    NOT NULL UNIQUE,
    password_hash TEXT    NOT NULL,              -- sha256(salt + ":" + pwd) 哈希
    display_name  TEXT,
    enabled       INTEGER NOT NULL DEFAULT 1,
    created_at    TEXT    NOT NULL,
    updated_at    TEXT    NOT NULL
);

-- 会话表：存储用户登录会话（当前使用 gf 内存会话，此表预留）
CREATE TABLE IF NOT EXISTS od_sessions (
    id         TEXT NOT NULL,
    username   TEXT NOT NULL,
    expires    TEXT NOT NULL,
    created_at TEXT NOT NULL,
    PRIMARY KEY (id)
);

-- 任务表：存储任务定义和配置
CREATE TABLE IF NOT EXISTS od_jobs (
    id                 INTEGER PRIMARY KEY AUTOINCREMENT,
    name               TEXT    NOT NULL UNIQUE,
    kind               TEXT    NOT NULL,           -- mysql | tdengine | minio | shell
    enabled            INTEGER NOT NULL DEFAULT 0, -- 模板默认禁用
    sort_order         INTEGER NOT NULL DEFAULT 0,
    cron               TEXT    NOT NULL,           -- gcron 6/7 位表达式（秒 分 时 日 月 周 [年]）
    depends_on         TEXT,                       -- 前置任务名（流水线依赖）
    description        TEXT,
    container          TEXT,
    database           TEXT,
    command            TEXT,                       -- 执行命令，支持 ${VAR}
    verify_command     TEXT,                       -- 检测命令，支持 ${VAR}（可选）
    min_free_mb        INTEGER NOT NULL DEFAULT 2048, -- 磁盘预检阈值(MB)
    out_dir            TEXT,
    retention_days     INTEGER NOT NULL DEFAULT 7,
    kv                 TEXT,                       -- 扩展参数 JSON
    notify_on_success  INTEGER NOT NULL DEFAULT 0,
    notify_on_failure  INTEGER NOT NULL DEFAULT 1,
    created_at         TEXT    NOT NULL,
    updated_at         TEXT    NOT NULL
);

-- 任务执行历史表：记录每次任务执行的详细信息
CREATE TABLE IF NOT EXISTS od_job_runs (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    job_id        INTEGER NOT NULL,
    status        TEXT    NOT NULL,              -- running | ok | failed | skipped
    started_at    TEXT    NOT NULL,
    finished_at   TEXT,
    elapsed_ms    INTEGER,
    exit_code     INTEGER,
    error         TEXT,                          -- 失败原因（执行/检测/依赖/预检）
    command       TEXT,                          -- 实际执行的命令
    output        TEXT,                          -- 执行命令输出(截断 4096)
    verify_status TEXT,                          -- ok | failed | skipped
    verify_output TEXT,
    size_mb       REAL,                          -- 备份产物体积(上报)
    disk_free_mb  REAL,                          -- 执行前磁盘剩余(预检)
    created_at    TEXT    NOT NULL
);

-- 任务执行历史索引：按任务ID和启动时间倒序查询
CREATE INDEX IF NOT EXISTS idx_runs_job ON od_job_runs(job_id, started_at DESC);

-- 通知配置表：存储飞书等通知渠道配置
CREATE TABLE IF NOT EXISTS od_notify_config (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    channel    TEXT    NOT NULL DEFAULT 'feishu', -- 通知渠道
    enabled    INTEGER NOT NULL DEFAULT 1,
    config     TEXT    NOT NULL,                  -- JSON：webhook、通知策略等
    updated_at TEXT    NOT NULL
);

-- ============================================================
-- 初始数据（可选，程序首次启动会自动 seed）
-- ============================================================

-- 默认管理员用户（实际由 .env 的 ADMIN_PASSWORD 配置）
-- INSERT INTO od_users (username, password_hash, display_name, enabled, created_at, updated_at)
-- VALUES ('admin', 'sha256_hash_here', '管理员', 1, datetime('now'), datetime('now'));

-- 默认飞书通知配置（webhook 需要用户配置）
-- INSERT INTO od_notify_config (channel, enabled, config, updated_at)
-- VALUES ('feishu', 1, '{"webhook":"","on_failure_only":true}', datetime('now'));
