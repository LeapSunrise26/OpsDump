package v1

import "github.com/gogf/gf/v2/frame/g"

// System module: liveness probe, status, and a demo endpoint.

type PingReq struct {
	g.Meta `path:"/ping" method:"all" tags:"System" summary:"健康检查(liveness)"`
}
type PingRes struct {
	Message string `json:"message" dc:"消息"`
}

type HealthzReq struct {
	g.Meta `path:"/healthz" method:"all" tags:"System" summary:"服务状态"`
}
type HealthzRes struct {
	Status  string `json:"status" dc:"状态"`
	Version string `json:"version" dc:"版本"`
}

type HelloReq struct {
	g.Meta `path:"/hello" method:"get" tags:"System" summary:"示例接口"`
	Name   string `v:"required" json:"name" dc:"姓名"`
	Age    int    `v:"required" json:"age" dc:"年龄"`
}
type HelloRes struct {
	Content string `json:"content" dc:"返回内容"`
}

// ShutdownReq gracefully stops the HTTP server. It is intentionally public
// (the desktop shell may call it before login) but requires the custom
// X-OpsDump-Control header: cross-origin browser requests cannot set custom
// headers without a CORS preflight, which this server never answers.
type ShutdownReq struct {
	g.Meta `path:"/api/system/shutdown" method:"post" tags:"System" summary:"优雅关闭服务(本机控制端)"`
}
type ShutdownRes struct {
	Message string `json:"message" dc:"消息"`
}
