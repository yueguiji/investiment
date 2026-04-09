package main

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"investment-platform/internal/asset"
	"investment-platform/internal/bridge"
	"investment-platform/internal/portfolio"
	"investment-platform/internal/quant"

	"go-stock/backend/data"
	"go-stock/backend/logger"
	"go-stock/backend/models"
	"go-stock/backend/runtimepaths"

	"github.com/coocood/freecache"
	"github.com/go-resty/resty/v2"
	"github.com/robfig/cron/v3"
)

const localNotificationTTLSeconds = 300

// App 主应用结构体 — 聚合所有模块
type App struct {
	ctx context.Context

	// go-stock 原有能力
	cache      *freecache.Cache
	cron       *cron.Cron
	cronEntrys map[string]cron.EntryID
	AiTools    []data.Tool

	// 平台新增模块服务
	AssetService     *asset.Service
	PortfolioService *portfolio.Service
	QuantService     *quant.Service
	Bridge           *bridge.Bridge
}

// NewApp 创建应用
func NewApp() *App {
	cacheSize := 512 * 1024
	cache := freecache.NewCache(cacheSize)
	c := cron.New(cron.WithSeconds())
	c.Start()

	return &App{
		cache:            cache,
		cron:             c,
		cronEntrys:       make(map[string]cron.EntryID),
		AssetService:     asset.NewService(),
		PortfolioService: portfolio.NewService(),
		QuantService:     quant.NewService(runtimepaths.QuantTemplatesDir()),
		Bridge:           bridge.NewBridge(),
	}
}

// startup 应用启动回调
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	logger.SugaredLogger.Info("投资分析平台启动中...")

	// 初始化默认数据
	go a.AssetService.InitDefaultCategories()
	go a.QuantService.InitDefaultCategories()
	go a.QuantService.SanitizeStoredTemplates()
	go a.ensureDefaultQuantPromptTemplate()
	go a.ensureDefaultHouseholdPromptTemplate()
	go a.ensureDefaultHouseholdChatPromptTemplate()
	go a.ensureDefaultFundPromptTemplates()
	go a.AssetService.InitDefaultHouseholdBenchmarks()
	go a.ensureDefaultStockPromptTemplates()
	go a.bootstrapGoStockRuntime()

	// 启动持仓价格定时刷新
	go a.startPortfolioRefresh()
}

// domReady DOM就绪回调
func (a *App) domReady(ctx context.Context) {
	logger.SugaredLogger.Info("前端DOM就绪")
}

func (a *App) bootstrapGoStockRuntime() {
	time.Sleep(1500 * time.Millisecond)

	config := a.GetConfig()

	go data.NewMarketNewsApi().TelegraphList(30)
	go data.NewMarketNewsApi().GetSinaNews(30)
	go data.NewMarketNewsApi().TradingViewNews()

	if len(data.NewStockDataApi().GetStockList("")) == 0 {
		logger.SugaredLogger.Warn("go-stock bootstrap: stock basic data is still empty")
	}

	if config.EnableFund {
		go data.NewFundApi().AllFund()
	}
}

// beforeClose 关闭前回调
func (a *App) beforeClose(ctx context.Context) (prevent bool) {
	// 保存当日收益快照
	go a.PortfolioService.SaveDailySnapshot()
	return false
}

// shutdown 关闭回调
func (a *App) shutdown(ctx context.Context) {
	logger.SugaredLogger.Info("投资分析平台关闭")
	a.cron.Stop()
}

// startPortfolioRefresh 启动持仓价格定时刷新
func (a *App) startPortfolioRefresh() {
	// 每30秒刷新一次持仓价格（仅在交易时间内）
	a.cron.AddFunc("*/30 * * * * *", func() {
		if !a.Bridge.IsMarketOpen() {
			return
		}
		holdings := a.PortfolioService.GetAllHoldings()
		if len(holdings) == 0 {
			return
		}
		a.PortfolioService.SyncPortfolioQuotes()
	})

	// 每日收盘后保存快照
	a.cron.AddFunc("0 5 15 * * 1-5", func() {
		a.PortfolioService.SaveDailySnapshot()
	})
}

// ==================== 暴露给前端的方法 ====================

// --- 资产分析 ---

func (a *App) GetAssetCategories() []asset.AssetCategory {
	return a.AssetService.GetCategories()
}

func (a *App) CreateAssetCategory(c asset.AssetCategory) *asset.AssetCategory {
	return a.AssetService.CreateCategory(c)
}

func (a *App) GetAssets() []asset.Asset {
	return a.AssetService.GetAssets()
}

func (a *App) GetAssetsByType(assetType string) []asset.Asset {
	return a.AssetService.GetAssetsByType(assetType)
}

func (a *App) CreateAsset(assetData asset.Asset) *asset.Asset {
	return a.AssetService.CreateAsset(assetData)
}

func (a *App) UpdateAsset(assetData asset.Asset) *asset.Asset {
	return a.AssetService.UpdateAsset(assetData)
}

func (a *App) DeleteAsset(id uint) bool {
	return a.AssetService.DeleteAsset(id)
}

func (a *App) GetAssetSummary() *asset.AssetSummary {
	return a.AssetService.GetAssetSummary()
}

func (a *App) GetHouseholdAccounts() []asset.HouseholdAccount {
	return a.AssetService.GetHouseholdAccounts()
}

func (a *App) GetHouseholdMembers() []asset.HouseholdMember {
	return a.AssetService.GetHouseholdMembers()
}

func (a *App) CreateHouseholdMember(item asset.HouseholdMember) *asset.HouseholdMember {
	return a.AssetService.CreateHouseholdMember(item)
}

func (a *App) UpdateHouseholdMember(item asset.HouseholdMember) *asset.HouseholdMember {
	return a.AssetService.UpdateHouseholdMember(item)
}

func (a *App) DeleteHouseholdMember(id uint) bool {
	return a.AssetService.DeleteHouseholdMember(id)
}

func (a *App) CreateHouseholdAccount(item asset.HouseholdAccount) *asset.HouseholdAccount {
	return a.AssetService.CreateHouseholdAccount(item)
}

func (a *App) UpdateHouseholdAccount(item asset.HouseholdAccount) *asset.HouseholdAccount {
	return a.AssetService.UpdateHouseholdAccount(item)
}

func (a *App) DeleteHouseholdAccount(id uint) bool {
	return a.AssetService.DeleteHouseholdAccount(id)
}

func (a *App) GetHouseholdFixedAssets() []asset.HouseholdFixedAsset {
	return a.AssetService.GetHouseholdFixedAssets()
}

func (a *App) CreateHouseholdFixedAsset(item asset.HouseholdFixedAsset) *asset.HouseholdFixedAsset {
	return a.AssetService.CreateHouseholdFixedAsset(item)
}

func (a *App) UpdateHouseholdFixedAsset(item asset.HouseholdFixedAsset) *asset.HouseholdFixedAsset {
	return a.AssetService.UpdateHouseholdFixedAsset(item)
}

func (a *App) DeleteHouseholdFixedAsset(id uint) bool {
	return a.AssetService.DeleteHouseholdFixedAsset(id)
}

func (a *App) GetHouseholdIncomes() []asset.HouseholdIncome {
	return a.AssetService.GetHouseholdIncomes()
}

func (a *App) CreateHouseholdIncome(item asset.HouseholdIncome) *asset.HouseholdIncome {
	return a.AssetService.CreateHouseholdIncome(item)
}

func (a *App) UpdateHouseholdIncome(item asset.HouseholdIncome) *asset.HouseholdIncome {
	return a.AssetService.UpdateHouseholdIncome(item)
}

func (a *App) DeleteHouseholdIncome(id uint) bool {
	return a.AssetService.DeleteHouseholdIncome(id)
}

func (a *App) GetHouseholdProtections() []asset.HouseholdProtection {
	return a.AssetService.GetHouseholdProtections()
}

func (a *App) CreateHouseholdProtection(item asset.HouseholdProtection) *asset.HouseholdProtection {
	return a.AssetService.CreateHouseholdProtection(item)
}

func (a *App) UpdateHouseholdProtection(item asset.HouseholdProtection) *asset.HouseholdProtection {
	return a.AssetService.UpdateHouseholdProtection(item)
}

func (a *App) DeleteHouseholdProtection(id uint) bool {
	return a.AssetService.DeleteHouseholdProtection(id)
}

func (a *App) GetHouseholdLiabilities() []asset.HouseholdLiability {
	return a.AssetService.GetHouseholdLiabilities()
}

func (a *App) CreateHouseholdLiability(item asset.HouseholdLiability) *asset.HouseholdLiability {
	return a.AssetService.CreateHouseholdLiability(item)
}

func (a *App) UpdateHouseholdLiability(item asset.HouseholdLiability) *asset.HouseholdLiability {
	return a.AssetService.UpdateHouseholdLiability(item)
}

func (a *App) DeleteHouseholdLiability(id uint) bool {
	return a.AssetService.DeleteHouseholdLiability(id)
}

func (a *App) GetHouseholdLiabilitySchedules(liabilityID uint) []asset.HouseholdLiabilitySchedule {
	return a.AssetService.GetHouseholdLiabilitySchedules(liabilityID)
}

func (a *App) RebuildHouseholdLiabilitySchedule(liabilityID uint) bool {
	return a.AssetService.RebuildHouseholdLiabilitySchedule(liabilityID)
}

func (a *App) GetHouseholdLiabilityTrend(monthsBack, monthsForward int) []asset.HouseholdLiabilityTrendPoint {
	return a.AssetService.GetHouseholdLiabilityTrend(monthsBack, monthsForward)
}

func (a *App) GetHouseholdLiquidAssetTrend(days int) []asset.HouseholdLiquidAssetTrendPoint {
	return a.AssetService.GetHouseholdLiquidAssetTrend(days)
}

func (a *App) GetHouseholdLiquidAssetDistribution() []asset.HouseholdLiquidAssetDistributionItem {
	return a.AssetService.GetHouseholdLiquidAssetDistribution()
}

func (a *App) GetHouseholdDashboardSummary() *asset.HouseholdDashboardSummary {
	return a.AssetService.GetHouseholdDashboardSummary()
}

func (a *App) SaveHouseholdSnapshot(triggerSource string) *asset.HouseholdSnapshot {
	return a.AssetService.SaveHouseholdSnapshot(triggerSource)
}

func (a *App) GetHouseholdSnapshots(days int) []asset.HouseholdSnapshot {
	return a.AssetService.GetHouseholdSnapshots(days)
}

func (a *App) GetLatestHouseholdAIAnalysis() *asset.HouseholdAIAnalysis {
	return a.AssetService.GetLatestHouseholdAIAnalysis()
}

func (a *App) GetHouseholdProfile() *asset.HouseholdProfile {
	return a.AssetService.GetHouseholdProfile()
}

func (a *App) UpsertHouseholdProfile(item asset.HouseholdProfile) *asset.HouseholdProfile {
	return a.AssetService.UpsertHouseholdProfile(item)
}

func (a *App) GetHouseholdBenchmarks(region string) []asset.HouseholdBenchmarkRecord {
	return a.AssetService.GetHouseholdBenchmarks(region)
}

func (a *App) UpsertHouseholdBenchmark(item asset.HouseholdBenchmarkRecord) *asset.HouseholdBenchmarkRecord {
	return a.AssetService.UpsertHouseholdBenchmark(item)
}

func (a *App) DeleteHouseholdBenchmark(id uint) bool {
	return a.AssetService.DeleteHouseholdBenchmark(id)
}

func (a *App) RunHouseholdAIAnalysis(region string, aiConfigId int, promptTemplateId int, triggerSource string) map[string]any {
	if strings.TrimSpace(region) == "" {
		region = "天津市"
	}
	if strings.TrimSpace(triggerSource) == "" {
		triggerSource = "manual"
	}
	contextData := a.AssetService.BuildHouseholdAIContext(region)
	inputPayload := a.AssetService.BuildHouseholdAIContextJSON(region)
	prompt := a.buildHouseholdAnalysisPrompt(region, inputPayload)

	if aiConfigId <= 0 {
		configs := a.GetAiConfigs()
		if len(configs) > 0 {
			aiConfigId = int(configs[0].ID)
		}
	}

	openAI := data.NewDeepSeekOpenAi(a.ctx, aiConfigId)
	if strings.TrimSpace(openAI.BaseUrl) == "" || strings.TrimSpace(openAI.ApiKey) == "" || strings.TrimSpace(openAI.Model) == "" {
		record := a.AssetService.SaveHouseholdAIAnalysis(asset.HouseholdAIAnalysis{
			TriggerSource:    triggerSource,
			Region:           region,
			BenchmarkVersion: contextData.BenchmarkVersion,
			AIConfigID:       aiConfigId,
			PromptTemplateID: promptTemplateId,
			ModelName:        "",
			Status:           "skipped",
			Prompt:           prompt,
			InputPayload:     inputPayload,
			AnalysisMarkdown: "",
			ErrorMessage:     "AI 源未配置完整",
		})
		return map[string]any{
			"success":   false,
			"message":   "AI 源未配置完整，请先在设置页补充接口地址、API Key 和模型名称",
			"analysis":  "",
			"prompt":    prompt,
			"record":    record,
			"aiEnabled": false,
		}
	}

	client := resty.New().
		SetBaseURL(strings.TrimSpace(openAI.BaseUrl)).
		SetHeader("Authorization", "Bearer "+strings.TrimSpace(openAI.ApiKey)).
		SetHeader("Content-Type", "application/json")
	if openAI.TimeOut <= 0 {
		openAI.TimeOut = 180
	}
	client.SetTimeout(time.Duration(openAI.TimeOut) * time.Second)
	if openAI.HttpProxyEnabled && strings.TrimSpace(openAI.HttpProxy) != "" {
		client.SetProxy(strings.TrimSpace(openAI.HttpProxy))
	}

	body := map[string]any{
		"model":       openAI.Model,
		"max_tokens":  openAI.MaxTokens,
		"temperature": 0.2,
		"stream":      false,
		"messages": []map[string]any{
			{
				"role":    "system",
				"content": a.resolveHouseholdSystemPrompt(promptTemplateId),
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
	}
	if openAI.MaxTokens <= 0 {
		body["max_tokens"] = 4096
	}

	resp, err := client.R().SetBody(body).Post("/chat/completions")
	if err != nil {
		record := a.AssetService.SaveHouseholdAIAnalysis(asset.HouseholdAIAnalysis{
			TriggerSource:    triggerSource,
			Region:           region,
			BenchmarkVersion: contextData.BenchmarkVersion,
			AIConfigID:       aiConfigId,
			PromptTemplateID: promptTemplateId,
			ModelName:        openAI.Model,
			Status:           "failed",
			Prompt:           prompt,
			InputPayload:     inputPayload,
			ErrorMessage:     err.Error(),
		})
		return map[string]any{
			"success":   false,
			"message":   err.Error(),
			"analysis":  "",
			"prompt":    prompt,
			"record":    record,
			"aiEnabled": true,
			"model":     openAI.Model,
		}
	}
	if resp.StatusCode() >= 400 {
		message := fmt.Sprintf("模型调用失败: HTTP %d %s", resp.StatusCode(), truncateText(resp.String(), 320))
		record := a.AssetService.SaveHouseholdAIAnalysis(asset.HouseholdAIAnalysis{
			TriggerSource:    triggerSource,
			Region:           region,
			BenchmarkVersion: contextData.BenchmarkVersion,
			AIConfigID:       aiConfigId,
			PromptTemplateID: promptTemplateId,
			ModelName:        openAI.Model,
			Status:           "failed",
			Prompt:           prompt,
			InputPayload:     inputPayload,
			ErrorMessage:     message,
		})
		return map[string]any{
			"success":   false,
			"message":   message,
			"analysis":  "",
			"prompt":    prompt,
			"record":    record,
			"aiEnabled": true,
			"model":     openAI.Model,
		}
	}

	var aiResp data.AiResponse
	if err := json.Unmarshal(resp.Body(), &aiResp); err != nil {
		message := "模型响应解析失败: " + err.Error()
		record := a.AssetService.SaveHouseholdAIAnalysis(asset.HouseholdAIAnalysis{
			TriggerSource:    triggerSource,
			Region:           region,
			BenchmarkVersion: contextData.BenchmarkVersion,
			AIConfigID:       aiConfigId,
			PromptTemplateID: promptTemplateId,
			ModelName:        openAI.Model,
			Status:           "failed",
			Prompt:           prompt,
			InputPayload:     inputPayload,
			ErrorMessage:     message,
		})
		return map[string]any{
			"success":   false,
			"message":   message,
			"analysis":  "",
			"prompt":    prompt,
			"record":    record,
			"aiEnabled": true,
			"model":     openAI.Model,
		}
	}
	if len(aiResp.Choices) == 0 {
		message := "模型没有返回分析内容"
		record := a.AssetService.SaveHouseholdAIAnalysis(asset.HouseholdAIAnalysis{
			TriggerSource:    triggerSource,
			Region:           region,
			BenchmarkVersion: contextData.BenchmarkVersion,
			AIConfigID:       aiConfigId,
			PromptTemplateID: promptTemplateId,
			ModelName:        openAI.Model,
			Status:           "failed",
			Prompt:           prompt,
			InputPayload:     inputPayload,
			ErrorMessage:     message,
		})
		return map[string]any{
			"success":   false,
			"message":   message,
			"analysis":  "",
			"prompt":    prompt,
			"record":    record,
			"aiEnabled": true,
			"model":     openAI.Model,
		}
	}

	content := strings.TrimSpace(aiResp.Choices[0].Message.Content)
	record := a.AssetService.SaveHouseholdAIAnalysis(asset.HouseholdAIAnalysis{
		TriggerSource:    triggerSource,
		Region:           region,
		BenchmarkVersion: contextData.BenchmarkVersion,
		AIConfigID:       aiConfigId,
		PromptTemplateID: promptTemplateId,
		ModelName:        openAI.Model,
		Status:           "success",
		Prompt:           prompt,
		InputPayload:     inputPayload,
		AnalysisMarkdown: content,
	})
	return map[string]any{
		"success":   true,
		"message":   "分析完成",
		"analysis":  content,
		"prompt":    prompt,
		"record":    record,
		"aiEnabled": true,
		"model":     openAI.Model,
	}
}

func (a *App) ChatWithHouseholdDigitalAnalysis(region string, aiConfigId int, promptTemplateId int, messages []asset.HouseholdChatMessage) map[string]any {
	if strings.TrimSpace(region) == "" {
		region = "天津市"
	}

	contextData := a.AssetService.BuildHouseholdAIContext(region)
	contextPayload := a.AssetService.BuildHouseholdAIContextJSON(region)
	contextNote := a.buildHouseholdChatContextNote(contextData, contextPayload)

	if aiConfigId <= 0 {
		configs := a.GetAiConfigs()
		if len(configs) > 0 {
			aiConfigId = int(configs[0].ID)
		}
	}

	openAI := data.NewDeepSeekOpenAi(a.ctx, aiConfigId)
	if strings.TrimSpace(openAI.BaseUrl) == "" || strings.TrimSpace(openAI.ApiKey) == "" || strings.TrimSpace(openAI.Model) == "" {
		return map[string]any{
			"success":   false,
			"message":   "AI 源未配置完整，请先在设置页补充接口地址、API Key 和模型名称",
			"reply":     "",
			"model":     "",
			"prompt":    contextNote,
			"aiEnabled": false,
		}
	}

	client := resty.New().
		SetBaseURL(strings.TrimSpace(openAI.BaseUrl)).
		SetHeader("Authorization", "Bearer "+strings.TrimSpace(openAI.ApiKey)).
		SetHeader("Content-Type", "application/json")
	if openAI.TimeOut <= 0 {
		openAI.TimeOut = 180
	}
	client.SetTimeout(time.Duration(openAI.TimeOut) * time.Second)
	if openAI.HttpProxyEnabled && strings.TrimSpace(openAI.HttpProxy) != "" {
		client.SetProxy(strings.TrimSpace(openAI.HttpProxy))
	}

	chatMessages := make([]map[string]any, 0, len(messages)+2)
	chatMessages = append(chatMessages,
		map[string]any{
			"role":    "system",
			"content": a.resolveHouseholdChatSystemPrompt(promptTemplateId),
		},
		map[string]any{
			"role":    "system",
			"content": contextNote,
		},
		map[string]any{
			"role":    "assistant",
			"content": a.buildHouseholdChatAssistantPrimer(contextData),
		},
	)
	for _, item := range messages {
		role := strings.ToLower(strings.TrimSpace(item.Role))
		if role != "user" && role != "assistant" {
			continue
		}
		content := strings.TrimSpace(item.Content)
		if content == "" {
			continue
		}
		chatMessages = append(chatMessages, map[string]any{
			"role":    role,
			"content": content,
		})
	}

	body := map[string]any{
		"model":       openAI.Model,
		"max_tokens":  openAI.MaxTokens,
		"temperature": 0.35,
		"stream":      false,
		"messages":    chatMessages,
	}
	if openAI.MaxTokens <= 0 {
		body["max_tokens"] = 4096
	}

	resp, err := client.R().SetBody(body).Post("/chat/completions")
	if err != nil {
		return map[string]any{
			"success":   false,
			"message":   err.Error(),
			"reply":     "",
			"model":     openAI.Model,
			"prompt":    contextNote,
			"aiEnabled": true,
		}
	}
	if resp.StatusCode() >= 400 {
		return map[string]any{
			"success":   false,
			"message":   fmt.Sprintf("模型调用失败: HTTP %d %s", resp.StatusCode(), truncateText(resp.String(), 320)),
			"reply":     "",
			"model":     openAI.Model,
			"prompt":    contextNote,
			"aiEnabled": true,
		}
	}

	var aiResp data.AiResponse
	if err := json.Unmarshal(resp.Body(), &aiResp); err != nil {
		return map[string]any{
			"success":   false,
			"message":   "模型响应解析失败: " + err.Error(),
			"reply":     "",
			"model":     openAI.Model,
			"prompt":    contextNote,
			"aiEnabled": true,
		}
	}
	if len(aiResp.Choices) == 0 {
		return map[string]any{
			"success":   false,
			"message":   "模型没有返回对话内容",
			"reply":     "",
			"model":     openAI.Model,
			"prompt":    contextNote,
			"aiEnabled": true,
		}
	}

	return map[string]any{
		"success":   true,
		"message":   "数字分析对话完成",
		"reply":     strings.TrimSpace(aiResp.Choices[0].Message.Content),
		"model":     openAI.Model,
		"prompt":    contextNote,
		"aiEnabled": true,
	}
}

// --- 持仓分析 ---

func (a *App) GetAllHoldings() []portfolio.Holding {
	return a.PortfolioService.GetAllHoldings()
}

func (a *App) GetHoldingsByType(holdingType string) []portfolio.Holding {
	return a.PortfolioService.GetHoldingsByType(holdingType)
}

func (a *App) CreateHolding(h portfolio.Holding) *portfolio.Holding {
	return a.PortfolioService.CreateHolding(h)
}

func (a *App) UpsertFundHoldingByAmount(input portfolio.FundPositionInput) *portfolio.Holding {
	return a.PortfolioService.UpsertFundHoldingByAmount(input)
}

func (a *App) UpdateHolding(h portfolio.Holding) *portfolio.Holding {
	return a.PortfolioService.UpdateHolding(h)
}

func (a *App) DeleteHolding(id uint) bool {
	return a.PortfolioService.DeleteHolding(id)
}

func (a *App) AddTransaction(tx portfolio.Transaction) *portfolio.Transaction {
	return a.PortfolioService.AddTransaction(tx)
}

func (a *App) GetTransactions(stockCode string, page, pageSize int) ([]portfolio.Transaction, int64) {
	return a.PortfolioService.GetTransactions(stockCode, page, pageSize)
}

func (a *App) GetPortfolioSummary() *portfolio.PortfolioSummary {
	return a.PortfolioService.GetPortfolioSummary()
}

func (a *App) GetFundPortfolioDashboard() *portfolio.FundPortfolioDashboard {
	return a.PortfolioService.GetFundDashboard()
}

func (a *App) GetFundProfile(code string) *portfolio.FundProfile {
	return a.PortfolioService.GetFundProfile(code)
}

func (a *App) RefreshFundProfile(code string) *portfolio.FundProfile {
	return a.PortfolioService.RefreshFundProfile(code)
}

func (a *App) GetFundScreener(query portfolio.FundScreenerQuery) *portfolio.FundScreenerResult {
	return a.PortfolioService.GetFundScreener(query)
}

func (a *App) RefreshFundScreenerData(limit int) map[string]any {
	return a.PortfolioService.RefreshFundScreenerData(limit)
}

func (a *App) EnsureFundUniverse() int64 {
	return a.PortfolioService.EnsureFundUniverse()
}

func (a *App) GetBetterFunds(query portfolio.BetterFundQuery) *portfolio.BetterFundResult {
	return a.PortfolioService.GetBetterFundsCached(query)
}

func (a *App) GetFundRecommendationRefreshStatus(autoStart bool) portfolio.FundRecommendationRefreshStatus {
	return a.PortfolioService.GetFundRecommendationRefreshStatus(autoStart)
}

func (a *App) CompareFunds(query portfolio.FundCompareQuery) *portfolio.FundCompareResult {
	return a.PortfolioService.CompareFunds(query)
}

func (a *App) AnalyzeFundWithAI(code string, aiConfigId int) map[string]any {
	code = strings.TrimSpace(code)
	if code == "" {
		return map[string]any{"success": false, "message": "基金代码不能为空"}
	}

	profile := a.PortfolioService.GetFundProfile(code)
	if profile == nil {
		return map[string]any{"success": false, "message": "未找到基金资料"}
	}

	peers := a.PortfolioService.GetBetterFundsCached(portfolio.BetterFundQuery{
		ReferenceCode: code,
		SameTypeOnly:  true,
		FeeFree7:      true,
		FeeFree30:     true,
		IncludeAClass: false,
		Page:          1,
		PageSize:      5,
	})

	payload, _ := json.MarshalIndent(map[string]any{
		"fundProfile":      profile,
		"betterCandidates": peers,
	}, "", "  ")

	prompt := strings.Join([]string{
		"请分析下面这只基金，输出 Markdown。",
		"重点回答：",
		"1. 近7天、近1月、近3月、近6月和回撤是否匹配。",
		"2. 同类排名、回撤修复和风险收益特征。",
		"3. 如果候选替代基金更优，请说明为什么更优、适合什么场景。",
		"4. 不给明确买卖指令，只给风格和风险提示。",
		"基金数据如下：",
		string(payload),
	}, "\n")

	systemPrompt, templateName := a.resolveFundSystemPrompt("single")
	return a.runFundAIAnalysis(aiConfigId, systemPrompt, templateName, prompt)
}

func (a *App) AnalyzeFundCollectionWithAI(scope string, aiConfigId int) map[string]any {
	scope = strings.TrimSpace(scope)
	if scope == "" {
		scope = "holdings"
	}

	var payload any
	switch scope {
	case "watchlist":
		payload = map[string]any{
			"scope": "watchlist",
			"funds": data.NewFundApi().GetFollowedFund(),
		}
	default:
		payload = map[string]any{
			"scope":     "holdings",
			"dashboard": a.PortfolioService.GetFundDashboard(),
		}
	}

	body, _ := json.MarshalIndent(payload, "", "  ")
	prompt := strings.Join([]string{
		"请分析这组基金，输出 Markdown。",
		"重点回答：",
		"1. 当前组合的收益/回撤结构。",
		"2. 债基、现金管理、权益基金的配置是否均衡。",
		"3. 哪几只基金贡献了主要收益，哪几只拖累了表现。",
		"4. 给出后续观察重点，不要给明确买卖指令。",
		"数据如下：",
		string(body),
	}, "\n")

	systemPrompt, templateName := a.resolveFundSystemPrompt("collection")
	return a.runFundAIAnalysis(aiConfigId, systemPrompt, templateName, prompt)
}

func (a *App) AnalyzeBetterFundsWithAI(query portfolio.BetterFundQuery, topN int, aiConfigId int) map[string]any {
	query.ReferenceCode = strings.TrimSpace(query.ReferenceCode)
	if query.ReferenceCode == "" {
		return map[string]any{"success": false, "message": "基金代码不能为空"}
	}
	if topN <= 0 {
		topN = 3
	}
	if topN > 5 {
		topN = 5
	}
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize < topN {
		query.PageSize = max(topN, 8)
	}

	result := a.PortfolioService.GetBetterFundsCached(query)
	if result == nil || len(result.Candidates) == 0 {
		return map[string]any{"success": false, "message": "当前栏目暂无可供 AI 分析的推荐基金"}
	}

	limit := topN
	if len(result.Candidates) < limit {
		limit = len(result.Candidates)
	}
	candidates := result.Candidates[:limit]

	payload, _ := json.MarshalIndent(map[string]any{
		"dimension":        result.Dimension,
		"dimensionLabel":   betterDimensionLabel(result.Dimension),
		"sortLabel":        result.SortLabel,
		"scopeLabel":       result.ScopeLabel,
		"comparedUniverse": result.ComparedUniverse,
		"sameTypeOnly":     query.SameTypeOnly,
		"sameSubTypeOnly":  query.SameSubTypeOnly,
		"feeFree7":         query.FeeFree7,
		"feeFree30":        query.FeeFree30,
		"includeAClass":    query.IncludeAClass,
		"onlyAClass":       query.OnlyAClass,
		"dataHint":         result.DataHint,
		"referenceFund":    result.Reference,
		"topCandidates":    candidates,
	}, "", "  ")

	prompt := strings.Join([]string{
		fmt.Sprintf("请比较当前“%s”栏目下的 Top%d 推荐基金，并输出 Markdown。", betterDimensionLabel(result.Dimension), limit),
		"重点回答：",
		"1. 先概括参考基金和 Top 候选基金在当前栏目维度下的差异。",
		"2. 对每只候选基金分别说明优势、短板、适合什么观察场景。",
		"3. 解释收益、回撤、夏普、Calmar、同类位置这些指标里，哪些最支持当前排序。",
		"4. 如果只能优先跟踪 1 到 2 只，请明确写出理由。",
		"5. 不给明确买卖指令，只给风格适配、风险提示和后续观察重点。",
		"6. 只能基于输入数据分析，不要编造外部数据或官方排名。",
		"数据如下：",
		string(payload),
	}, "\n")

	systemPrompt, templateName := a.resolveFundSystemPrompt("recommendation")
	return a.runFundAIAnalysis(aiConfigId, systemPrompt, templateName, prompt)
}

func (a *App) SyncPortfolioQuotes() *portfolio.PortfolioSummary {
	return a.PortfolioService.SyncPortfolioQuotes()
}

func (a *App) GetProfitHistory(days int) []portfolio.ProfitSnapshot {
	return a.PortfolioService.GetProfitHistory(days)
}

func (a *App) SavePortfolioSnapshot() *portfolio.ProfitSnapshot {
	return a.PortfolioService.SaveAndReturnDailySnapshot()
}

func (a *App) runFundAIAnalysis(aiConfigId int, systemPrompt string, templateName string, userPrompt string) map[string]any {
	if aiConfigId <= 0 {
		configs := a.GetAiConfigs()
		if len(configs) > 0 {
			aiConfigId = int(configs[0].ID)
		}
	}

	openAI := data.NewDeepSeekOpenAi(a.ctx, aiConfigId)
	if strings.TrimSpace(openAI.BaseUrl) == "" || strings.TrimSpace(openAI.ApiKey) == "" || strings.TrimSpace(openAI.Model) == "" {
		return map[string]any{
			"success":   false,
			"message":   "AI 源未配置完整，请先在设置页补充接口地址、API Key 和模型名称",
			"analysis":  "",
			"prompt":    userPrompt,
			"aiEnabled": false,
			"template":  templateName,
		}
	}

	client := resty.New().
		SetBaseURL(strings.TrimSpace(openAI.BaseUrl)).
		SetHeader("Authorization", "Bearer "+strings.TrimSpace(openAI.ApiKey)).
		SetHeader("Content-Type", "application/json")
	if openAI.TimeOut <= 0 {
		openAI.TimeOut = 180
	}
	client.SetTimeout(time.Duration(openAI.TimeOut) * time.Second)
	if openAI.HttpProxyEnabled && strings.TrimSpace(openAI.HttpProxy) != "" {
		client.SetProxy(strings.TrimSpace(openAI.HttpProxy))
	}

	body := map[string]any{
		"model":       openAI.Model,
		"max_tokens":  4096,
		"temperature": 0.2,
		"stream":      false,
		"messages": []map[string]any{
			{
				"role":    "system",
				"content": systemPrompt,
			},
			{
				"role":    "user",
				"content": userPrompt,
			},
		},
	}
	if openAI.MaxTokens > 0 {
		body["max_tokens"] = openAI.MaxTokens
	}

	resp, err := client.R().SetBody(body).Post("/chat/completions")
	if err != nil {
		return map[string]any{
			"success":   false,
			"message":   err.Error(),
			"analysis":  "",
			"prompt":    userPrompt,
			"aiEnabled": true,
			"model":     openAI.Model,
			"template":  templateName,
		}
	}
	if resp.StatusCode() >= 400 {
		return map[string]any{
			"success":   false,
			"message":   fmt.Sprintf("模型调用失败: HTTP %d %s", resp.StatusCode(), truncateText(resp.String(), 320)),
			"analysis":  "",
			"prompt":    userPrompt,
			"aiEnabled": true,
			"model":     openAI.Model,
			"template":  templateName,
		}
	}

	var aiResp data.AiResponse
	if err := json.Unmarshal(resp.Body(), &aiResp); err != nil {
		return map[string]any{
			"success":   false,
			"message":   "模型响应解析失败: " + err.Error(),
			"analysis":  "",
			"prompt":    userPrompt,
			"aiEnabled": true,
			"model":     openAI.Model,
			"template":  templateName,
		}
	}
	if len(aiResp.Choices) == 0 {
		return map[string]any{
			"success":   false,
			"message":   "模型没有返回分析内容",
			"analysis":  "",
			"prompt":    userPrompt,
			"aiEnabled": true,
			"model":     openAI.Model,
			"template":  templateName,
		}
	}

	content := strings.TrimSpace(aiResp.Choices[0].Message.Content)
	return map[string]any{
		"success":   true,
		"message":   "分析完成",
		"analysis":  content,
		"prompt":    userPrompt,
		"aiEnabled": true,
		"model":     openAI.Model,
		"template":  templateName,
	}
}

// --- 量化模板 ---

func (a *App) GetQuantCategories() []quant.TemplateCategory {
	return a.QuantService.GetCategories()
}

func (a *App) CreateQuantCategory(c quant.TemplateCategory) *quant.TemplateCategory {
	return a.QuantService.CreateCategory(c)
}

func (a *App) GetQuantTemplates(categoryId uint, status string, page, pageSize int) ([]quant.Template, int64) {
	return a.QuantService.GetTemplates(categoryId, status, page, pageSize)
}

func (a *App) GetQuantTemplate(id uint) *quant.Template {
	return a.QuantService.GetTemplate(id)
}

func (a *App) CreateQuantTemplate(t quant.Template) *quant.Template {
	return a.QuantService.CreateTemplate(t)
}

func (a *App) UpdateQuantTemplate(t quant.Template) *quant.Template {
	return a.QuantService.UpdateTemplate(t)
}

func (a *App) DeleteQuantTemplate(id uint) bool {
	return a.QuantService.DeleteTemplate(id)
}

func (a *App) ActivateQuantTemplate(id uint) bool {
	return a.QuantService.ActivateTemplate(id)
}

func (a *App) ExportQuantTemplate(id uint) (string, error) {
	return a.QuantService.ExportTemplate(id)
}

func (a *App) BuildQuantPrompt(req quant.GenerateRequest) string {
	return a.QuantService.BuildGeneratePromptWithContext(req)
}

func (a *App) GenerateQuantCode(req quant.GenerateRequest, aiConfigId int) map[string]any {
	if aiConfigId <= 0 {
		configs := a.GetAiConfigs()
		if len(configs) > 0 {
			aiConfigId = int(configs[0].ID)
		}
	}

	openAI := data.NewDeepSeekOpenAi(a.ctx, aiConfigId)
	if strings.TrimSpace(openAI.BaseUrl) == "" || strings.TrimSpace(openAI.ApiKey) == "" || strings.TrimSpace(openAI.Model) == "" {
		return map[string]any{
			"success": false,
			"message": "当前 AI 源未完整配置，请先在设置中填写接口地址、API Key 和模型名称",
		}
	}

	prompt := a.QuantService.BuildGeneratePromptWithContext(req)
	systemPrompt := a.resolveQuantSystemPrompt(req.PromptTemplateID)

	messages := []map[string]any{
		{
			"role":    "system",
			"content": systemPrompt,
		},
		{
			"role":    "user",
			"content": prompt,
		},
	}

	client := resty.New().
		SetBaseURL(strings.TrimSpace(openAI.BaseUrl)).
		SetHeader("Authorization", "Bearer "+strings.TrimSpace(openAI.ApiKey)).
		SetHeader("Content-Type", "application/json")
	if openAI.TimeOut <= 0 {
		openAI.TimeOut = 180
	}
	client.SetTimeout(time.Duration(openAI.TimeOut) * time.Second)
	if openAI.HttpProxyEnabled && strings.TrimSpace(openAI.HttpProxy) != "" {
		client.SetProxy(strings.TrimSpace(openAI.HttpProxy))
	}

	body := map[string]any{
		"model":       openAI.Model,
		"max_tokens":  openAI.MaxTokens,
		"temperature": openAI.Temperature,
		"stream":      false,
		"messages":    messages,
	}
	if openAI.MaxTokens <= 0 {
		body["max_tokens"] = 4096
	}
	if openAI.Temperature < 0 {
		body["temperature"] = 0.2
	}

	resp, err := client.R().SetBody(body).Post("/chat/completions")
	if err != nil {
		return map[string]any{
			"success": false,
			"message": err.Error(),
			"prompt":  prompt,
			"model":   openAI.Model,
		}
	}
	if resp.StatusCode() >= 400 {
		return map[string]any{
			"success": false,
			"message": fmt.Sprintf("模型调用失败: HTTP %d %s", resp.StatusCode(), truncateText(resp.String(), 320)),
			"prompt":  prompt,
			"model":   openAI.Model,
		}
	}

	var aiResp data.AiResponse
	if err := json.Unmarshal(resp.Body(), &aiResp); err != nil {
		return map[string]any{
			"success": false,
			"message": "模型响应解析失败: " + err.Error(),
			"prompt":  prompt,
			"model":   openAI.Model,
		}
	}
	if len(aiResp.Choices) == 0 {
		return map[string]any{
			"success": false,
			"message": "模型没有返回内容",
			"prompt":  prompt,
			"model":   openAI.Model,
		}
	}

	content := aiResp.Choices[0].Message.Content
	code := extractPythonBlock(content)
	if code == "" {
		code = strings.TrimSpace(content)
	}

	return map[string]any{
		"success":       true,
		"message":       "生成成功",
		"code":          code,
		"rawContent":    content,
		"explanation":   extractMarkdownSection(content, "策略说明"),
		"riskWarning":   extractMarkdownSection(content, "风险提示"),
		"suggestedName": extractMarkdownSection(content, "建议名称"),
		"prompt":        prompt,
		"model":         openAI.Model,
	}
}

func (a *App) GetQuantTagTaxonomy() []quant.TagGroup {
	return a.QuantService.GetTagTaxonomy()
}

func (a *App) BuildQuantScriptSearchLinks(query string) []quant.SearchLink {
	return a.QuantService.BuildScriptSearchLinks(query)
}

func (a *App) SearchQuantScriptsWithAI(req quant.ScriptSearchRequest, aiConfigId int) map[string]any {
	hits, err := a.QuantService.SearchScriptSources(req)
	if err != nil {
		return map[string]any{
			"success": false,
			"message": err.Error(),
			"hits":    []quant.SearchHit{},
		}
	}

	if aiConfigId <= 0 {
		configs := a.GetAiConfigs()
		if len(configs) > 0 {
			aiConfigId = int(configs[0].ID)
		}
	}

	prompt := a.QuantService.BuildSearchAgentCandidatePrompt(req, hits)
	if aiConfigId <= 0 {
		return map[string]any{
			"success":   true,
			"message":   "未配置 AI 源，已返回聚合搜索结果",
			"hits":      hits,
			"analysis":  "",
			"prompt":    prompt,
			"model":     "",
			"aiEnabled": false,
		}
	}

	openAI := data.NewDeepSeekOpenAi(a.ctx, aiConfigId)
	if strings.TrimSpace(openAI.BaseUrl) == "" || strings.TrimSpace(openAI.ApiKey) == "" || strings.TrimSpace(openAI.Model) == "" {
		return map[string]any{
			"success":   true,
			"message":   "AI 源配置不完整，已返回聚合搜索结果",
			"hits":      hits,
			"analysis":  "",
			"prompt":    prompt,
			"model":     "",
			"aiEnabled": false,
		}
	}

	client := resty.New().
		SetBaseURL(strings.TrimSpace(openAI.BaseUrl)).
		SetHeader("Authorization", "Bearer "+strings.TrimSpace(openAI.ApiKey)).
		SetHeader("Content-Type", "application/json")
	if openAI.TimeOut <= 0 {
		openAI.TimeOut = 180
	}
	client.SetTimeout(time.Duration(openAI.TimeOut) * time.Second)
	if openAI.HttpProxyEnabled && strings.TrimSpace(openAI.HttpProxy) != "" {
		client.SetProxy(strings.TrimSpace(openAI.HttpProxy))
	}

	body := map[string]any{
		"model":       openAI.Model,
		"max_tokens":  openAI.MaxTokens,
		"temperature": 0.2,
		"stream":      false,
		"messages": []map[string]any{
			{
				"role":    "system",
				"content": "你是一名擅长量化平台脚本检索、归纳、筛选的研究助理。你需要从候选结果中快速找出最值得看的脚本来源，优先关注可运行性、平台适配、策略逻辑清晰度和研究价值。",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
	}
	if openAI.MaxTokens <= 0 {
		body["max_tokens"] = 2048
	}

	resp, err := client.R().SetBody(body).Post("/chat/completions")
	if err != nil {
		return map[string]any{
			"success":   true,
			"message":   "AI 检索总结失败，已返回聚合搜索结果: " + err.Error(),
			"hits":      hits,
			"analysis":  "",
			"prompt":    prompt,
			"model":     openAI.Model,
			"aiEnabled": false,
		}
	}
	if resp.StatusCode() >= 400 {
		return map[string]any{
			"success":   true,
			"message":   fmt.Sprintf("AI 检索总结失败: HTTP %d %s", resp.StatusCode(), truncateText(resp.String(), 240)),
			"hits":      hits,
			"analysis":  "",
			"prompt":    prompt,
			"model":     openAI.Model,
			"aiEnabled": false,
		}
	}

	var aiResp data.AiResponse
	if err := json.Unmarshal(resp.Body(), &aiResp); err != nil || len(aiResp.Choices) == 0 {
		return map[string]any{
			"success":   true,
			"message":   "AI 检索总结解析失败，已返回聚合搜索结果",
			"hits":      hits,
			"analysis":  "",
			"prompt":    prompt,
			"model":     openAI.Model,
			"aiEnabled": false,
		}
	}

	return map[string]any{
		"success":   true,
		"message":   "检索完成",
		"hits":      hits,
		"analysis":  strings.TrimSpace(aiResp.Choices[0].Message.Content),
		"prompt":    prompt,
		"model":     openAI.Model,
		"aiEnabled": true,
	}
}

func (a *App) ensureDefaultQuantPromptTemplate() {
	api := data.NewPromptTemplateApi()
	existing := api.GetPromptTemplates(defaultQuantPromptTemplateName, defaultQuantPromptTemplateType)
	if existing != nil && len(*existing) > 0 {
		return
	}

	result := api.AddPrompt(models.PromptTemplate{
		Name:    defaultQuantPromptTemplateName,
		Type:    defaultQuantPromptTemplateType,
		Content: defaultQuantPromptTemplateContent,
	})
	logger.SugaredLogger.Infof("seed quant prompt template: %s", result)
}

func (a *App) resolveQuantSystemPrompt(promptTemplateID int) string {
	templateContent := ""
	if promptTemplateID > 0 {
		templateContent = strings.TrimSpace(data.NewPromptTemplateApi().GetPromptTemplateByID(promptTemplateID))
	}
	if templateContent == "" {
		api := data.NewPromptTemplateApi()
		existing := api.GetPromptTemplates(defaultQuantPromptTemplateName, defaultQuantPromptTemplateType)
		if existing != nil && len(*existing) > 0 {
			templateContent = strings.TrimSpace((*existing)[0].Content)
		}
	}
	if templateContent == "" {
		return defaultQuantPromptTemplateContent
	}
	return templateContent
}

func (a *App) ensureDefaultHouseholdPromptTemplate() {
	api := data.NewPromptTemplateApi()
	existing := api.GetPromptTemplates(defaultHouseholdPromptTemplateName, defaultHouseholdPromptTemplateType)
	if existing != nil && len(*existing) > 0 {
		return
	}

	result := api.AddPrompt(models.PromptTemplate{
		Name:    defaultHouseholdPromptTemplateName,
		Type:    defaultHouseholdPromptTemplateType,
		Content: defaultHouseholdPromptTemplateContent,
	})
	logger.SugaredLogger.Infof("seed household prompt template: %s", result)
}

func (a *App) ensureDefaultHouseholdChatPromptTemplate() {
	api := data.NewPromptTemplateApi()
	existing := api.GetPromptTemplates(defaultHouseholdChatPromptTemplateName, defaultHouseholdPromptTemplateType)
	if existing != nil && len(*existing) > 0 {
		return
	}

	result := api.AddPrompt(models.PromptTemplate{
		Name:    defaultHouseholdChatPromptTemplateName,
		Type:    defaultHouseholdPromptTemplateType,
		Content: defaultHouseholdChatPromptTemplateContent,
	})
	logger.SugaredLogger.Infof("seed household chat prompt template: %s", result)
}

func (a *App) ensureDefaultFundPromptTemplates() {
	a.seedPromptTemplate(defaultFundAnalysisPromptTemplateName, defaultFundPromptTemplateType, defaultFundAnalysisSystemPrompt)
	a.seedPromptTemplate(defaultFundCollectionPromptTemplateName, defaultFundPromptTemplateType, defaultFundCollectionSystemPrompt)
	a.seedPromptTemplate(defaultFundRecommendationPromptTemplateName, defaultFundPromptTemplateType, defaultFundRecommendationSystemPrompt)
}

func (a *App) resolveHouseholdSystemPrompt(promptTemplateID int) string {
	templateContent := ""
	if promptTemplateID > 0 {
		templateContent = strings.TrimSpace(data.NewPromptTemplateApi().GetPromptTemplateByID(promptTemplateID))
	}
	if templateContent == "" {
		api := data.NewPromptTemplateApi()
		existing := api.GetPromptTemplates(defaultHouseholdPromptTemplateName, defaultHouseholdPromptTemplateType)
		if existing != nil && len(*existing) > 0 {
			templateContent = strings.TrimSpace((*existing)[0].Content)
		}
	}
	if templateContent == "" {
		return defaultHouseholdPromptTemplateContent
	}
	return templateContent
}

func (a *App) resolveHouseholdChatSystemPrompt(promptTemplateID int) string {
	templateContent := ""
	if promptTemplateID > 0 {
		templateContent = strings.TrimSpace(data.NewPromptTemplateApi().GetPromptTemplateByID(promptTemplateID))
	}
	if templateContent == "" {
		api := data.NewPromptTemplateApi()
		existing := api.GetPromptTemplates(defaultHouseholdChatPromptTemplateName, defaultHouseholdPromptTemplateType)
		if existing != nil && len(*existing) > 0 {
			templateContent = strings.TrimSpace((*existing)[0].Content)
		}
	}
	if templateContent == "" {
		return defaultHouseholdChatPromptTemplateContent
	}
	return templateContent
}

func (a *App) resolveFundSystemPrompt(mode string) (string, string) {
	defaultName := defaultFundAnalysisPromptTemplateName
	fallback := defaultFundAnalysisSystemPrompt
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "collection":
		defaultName = defaultFundCollectionPromptTemplateName
		fallback = defaultFundCollectionSystemPrompt
	case "recommendation":
		defaultName = defaultFundRecommendationPromptTemplateName
		fallback = defaultFundRecommendationSystemPrompt
	}

	api := data.NewPromptTemplateApi()
	existing := api.GetPromptTemplates(defaultName, defaultFundPromptTemplateType)
	if existing != nil && len(*existing) > 0 {
		templateContent := strings.TrimSpace((*existing)[0].Content)
		if templateContent != "" {
			return templateContent, defaultName
		}
	}
	return fallback, defaultName
}

func betterDimensionLabel(dimension string) string {
	switch strings.ToLower(strings.TrimSpace(dimension)) {
	case "lower_drawdown":
		return "回撤更低"
	case "higher_return":
		return "收益更高"
	default:
		return "实力均衡更优"
	}
}

func (a *App) buildHouseholdAnalysisPrompt(region string, inputPayload string) string {
	contextData := a.AssetService.BuildHouseholdAIContext(region)
	return strings.Join([]string{
		"请基于下面的家庭资产输入数据，输出一份结构化资产分析报告。",
		"要求：",
		"1. 必须覆盖总资产、净资产、流动资产、固定资产、总负债、负债率、月供压力、保障结构。",
		"2. 必须识别风险点和积极信号，并给出可执行建议。",
		"3. 必须结合输入里的 benchmark 数据，对用户在 " + region + " / 全国的收入与资产负债水平做谨慎对比，并给出“大致处于什么层级”的结论。",
		"4. 如果基准不足，请明确写“基准不足，不能下结论”。",
		"5. 输出结构固定为：核心结论、关键指标表、风险点、优化建议、地区/全国对比、后续关注项。",
		"6. 不要编造外部数据，不要给投资荐股结论。",
		"7. 若可推断，请在“地区/全国对比”中明确写出资产水平和负债率的大致层级；若不能推断，要明确说明基准不足。",
		"",
		"辅助层级判断提示：",
		a.buildHouseholdRankingHints(contextData),
		"",
		"输入数据(JSON)：",
		inputPayload,
	}, "\n")
}

func (a *App) buildHouseholdChatContextNote(contextData *asset.HouseholdAIContext, contextPayload string) string {
	return strings.Join([]string{
		"以下是本轮对话必须使用的家庭数字分析上下文。",
		"系统消息中的 FAMILY_CONTEXT_JSON 就是当前真实家庭数据，已经包含家庭成员年龄/职业/每个人的保障情况、全部资产细项、流动资金趋势、流动资金分布、负债明细、月收入明细、家庭月均支出、历史快照和地区基准。",
		"回答时必须优先引用这些数字，不能再说“用户没有提供家庭资产或负债数据”，除非 FAMILY_CONTEXT_JSON 里确实缺失。",
		"如果某项数据在 FAMILY_CONTEXT_JSON 中不存在，才可以说明“数据不足，不能下结论”。",
		"",
		"FAMILY_CONTEXT_JSON:",
		contextPayload,
	}, "\n")
}

func (a *App) buildHouseholdChatAssistantPrimer(contextData *asset.HouseholdAIContext) string {
	if contextData == nil || contextData.Summary == nil {
		return "我已接收当前家庭数字分析上下文；如果某项数据缺失，我会明确说明。"
	}
	memberCount := len(contextData.Members)
	if memberCount <= 0 {
		memberCount = 1
	}
	memberLines := make([]string, 0, len(contextData.MemberProfiles))
	for _, item := range contextData.MemberProfiles {
		protectionParts := make([]string, 0, 2)
		if item.ProtectionStatus.HasSocialInsurance {
			protectionParts = append(protectionParts, "有五险")
		}
		if item.ProtectionStatus.HasHousingFund {
			protectionParts = append(protectionParts, "有公积金")
		}
		if len(item.ProtectionStatus.CommercialCoverage) > 0 {
			protectionParts = append(protectionParts, "商业保障:"+strings.Join(item.ProtectionStatus.CommercialCoverage, "、"))
		}
		if len(protectionParts) == 0 {
			protectionParts = append(protectionParts, "保障未录入")
		}
		ageLabel := "年龄未录入"
		if item.Age > 0 {
			ageLabel = fmt.Sprintf("%d岁", item.Age)
		}
		memberLines = append(memberLines, fmt.Sprintf("%s(%s,%s,%s,%s)", item.Name, item.Relationship, ageLabel, fallbackText(item.Occupation, "职业未录入"), strings.Join(protectionParts, " / ")))
	}

	assetLines := make([]string, 0, len(contextData.AssetDetails))
	for _, item := range contextData.AssetDetails {
		assetLines = append(assetLines, fmt.Sprintf("%s[%s] %s，占总资产%.2f%%", item.Name, item.Category, formatMoneyForPrompt(item.Value), item.ShareOfTotal))
	}

	incomeLines := make([]string, 0, len(contextData.IncomeDetails))
	for _, item := range contextData.IncomeDetails {
		line := fmt.Sprintf("%s/%s：税前%s，税后%s，个税%s", fallbackText(item.Owner, item.Name), item.Type, formatMoneyForPrompt(item.MonthlyGross), formatMoneyForPrompt(item.MonthlyNet), formatMoneyForPrompt(item.MonthlyTax))
		if strings.TrimSpace(item.FormulaText) != "" {
			line += "；公式：" + item.FormulaText
		}
		incomeLines = append(incomeLines, line)
	}

	liabilityLines := make([]string, 0, len(contextData.LiabilityDetails))
	for _, item := range contextData.LiabilityDetails {
		liabilityLines = append(liabilityLines, fmt.Sprintf("%s[%s] 剩余本金%s，月供%s，利率%.2f%%，剩余%d个月", item.Name, item.Type, formatMoneyForPrompt(item.OutstandingPrincipal), formatMoneyForPrompt(item.MonthlyPayment), item.AnnualRate, item.RemainingMonths))
	}

	liquidTrendLines := make([]string, 0, len(contextData.LiquidAssetTrend))
	for _, item := range contextData.LiquidAssetTrend {
		liquidTrendLines = append(liquidTrendLines, fmt.Sprintf("%s 流动资产%s", item.Date, formatMoneyForPrompt(item.TotalLiquidAssets)))
	}

	liquidDistributionLines := make([]string, 0, len(contextData.LiquidAssetDistribution))
	for _, item := range contextData.LiquidAssetDistribution {
		liquidDistributionLines = append(liquidDistributionLines, fmt.Sprintf("%s[%s] %s，占流动资产%.2f%%", item.Name, item.AccountType, formatMoneyForPrompt(item.Balance), item.ShareOfLiquid))
	}

	monthlyHouseholdSpend := 0.0
	if contextData.Profile != nil {
		monthlyHouseholdSpend = contextData.Profile.MonthlyHouseholdSpend
	}
	monthlyFreeCashflow := contextData.Summary.MonthlyNetIncome - contextData.Summary.MonthlyEffectiveDebtPayment - monthlyHouseholdSpend

	return strings.Join([]string{
		fmt.Sprintf("我已读取当前家庭数据：总资产约%s，净资产约%s，总负债约%s，负债率约%.2f%%，月税后收入约%s，月供约%s，扣除公积金回流后的真实月供占用约%s，家庭月均支出约%s，可支配结余约%s；家庭成员 %d 人，地区 %s。",
			formatMoneyForPrompt(contextData.Summary.TotalAssets),
			formatMoneyForPrompt(contextData.Summary.NetAssets),
			formatMoneyForPrompt(contextData.Summary.TotalLiabilities),
			contextData.Summary.DebtRatio,
			formatMoneyForPrompt(contextData.Summary.MonthlyNetIncome),
			formatMoneyForPrompt(contextData.Summary.MonthlyDebtPayment),
			formatMoneyForPrompt(contextData.Summary.MonthlyEffectiveDebtPayment),
			formatMoneyForPrompt(monthlyHouseholdSpend),
			formatMoneyForPrompt(monthlyFreeCashflow),
			memberCount,
			contextData.Region,
		),
		"成员画像：" + joinOrFallback(memberLines),
		"全部资产明细：" + joinOrFallback(assetLines),
		"流动资金趋势：" + joinOrFallback(liquidTrendLines),
		"流动资金分布：" + joinOrFallback(liquidDistributionLines),
		"月收入明细：" + joinOrFallback(incomeLines),
		"负债明细：" + joinOrFallback(liabilityLines),
		"接下来我会基于这些上下文直接回答，不再要求你重复提供已录入的资产负债、成员和保障数据。",
	}, "\n")
}

func (a *App) buildHouseholdRankingHints(contextData *asset.HouseholdAIContext) string {
	if contextData == nil || contextData.Summary == nil {
		return "- 基准不足，不能下结论。"
	}

	members := 1
	if contextData.Profile != nil && contextData.Profile.MembersCount > 0 {
		members = contextData.Profile.MembersCount
	}
	netAssetsPerAdult := contextData.Summary.NetAssets / float64(members)
	debtRatio := contextData.Summary.DebtRatio

	lines := []string{
		a.buildAssetRankingHintLine("天津市", netAssetsPerAdult, contextData.Benchmarks),
		a.buildAssetRankingHintLine("全国", netAssetsPerAdult, contextData.Benchmarks),
		a.buildDebtRankingHintLine("天津市", debtRatio, contextData.Benchmarks),
		a.buildDebtRankingHintLine("全国", debtRatio, contextData.Benchmarks),
	}
	return strings.Join(lines, "\n")
}

func formatMoneyForPrompt(value float64) string {
	return fmt.Sprintf("%.2f 元", value)
}

func joinOrFallback(items []string) string {
	if len(items) == 0 {
		return "暂无明细"
	}
	return strings.Join(items, "；")
}

func fallbackText(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return strings.TrimSpace(value)
}

func (a *App) buildAssetRankingHintLine(region string, netAssetsPerAdult float64, benchmarks []asset.HouseholdBenchmark) string {
	if netAssetsPerAdult <= 0 {
		return "- " + region + "资产水平：基准不足，不能下结论。"
	}
	benchmark := findHouseholdBenchmark(benchmarks, region, []string{"天津居民人均可支配收入", "全国居民人均可支配收入"})
	if benchmark == nil || benchmark.Value <= 0 {
		return "- " + region + "资产水平：基准不足，不能下结论。"
	}
	multiple := netAssetsPerAdult / benchmark.Value
	level := approximateAssetLevel(multiple)
	return fmt.Sprintf("- %s资产水平：按人均净资产 %.2f 元 / 人均可支配收入 %.2f 元 = %.1f 倍，粗略对应 %s。", region, netAssetsPerAdult, benchmark.Value, multiple, level)
}

func (a *App) buildDebtRankingHintLine(region string, debtRatio float64, benchmarks []asset.HouseholdBenchmark) string {
	benchmark := findHouseholdBenchmark(benchmarks, region, []string{"天津负债率参考", "全国负债率参考", "天津贷款/存款余额比", "全国住户贷款/存款余额比"})
	if benchmark == nil || benchmark.Value <= 0 {
		return "- " + region + "负债率水平：基准不足，不能下结论。"
	}
	relative := debtRatio / benchmark.Value
	level := approximateDebtLevel(relative)
	return fmt.Sprintf("- %s负债率水平：当前负债率 %.2f%% / 代理杠杆 %.2f%% = %.2f，粗略对应 %s。", region, debtRatio, benchmark.Value, relative, level)
}

func findHouseholdBenchmark(items []asset.HouseholdBenchmark, region string, names []string) *asset.HouseholdBenchmark {
	nameSet := make(map[string]struct{}, len(names))
	for _, name := range names {
		nameSet[name] = struct{}{}
	}
	for _, item := range items {
		if item.Region != region {
			continue
		}
		if _, ok := nameSet[item.Name]; ok {
			record := item
			return &record
		}
	}
	return nil
}

func approximateAssetLevel(multiple float64) string {
	switch {
	case multiple < 1:
		return "约后 50%"
	case multiple < 2:
		return "约前 50%-35%"
	case multiple < 4:
		return "约前 35%-20%"
	case multiple < 8:
		return "约前 20%-10%"
	case multiple < 15:
		return "约前 10%-5%"
	default:
		return "约前 5%以内"
	}
}

func approximateDebtLevel(relative float64) string {
	switch {
	case relative <= 0.3:
		return "优于约 90%"
	case relative <= 0.6:
		return "优于约 75%"
	case relative <= 1:
		return "优于约 60%"
	case relative <= 1.4:
		return "接近中位"
	case relative <= 2:
		return "约后 40%"
	default:
		return "约后 20%"
	}
}

const defaultQuantPromptTemplateName = "量化脚本生成-标准模板"
const defaultQuantPromptTemplateType = "模型系统Prompt"
const defaultHouseholdPromptTemplateName = "家庭资产分析-标准模板"
const defaultHouseholdChatPromptTemplateName = "家庭数字分析-连续对话模板"
const defaultHouseholdPromptTemplateType = "模型系统Prompt"

const defaultFundAnalysisPromptTemplateName = "基金分析-标准模板"
const defaultFundCollectionPromptTemplateName = "基金组合分析-标准模板"
const defaultFundRecommendationPromptTemplateName = "基金对比推荐分析-标准模板"
const defaultFundPromptTemplateType = "模型系统Prompt"

var defaultQuantPromptTemplateContent = strings.Join([]string{
	"你是一名资深量化研究员、Python 量化工程师和交易系统架构师。",
	"",
	"你的唯一任务是：根据用户提供的策略目标、标签、场景和资金约束，输出一份可以直接运行的 Python 量化脚本草案，并补充简洁的策略说明。",
	"",
	"请严格遵守以下要求：",
	"1. 输出顺序必须固定：",
	"   - 先输出一个 ```python 代码块",
	"   - 再输出一个“## 策略说明”段落",
	"   - 再输出一个“## 风险提示”段落",
	"   - 再输出一个“## 建议名称”段落",
	"2. Python 代码必须尽量可直接运行，禁止输出伪代码、TODO、占位符实现、只写函数签名不写逻辑。",
	"3. 默认仅使用 Python 标准库、pandas、numpy；不要依赖聚宽、掘金、QMT、vnpy、tushare、akshare 等外部交易或数据 SDK。",
	"4. 如果用户提到某个平台，请在代码中通过适配器或注释留好接入点，但主体脚本仍需在本地环境下可独立运行。",
	"5. 如果缺少真实行情数据，请在脚本里自动提供一份最小可运行的样例数据或模拟数据，确保脚本运行后至少能完成：",
	"   - 数据准备",
	"   - 因子计算",
	"   - 信号生成",
	"   - 仓位管理",
	"   - 风险控制",
	"   - 回测或运行入口",
	"6. 代码必须包含：",
	"   - 清晰的参数区 CONFIG 或 dataclass",
	"   - 日志输出",
	"   - 核心策略类或核心主流程函数",
	"   - if __name__ == \"__main__\": 入口",
	"   - 结果打印或回测摘要输出",
	"7. 优先生成结构清晰、注释充分、便于二次修改的脚本，而不是追求复杂炫技。",
	"8. 如果用户需求存在歧义，请做最稳妥的工程化假设，不要反问。",
	"9. 所有中文说明要简洁，代码中的变量名、函数名、类名使用英文。",
	"10. “## 建议名称”只输出一个适合保存到策略库的名称，长度尽量控制在 10 到 24 个中文字符或等效英文长度。",
	"",
	"代码质量要求：",
	"- 允许使用面向对象或函数式结构，但必须自洽。",
	"- 需要给出买入、卖出、止损、仓位控制的基本逻辑。",
	"- 若用户给出情绪、量能、场景、资金规模等标签，要明确反映到参数和信号逻辑中。",
	"- 若用户给出股票代码，则优先围绕这些标的构造示例；若没有给出，则使用通用示例标的。",
	"- 不要输出 Markdown 表格。",
	"- 不要输出多段代码块。",
	"- 不要解释你为什么这么做，只按规定结构输出结果。",
}, "\n")

var defaultHouseholdPromptTemplateContent = strings.Join([]string{
	"你是家庭资产分析助手，需要基于用户提供的家庭资产、负债、收入、保障、历史快照和地区基准数据，输出严谨、可执行、可复核的分析报告。",
	"",
	"请严格遵守：",
	"1. 只基于输入数据分析，不要臆测未提供的数据。",
	"2. 输出必须使用中文，结构固定为：",
	"   - 核心结论",
	"   - 关键指标表",
	"   - 风险点",
	"   - 优化建议",
	"   - 地区/全国对比",
	"   - 后续关注项",
	"3. 若 benchmark 数据不足以支持天津市或全国对比，必须明确说明“基准不足，不能下结论”。",
	"4. 不要给出具体投资标的推荐，不要使用夸张措辞。",
	"5. 优化建议要优先围绕流动性、负债结构、保障配置、收入结构和资产集中度。",
}, "\n")

var defaultHouseholdChatPromptTemplateContent = strings.Join([]string{
	"你是家庭数字分析助手，需要基于系统消息中提供的家庭资产、负债、收入、保障、家庭成员、快照和地区基准数据进行连续对话。",
	"",
	"对话规则：",
	"1. 系统消息里的 FAMILY_CONTEXT_JSON 就是当前真实家庭数据，必须优先使用，不要忽略，也不要要求用户重复提供已经录入的资产负债数据。",
	"2. 回答必须使用中文，默认简洁、清楚、可执行。",
	"3. 如果用户问资产、负债、偿债压力、税后收入、流动性、保障、地区对比，优先引用 FAMILY_CONTEXT_JSON 里的实际数字，并明确点出数字来源。",
	"4. 只有在 FAMILY_CONTEXT_JSON 里确实缺少某项数据时，才可以写“数据不足，不能下结论”，然后说明还缺什么。",
	"5. 不给出具体股票、基金或高风险产品推荐，不夸张渲染。",
	"6. 可以结合天津市/全国 benchmark 做粗略比较，但要注明这是基于当前基准数据的辅助判断。",
	"7. 如果用户问后续动作，优先给出 3-5 条家庭财务层面的建议，围绕现金流、负债、保障和资产结构展开。",
}, "\n")

var defaultFundAnalysisSystemPrompt = strings.Join([]string{
	"你是一名偏审慎风格的基金研究助手，擅长分析公募基金的阶段收益、回撤、同类排名和风格适配。",
	"输出必须用简体中文 Markdown，结论要清楚，但不要制造确定性的买卖建议。",
	"优先解释收益和回撤是否匹配、同类位置是否稳定、适合什么资金属性。",
	"如果给出替代基金比较，只能基于输入数据说明优劣，不要编造外部数据。",
}, "\n")

var defaultFundCollectionSystemPrompt = strings.Join([]string{
	"你是一名偏审慎风格的基金组合顾问，负责总结基金持仓或基金自选的结构、收益来源、风险暴露和后续观察重点。",
	"输出必须用简体中文 Markdown。",
	"不要给出确定性的买卖建议，不要承诺收益，不要编造输入之外的数据。",
	"重点说明组合里谁贡献收益、谁拖累表现、回撤风险来自哪里、保守资产是否足够。",
}, "\n")

var defaultFundRecommendationSystemPrompt = strings.Join([]string{
	"你是一名偏审慎风格的基金比较分析助手，擅长基于给定的收益、回撤、夏普、Calmar、同类位置和推荐理由，比较多只基金谁更适合当前观察目标。",
	"输出必须使用简体中文 Markdown，先给结论，再解释依据。",
	"你只能基于输入数据分析，不要编造外部评级、官方排名、基金经理观点或未提供的行情数据。",
	"要把每只候选基金的优势、短板、适配场景写清楚，并解释为什么当前排序成立或不成立。",
	"不要给确定性的买卖指令，只给风险提示、观察重点和更适合的资金属性。",
}, "\n")

func extractPythonBlock(content string) string {
	re := regexp.MustCompile("(?s)```(?:python)?\\s*(.*?)```")
	match := re.FindStringSubmatch(content)
	if len(match) < 2 {
		return ""
	}
	return strings.TrimSpace(match[1])
}

func extractMarkdownSection(content string, title string) string {
	re := regexp.MustCompile(fmt.Sprintf(`(?s)(?:^|\n)#{1,3}\s*%s\s*(.*?)(?:(?:\n#{1,3}\s*)|$)`, regexp.QuoteMeta(title)))
	match := re.FindStringSubmatch(content)
	if len(match) < 2 {
		return ""
	}
	return strings.TrimSpace(match[1])
}

func truncateText(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	return value[:limit] + "..."
}

func extractJSONObject(content string) string {
	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start < 0 || end <= start {
		return ""
	}
	return strings.TrimSpace(content[start : end+1])
}

func (a *App) SendLocalNotification(title string, content string, cacheKey string) bool {
	title = strings.TrimSpace(title)
	content = strings.TrimSpace(content)
	cacheKey = strings.TrimSpace(cacheKey)
	if title == "" || content == "" {
		return false
	}
	config := a.GetConfig()
	if config != nil && !config.LocalPushEnable {
		return false
	}
	if cacheKey != "" {
		ttl, _ := a.cache.TTL([]byte(cacheKey))
		if ttl > 0 {
			return false
		}
		if err := a.cache.Set([]byte(cacheKey), []byte("1"), localNotificationTTLSeconds); err != nil {
			logger.SugaredLogger.Errorf("set local notification cache failed: %v", err)
		}
	}
	go data.NewAlertWindowsApi(title, "alert", content, "").SendNotification()
	return true
}

func (a *App) AnalyzeQuantLinkageWithAI(req quant.LinkageAIRequest, aiConfigId int) map[string]any {
	prompt := a.QuantService.BuildLinkageAIPrompt(req.Summary, req.Templates)

	if aiConfigId <= 0 {
		configs := a.GetAiConfigs()
		if len(configs) > 0 {
			aiConfigId = int(configs[0].ID)
		}
	}

	if aiConfigId <= 0 {
		return map[string]any{
			"success":   true,
			"message":   "未配置 AI 源，已跳过 AI 联动推荐分析",
			"analysis":  "",
			"parsed":    map[string]any{},
			"prompt":    prompt,
			"model":     "",
			"aiEnabled": false,
		}
	}

	openAI := data.NewDeepSeekOpenAi(a.ctx, aiConfigId)
	if strings.TrimSpace(openAI.BaseUrl) == "" || strings.TrimSpace(openAI.ApiKey) == "" || strings.TrimSpace(openAI.Model) == "" {
		return map[string]any{
			"success":   true,
			"message":   "AI 源配置不完整，已跳过 AI 联动推荐分析",
			"analysis":  "",
			"parsed":    map[string]any{},
			"prompt":    prompt,
			"model":     "",
			"aiEnabled": false,
		}
	}

	client := resty.New().
		SetBaseURL(strings.TrimSpace(openAI.BaseUrl)).
		SetHeader("Authorization", "Bearer "+strings.TrimSpace(openAI.ApiKey)).
		SetHeader("Content-Type", "application/json")
	if openAI.TimeOut <= 0 {
		openAI.TimeOut = 180
	}
	client.SetTimeout(time.Duration(openAI.TimeOut) * time.Second)
	if openAI.HttpProxyEnabled && strings.TrimSpace(openAI.HttpProxy) != "" {
		client.SetProxy(strings.TrimSpace(openAI.HttpProxy))
	}

	body := map[string]any{
		"model":       openAI.Model,
		"max_tokens":  openAI.MaxTokens,
		"temperature": 0.2,
		"stream":      false,
		"messages": []map[string]any{
			{
				"role":    "system",
				"content": "你是一名量化研究助理，需要在脚本联动推荐场景下给出严谨、可执行的脚本切换建议。你只能从候选脚本中选择，并且必须按要求输出 JSON。",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
	}
	if openAI.MaxTokens <= 0 {
		body["max_tokens"] = 2048
	}

	resp, err := client.R().SetBody(body).Post("/chat/completions")
	if err != nil {
		return map[string]any{
			"success":   true,
			"message":   "AI 联动推荐分析失败，已保留规则推荐结果: " + err.Error(),
			"analysis":  "",
			"parsed":    map[string]any{},
			"prompt":    prompt,
			"model":     openAI.Model,
			"aiEnabled": false,
		}
	}
	if resp.StatusCode() >= 400 {
		return map[string]any{
			"success":   true,
			"message":   fmt.Sprintf("AI 联动推荐分析失败: HTTP %d %s", resp.StatusCode(), truncateText(resp.String(), 240)),
			"analysis":  "",
			"parsed":    map[string]any{},
			"prompt":    prompt,
			"model":     openAI.Model,
			"aiEnabled": false,
		}
	}

	var aiResp data.AiResponse
	if err := json.Unmarshal(resp.Body(), &aiResp); err != nil || len(aiResp.Choices) == 0 {
		return map[string]any{
			"success":   true,
			"message":   "AI 联动推荐分析解析失败，已保留规则推荐结果",
			"analysis":  "",
			"parsed":    map[string]any{},
			"prompt":    prompt,
			"model":     openAI.Model,
			"aiEnabled": false,
		}
	}

	content := strings.TrimSpace(aiResp.Choices[0].Message.Content)
	parsed := map[string]any{}
	if jsonText := extractJSONObject(content); jsonText != "" {
		if err := json.Unmarshal([]byte(jsonText), &parsed); err != nil {
			logger.SugaredLogger.Warnf("parse linkage ai json failed: %v", err)
		}
	}

	return map[string]any{
		"success":   true,
		"message":   "联动推荐分析完成",
		"analysis":  content,
		"parsed":    parsed,
		"prompt":    prompt,
		"model":     openAI.Model,
		"aiEnabled": true,
	}
}

func (a *App) ensureDefaultStockPromptTemplates() {
	for _, item := range defaultStockPromptTemplates {
		a.seedPromptTemplate(item.Name, item.Type, item.Content)
	}
}

func (a *App) seedPromptTemplate(name string, promptType string, content string) {
	api := data.NewPromptTemplateApi()
	existing := api.GetPromptTemplates(name, promptType)
	if existing != nil && len(*existing) > 0 {
		return
	}

	result := api.AddPrompt(models.PromptTemplate{
		Name:    name,
		Type:    promptType,
		Content: content,
	})
	logger.SugaredLogger.Infof("seed prompt template [%s/%s]: %s", promptType, name, result)
}

var defaultStockPromptTemplates = []models.PromptTemplate{
	{
		Name: "股票分析-专业研究框架",
		Type: "模型系统Prompt",
		Content: strings.Join([]string{
			"你是一名专业股票研究员、行业分析师和风险控制顾问。",
			"",
			"你的任务是基于用户给出的股票名称、代码和上下文数据，输出结构清晰、可执行、可复核的中文股票分析报告。",
			"",
			"输出要求：",
			"1. 不要复述用户问题，不要输出“用户要求我分析”这类开场白。",
			"2. 直接进入分析正文，使用 Markdown 标题分层输出。",
			"3. 优先给出结论摘要，再展开论证。",
			"4. 分析必须明确区分：事实、推断、风险、交易应对。",
			"5. 不确定的数据要明确写“若当前数据未完整更新，则以下判断基于已有公开信息推断”。",
			"6. 不做收益承诺，不使用夸张营销措辞。",
			"",
			"建议结构：",
			"# 股票名称（代码）分析报告",
			"## 一、核心结论",
			"## 二、公司与业务概况",
			"## 三、基本面与财务质量",
			"## 四、行业位置与竞争优势",
			"## 五、估值与预期差",
			"## 六、技术面与交易节奏",
			"## 七、催化剂与风险点",
			"## 八、操作建议",
			"",
			"写作要求：",
			"- 每个二级标题下尽量使用 3 到 6 个要点。",
			"- 如果出现估值、趋势、景气度等判断，必须说明驱动逻辑。",
			"- 如果信息足够，可附“短线 / 中线 / 长线”三种观察视角。",
			"- 结尾必须附上“风险提示”。",
		}, "\n"),
	},
	{
		Name: "股票分析-深度研究报告",
		Type: "模型用户Prompt",
		Content: strings.Join([]string{
			"请按照“专业卖方/买方研究报告”的方式分析 {{stockName}}（{{stockCode}}）。",
			"",
			"重点覆盖：",
			"1. 公司主营业务、收入结构、盈利来源和关键经营指标。",
			"2. 最近几个报告期的营收、利润、毛利率、现金流、负债与资本开支变化。",
			"3. 当前行业景气度、竞争格局、公司护城河和相对同行位置。",
			"4. 当前估值水平处于什么区间，市场预期差可能在哪里。",
			"5. 未来 1 到 4 个季度值得跟踪的催化剂和核心风险。",
			"6. 给出结论：适合观察、低吸、右侧跟随，还是暂时回避。",
			"",
			"输出要求：",
			"- 先给“核心结论”和“投资看点”。",
			"- 再展开基本面、估值、风险和交易策略。",
			"- 语言专业，但不要空话，不要套模板式废话。",
			"- 如果某些关键数据无法确认，请明确说明缺失项及其对判断的影响。",
		}, "\n"),
	},
	{
		Name: "股票分析-护城河与估值判断",
		Type: "模型用户Prompt",
		Content: strings.Join([]string{
			"请使用“护城河 + 资本回报 + 估值 + 不确定性”的框架分析 {{stockName}}（{{stockCode}}）。",
			"",
			"请重点回答：",
			"1. 这家公司是否具备品牌、成本、网络效应、转换成本或规模优势中的一种或多种护城河？",
			"2. 护城河是增强、维持还是削弱？主要证据是什么？",
			"3. 公司的盈利质量、自由现金流、资本开支压力和 ROE / ROIC 表现如何？",
			"4. 当前估值大致偏贵、合理还是低估？背后的关键假设是什么？",
			"5. 不确定性主要来自行业周期、政策、商品价格、管理层执行还是财务结构？",
			"6. 更适合中长期配置，还是只适合阶段性观察？",
			"",
			"输出要求：",
			"- 结论必须清楚写出“护城河判断、估值判断、不确定性等级”。",
			"- 用条目化方式写，不要大段空泛叙述。",
			"- 如有必要，可补充“触发重新评估的关键变量”。",
		}, "\n"),
	},
	{
		Name: "股票分析-技术面与交易计划",
		Type: "模型用户Prompt",
		Content: strings.Join([]string{
			"请站在短中线交易员视角，分析 {{stockName}}（{{stockCode}}）当前是否具备可执行的交易机会。",
			"",
			"请重点覆盖：",
			"1. 趋势状态：上升、震荡、下跌，处于主升、回踩还是反弹阶段？",
			"2. 量价关系、关键均线、支撑阻力、阶段高低点的意义。",
			"3. 若结合事件、业绩、政策或板块轮动，当前交易逻辑是否成立。",
			"4. 给出更具体的交易计划：关注价位、确认信号、止损位、止盈位、仓位建议。",
			"5. 明确什么情况下继续跟踪，什么情况下放弃。",
			"",
			"输出要求：",
			"- 先给一句话交易判断。",
			"- 再按“趋势、量价、催化、计划、风险”输出。",
			"- 如果没有明显机会，要明确说明“当前不建议交易”的理由。",
			"- 结尾补一段“适合什么类型的交易者”。",
		}, "\n"),
	},
}
