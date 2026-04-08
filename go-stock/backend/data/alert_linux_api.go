//go:build linux
// +build linux

package data

import (
	"go-stock/backend/logger"
	"os/exec"
)

// AlertWindowsApi keeps the existing API name used by the app layer.
type AlertWindowsApi struct {
	AppID   string
	Title   string
	Content string
	Icon    string
}

func NewAlertWindowsApi(AppID string, Title string, Content string, Icon string) *AlertWindowsApi {
	return &AlertWindowsApi{
		AppID:   AppID,
		Title:   Title,
		Content: Content,
		Icon:    Icon,
	}
}

func (a AlertWindowsApi) SendNotification() bool {
	if GetSettingConfig().LocalPushEnable == false {
		logger.SugaredLogger.Error("本地推送未开启")
		return false
	}

	if _, err := exec.LookPath("notify-send"); err != nil {
		logger.SugaredLogger.Warn("notify-send not found, skip local notification")
		return false
	}

	cmd := exec.Command("notify-send", a.Title, a.Content)
	if err := cmd.Run(); err != nil {
		logger.SugaredLogger.Error(err)
		return false
	}

	return true
}
