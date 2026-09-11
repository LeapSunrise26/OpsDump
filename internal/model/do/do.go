package do

// User is the data object for dao user operations.
type User struct {
	Username     any `json:"username"`
	PasswordHash any `json:"password_hash"`
	DisplayName  any `json:"display_name"`
	Enabled      any `json:"enabled"`
	CreatedAt    any `json:"created_at"`
	UpdatedAt    any `json:"updated_at"`
}

// Job is the data object for dao job operations.
type Job struct {
	Name          any `json:"name"`
	Kind          any `json:"kind"`
	Enabled       any `json:"enabled"`
	Cron          any `json:"cron"`
	DependsOn     any `json:"depends_on"`
	Description   any `json:"description"`
	Container     any `json:"container"`
	Database      any `json:"database"`
	Command       any `json:"command"`
	VerifyCommand any `json:"verify_command"`
	MinFreeMb     any `json:"min_free_mb"`
	OutDir        any `json:"out_dir"`
	RetentionDays any `json:"retention_days"`
	Kv            any `json:"kv"`
	UpdatedAt     any `json:"updated_at"`
}

// JobRun is the data object for dao job_runs operations.
type JobRun struct {
	JobId        any `json:"job_id"`
	Status       any `json:"status"`
	StartedAt    any `json:"started_at"`
	FinishedAt   any `json:"finished_at"`
	ElapsedMs    any `json:"elapsed_ms"`
	ExitCode     any `json:"exit_code"`
	Error        any `json:"error"`
	Output       any `json:"output"`
	VerifyStatus any `json:"verify_status"`
	VerifyOutput any `json:"verify_output"`
	SizeMb       any `json:"size_mb"`
	DiskFreeMb   any `json:"disk_free_mb"`
	CreatedAt    any `json:"created_at"`
}