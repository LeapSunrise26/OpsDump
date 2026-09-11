package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"

	"ops-dump/internal/consts"
	"ops-dump/internal/dao"
	"ops-dump/internal/model"
	"ops-dump/internal/model/entity"
)

const pwdSalt = "ops-dump.salt.v1"

func hashPassword(pwd string) string {
	sum := sha256.Sum256([]byte(pwdSalt + ":" + pwd))
	return hex.EncodeToString(sum[:])
}

// AuthService handles administrator authentication and seeding.
type AuthService struct{}

// NewAuth creates an AuthService instance.
func NewAuth() *AuthService { return &AuthService{} }

// SeedAdmin creates the initial administrator only if no user exists yet. The
// password comes from .env ADMIN_PASSWORD and is set ONCE; existing accounts
// are never overwritten, so a changed password stays in effect across restarts.
// To recover a forgotten password, delete ops-dump.db (re-seeds with the
// default) or provide a dedicated reset mechanism.
func (s *AuthService) SeedAdmin(ctx context.Context) error {
	n, err := dao.Users.Ctx(ctx).Count()
	if err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	username := Cfg.Admin.Username
	if username == "" {
		username = consts.DefaultAdminUser
	}
	pwd := Env.Get(consts.EnvAdminPassword)
	if pwd == "" {
		pwd = consts.DefaultAdminPwd
	}
	_, err = dao.Users.Ctx(ctx).Insert(dataNoId(entity.User{
		Username:     username,
		PasswordHash: hashPassword(pwd),
		DisplayName:  username,
		Enabled:      1,
		CreatedAt:    now(),
		UpdatedAt:    now(),
	}))
	return err
}

// Login verifies the username/password and returns the user on success.
func (s *AuthService) Login(ctx context.Context, in model.LoginInput) (*entity.User, error) {
	var user entity.User
	err := dao.Users.Ctx(ctx).Where("username", in.Username).Scan(&user)
	if err != nil {
		return nil, err
	}
	if user.Id == 0 {
		return nil, gerror.NewCode(gcode.CodeOperationFailed, "密码错误")
	}
	if user.Enabled != 1 {
		return nil, gerror.NewCode(gcode.CodeOperationFailed, "用户已禁用")
	}
	if user.PasswordHash != hashPassword(in.Password) {
		return nil, gerror.NewCode(gcode.CodeOperationFailed, "密码错误")
	}
	return &user, nil
}
