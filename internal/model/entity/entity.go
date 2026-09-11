package entity

// User maps table users.
type User struct {
	Id           int64  `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"password_hash,omitempty"`
	DisplayName  string `json:"display_name"`
	Enabled      int    `json:"enabled"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

// Job maps table jobs.
type Job struct {
	Id               int64  `json:"id"`
	Name             string `json:"name"`
	Kind             string `json:"kind"`
	Enabled          int    `json:"enabled"`
	SortOrder        int    `json:"sort_order"`
	Cron             string `json:"cron"`
	DependsOn        string `json:"depends_on"`
	Description      string `json:"description"`
	Command          string `json:"command"`
	VerifyCommand    string `json:"verify_command"`
	OutDir           string `json:"out_dir"`
	RetentionDays    int    `json:"retention_days"`
	Kv               string `json:"kv"`
	NotifyOnSuccess  int    `json:"notify_on_success"`
	NotifyOnFailure  int    `json:"notify_on_failure"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
}

// JobRun maps table job_runs.
type JobRun struct {
	Id           int64   `json:"id"`
	JobId        int64   `json:"job_id"`
	JobName      string  `json:"job_name,omitempty"`
	Status       string  `json:"status"`
	StartedAt    string  `json:"started_at"`
	FinishedAt   string  `json:"finished_at"`
	ElapsedMs    int64   `json:"elapsed_ms"`
	ExitCode     int     `json:"exit_code"`
	Error        string  `json:"error"`
	Command      string  `json:"command"`
	Output       string  `json:"output"`
	VerifyStatus string  `json:"verify_status"`
	VerifyOutput string  `json:"verify_output"`
	SizeMb       float64 `json:"size_mb"`
	DiskFreeMb   float64 `json:"disk_free_mb"`
	CreatedAt    string  `json:"created_at"`
}

// NotifyConfig maps table notify_config.
type NotifyConfig struct {
	Id        int64  `json:"id"`
	Channel   string `json:"channel"`
	Enabled   int    `json:"enabled"`
	Config    string `json:"config"`
	UpdatedAt string `json:"updated_at"`
}

// Session maps table sessions.
type Session struct {
	Id        string `json:"id"`
	Username  string `json:"username"`
	Expires   string `json:"expires"`
	CreatedAt string `json:"created_at"`
}
