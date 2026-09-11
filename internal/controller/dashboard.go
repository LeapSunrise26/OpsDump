package controller

import (
	"context"

	v1 "ops-dump/api/v1"
	"ops-dump/internal/service"
)

// DashboardController handles dashboard statistics.
type DashboardController struct{}

func NewDashboard() *DashboardController { return &DashboardController{} }

func (c *DashboardController) Get(ctx context.Context, req *v1.DashboardReq) (res *v1.DashboardRes, err error) {
	dJobs, stats, err := service.NewJob().Dashboard(ctx)
	if err != nil {
		return
	}
	res = &v1.DashboardRes{
		Stats: v1.DashboardStats{
			Total:   stats.Total,
			Enabled: stats.Enabled,
			Ok:      stats.Ok,
			Fail:    stats.Fail,
			Rate:    stats.Rate,
		},
		Jobs: make([]v1.DashboardJob, 0, len(dJobs)),
	}
	for _, dj := range dJobs {
		res.Jobs = append(res.Jobs, v1.DashboardJob{
			Id:           dj.Id,
			Name:         dj.Name,
			Kind:         dj.Kind,
			Enabled:      dj.Enabled,
			Cron:         dj.Cron,
			LastStatus:   dj.LastRun.Status,
			LastTime:     dj.LastRun.FinishedAt,
			TodayOk:      dj.TodayOk,
			TodayFail:    dj.TodayFail,
			TodayRate:    dj.TodayRate,
			TodayAvgTime: dj.TodayAvgTime,
		})
	}
	return
}
