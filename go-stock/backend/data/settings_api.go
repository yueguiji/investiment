package data

import (
	"encoding/json"
	"errors"
	"go-stock/backend/db"
	"go-stock/backend/logger"
	"go-stock/backend/runtimeconfig"
	"strings"
	"time"

	"github.com/samber/lo"
	"gorm.io/gorm"
)

type Settings struct {
	gorm.Model
	TushareToken           string `json:"tushareToken"`
	LocalPushEnable        bool   `json:"localPushEnable"`
	DingPushEnable         bool   `json:"dingPushEnable"`
	DingRobot              string `json:"dingRobot"`
	FeishuPushEnable       bool   `json:"feishuPushEnable"`
	FeishuRobot            string `json:"feishuRobot"`
	FeishuSecret           string `json:"feishuSecret" gorm:"column:feishu_secret"`
	UpdateBasicInfoOnStart bool   `json:"updateBasicInfoOnStart"`
	RefreshInterval        int64  `json:"refreshInterval"`
	OpenAiEnable           bool   `json:"openAiEnable"`
	Prompt                 string `json:"prompt"`
	CheckUpdate            bool   `json:"checkUpdate"`
	QuestionTemplate       string `json:"questionTemplate"`
	CrawlTimeOut           int64  `json:"crawlTimeOut"`
	KDays                  int64  `json:"kDays"`
	EnableDanmu            bool   `json:"enableDanmu"`
	BrowserPath            string `json:"browserPath"`
	EnableNews             bool   `json:"enableNews"`
	DarkTheme              bool   `json:"darkTheme"`
	BrowserPoolSize        int    `json:"browserPoolSize"`
	EnableFund             bool   `json:"enableFund"`
	EnablePushNews         bool   `json:"enablePushNews"`
	EnableOnlyPushRedNews  bool   `json:"enableOnlyPushRedNews"`
	SponsorCode            string `json:"sponsorCode"`
	HttpProxy              string `json:"httpProxy"`
	HttpProxyEnabled       bool   `json:"httpProxyEnabled"`
	EnableAgent            bool   `json:"enableAgent"`
	QgqpBId                string `json:"qgqpBId" gorm:"column:qgqp_b_id"`
	IwencaiApiKey          string `json:"iwencaiApiKey" gorm:"column:iwencai_api_key"`
	EmApiKey               string `json:"emApiKey" gorm:"column:em_api_key"`
	PromptPlazaApiBase     string `json:"promptPlazaApiBase" gorm:"column:prompt_plaza_api_base"`
	AssetUnlockPassword    string `json:"assetUnlockPassword"`
	FundEstimateSource     string `json:"fundEstimateSource"`
}

func (receiver Settings) TableName() string {
	return "settings"
}

type AIConfig struct {
	ID               uint `gorm:"primarykey"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
	Name             string  `json:"name"`
	BaseUrl          string  `json:"baseUrl"`
	ApiKey           string  `json:"apiKey" `
	ModelName        string  `json:"modelName"`
	MaxTokens        int     `json:"maxTokens"`
	Temperature      float64 `json:"temperature"`
	TimeOut          int     `json:"timeOut"`
	HttpProxy        string  `json:"httpProxy"`
	HttpProxyEnabled bool    `json:"httpProxyEnabled"`
	SessionId        string  `json:"sessionId" gorm:"index;size:64"`
	Thinking         bool    `json:"thinking"`
}

func (AIConfig) TableName() string {
	return "ai_config"
}

type SettingConfig struct {
	*Settings
	AiConfigs []*AIConfig `json:"aiConfigs"`
}

func (c *SettingConfig) GetAIConfigThinking(aiConfigId int) bool {
	if aiConfigId <= 0 && len(c.AiConfigs) > 0 {
		return c.AiConfigs[0].Thinking
	}
	for _, cfg := range c.AiConfigs {
		if int(cfg.ID) == aiConfigId {
			return cfg.Thinking
		}
	}
	return false
}

type SettingsApi struct {
	Config *SettingConfig
}

func NewSettingsApi() *SettingsApi {
	return &SettingsApi{
		Config: GetSettingConfig(),
	}
}

func (s *SettingsApi) Export() string {
	d, _ := json.MarshalIndent(s.Config, "", "    ")
	return string(d)
}

func UpdateConfig(s *SettingConfig) string {
	count := int64(0)
	db.Dao.Model(&Settings{}).Count(&count)
	if count > 0 {
		db.Dao.Model(&Settings{}).Where("id=?", s.ID).Updates(map[string]any{
			"local_push_enable":          s.LocalPushEnable,
			"ding_push_enable":           s.DingPushEnable,
			"ding_robot":                 s.DingRobot,
			"update_basic_info_on_start": s.UpdateBasicInfoOnStart,
			"refresh_interval":           s.RefreshInterval,
			"open_ai_enable":             s.OpenAiEnable,
			"tushare_token":              s.TushareToken,
			"prompt":                     s.Prompt,
			"check_update":               s.CheckUpdate,
			"question_template":          s.QuestionTemplate,
			"crawl_time_out":             s.CrawlTimeOut,
			"k_days":                     s.KDays,
			"enable_danmu":               s.EnableDanmu,
			"browser_path":               s.BrowserPath,
			"enable_news":                s.EnableNews,
			"dark_theme":                 s.DarkTheme,
			"enable_fund":                s.EnableFund,
			"enable_push_news":           s.EnablePushNews,
			"enable_only_push_red_news":  s.EnableOnlyPushRedNews,
			"sponsor_code":               s.SponsorCode,
			"http_proxy":                 s.HttpProxy,
			"http_proxy_enabled":         s.HttpProxyEnabled,
			"enable_agent":               s.EnableAgent,
			"qgqp_b_id":                  s.QgqpBId,
			"asset_unlock_password":      s.AssetUnlockPassword,
			"fund_estimate_source":       NormalizeFundEstimateSource(s.FundEstimateSource),
		})

		//更新AiConfig
		err := updateAiConfigs(s.AiConfigs)
		if err != nil {
			logger.SugaredLogger.Errorf("更新AI模型服务配置失败: %v", err)
			return "更新AI模型服务配置失败: " + err.Error()
		}
	} else {
		logger.SugaredLogger.Infof("未找到配置，创建默认配置")
		// 创建主配置
		result := db.Dao.Model(&Settings{}).Create(&Settings{})
		if result.Error != nil {
			logger.SugaredLogger.Error("创建配置失败:", result.Error)
			return "创建配置失败: " + result.Error.Error()
		}
	}
	return "保存成功！"
}

func NormalizeFundEstimateSource(source string) string {
	switch strings.TrimSpace(source) {
	case "eastmoney_mobile":
		return "eastmoney_mobile"
	case "danjuan_position":
		return "danjuan_position"
	case "ai_corrected":
		return "ai_corrected"
	default:
		return "eastmoney_js"
	}
}

func UpdateFundEstimateSource(source string) string {
	normalized := NormalizeFundEstimateSource(source)
	var settings Settings
	result := db.Dao.Model(&Settings{}).First(&settings)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			settings.FundEstimateSource = normalized
			if err := db.Dao.Create(&settings).Error; err != nil {
				return "保存基金估算源失败: " + err.Error()
			}
			return normalized
		}
		return "保存基金估算源失败: " + result.Error.Error()
	}
	if err := db.Dao.Model(&settings).Update("fund_estimate_source", normalized).Error; err != nil {
		return "保存基金估算源失败: " + err.Error()
	}
	return normalized
}

func updateAiConfigs(aiConfigs []*AIConfig) error {
	if len(aiConfigs) == 0 {
		err := db.Dao.Exec("DELETE FROM ai_config").Error
		if err != nil {
			return err
		}
		return db.Dao.Exec("DELETE FROM sqlite_sequence WHERE name='ai_config'").Error
	}
	var ids []uint
	lo.ForEach(aiConfigs, func(item *AIConfig, index int) {
		ids = append(ids, item.ID)
	})
	var existAiConfigs []*AIConfig
	err := db.Dao.Model(&AIConfig{}).Select("id").Where("id in (?) ", ids).Find(&existAiConfigs).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	idMap := make(map[uint]bool)
	lo.ForEach(existAiConfigs, func(item *AIConfig, index int) {
		idMap[item.ID] = true
	})
	var addAiConfigs []*AIConfig
	var notDeleteIds []uint
	var e error
	lo.ForEach(aiConfigs, func(item *AIConfig, index int) {
		if e != nil {
			return
		}
		if !idMap[item.ID] {
			addAiConfigs = append(addAiConfigs, item)
		} else {
			notDeleteIds = append(notDeleteIds, item.ID)
			e = db.Dao.Model(&AIConfig{}).Where("id=?", item.ID).Updates(map[string]interface{}{
				"name":               item.Name,
				"base_url":           item.BaseUrl,
				"api_key":            item.ApiKey,
				"model_name":         item.ModelName,
				"max_tokens":         item.MaxTokens,
				"temperature":        item.Temperature,
				"time_out":           item.TimeOut,
				"http_proxy":         item.HttpProxy,
				"http_proxy_enabled": item.HttpProxyEnabled,
			}).Error
			if e != nil {
				return
			}
		}
	})
	if e != nil {
		return e
	}
	//删除旧的配置
	if len(notDeleteIds) > 0 {
		err = db.Dao.Exec("DELETE FROM ai_config WHERE id NOT IN ?", notDeleteIds).Error
		if err != nil {
			return err
		}
	}
	logger.SugaredLogger.Infof("更新aiConfigs +%d", len(addAiConfigs))
	//批量新增的配置
	err = db.Dao.CreateInBatches(addAiConfigs, len(addAiConfigs)).Error
	return err
}

func GetSettingConfig() *SettingConfig {
	if db.Dao == nil {
		return &SettingConfig{
			Settings: &Settings{
				CrawlTimeOut:       60,
				KDays:              60,
				BrowserPoolSize:    1,
				FundEstimateSource: NormalizeFundEstimateSource(""),
			},
			AiConfigs: []*AIConfig{},
		}
	}
	settingConfig := &SettingConfig{}
	settings := &Settings{}
	aiConfigs := make([]*AIConfig, 0)
	// 处理数据库查询可能返回的空结果
	result := db.Dao.Model(&Settings{}).First(settings)
	if settings.OpenAiEnable {
		// 处理AI配置查询可能出现的错误
		result = db.Dao.Model(&AIConfig{}).Find(&aiConfigs)
		if result.Error != nil {
			logger.SugaredLogger.Error("查询AI配置失败:", result.Error)
		} else if len(aiConfigs) > 0 {
			lo.ForEach(aiConfigs, func(item *AIConfig, index int) {
				if item.TimeOut <= 0 {
					item.TimeOut = 60 * 5
				}
			})
		}
		if settings.CrawlTimeOut <= 0 {
			settings.CrawlTimeOut = 60
		}
		if settings.KDays < 30 {
			settings.KDays = 60
		}
	}
	if settings.BrowserPath == "" {
		settings.BrowserPath, _ = CheckBrowser()
	}
	if settings.BrowserPoolSize <= 0 {
		settings.BrowserPoolSize = 1
	}
	settings.FundEstimateSource = NormalizeFundEstimateSource(settings.FundEstimateSource)
	applySettingsRuntimeDefaults(settings)
	settingConfig.Settings = settings
	settingConfig.AiConfigs = aiConfigs

	return settingConfig
}

func applySettingsRuntimeDefaults(settings *Settings) {
	if settings == nil {
		return
	}
	if strings.TrimSpace(settings.AssetUnlockPassword) == "" {
		settings.AssetUnlockPassword = strings.TrimSpace(runtimeconfig.Current().AssetUnlockPassword)
		if settings.ID > 0 && settings.AssetUnlockPassword != "" {
			db.Dao.Model(&Settings{}).Where("id = ?", settings.ID).Update("asset_unlock_password", settings.AssetUnlockPassword)
		}
	}
}

func AssetUnlockEnabled() bool {
	return strings.TrimSpace(GetSettingConfig().AssetUnlockPassword) != ""
}

func VerifyAssetUnlockPassword(password string) bool {
	expected := strings.TrimSpace(GetSettingConfig().AssetUnlockPassword)
	if expected == "" {
		return true
	}
	return strings.TrimSpace(password) == expected
}
