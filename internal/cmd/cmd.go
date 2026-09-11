package cmd

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"

	"ops-dump/internal/model"
	"ops-dump/internal/service"
)

// Main is the program entrypoint.
func Main() {
	ctx := gctx.New()
	var cfg model.AppConfig

	if err := gconv.Struct(g.Cfg().MustGet(ctx, "ops-dump").Val(), &cfg); err != nil {
		g.Log().Fatalf(ctx, "load config failed: %+v", err)
	}

	if cfg.Timezone != "" {
		_ = gtime.SetTimeZone(cfg.Timezone)
	}

	if err := service.NewBoot().Init(ctx, &cfg); err != nil {
		g.Log().Fatalf(ctx, "init failed: %+v", err)
	}
	g.Log().Infof(ctx, "database ready at %s", cfg.DbPath)

	if err := service.NewJob().SchedulerReload(ctx); err != nil {
		g.Log().Errorf(ctx, "scheduler reload failed: %+v", err)
	}

	s := g.Server()
	Register(s)

	g.Log().Infof(ctx, "opsdump started")
	s.Run()
}
