package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	v1 "ops-dump/api/v1"
)

// Version is the running build version.
const Version = "0.1.0"

// SystemController exposes liveness/status and a demo endpoint.
type SystemController struct{}

func NewSystem() *SystemController { return &SystemController{} }

func (c *SystemController) Ping(ctx context.Context, req *v1.PingReq) (res *v1.PingRes, err error) {
	g.RequestFromCtx(ctx).Response.Write("pong")
	return
}

func (c *SystemController) Healthz(ctx context.Context, req *v1.HealthzReq) (res *v1.HealthzRes, err error) {
	res = &v1.HealthzRes{
		Status:  "ok",
		Version: Version,
	}
	return
}

func (c *SystemController) Hello(ctx context.Context, req *v1.HelloReq) (res *v1.HelloRes, err error) {
	res = &v1.HelloRes{
		Content: fmt.Sprintf("Hello %s! Your Age is %d", req.Name, req.Age),
	}
	return
}

// Shutdown gracefully stops the server (used by the Electron desktop shell).
// Requires header X-OpsDump-Control: 1 — a guard against web-page CSRF:
// cross-site fetches cannot send custom headers without a CORS preflight,
// and this server never approves preflights.
func (c *SystemController) Shutdown(ctx context.Context, req *v1.ShutdownReq) (res *v1.ShutdownRes, err error) {
	r := g.RequestFromCtx(ctx)
	if r.Header.Get("X-OpsDump-Control") != "1" {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "missing X-OpsDump-Control header")
	}
	res = &v1.ShutdownRes{Message: "shutting down"}
	go func() {
		// Let the response flush before the listener closes.
		time.Sleep(300 * time.Millisecond)
		_ = g.Server().Shutdown()
	}()
	return
}
