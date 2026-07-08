package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	stockagent "go-stock/backend/agent"
	"go-stock/backend/data"
	"go-stock/backend/db"
	"go-stock/backend/logger"
	"go-stock/backend/machineid"
	"go-stock/backend/models"

	"github.com/duke-git/lancet/v2/convertor"
	"github.com/robfig/cron/v3"
	"github.com/samber/lo"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) setCronEntry(key string, id cron.EntryID) {
	a.cronEntrysMu.Lock()
	a.cronEntrys[key] = id
	a.cronEntrysMu.Unlock()
}

func (a *App) getCronEntry(key string) (cron.EntryID, bool) {
	a.cronEntrysMu.Lock()
	id, exists := a.cronEntrys[key]
	a.cronEntrysMu.Unlock()
	return id, exists
}

func (a *App) removeCronEntry(key string) {
	a.cronEntrysMu.Lock()
	delete(a.cronEntrys, key)
	a.cronEntrysMu.Unlock()
}

func (a *App) GetEffectiveSponsorVip() map[string]any {
	level, active := data.EffectiveSponsorVipLevel()
	return map[string]any{"vipLevel": level, "active": active}
}

func (a *App) GetMachineId() string {
	return machineid.GetMachineId()
}

func (a *App) CheckDeviceBinding(token string, apiBase string) map[string]any {
	uuid := a.GetMachineId()
	result := map[string]any{"bound": false, "deviceCount": 0, "maxDevices": 5}
	if token == "" || apiBase == "" {
		return result
	}
	url := fmt.Sprintf("%s/user/device-check?uuid=%s", strings.TrimRight(apiBase, "/"), uuid)
	resp, err := data.SharedHTTPClient.R().SetHeader("Authorization", "Bearer "+token).Get(url)
	if err != nil {
		return result
	}
	var respData struct {
		Code int `json:"code"`
		Data struct {
			Bound       bool `json:"bound"`
			DeviceCount int  `json:"deviceCount"`
			MaxDevices  int  `json:"maxDevices"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp.Body(), &respData); err != nil || respData.Code != 0 {
		return result
	}
	result["bound"] = respData.Data.Bound
	result["deviceCount"] = respData.Data.DeviceCount
	result["maxDevices"] = respData.Data.MaxDevices
	return result
}

func (a *App) PromptPlazaRequest(method, apiBase, path string, query map[string]any, body, token string) map[string]any {
	result := map[string]any{"code": -1, "message": "", "data": nil}
	if apiBase == "" {
		result["message"] = "apiBase is empty"
		return result
	}
	url := strings.TrimRight(apiBase, "/") + path
	req := data.SharedHTTPClient.R().SetHeader("Content-Type", "application/json")
	if token != "" {
		req = req.SetHeader("Authorization", "Bearer "+token)
	}
	if len(query) > 0 {
		params := make(map[string]string, len(query))
		for k, v := range query {
			if v == nil {
				continue
			}
			s := fmt.Sprintf("%v", v)
			if s != "" {
				params[k] = s
			}
		}
		req = req.SetQueryParams(params)
	}
	if body != "" {
		req = req.SetBody(body)
	}
	resp, err := req.Execute(strings.ToUpper(method), url)
	if err != nil {
		result["message"] = err.Error()
		return result
	}
	var respData map[string]any
	if err := json.Unmarshal(resp.Body(), &respData); err != nil {
		result["message"] = "response parse failed: " + err.Error()
		return result
	}
	if _, ok := respData["code"]; !ok {
		respData["code"] = -1
	}
	return respData
}

func (a *App) QuitApp() {
	if a.ctx != nil {
		if a.cron != nil {
			a.cron.Stop()
		}
		runtime.Quit(a.ctx)
	}
}

func (a *App) AbortChatWithAgent() {
	a.cancelAgentTask()
	runtime.EventsEmit(a.ctx, "agent-message", map[string]any{
		"role":          "assistant",
		"response_meta": map[string]any{"finish_reason": "stop"},
	})
}

func (a *App) AbortSummaryStockNews() {
	a.cancelSummaryTask()
	runtime.EventsEmit(a.ctx, "summaryStockNews", "DONE")
}

func (a *App) setAgentCancel(cancel context.CancelFunc) uint64 {
	a.agentCancelMu.Lock()
	if a.agentCancel != nil {
		a.agentCancel()
	}
	a.agentCancelSeq++
	a.agentCancel = cancel
	token := a.agentCancelSeq
	a.agentCancelMu.Unlock()
	return token
}

func (a *App) clearAgentCancel(token uint64) {
	a.agentCancelMu.Lock()
	if a.agentCancelSeq == token {
		a.agentCancel = nil
	}
	a.agentCancelMu.Unlock()
}

func (a *App) cancelAgentTask() {
	a.agentCancelMu.Lock()
	cancel := a.agentCancel
	a.agentCancel = nil
	a.agentCancelMu.Unlock()
	if cancel != nil {
		cancel()
	}
}

func (a *App) setSummaryCancel(cancel context.CancelFunc) uint64 {
	a.agentCancelMu.Lock()
	if a.summaryCancel != nil {
		a.summaryCancel()
	}
	a.summaryCancelSeq++
	a.summaryCancel = cancel
	token := a.summaryCancelSeq
	a.agentCancelMu.Unlock()
	return token
}

func (a *App) clearSummaryCancel(token uint64) {
	a.agentCancelMu.Lock()
	if a.summaryCancelSeq == token {
		a.summaryCancel = nil
	}
	a.agentCancelMu.Unlock()
}

func (a *App) cancelSummaryTask() {
	a.agentCancelMu.Lock()
	cancel := a.summaryCancel
	a.summaryCancel = nil
	a.agentCancelMu.Unlock()
	if cancel != nil {
		cancel()
	}
}

func (a *App) ShareText(text, title string) string {
	text = strings.TrimSpace(text)
	title = strings.TrimSpace(title)
	if text == "" {
		return "内容为空"
	}
	if title == "" {
		title = "AI助手"
	}
	response, err := data.SharedHTTPClient.R().SetHeader("ua-x", "go-stock").SetFormData(map[string]string{
		"text":         text,
		"stockCode":    title,
		"stockName":    title,
		"analysisTime": time.Now().Format("2006/01/02"),
	}).Post("http://go-stock.sparkmemory.top:16688/upload")
	if err != nil {
		return err.Error()
	}
	return response.String()
}

func (a *App) SetTradingPrice(stockCode string, entryPrice, takeProfitPrice, stopLossPrice, costPrice float64) string {
	return data.NewStockDataApi().SetTradingPrice(entryPrice, takeProfitPrice, stopLossPrice, costPrice, stockCode)
}

func (a *App) GetFundHistoryNetValue(fundCode string, pageSize int, startDate string, endDate string) []data.FundHistoryNetValue {
	res, _ := data.NewFundApi().GetFundHistoryNetValue(fundCode, 1, pageSize, startDate, endDate)
	if res == nil {
		return []data.FundHistoryNetValue{}
	}
	return res
}

func (a *App) GetFundRanking(marketType, fundType, sortField, sortOrder string, pageIndex, pageSize int) map[string]any {
	result, err := data.NewFundApi().GetFundRanking(marketType, fundType, sortField, sortOrder, pageIndex, pageSize)
	if err != nil || result == nil {
		logger.SugaredLogger.Errorf("GetFundRanking failed: %v", err)
		if pageIndex <= 0 {
			pageIndex = 1
		}
		if pageSize <= 0 {
			pageSize = 50
		}
		return map[string]any{
			"items":      []data.FundRankingItem{},
			"totalCount": 0,
			"pageIndex":  pageIndex,
			"pageSize":   pageSize,
			"totalPages": 0,
		}
	}
	return map[string]any{
		"items":      result.Items,
		"totalCount": result.TotalCount,
		"pageIndex":  result.PageIndex,
		"pageSize":   result.PageSize,
		"totalPages": result.TotalPages,
	}
}

func (a *App) SearchFundCodes(keyword string) []map[string]string {
	searchItems := data.NewFundApi().SearchFundCodes(keyword)
	items := make([]map[string]string, 0, len(searchItems))
	for _, fund := range searchItems {
		items = append(items, map[string]string{
			"code": fund.Code,
			"name": fund.Name,
			"type": fund.Type,
		})
	}
	return items
}

func (a *App) GetFollowedFundPaged(pageIndex, pageSize int, keyword string) map[string]any {
	result := data.NewFundApi().GetFollowedFundPaged(pageIndex, pageSize, keyword)
	if result == nil {
		result = &data.FollowedFundPagedResult{}
	}
	return map[string]any{
		"items":      result.Items,
		"totalCount": result.TotalCount,
		"pageIndex":  result.PageIndex,
		"pageSize":   result.PageSize,
		"totalPages": result.TotalPages,
	}
}

func (a *App) CreateCronTask(task *models.CronTask) string {
	if err := stockagent.NewCronTaskApi().Create(task); err != nil {
		return fmt.Sprintf("创建失败: %v", err)
	}
	return a.scheduleCronTask(task, "创建成功")
}

func (a *App) UpdateCronTask(task *models.CronTask) string {
	if err := stockagent.NewCronTaskApi().Update(task); err != nil {
		return fmt.Sprintf("更新失败: %v", err)
	}
	if entryID, exists := a.getCronEntry(convertor.ToString(task.ID) + "_" + task.Name); exists {
		a.cron.Remove(entryID)
	}
	return a.scheduleCronTask(task, "更新成功")
}

func (a *App) DeleteCronTask(id uint) string {
	task, _ := stockagent.NewCronTaskApi().GetByID(id)
	if err := stockagent.NewCronTaskApi().Delete(id); err != nil {
		return fmt.Sprintf("删除失败: %v", err)
	}
	if task != nil {
		if entryID, exists := a.getCronEntry(convertor.ToString(id) + "_" + task.Name); exists {
			a.cron.Remove(entryID)
			a.removeCronEntry(convertor.ToString(id) + "_" + task.Name)
		}
	}
	return "删除成功"
}

func (a *App) GetCronTaskByID(id uint) *models.CronTask {
	task, err := stockagent.NewCronTaskApi().GetByID(id)
	if err != nil {
		return nil
	}
	return task
}

func (a *App) GetCronTaskList(query *models.CronTaskQuery) *models.CronTaskPageResp {
	return stockagent.NewCronTaskApi().List(query)
}

func (a *App) EnableCronTask(id uint, enable bool) string {
	if err := stockagent.NewCronTaskApi().EnableTask(id, enable); err != nil {
		return fmt.Sprintf("操作失败: %v", err)
	}
	task, err := stockagent.NewCronTaskApi().GetByID(id)
	if err == nil && task != nil {
		key := convertor.ToString(id) + "_" + task.Name
		if entryID, exists := a.getCronEntry(key); exists {
			a.cron.Remove(entryID)
			a.removeCronEntry(key)
		}
		if enable {
			return a.scheduleCronTask(task, "操作成功")
		}
	}
	return "操作成功"
}

func (a *App) ExecuteCronTaskNow(id uint) string {
	task, err := stockagent.NewCronTaskApi().GetByID(id)
	if err != nil {
		return fmt.Sprintf("任务不存在: %v", err)
	}
	go func() {
		if err := stockagent.NewCronTaskApi().ExecuteTask(a.ctx, task); err != nil {
			logger.SugaredLogger.Errorf("execute cron task failed: %v %s", err, task.Name)
		}
	}()
	return "任务执行中"
}

func (a *App) GetCronTaskTypes() []lo.Tuple2[string, string] {
	return stockagent.NewCronTaskApi().GetTaskTypes()
}

func (a *App) ValidateCronExpr(expr string) string {
	if err := stockagent.NewCronTaskApi().ValidateCronExpr(expr); err != nil {
		return fmt.Sprintf("无效表达式: %v", err)
	}
	return "有效表达式"
}

func (a *App) SearchCronTasks(keyword string) []models.CronTask {
	return stockagent.NewCronTaskApi().SearchTasks(keyword)
}

func (a *App) CalculateNextRunTime(cronExpr string) string {
	return stockagent.NewCronTaskApi().CalculateNextRunTime(cronExpr).Format("2006-01-02 15:04:05")
}

func (a *App) CalculateNextRunTimes(cronExpr string, count int) []string {
	times := stockagent.NewCronTaskApi().CalculateNextRunTimes(cronExpr, count)
	result := make([]string, 0, len(times))
	for _, t := range times {
		result = append(result, t.Format("2006-01-02 15:04:05"))
	}
	return result
}

func (a *App) scheduleCronTask(task *models.CronTask, okMsg string) string {
	if task == nil {
		return "任务为空"
	}
	taskCopy := *task
	entryID, err := a.cron.AddFunc(taskCopy.CronExpr, func() {
		if err := stockagent.NewCronTaskApi().ExecuteTask(a.ctx, &taskCopy); err != nil {
			logger.SugaredLogger.Errorf("execute cron task failed: %v %s", err, taskCopy.Name)
		}
	})
	if err != nil {
		return okMsg + ", 但定时调度失败"
	}
	a.setCronEntry(convertor.ToString(taskCopy.ID)+"_"+taskCopy.Name, entryID)
	return okMsg
}

func (a *App) AddTradingRecord(record data.TradingRecord) (uint, error) {
	return data.NewStockDataApi().AddTradingRecord(record)
}

func (a *App) GetTradingRecordList(query data.TradingRecordListQuery) *data.TradingRecordPageData {
	page, err := data.NewStockDataApi().GetTradingRecordList(query)
	if err != nil {
		return &data.TradingRecordPageData{}
	}
	return page
}

func (a *App) GetTradingRecordStatistics() *data.TradingRecordStatistics {
	stats, err := data.NewStockDataApi().GetTradingRecordStatistics()
	if err != nil {
		return &data.TradingRecordStatistics{}
	}
	return stats
}

func (a *App) UpdateTradingRecord(record data.TradingRecord) error {
	return data.NewStockDataApi().UpdateTradingRecord(record)
}

func (a *App) DeleteTradingRecord(id uint) error {
	return data.NewStockDataApi().DeleteTradingRecord(id)
}

func (a *App) CheckFrequentTrading(stockCode string) map[string]any {
	canTrade, msg := data.NewStockDataApi().CheckFrequentTrading(stockCode)
	return map[string]any{"canTrade": canTrade, "msg": msg}
}

func (a *App) GetStockRealTimePrice(stockCode string) map[string]any {
	list, err := data.NewStockDataApi().GetStockCodeRealTimeData(stockCode)
	if err != nil || list == nil || len(*list) == 0 {
		return map[string]any{"price": 0, "success": false}
	}
	price, _ := convertor.ToFloat((*list)[0].Price)
	return map[string]any{"price": price, "success": true, "data": (*list)[0]}
}

func (a *App) GetAllCustomStrategies() *[]models.CustomStrategy {
	return data.NewCustomStrategyApi().GetAllCustomStrategies()
}

func (a *App) SaveCustomStrategy(strategy models.CustomStrategy) string {
	return data.NewCustomStrategyApi().SaveCustomStrategy(strategy)
}

func (a *App) DeleteCustomStrategy(id uint) string {
	return data.NewCustomStrategyApi().DeleteCustomStrategy(id)
}

func (a *App) UpdateAiRecommendStocksAlert(id uint, enableAlert bool) string {
	if err := db.Dao.Model(&models.AiRecommendStocks{}).Where("id = ?", id).Update("enable_alert", enableAlert).Error; err != nil {
		return "更新失败"
	}
	return "更新成功"
}

func (a *App) GetBKFundFlowListByDate(code string, date string) []models.BKFundFlowPoint {
	return data.NewBKFundFlowApi().GetBKFundFlowListByDate(code, date)
}

func (a *App) GetBKFundFlowTopListByDate(date string, topN int) []models.BKFundFlow {
	return data.NewBKFundFlowApi().GetBKFundFlowTopListByDate(date, topN)
}

func (a *App) FetchBKFundFlowNow() int {
	count, err := data.NewBKFundFlowApi().FetchAndSave()
	if err != nil {
		logger.SugaredLogger.Errorf("FetchBKFundFlowNow failed: %v", err)
		return 0
	}
	return count
}

func (a *App) GetAllBKCodes() []map[string]string {
	return data.NewBKFundFlowApi().GetAllBKCodes()
}

func (a *App) GetConceptFundFlowListByDate(code string, date string) []models.ConceptFundFlowPoint {
	return data.NewConceptFundFlowApi().GetConceptFundFlowListByDate(code, date)
}

func (a *App) GetConceptFundFlowTopListByDate(date string, topN int) []models.ConceptFundFlow {
	return data.NewConceptFundFlowApi().GetConceptFundFlowTopListByDate(date, topN)
}

func (a *App) FetchConceptFundFlowNow() int {
	count, err := data.NewConceptFundFlowApi().FetchAndSave()
	if err != nil {
		logger.SugaredLogger.Errorf("FetchConceptFundFlowNow failed: %v", err)
		return 0
	}
	return count
}

func (a *App) GetAllConceptCodes() []map[string]string {
	return data.NewConceptFundFlowApi().GetAllConceptCodes()
}

func (a *App) GetDailyChangeStats(days int) []data.DailyChangeStats {
	res, _ := data.NewStockChangeHistoryService().GetDailyChangeStats(days)
	return res
}

func (a *App) GetChangeTypeDailyStats(days int) []data.ChangeTypeDailyStats {
	res, _ := data.NewStockChangeHistoryService().GetChangeTypeDailyStats(days)
	return res
}

func (a *App) GetChangeRank(days int, topN int) *data.ChangeRankResult {
	res, _ := data.NewStockChangeHistoryService().GetChangeRank(days, topN)
	if res == nil {
		return &data.ChangeRankResult{}
	}
	return res
}

func (a *App) GetDailyDimensionStats(dimension string, name string, days int) []data.DailyDimensionStats {
	res, _ := data.NewStockChangeHistoryService().GetDailyDimensionStats(dimension, name, days)
	return res
}

func (a *App) GetTypeStatsByDate(date string) []data.TypeCountStats {
	res, _ := data.NewStockChangeHistoryService().GetTypeStatsByDate(date)
	return res
}

func (a *App) GetUplimitHot(date string, limit int) map[string]any {
	return data.NewMarketNewsApi().GetUplimitHot(date, limit)
}

func (a *App) IsTradingTime() bool {
	return isTradingTime(time.Now())
}

func (a *App) IsHKTradingTime() bool {
	return IsHKTradingTime(time.Now())
}

func (a *App) IsUSTradingTime() bool {
	return IsUSTradingTime(time.Now())
}

func (a *App) IsTradingDay(date string) bool {
	if date == "" {
		return isTradingDay(time.Now())
	}
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return false
	}
	return isTradingDay(t)
}

func (a *App) GetLatestTradingDay() string {
	t := time.Now()
	for !isTradingDay(t) {
		t = t.AddDate(0, 0, -1)
	}
	return t.Format("2006-01-02")
}
