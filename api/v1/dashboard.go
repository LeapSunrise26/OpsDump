package v1

import "github.com/gogf/gf/v2/frame/g"

// Dashboard module: overview statistics.

type DashboardStats struct {
	Total   int    `json:"total" dc:"任务总数"`
	Enabled int    `json:"enabled" dc:"已启用"`
	Ok      int    `json:"ok" dc:"今日成功"`
	Fail    int    `json:"fail" dc:"今日失败"`
	Rate    string `json:"rate" dc:"成功率"`
}

type DashboardJobRun struct {
	JobId        int64  `json:"job_id"`
	Status       string `json:"status"`
	ElapsedMs    int64  `json:"elapsed_ms"`
	FinishedAt   string `json:"finished_at"`
}

type DashboardJob struct {
	Id           int64   `json:"id"`
	Name         string  `json:"name"`
	Kind         string  `json:"kind"`
	Enabled      int     `json:"enabled"`
	Cron         string  `json:"cron"`
	LastStatus   string  `json:"last_status"`
	LastTime     string  `json:"last_time"`
	TodayOk      int     `json:"today_ok"`
	TodayFail    int     `json:"today_fail"`
	TodayRate    string  `json:"today_rate"`
	TodayAvgTime string  `json:"today_avg_time"`
}

type DashboardReq struct {
	g.Meta `path:"dashboard" method:"get" tags:"Dashboard" summary:"概览数据"`
}
type DashboardRes struct {
	Stats DashboardStats  `json:"stats" dc:"统计数据"`
	Jobs  []DashboardJob `json:"jobs" dc:"任务列表"`
}
