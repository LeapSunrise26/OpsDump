package controller

import (
	"context"

	v1 "ops-dump/api/v1"
	"ops-dump/internal/service"
)

// NotifyController handles notification configuration.
type NotifyController struct{}

func NewNotify() *NotifyController { return &NotifyController{} }

func (c *NotifyController) GetConfig(ctx context.Context, req *v1.NotifyGetReq) (res *v1.NotifyGetRes, err error) {
	cfgs, err := service.NewNotify().GetConfig(ctx)
	if err != nil {
		return
	}
	res = &v1.NotifyGetRes{Configs: &v1.NotifyConfigs{
		Feishu: &v1.NotifyChannelConfig{
			Enabled: cfgs.Feishu.Enabled,
			Webhook: cfgs.Feishu.Webhook,
		},
		Dingtalk: &v1.NotifyChannelConfig{
			Enabled: cfgs.Dingtalk.Enabled,
			Webhook: cfgs.Dingtalk.Webhook,
		},
	}}
	return
}

func (c *NotifyController) SaveConfig(ctx context.Context, req *v1.NotifySaveReq) (res *v1.NotifySaveRes, err error) {
	err = service.NewNotify().SaveConfig(ctx, req.Channel, &service.NotifyChannelConfig{
		Enabled: req.Enabled,
		Webhook: req.Webhook,
	})
	return
}

func (c *NotifyController) Test(ctx context.Context, req *v1.NotifyTestReq) (res *v1.NotifyTestRes, err error) {
	ok, msg := service.NewNotify().TestSend(ctx, req.Channel)
	res = &v1.NotifyTestRes{Success: ok, Message: msg}
	return
}
