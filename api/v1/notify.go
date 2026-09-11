package v1

import "github.com/gogf/gf/v2/frame/g"

// Notify module: notification channel configuration.

type NotifyChannelConfig struct {
	Enabled int    `json:"enabled" dc:"是否启用"`
	Webhook string `json:"webhook" dc:"Webhook URL"`
}

type NotifyConfigs struct {
	Feishu   *NotifyChannelConfig `json:"feishu" dc:"飞书配置"`
	Dingtalk *NotifyChannelConfig `json:"dingtalk" dc:"钉钉配置"`
}

type NotifyGetReq struct {
	g.Meta `path:"notify/config" method:"get" tags:"Notify" summary:"获取通知配置"`
}
type NotifyGetRes struct {
	Configs *NotifyConfigs `json:"configs" dc:"通知配置"`
}

type NotifySaveReq struct {
	g.Meta `path:"notify/config" method:"post" tags:"Notify" summary:"保存通知配置"`
	Channel string `json:"channel" dc:"通知渠道"`
	NotifyChannelConfig
}
type NotifySaveRes struct{}

type NotifyTestReq struct {
	g.Meta   `path:"notify/test" method:"post" tags:"Notify" summary:"测试通知"`
	Channel  string `json:"channel" dc:"测试渠道"`
}
type NotifyTestRes struct {
	Success bool   `json:"success" dc:"是否成功"`
	Message string `json:"message" dc:"结果消息"`
}
