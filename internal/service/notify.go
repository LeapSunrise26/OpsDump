package service

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gogf/gf/v2/frame/g"

	"ops-dump/internal/consts"
	"ops-dump/internal/dao"
	"ops-dump/internal/model/entity"
)

// feishuConfig mirrors the JSON stored in notify_config.config.
type feishuConfig struct {
	Webhook string `json:"webhook"`
}

// dingtalkConfig mirrors the JSON stored in notify_config.config for dingtalk.
type dingtalkConfig struct {
	Webhook string `json:"webhook"`
}

// NotifyService sends job outcome notifications.
type NotifyService struct{}

// NewNotify creates a NotifyService instance.
func NewNotify() *NotifyService { return &NotifyService{} }

// Notify sends a notification about a job outcome.
// The onSuccess/onFailure flags come from the job's own configuration.
func (s *NotifyService) Notify(ctx context.Context, jobName, status string, jobOnSuccess, jobOnFailure bool, err error) {
	if status == "ok" && !jobOnSuccess {
		return
	}
	if status != "ok" && !jobOnFailure {
		return
	}
	timeStr := time.Now().Format("2006-01-02 15:04:05")

	// Send to feishu
	if cfg, ok := loadFeishuConfig(ctx); ok && cfg.Webhook != "" {
		if err := sendFeishuCard(ctx, cfg.Webhook, jobName, status, timeStr); err != nil {
			g.Log().Errorf(ctx, "notify feishu failed: %+v", err)
		}
	}

	// Send to dingtalk
	if cfg, ok := loadDingtalkConfig(ctx); ok && cfg.Webhook != "" {
		if err := sendDingtalk(ctx, cfg.Webhook, jobName, status, timeStr); err != nil {
			g.Log().Errorf(ctx, "notify dingtalk failed: %+v", err)
		}
	}
}

// GetConfig returns the current notification configurations.
func (s *NotifyService) GetConfig(ctx context.Context) (*NotifyConfigs, error) {
	fc, fe := loadAnyFeishuConfig(ctx)
	dc, de := loadAnyDingtalkConfig(ctx)
	feishu := &NotifyChannelConfig{Enabled: fe}
	dingtalk := &NotifyChannelConfig{Enabled: de}
	if fc != nil {
		feishu.Webhook = fc.Webhook
	}
	if dc != nil {
		dingtalk.Webhook = dc.Webhook
	}
	return &NotifyConfigs{
		Feishu:   feishu,
		Dingtalk: dingtalk,
	}, nil
}

// SaveConfig saves a notification channel configuration.
func (s *NotifyService) SaveConfig(ctx context.Context, channel string, c *NotifyChannelConfig) error {
	var cfgJSON []byte
	switch channel {
	case consts.ChannelFeishu:
		cfgJSON, _ = json.Marshal(feishuConfig{Webhook: c.Webhook})
	case consts.ChannelDingtalk:
		cfgJSON, _ = json.Marshal(dingtalkConfig{Webhook: c.Webhook})
	default:
		return nil
	}
	var existing entity.NotifyConfig
	err := dao.NotifyConfig.Ctx(ctx).Where("channel", channel).Scan(&existing)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if existing.Id > 0 {
		_, err = dao.NotifyConfig.Ctx(ctx).Where("id", existing.Id).Data(g.Map{
			"enabled":    c.Enabled,
			"config":     string(cfgJSON),
			"updated_at": now(),
		}).Update()
	} else {
		_, err = dao.NotifyConfig.Ctx(ctx).Insert(g.Map{
			"channel":    channel,
			"enabled":    c.Enabled,
			"config":     string(cfgJSON),
			"updated_at": now(),
		})
	}
	return err
}

// TestSend sends a test notification to the specified channel.
func (s *NotifyService) TestSend(ctx context.Context, channel string) (bool, string) {
	timeStr := time.Now().Format("2006-01-02 15:04:05")
	switch channel {
	case consts.ChannelFeishu:
		fc, ok := loadFeishuConfig(ctx)
		if !ok || fc.Webhook == "" {
			return false, "飞书 Webhook 未配置"
		}
		if err := sendFeishuCard(ctx, fc.Webhook, "测试任务", "ok", timeStr); err != nil {
			return false, "发送失败: " + err.Error()
		}
		return true, "发送成功"
	case consts.ChannelDingtalk:
		dc, ok := loadDingtalkConfig(ctx)
		if !ok || dc.Webhook == "" {
			return false, "钉钉 Webhook 未配置"
		}
		if err := sendDingtalk(ctx, dc.Webhook, "测试任务", "ok", timeStr); err != nil {
			return false, "发送失败: " + err.Error()
		}
		return true, "发送成功"
	}
	return false, "未知渠道"
}

// NotifyConfigs holds all channel configurations.
type NotifyConfigs struct {
	Feishu   *NotifyChannelConfig `json:"feishu"`
	Dingtalk *NotifyChannelConfig `json:"dingtalk"`
}

// NotifyChannelConfig is the API response struct for a single notification channel.
type NotifyChannelConfig struct {
	Enabled int    `json:"enabled"`
	Webhook string `json:"webhook"`
}

func loadFeishuConfig(ctx context.Context) (*feishuConfig, bool) {
	var cfg entity.NotifyConfig
	err := dao.NotifyConfig.Ctx(ctx).Where("channel", consts.ChannelFeishu).Where("enabled", 1).Scan(&cfg)
	if err != nil || cfg.Id == 0 {
		return nil, false
	}
	var fc feishuConfig
	if err := json.Unmarshal([]byte(cfg.Config), &fc); err != nil {
		return nil, false
	}
	return &fc, true
}

func loadDingtalkConfig(ctx context.Context) (*dingtalkConfig, bool) {
	var cfg entity.NotifyConfig
	err := dao.NotifyConfig.Ctx(ctx).Where("channel", consts.ChannelDingtalk).Where("enabled", 1).Scan(&cfg)
	if err != nil || cfg.Id == 0 {
		return nil, false
	}
	var dc dingtalkConfig
	if err := json.Unmarshal([]byte(cfg.Config), &dc); err != nil {
		return nil, false
	}
	return &dc, true
}

// loadAnyFeishuConfig loads feishu config regardless of enabled status (for GetConfig).
func loadAnyFeishuConfig(ctx context.Context) (*feishuConfig, int) {
	var cfg entity.NotifyConfig
	err := dao.NotifyConfig.Ctx(ctx).Where("channel", consts.ChannelFeishu).OrderDesc("id").Scan(&cfg)
	if err != nil || cfg.Id == 0 {
		return nil, 0
	}
	var fc feishuConfig
	_ = json.Unmarshal([]byte(cfg.Config), &fc)
	return &fc, cfg.Enabled
}

// loadAnyDingtalkConfig loads dingtalk config regardless of enabled status (for GetConfig).
func loadAnyDingtalkConfig(ctx context.Context) (*dingtalkConfig, int) {
	var cfg entity.NotifyConfig
	err := dao.NotifyConfig.Ctx(ctx).Where("channel", consts.ChannelDingtalk).OrderDesc("id").Scan(&cfg)
	if err != nil || cfg.Id == 0 {
		return nil, 0
	}
	var dc dingtalkConfig
	_ = json.Unmarshal([]byte(cfg.Config), &dc)
	return &dc, cfg.Enabled
}

// sendFeishu POSTs a text message to the feishu webhook.
func sendFeishu(ctx context.Context, webhook, text string) error {
	body, _ := json.Marshal(map[string]any{
		"msg_type": "text",
		"content": map[string]string{
			"text": text,
		},
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhook, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_ = resp
	return nil
}

// sendFeishuCard sends an interactive card message with colored header.
func sendFeishuCard(ctx context.Context, webhook, jobName, status, timeStr string) error {
	icon := "✅"
	title := "任务成功"
	headerColor := "blue"
	if status != "ok" {
		icon = "❌"
		title = "任务失败"
		headerColor = "red"
	}
	body, _ := json.Marshal(map[string]any{
		"msg_type": "interactive",
		"card": map[string]any{
			"header": map[string]any{
				"title": map[string]any{
					"tag":     "plain_text",
					"content": icon + " " + title,
				},
				"template": headerColor,
			},
			"elements": []map[string]any{
				{
					"tag": "div",
					"fields": []map[string]any{
						{
							"is_short": true,
							"text": map[string]any{
								"tag":     "lark_md",
								"content": "**任务名称**\n" + jobName,
							},
						},
						{
							"is_short": true,
							"text": map[string]any{
								"tag":     "lark_md",
								"content": "**执行时间**\n" + timeStr,
							},
						},
					},
				},
			},
		},
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhook, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_ = resp
	return nil
}

// sendDingtalk sends a markdown message to dingtalk webhook.
func sendDingtalk(ctx context.Context, webhook, jobName, status, timeStr string) error {
	icon := "✅"
	title := "任务成功"
	if status != "ok" {
		icon = "❌"
		title = "任务失败"
	}
	body, _ := json.Marshal(map[string]any{
		"msgtype": "markdown",
		"markdown": map[string]any{
			"title": icon + " " + title,
			"text": "### " + icon + " " + title + "\n\n" +
				"**任务名称：** " + jobName + "\n\n" +
				"**执行时间：** " + timeStr + "\n",
		},
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhook, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_ = resp
	return nil
}
