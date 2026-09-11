package consts

const (
	// Table names.
	TableUsers        = "od_users"
	TableSessions     = "od_sessions"
	TableJobs         = "od_jobs"
	TableJobRuns      = "od_job_runs"
	TableNotifyConfig = "od_notify_config"

	// Job kinds.
	JobKindMysql    = "mysql"
	JobKindTdengine = "tdengine"
	JobKindMinio    = "minio"
	JobKindShell    = "shell"

	// Job status.
	StatusRunning = "running"
	StatusOk      = "ok"
	StatusFailed  = "failed"
	StatusSkipped = "skipped"

	// Default admin.
	DefaultAdminUser = "admin"
	DefaultAdminPwd  = "Opsdump1!"
	EnvAdminPassword = "ADMIN_PASSWORD"

	// Session marker key.
	SessionUserId   = "uid"
	SessionUsername = "username"

	// Notification channel.
	ChannelFeishu   = "feishu"
	ChannelDingtalk = "dingtalk"

	// Page routes.
	PageLogin = "/login"
)
