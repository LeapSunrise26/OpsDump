package model

// AppConfig is the process-level configuration loaded from the `ops-dump`
// node in manifest/config/config.yaml.
type AppConfig struct {
	Timezone    string `json:"timezone"`
	DbPath      string `json:"dbPath"`
	Credentials struct {
		EnvFile string `json:"envFile"`
	} `json:"credentials"`
	Admin struct {
		Username string `json:"username"`
	} `json:"admin"`
}

// LoginInput is the internal login request payload.
type LoginInput struct {
	Username string `p:"username" v:"required"`
	Password string `p:"password" v:"required"`
}

// LoginOutput is the login response.
type LoginOutput struct {
	Username string `json:"username"`
}

// DashboardStats is rendered on the dashboard page.
type DashboardStats struct {
	Today map[string]int
	Total int
}

// JobParam is a struct used when creating/updating a job from the API.
type JobParam struct {
	Id              int64  `p:"id"`
	Name            string `p:"name"`
	Kind            string `p:"kind"`
	Enabled         int    `p:"enabled"`
	SortOrder       int    `p:"sort_order"`
	Cron            string `p:"cron"`
	DependsOn       string `p:"depends_on"`
	Description     string `p:"description"`
	Command         string `p:"command"`
	VerifyCommand   string `p:"verify_command"`
	OutDir          string `p:"out_dir"`
	RetentionDays   int    `p:"retention_days"`
	Kv              string `p:"kv"`
	NotifyOnSuccess int    `p:"notify_on_success"`
	NotifyOnFailure int    `p:"notify_on_failure"`
}
