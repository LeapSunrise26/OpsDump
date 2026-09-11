package controller

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	v1 "ops-dump/api/v1"
	"ops-dump/internal/consts"
	"ops-dump/internal/model"
	"ops-dump/internal/service"
)

// AuthController handles login, logout and current-user.
type AuthController struct{}

func NewAuth() *AuthController { return &AuthController{} }

func (c *AuthController) Login(ctx context.Context, req *v1.LoginReq) (res *v1.LoginRes, err error) {
	user, err := service.NewAuth().Login(ctx, model.LoginInput{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		return
	}
	r := g.RequestFromCtx(ctx)
	_ = r.Session.Set(consts.SessionUserId, user.Id)
	_ = r.Session.Set(consts.SessionUsername, user.Username)
	res = &v1.LoginRes{Username: user.Username}
	return
}

func (c *AuthController) Logout(ctx context.Context, req *v1.LogoutReq) (res *v1.LogoutRes, err error) {
	_ = g.RequestFromCtx(ctx).Session.RemoveAll()
	return
}

func (c *AuthController) CurrentUser(ctx context.Context, req *v1.CurrentUserReq) (res *v1.CurrentUserRes, err error) {
	name := g.RequestFromCtx(ctx).Session.MustGet(consts.SessionUsername).String()
	res = &v1.CurrentUserRes{Username: name}
	return
}
