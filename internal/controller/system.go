package controller

import (
	"context"
	"fmt"

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
