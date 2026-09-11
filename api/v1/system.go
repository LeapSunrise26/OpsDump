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
