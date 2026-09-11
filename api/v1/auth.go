package v1

import "github.com/gogf/gf/v2/frame/g"

// Auth module: login/logout/session.

type LoginReq struct {
	g.Meta   `path:"login" method:"post" tags:"Auth" summary:"登录"`
	Username string `v:"required" json:"username" dc:"用户名"`
	Password string `v:"required" json:"password" dc:"密码"`
}
type LoginRes struct {
	Username string `json:"username" dc:"用户名"`
}

type LogoutReq struct {
	g.Meta `path:"logout" method:"post" tags:"Auth" summary:"登出"`
}
type LogoutRes struct{}

type CurrentUserReq struct {
	g.Meta `path:"current-user" method:"get" tags:"Auth" summary:"当前用户"`
}
type CurrentUserRes struct {
	Username string `json:"username" dc:"用户名"`
}
