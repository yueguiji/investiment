package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	stockagent "go-stock/backend/agent"
	"go-stock/backend/data"
	"go-stock/backend/db"
	"go-stock/backend/logger"
	"go-stock/backend/models"
	"go-stock/backend/runtimeconfig"

	"github.com/PuerkitoBio/goquery"
	"github.com/duke-git/lancet/v2/convertor"
	"github.com/duke-git/lancet/v2/mathutil"
	lancetslice "github.com/duke-git/lancet/v2/slice"
	"github.com/duke-git/lancet/v2/strutil"
	"github.com/go-resty/resty/v2"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func GetImageBase(src []byte) string {
	return "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(src)
}

func GenNotificationMsg(stockInfo *data.StockInfo) string {
	price, _ := convertor.ToFloat(stockInfo.Price)
	preClose, _ := convertor.ToFloat(stockInfo.PreClose)
	var rf float64
	if preClose > 0 {
		rf = mathutil.RoundToFloat(((price - preClose) / preClose * 100), 2)
	}
	return "[" + stockInfo.Name + "] " + stockInfo.Price + " " + convertor.ToString(rf) + "% " + stockInfo.Date + " " + stockInfo.Time
}

func getMsgTypeTTL(msgType int) int {
	switch msgType {
	case 1:
		return 60 * 5
	case 2, 3:
		return 60 * 30
	default:
		return 60 * 5
	}
}

func getMsgTypeName(msgType int) string {
	switch msgType {
	case 1:
		return "涨跌报警"
	case 2:
		return "股价报警"
	case 3:
		return "成本价报警"
	default:
		return "报警"
	}
}

func isTradingDay(date time.Time) bool {
	weekday := date.Weekday()
	return weekday != time.Saturday && weekday != time.Sunday
}

func isTradingTime(date time.Time) bool {
	if !isTradingDay(date) {
		return false
	}
	hour, minute, _ := date.Clock()
	if (hour == 9 && minute >= 15) || hour == 10 || (hour == 11 && minute <= 30) {
		return true
	}
	return hour == 13 || hour == 14 || (hour == 15 && minute == 0)
}

func IsHKTradingTime(date time.Time) bool {
	hour, minute, _ := date.Clock()
	if hour == 9 && minute <= 30 {
		return true
	}
	if (hour == 9 && minute > 30) || (hour >= 10 && hour < 12) || (hour == 12 && minute == 0) {
		return true
	}
	if hour == 13 || (hour >= 14 && hour < 16) || (hour == 16 && minute <= 10) {
		return true
	}
	return false
}

func IsUSTradingTime(date time.Time) bool {
	est, err := time.LoadLocation("America/New_York")
	estTime := date.Add(-12 * time.Hour)
	if err == nil {
		estTime = date.In(est)
	}
	weekday := estTime.Weekday()
	if weekday == time.Saturday || weekday == time.Sunday {
		return false
	}
	hour, minute, _ := estTime.Clock()
	if hour >= 4 && hour < 9 {
		return true
	}
	if hour == 9 && minute < 30 {
		return true
	}
	if (hour == 9 && minute >= 30) || (hour >= 10 && hour < 16) || (hour == 16 && minute == 0) {
		return true
	}
	return (hour == 16 && minute > 0) || (hour >= 17 && hour < 20) || (hour == 20 && minute == 0)
}

func addStockFollowData(follow data.FollowedStock, stockData *data.StockInfo) {
	stockData.PrePrice = follow.Price
	stockData.Sort = follow.Sort
	stockData.CostPrice = follow.CostPrice
	stockData.CostVolume = follow.Volume
	stockData.AlarmChangePercent = follow.AlarmChangePercent
	stockData.AlarmPrice = follow.AlarmPrice
	stockData.Groups = follow.Groups

	price, _ := convertor.ToFloat(stockData.Price)
	if price == 0 {
		price, _ = convertor.ToFloat(stockData.A1P)
	}
	if price == 0 {
		price, _ = convertor.ToFloat(stockData.B1P)
	}
	preClosePrice, _ := convertor.ToFloat(stockData.PreClose)
	if price == 0 {
		price = preClosePrice
	}
	highPrice, _ := convertor.ToFloat(stockData.High)
	if highPrice == 0 {
		highPrice, _ = convertor.ToFloat(stockData.Open)
	}
	lowPrice, _ := convertor.ToFloat(stockData.Low)
	if lowPrice == 0 {
		lowPrice, _ = convertor.ToFloat(stockData.Open)
	}

	if price > 0 && preClosePrice > 0 {
		stockData.ChangePrice = mathutil.RoundToFloat(price-preClosePrice, 2)
		stockData.ChangePercent = mathutil.RoundToFloat(mathutil.Div(price-preClosePrice, preClosePrice)*100, 3)
	}
	if highPrice > 0 && preClosePrice > 0 {
		stockData.HighRate = mathutil.RoundToFloat(mathutil.Div(highPrice-preClosePrice, preClosePrice)*100, 3)
	}
	if lowPrice > 0 && preClosePrice > 0 {
		stockData.LowRate = mathutil.RoundToFloat(mathutil.Div(lowPrice-preClosePrice, preClosePrice)*100, 3)
	}
	if follow.CostPrice > 0 && follow.Volume > 0 {
		if price > 0 {
			stockData.Profit = mathutil.RoundToFloat(mathutil.Div(price-follow.CostPrice, follow.CostPrice)*100, 3)
			stockData.ProfitAmount = mathutil.RoundToFloat((price-follow.CostPrice)*float64(follow.Volume), 2)
			stockData.ProfitAmountToday = mathutil.RoundToFloat((price-preClosePrice)*float64(follow.Volume), 2)
		} else {
			stockData.Profit = mathutil.RoundToFloat(mathutil.Div(preClosePrice-follow.CostPrice, follow.CostPrice)*100, 3)
			stockData.ProfitAmount = mathutil.RoundToFloat((preClosePrice-follow.CostPrice)*float64(follow.Volume), 2)
			stockData.ProfitAmountToday = 0
		}
	}

	if follow.Price != price && price > 0 {
		go db.Dao.Model(follow).Where("stock_code = ?", follow.StockCode).Updates(map[string]interface{}{"price": price})
	}
}

func getStockInfo(follow data.FollowedStock) *data.StockInfo {
	stockDatas, err := data.NewStockDataApi().GetStockCodeRealTimeData(follow.StockCode)
	if err != nil || len(*stockDatas) == 0 {
		return &data.StockInfo{}
	}
	stockData := (*stockDatas)[0]
	addStockFollowData(follow, &stockData)
	return &stockData
}

func GetStockInfos(follows ...data.FollowedStock) *[]data.StockInfo {
	stockInfos := make([]data.StockInfo, 0)
	stockCodes := make([]string, 0)
	for _, follow := range follows {
		if strutil.HasPrefixAny(follow.StockCode, []string{"SZ", "SH", "sh", "sz"}) && !isTradingTime(time.Now()) {
			continue
		}
		if strutil.HasPrefixAny(follow.StockCode, []string{"hk", "HK"}) && !IsHKTradingTime(time.Now()) {
			continue
		}
		if strutil.HasPrefixAny(follow.StockCode, []string{"us", "US", "gb_"}) && !IsUSTradingTime(time.Now()) {
			continue
		}
		stockCodes = append(stockCodes, follow.StockCode)
	}
	stockData, _ := data.NewStockDataApi().GetStockCodeRealTimeData(stockCodes...)
	for _, info := range *stockData {
		v, ok := lancetslice.FindBy(follows, func(_ int, follow data.FollowedStock) bool {
			if strutil.HasPrefixAny(follow.StockCode, []string{"US", "us"}) {
				return strings.ToLower(strings.Replace(follow.StockCode, "us", "gb_", 1)) == info.Code
			}
			return follow.StockCode == info.Code
		})
		if ok {
			addStockFollowData(v, &info)
			stockInfos = append(stockInfos, info)
		}
	}
	return &stockInfos
}

func (a *App) Greet(stockCode string) *data.StockInfo {
	follow := &data.FollowedStock{StockCode: stockCode}
	db.Dao.Model(follow).Where("stock_code = ?", stockCode).Preload("Groups").Preload("Groups.GroupInfo").First(follow)
	return getStockInfo(*follow)
}

func (a *App) Follow(stockCode string) string   { return data.NewStockDataApi().Follow(stockCode) }
func (a *App) UnFollow(stockCode string) string { return data.NewStockDataApi().UnFollow(stockCode) }
func (a *App) GetFollowList(groupId int) *[]data.FollowedStock {
	return data.NewStockDataApi().GetFollowList(groupId)
}
func (a *App) GetStockList(key string) []data.StockBasic {
	return data.NewStockDataApi().GetStockList(key)
}
func (a *App) SetCostPriceAndVolume(stockCode string, price float64, volume int64) string {
	return data.NewStockDataApi().SetCostPriceAndVolume(price, volume, stockCode)
}
func (a *App) SetAlarmChangePercent(val, alarmPrice float64, stockCode string) string {
	return data.NewStockDataApi().SetAlarmChangePercent(val, alarmPrice, stockCode)
}
func (a *App) SetStockSort(sort int64, stockCode string) {
	data.NewStockDataApi().SetStockSort(sort, stockCode)
}

func (a *App) SendDingDingMessage(message string, stockCode string) string {
	ttl, _ := a.cache.TTL([]byte(stockCode))
	logger.SugaredLogger.Infof("stockCode %s ttl:%d", stockCode, ttl)
	if ttl > 0 {
		return ""
	}
	if err := a.cache.Set([]byte(stockCode), []byte("1"), 60*5); err != nil {
		logger.SugaredLogger.Errorf("set cache error: %s", err.Error())
		return ""
	}
	return data.NewDingDingAPI().SendDingDingMessage(message)
}

func (a *App) SendDingDingMessageByType(message string, stockCode string, msgType int) string {
	if strutil.HasPrefixAny(stockCode, []string{"SZ", "SH", "sh", "sz"}) && !isTradingTime(time.Now()) {
		return "not in A-share trading hours"
	}
	if strutil.HasPrefixAny(stockCode, []string{"hk", "HK"}) && !IsHKTradingTime(time.Now()) {
		return "not in HK trading hours"
	}
	if strutil.HasPrefixAny(stockCode, []string{"us", "US", "gb_"}) && !IsUSTradingTime(time.Now()) {
		return "not in US trading hours"
	}
	ttl, _ := a.cache.TTL([]byte(stockCode))
	if ttl > 0 {
		return ""
	}
	if err := a.cache.Set([]byte(stockCode), []byte("1"), getMsgTypeTTL(msgType)); err != nil {
		logger.SugaredLogger.Errorf("set cache error: %s", err.Error())
		return ""
	}
	stockInfo := &data.StockInfo{}
	db.Dao.Model(stockInfo).Where("code = ?", stockCode).First(stockInfo)
	go data.NewAlertWindowsApi("go-stock message", "alert", GenNotificationMsg(stockInfo), "").SendNotification()
	return data.NewDingDingAPI().SendDingDingMessage(message)
}

func (a *App) NewChatStream(stock, stockCode, question string, aiConfigId int, sysPromptId *int, enableTools bool, think bool) {
	var msgs <-chan map[string]any
	if enableTools {
		msgs = data.NewDeepSeekOpenAi(a.ctx, aiConfigId).NewChatStream(stock, stockCode, question, sysPromptId, a.AiTools, think)
	} else {
		msgs = data.NewDeepSeekOpenAi(a.ctx, aiConfigId).NewChatStream(stock, stockCode, question, sysPromptId, []data.Tool{}, think)
	}
	for msg := range msgs {
		runtime.EventsEmit(a.ctx, "newChatStream", msg)
	}
	runtime.EventsEmit(a.ctx, "newChatStream", "DONE")
}

func (a *App) SaveAIResponseResult(stockCode, stockName, result, chatId, question string, aiConfigId int) {
	data.NewDeepSeekOpenAi(a.ctx, aiConfigId).SaveAIResponseResult(stockCode, stockName, result, chatId, question)
}

func (a *App) GetAIResponseResult(stock string) *models.AIResponseResult {
	res := data.NewDeepSeekOpenAi(a.ctx, 0).GetAIResponseResult(stock)
	if res == nil {
		return &models.AIResponseResult{}
	}
	return res
}

func (a *App) GetVersionInfo() *models.VersionInfo {
	return &models.VersionInfo{
		Version:            Version,
		Icon:               GetImageBase(icon),
		Alipay:             GetImageBase(alipay),
		Wxpay:              GetImageBase(wxpay),
		Wxgzh:              GetImageBase(wxgzh),
		Content:            VersionCommit,
		OfficialStatement:  OFFICIAL_STATEMENT,
		DanmuWebsocketURL:  runtimeconfig.Current().ResolveDanmuWebsocketURL(),
		MessageWallURL:     runtimeconfig.Current().ResolveMessageWallURL(),
		AssetUnlockEnabled: data.AssetUnlockEnabled(),
	}
}

func (a *App) VerifyAssetUnlockPassword(password string) bool {
	return data.VerifyAssetUnlockPassword(password)
}

func (a *App) UpdateConfig(settingConfig *data.SettingConfig) string {
	return data.UpdateConfig(settingConfig)
}

func (a *App) ExportConfig() string {
	return data.NewSettingsApi().Export()
}

func (a *App) GetConfig() *data.SettingConfig {
	cfg := data.GetSettingConfig()
	if cfg == nil {
		return &data.SettingConfig{}
	}
	if len(cfg.AiConfigs) == 0 {
		cfg.AiConfigs = []*data.AIConfig{{Name: "Default", ModelName: "default"}}
	}
	return cfg
}

func (a *App) CheckSponsorCode(sponsorCode string) map[string]any {
	if strings.TrimSpace(sponsorCode) == "" {
		return map[string]any{
			"code": 0,
			"msg":  "赞助码不能为空",
		}
	}
	return map[string]any{
		"code": 1,
		"msg":  "褰撳墠鐗堟湰鏈惎鐢ㄨ禐鍔╃爜鏍￠獙锛屽凡鍏佽淇濆瓨",
	}
}

func (a *App) CheckUpdate(flag int) string {
	return "当前版本未启用在线更新检查"
}

func (a *App) GetSponsorInfo() map[string]any { return map[string]any{} }

func (a *App) ShareAnalysis(stockCode, stockName string) string {
	res := data.NewDeepSeekOpenAi(a.ctx, 0).GetAIResponseResult(stockCode)
	if res == nil || len(res.Content) <= 100 {
		return "analysis content is empty"
	}
	analysisTime := res.CreatedAt.Format("2006/01/02")
	response, err := resty.New().SetHeader("ua-x", "go-stock").R().SetFormData(map[string]string{
		"text":         res.Content,
		"stockCode":    stockCode,
		"stockName":    stockName,
		"analysisTime": analysisTime,
	}).Post(runtimeconfig.Current().ResolveShareUploadURL())
	if err != nil {
		return err.Error()
	}
	return response.String()
}

func (a *App) SaveAsMarkdown(stockCode, stockName string) string {
	res := data.NewDeepSeekOpenAi(a.ctx, 0).GetAIResponseResult(stockCode)
	if res == nil || len(res.Content) <= 100 {
		return "analysis content is empty"
	}
	analysisTime := res.CreatedAt.Format("2006-01-02_15_04_05")
	file, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Save Markdown",
		DefaultFilename: fmt.Sprintf("%s[%s]_%s.md", stockName, stockCode, analysisTime),
		Filters:         []runtime.FileFilter{{DisplayName: "Markdown", Pattern: "*.md;*.markdown"}},
	})
	if err != nil {
		return err.Error()
	}
	if err := os.WriteFile(file, []byte(res.Content), 0644); err != nil {
		return err.Error()
	}
	return "saved to " + file
}

func (a *App) SaveImage(name, base64Data string) string {
	filePath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Save Image",
		DefaultFilename: name + "_analysis.png",
		Filters:         []runtime.FileFilter{{DisplayName: "PNG", Pattern: "*.png"}},
	})
	if err != nil || filePath == "" {
		return "save canceled"
	}
	decodeString, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return "invalid image data"
	}
	if err := os.WriteFile(filepath.Clean(filePath), decodeString, os.ModePerm); err != nil {
		return err.Error()
	}
	return filePath
}

func (a *App) SaveWordFile(filename string, base64Data string) string {
	filePath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Save Word",
		DefaultFilename: filename,
		Filters:         []runtime.FileFilter{{DisplayName: "Word", Pattern: "*.docx"}},
	})
	if err != nil || filePath == "" {
		return "save canceled"
	}
	decodeString, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return "invalid word data"
	}
	if err := os.WriteFile(filepath.Clean(filePath), decodeString, 0777); err != nil {
		return err.Error()
	}
	return filePath
}

func (a *App) GetPromptTemplates(name, promptType string) *[]models.PromptTemplate {
	res := data.NewPromptTemplateApi().GetPromptTemplates(name, promptType)
	if res == nil {
		empty := []models.PromptTemplate{}
		return &empty
	}
	return res
}

func (a *App) SetStockAICron(cronText, stockCode string) {
	data.NewStockDataApi().SetStockAICron(cronText, stockCode)
	stockCode = strings.TrimSpace(stockCode)
	if strutil.HasPrefixAny(stockCode, []string{"gb_"}) {
		stockCode = strings.ToUpper(stockCode)
		stockCode = strings.Replace(stockCode, "gb_", "us", 1)
		stockCode = strings.Replace(stockCode, "GB_", "us", 1)
	}
	if entryID, exists := a.getCronEntry(stockCode); exists {
		a.cron.Remove(entryID)
		a.removeCronEntry(stockCode)
	}
	follow := data.NewStockDataApi().GetFollowedStockByStockCode(stockCode)
	if follow.StockCode == "" || strings.TrimSpace(cronText) == "" {
		return
	}
	id, err := a.cron.AddFunc(cronText, a.AddCronTask(follow))
	if err != nil {
		logger.SugaredLogger.Errorf("add stock AI cron failed: %s cron=%s err=%v", follow.StockCode, cronText, err)
		return
	}
	a.setCronEntry(stockCode, id)
}

func (a *App) InitCronTasks() {
	if a.cron == nil {
		return
	}
	followList := data.NewStockDataApi().GetFollowList(0)
	if followList != nil {
		for _, follow := range *followList {
			if follow.Cron == nil || strings.TrimSpace(*follow.Cron) == "" || strings.TrimSpace(follow.StockCode) == "" {
				continue
			}
			if entryID, exists := a.getCronEntry(follow.StockCode); exists {
				a.cron.Remove(entryID)
				a.removeCronEntry(follow.StockCode)
			}
			entryID, err := a.cron.AddFunc(*follow.Cron, a.AddCronTask(follow))
			if err != nil {
				logger.SugaredLogger.Errorf("add stock AI cron failed: %s cron=%s err=%v", follow.StockCode, *follow.Cron, err)
				continue
			}
			a.setCronEntry(follow.StockCode, entryID)
		}
	}
	if _, exists := a.getCronEntry("CheckStockBaseInfo"); !exists {
		entryID, err := a.cron.AddFunc("0 0 2 * * *", func() {
			a.CheckStockBaseInfo(a.ctx)
		})
		if err == nil {
			a.setCronEntry("CheckStockBaseInfo", entryID)
		} else {
			logger.SugaredLogger.Errorf("add CheckStockBaseInfo cron failed: %v", err)
		}
	}
}

func (a *App) CheckStockBaseInfo(ctx context.Context) {
	defer func() {
		go runtime.EventsEmit(ctx, "loadingMsg", "done")
	}()

	stockBasics := &[]data.StockBasic{}
	_, err := resty.New().R().
		SetHeader("user", "go-stock").
		SetResult(stockBasics).
		Get(runtimeconfig.Current().ResolveStockBasicURL())
	if err != nil {
		logger.SugaredLogger.Errorf("fetch StockBasic failed: %v", err)
	} else {
		for _, stock := range *stockBasics {
			if strings.TrimSpace(stock.TsCode) == "" {
				continue
			}
			var existing data.StockBasic
			if err := db.Dao.Where("ts_code = ?", stock.TsCode).Assign(stock).FirstOrCreate(&existing).Error; err != nil {
				logger.SugaredLogger.Errorf("upsert StockBasic failed ts_code=%s: %v", stock.TsCode, err)
			}
		}
	}

	stockHKBasics := &[]models.StockInfoHK{}
	_, err = resty.New().R().
		SetHeader("user", "go-stock").
		SetResult(stockHKBasics).
		Get(runtimeconfig.Current().ResolveStockBaseInfoHKURL())
	if err != nil {
		logger.SugaredLogger.Errorf("fetch StockInfoHK failed: %v", err)
	} else {
		for _, stock := range *stockHKBasics {
			if strings.TrimSpace(stock.Code) == "" {
				continue
			}
			var existing models.StockInfoHK
			if err := db.Dao.Where("code = ?", stock.Code).Assign(stock).FirstOrCreate(&existing).Error; err != nil {
				logger.SugaredLogger.Errorf("upsert StockInfoHK failed code=%s: %v", stock.Code, err)
			}
		}
	}

	stockUSBasics := &[]models.StockInfoUS{}
	_, err = resty.New().R().
		SetHeader("user", "go-stock").
		SetResult(stockUSBasics).
		Get(runtimeconfig.Current().ResolveStockBaseInfoUSURL())
	if err != nil {
		logger.SugaredLogger.Errorf("fetch StockInfoUS failed: %v", err)
	} else {
		for _, stock := range *stockUSBasics {
			if strings.TrimSpace(stock.Code) == "" {
				continue
			}
			var existing models.StockInfoUS
			if err := db.Dao.Where("code = ?", stock.Code).Assign(stock).FirstOrCreate(&existing).Error; err != nil {
				logger.SugaredLogger.Errorf("upsert StockInfoUS failed code=%s: %v", stock.Code, err)
			}
		}
	}
}

func (a *App) NewsPush(news *[]models.Telegraph) {
	if news == nil {
		return
	}
	follows := data.NewStockDataApi().GetFollowList(0)
	stockNames := make([]string, 0)
	if follows != nil {
		for _, item := range *follows {
			if strings.TrimSpace(item.Name) != "" {
				stockNames = append(stockNames, item.Name)
			}
		}
	}
	config := a.GetConfig()
	for _, telegraph := range *news {
		if config.EnableOnlyPushRedNews {
			if telegraph.IsRed || strutil.ContainsAny(telegraph.Content, stockNames) {
				go runtime.EventsEmit(a.ctx, "newsPush", telegraph)
			}
			continue
		}
		go runtime.EventsEmit(a.ctx, "newsPush", telegraph)
	}
}

func (a *App) AddCronTask(follow data.FollowedStock) func() {
	return func() {
		go runtime.EventsEmit(a.ctx, "warnMsg", "开始自动分析"+follow.Name+"_"+follow.StockCode)
		ai := data.NewDeepSeekOpenAi(a.ctx, follow.AiConfigId)
		msgs := ai.NewChatStream(follow.Name, follow.StockCode, "", nil, a.AiTools, true)
		var res strings.Builder
		chatId := ""
		question := ""
		for msg := range msgs {
			if value, ok := msg["extraContent"].(string); ok {
				res.WriteString(value)
				res.WriteString("\n")
			}
			if value, ok := msg["content"].(string); ok {
				res.WriteString(value)
			}
			if value, ok := msg["chatId"].(string); ok {
				chatId = value
			}
			if value, ok := msg["question"].(string); ok {
				question = value
			}
		}
		data.NewDeepSeekOpenAi(a.ctx, follow.AiConfigId).SaveAIResponseResult(follow.StockCode, follow.Name, res.String(), chatId, question)
		go runtime.EventsEmit(a.ctx, "warnMsg", "AI分析完成："+follow.Name+"_"+follow.StockCode)
	}
}
func (a *App) AddGroup(group data.Group) string {
	if data.NewStockGroupApi(db.Dao).AddGroup(group) {
		return "success"
	}
	return "failed"
}
func (a *App) GetGroupList() []data.Group { return data.NewStockGroupApi(db.Dao).GetGroupList() }
func (a *App) UpdateGroupSort(id int, newSort int) bool {
	return data.NewStockGroupApi(db.Dao).UpdateGroupSort(id, newSort)
}
func (a *App) InitializeGroupSort() bool { return data.NewStockGroupApi(db.Dao).InitializeGroupSort() }
func (a *App) GetGroupStockList(groupId int) []data.GroupStock {
	return data.NewStockGroupApi(db.Dao).GetGroupStockByGroupId(groupId)
}
func (a *App) GetAllGroupStocks() []data.GroupStock {
	return data.NewStockGroupApi(db.Dao).GetAllGroupStocks()
}
func (a *App) AddStockGroup(groupId int, stockCode string) string {
	if data.NewStockGroupApi(db.Dao).AddStockGroup(groupId, stockCode) {
		return "success"
	}
	return "failed"
}
func (a *App) RemoveStockGroup(code, name string, groupId int) string {
	if data.NewStockGroupApi(db.Dao).RemoveStockGroup(code, name, groupId) {
		return "success"
	}
	return "failed"
}
func (a *App) RemoveGroup(groupId int) string {
	if data.NewStockGroupApi(db.Dao).RemoveGroup(groupId) {
		return "success"
	}
	return "failed"
}
func (a *App) UpdateGroup(id int, name string) string {
	if data.NewStockGroupApi(db.Dao).UpdateGroup(id, name) {
		return "success"
	}
	return "failed"
}

func (a *App) GetStockKLine(stockCode, stockName string, days int64) *[]data.KLineData {
	return data.NewStockDataApi().GetStockKLineWithFallback(stockCode, "day", days)
}
func (a *App) GetStockCommonKLine(stockCode, stockName string, days int64) *[]data.KLineData {
	return data.NewStockDataApi().GetCommonKLineData(stockCode, "day", days)
}

func (a *App) GetStockEastMoneyKLine(stockCode, stockName string, klt string, limit int) *[]data.KLineData {
	return a.GetStockEastMoneyKLinePage(stockCode, stockName, klt, limit, "")
}

func (a *App) GetStockEastMoneyKLinePage(stockCode, stockName string, klt string, limit int, end string) *[]data.KLineData {
	if limit <= 0 {
		limit = 500
	}
	if limit > 5000 {
		limit = 5000
	}
	klt = strings.TrimSpace(klt)
	if klt == "" {
		klt = "1"
	}
	end = strings.TrimSpace(end)
	return data.NewEastMoneyKLineApi(data.GetSettingConfig()).GetKLineDataBefore(stockCode, klt, "", limit, end)
}

func (a *App) GetStockKLineWithFallback(stockCode, stockName string, klt string, limit int, adjustFlag string) *data.KLineSourceResult {
	if limit <= 0 {
		limit = 500
	}
	if limit > 5000 {
		limit = 5000
	}
	klt = strings.TrimSpace(klt)
	if klt == "" {
		klt = "101"
	}
	return data.FetchKLineWithFallback(stockCode, stockName, klt, limit, "", adjustFlag)
}

func (a *App) GetStockKLinePageWithFallback(stockCode, stockName string, klt string, limit int, end string, adjustFlag string) *data.KLineSourceResult {
	if limit <= 0 {
		limit = 500
	}
	if limit > 5000 {
		limit = 5000
	}
	klt = strings.TrimSpace(klt)
	if klt == "" {
		klt = "101"
	}
	return data.FetchKLineWithFallback(stockCode, stockName, klt, limit, strings.TrimSpace(end), adjustFlag)
}

func (a *App) GetChipDistribution(stockCode string, days int, bins int, adjustFlag string) (*data.ChipDistributionResult, error) {
	stockCode = strings.TrimSpace(stockCode)
	if stockCode == "" {
		return nil, fmt.Errorf("stockCode cannot be empty")
	}
	if days <= 0 {
		days = 120
	}
	if bins <= 0 {
		bins = 80
	}
	adjustFlag = strings.TrimSpace(strings.ToLower(adjustFlag))
	if adjustFlag != "" && adjustFlag != "qfq" && adjustFlag != "hfq" {
		adjustFlag = "qfq"
	}

	api := data.NewEastMoneyKLineApi(data.GetSettingConfig())
	if !api.ValidateStockCode(stockCode) {
		return nil, fmt.Errorf("invalid stockCode: %s", stockCode)
	}

	var kLines *[]data.KLineData
	if adjustFlag != "" {
		kLines = api.GetKLineData(stockCode, "101", adjustFlag, days)
	} else {
		result := data.FetchKLineWithFallback(stockCode, "", "101", days, "")
		if result != nil && result.Data != nil {
			kLines = result.Data
		}
	}
	if kLines == nil || len(*kLines) == 0 {
		return nil, fmt.Errorf("no kline data")
	}
	return data.NewChipDistributionCalculator().Calculate(stockCode, *kLines, bins)
}

func (a *App) GetTdxCallAuction(stockCode string, start uint32, count uint32) *[]data.TdxCallAuctionData {
	if count <= 0 {
		count = 500
	}
	return data.NewTdxKLineApi().GetCallAuction(stockCode, start, count)
}

func (a *App) GetTdxCompanyInfo(stockCode string) *data.TdxCompanyInfoBundle {
	return data.NewTdxKLineApi().GetF10Data(stockCode)
}

func (a *App) GetTdxFinanceInfo(stockCode string) *data.TdxFinanceInfo {
	return data.NewTdxKLineApi().GetFinanceInfo(stockCode)
}

func (a *App) GetTdxXDXRInfo(stockCode string) *[]data.TdxXDXRItem {
	return data.NewTdxKLineApi().GetXDXRInfo(stockCode)
}

func (a *App) GetTdxCompanyCategoryList(stockCode string) *[]data.TdxCompanyCategory {
	return data.NewTdxKLineApi().GetF10CategoryList(stockCode)
}

func (a *App) GetTdxCompanyCategoryContent(stockCode string, categoryName string) *data.TdxCompanyInfoSection {
	return data.NewTdxKLineApi().GetF10CategoryContent(stockCode, categoryName)
}

func (a *App) GetTdxSymbolBelongBoard(stockCode string) *[]data.MACBelongBoardItem {
	return data.NewTdxKLineApi().GetMACSymbolBelongBoard(stockCode)
}

func (a *App) GetStockMinutePriceLineData(stockCode, stockName string) map[string]any {
	priceData, date := data.NewStockDataApi().GetStockMinutePriceData(stockCode)
	return map[string]any{
		"priceData": priceData,
		"date":      date,
		"stockName": stockName,
		"stockCode": stockCode,
	}
}
func (a *App) GetTelegraphList(source string) *[]*models.Telegraph {
	res := data.NewMarketNewsApi().GetTelegraphList(source)
	if res == nil {
		empty := []*models.Telegraph{}
		return &empty
	}
	return res
}

func (a *App) ReFleshTelegraphList(source string) *[]*models.Telegraph {
	newsAPI := data.NewMarketNewsApi()
	switch strings.TrimSpace(source) {
	case "财联社电报":
		newsAPI.TelegraphList(30)
	case "新浪财经":
		newsAPI.GetSinaNews(30)
	case "外媒":
		newsAPI.TradingViewNews()
	default:
		newsAPI.TelegraphList(30)
		newsAPI.GetSinaNews(30)
		newsAPI.TradingViewNews()
	}

	res := newsAPI.GetTelegraphList(source)
	if res == nil {
		empty := []*models.Telegraph{}
		return &empty
	}
	return res
}
func (a *App) GlobalStockIndexes() map[string]any {
	return data.NewMarketNewsApi().GlobalStockIndexes(30)
}
func (a *App) SummaryStockNews(question string, aiConfigId int, sysPromptId *int, enableTools bool, think bool) {
	ctx, cancel := context.WithCancel(a.ctx)
	cancelToken := a.setSummaryCancel(cancel)
	defer a.clearSummaryCancel(cancelToken)
	defer cancel()

	var msgs <-chan map[string]any
	if enableTools {
		msgs = data.NewDeepSeekOpenAi(ctx, aiConfigId).NewSummaryStockNewsStreamWithTools(question, sysPromptId, a.AiTools, think, nil)
	} else {
		msgs = data.NewDeepSeekOpenAi(ctx, aiConfigId).NewSummaryStockNewsStream(question, sysPromptId, think, nil)
	}
	for {
		select {
		case <-ctx.Done():
			runtime.EventsEmit(a.ctx, "summaryStockNews", "DONE")
			return
		case msg, ok := <-msgs:
			if !ok {
				runtime.EventsEmit(a.ctx, "summaryStockNews", "DONE")
				return
			}
			runtime.EventsEmit(a.ctx, "summaryStockNews", msg)
		}
	}
}
func (a *App) GetIndustryRank(sort string, cnt int) []any {
	res := data.NewMarketNewsApi().GetIndustryRank(sort, cnt)
	return res["data"].([]any)
}
func (a *App) GetIndustryMoneyRankSina(fenlei, sort string) []map[string]any {
	return data.NewMarketNewsApi().GetIndustryMoneyRankSina(fenlei, sort)
}
func (a *App) GetMoneyRankSina(sort string) []map[string]any {
	return data.NewMarketNewsApi().GetMoneyRankSina(sort)
}
func (a *App) GetStockMoneyTrendByDay(stockCode string, days int) []map[string]any {
	res := data.NewMarketNewsApi().GetStockMoneyTrendByDay(stockCode, days)
	lancetslice.Reverse(res)
	return res
}
func (a *App) OpenURL(url string) { runtime.BrowserOpenURL(a.ctx, url) }
func (a *App) GetAiConfigs() []*data.AIConfig {
	cfg := a.GetConfig()
	if len(cfg.AiConfigs) == 0 {
		return []*data.AIConfig{{Name: "Default", ModelName: "default"}}
	}
	return cfg.AiConfigs
}

func (a *App) GetUserManual() string {
	return string(userManual)
}

// AiModelInfo describes model context limits returned to the upstream settings UI.
type AiModelInfo struct {
	ModelName string `json:"modelName"`
	MaxTokens int    `json:"maxTokens"`
	Source    string `json:"source"`
}

func (a *App) FetchAiModels(baseUrl, apiKey string) []string {
	baseUrl = strutil.Trim(baseUrl)
	apiKey = strutil.Trim(apiKey)
	if baseUrl == "" || apiKey == "" {
		return []string{}
	}

	type modelItem struct {
		ID string `json:"id"`
	}
	var respData struct {
		Data []modelItem `json:"data"`
	}

	client := data.SharedHTTPClient
	client.SetBaseURL(baseUrl)

	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+apiKey).
		SetHeader("Content-Type", "application/json").
		SetResult(&respData).
		Get("/models")
	if err != nil {
		logger.SugaredLogger.Errorf("FetchAiModels error: %v", err)
		return []string{}
	}
	if resp.IsError() {
		logger.SugaredLogger.Errorf("FetchAiModels http error: %s", resp.Status())
		return []string{}
	}

	modelsList := make([]string, 0, len(respData.Data))
	for _, m := range respData.Data {
		if strings.TrimSpace(m.ID) != "" {
			modelsList = append(modelsList, m.ID)
		}
	}
	return modelsList
}

func (a *App) FetchAiModelInfo(baseUrl, apiKey, modelName string) *AiModelInfo {
	baseUrl = strutil.Trim(baseUrl)
	modelName = strutil.Trim(modelName)
	if baseUrl == "" || modelName == "" {
		return nil
	}

	info := &AiModelInfo{
		ModelName: modelName,
		MaxTokens: 0,
		Source:    "",
	}

	if apiKey != "" {
		type modelDetail struct {
			ID             string `json:"id"`
			MaxContextLen  int    `json:"max_context_length"`
			ContextLength  int    `json:"context_length"`
			MaxOutputTok   int    `json:"max_output_tokens"`
			MaxTokensField int    `json:"max_tokens"`
		}
		var detail modelDetail

		client := data.SharedHTTPClient
		client.SetBaseURL(baseUrl)

		resp, err := client.R().
			SetHeader("Authorization", "Bearer "+apiKey).
			SetHeader("Content-Type", "application/json").
			SetResult(&detail).
			Get("/models/" + modelName)

		if err == nil && !resp.IsError() && detail.ID != "" {
			switch {
			case detail.MaxContextLen > 0:
				info.MaxTokens = detail.MaxContextLen
				info.Source = "api"
			case detail.ContextLength > 0:
				info.MaxTokens = detail.ContextLength
				info.Source = "api"
			case detail.MaxOutputTok > 0:
				info.MaxTokens = detail.MaxOutputTok
				info.Source = "api"
			case detail.MaxTokensField > 0:
				info.MaxTokens = detail.MaxTokensField
				info.Source = "api"
			}
		}
	}

	if info.MaxTokens == 0 {
		if maxTokens := getBuiltinModelMaxTokens(modelName); maxTokens > 0 {
			info.MaxTokens = maxTokens
			info.Source = "builtin"
		}
	}

	return info
}

func getBuiltinModelMaxTokens(modelName string) int {
	modelTokenMap := map[string]int{
		"deepseek-chat":        65536,
		"deepseek-reasoner":    65536,
		"deepseek-coder":       16384,
		"deepseek-v3":          65536,
		"deepseek-r1":          65536,
		"gpt-4o":               16384,
		"gpt-4o-mini":          16384,
		"gpt-4o-2024-05-13":    4096,
		"gpt-4-turbo":          4096,
		"gpt-4-turbo-preview":  4096,
		"gpt-4":                8192,
		"gpt-4-32k":            32768,
		"gpt-3.5-turbo":        4096,
		"gpt-3.5-turbo-16k":    16384,
		"gpt-4.1":              32768,
		"gpt-4.1-mini":         32768,
		"gpt-4.1-nano":         32768,
		"o1":                   100000,
		"o1-mini":              65536,
		"o1-preview":           32768,
		"o3-mini":              100000,
		"o4-mini":              100000,
		"claude-3-5-sonnet":    8192,
		"claude-3-5-haiku":     8192,
		"claude-3-opus":        4096,
		"claude-3-sonnet":      4096,
		"claude-3-haiku":       4096,
		"glm-4":                8192,
		"glm-4-plus":           4096,
		"glm-4-air":            4096,
		"glm-4-flash":          4096,
		"glm-4-long":           4096,
		"chatglm-turbo":        4096,
		"moonshot-v1-8k":       8192,
		"moonshot-v1-32k":      32768,
		"moonshot-v1-128k":     131072,
		"qwen-turbo":           8192,
		"qwen-plus":            131072,
		"qwen-max":             8192,
		"qwen-long":            65536,
		"qwen2.5-72b-instruct": 32768,
		"hunyuan-lite":         4096,
		"hunyuan-standard":     4096,
		"hunyuan-pro":          4096,
		"hunyuan-turbo":        4096,
		"spark-lite":           4096,
		"spark-pro":            4096,
		"spark-max":            4096,
		"spark-4.0-ultra":      4096,
		"yi-light":             16384,
		"yi-large":             16384,
		"yi-medium":            16384,
		"yi-spark":             16384,
		"yi-vision":            16384,
		"abab6.5-chat":         8192,
		"abab6.5s-chat":        8192,
		"abab5.5-chat":         4096,
		"baichuan2-turbo":      4096,
		"baichuan2-53b":        4096,
		"ernie-4.0":            4096,
		"ernie-3.5":            4096,
		"ernie-speed":          4096,
		"ernie-lite":           4096,
	}
	if maxTokens, ok := modelTokenMap[modelName]; ok {
		return maxTokens
	}

	for prefix, maxTokens := range map[string]int{
		"deepseek":      65536,
		"gpt-4o":        16384,
		"gpt-4-turbo":   4096,
		"gpt-4-":        8192,
		"gpt-3.5":       4096,
		"gpt-4.1":       32768,
		"o1-":           65536,
		"o3-":           100000,
		"o4-":           100000,
		"claude-3":      8192,
		"glm-4":         8192,
		"chatglm":       4096,
		"moonshot-v1":   8192,
		"qwen-":         8192,
		"qwen2":         32768,
		"hunyuan-":      4096,
		"spark-":        4096,
		"yi-":           16384,
		"abab":          8192,
		"baichuan":      4096,
		"ernie-":        4096,
		"llama-3":       8192,
		"llama3":        8192,
		"mistral-":      8192,
		"mixtral-":      32768,
		"codestral-":    32768,
		"gemini-1.5":    8192,
		"gemini-2":      8192,
		"command-r":     4096,
		"Qwen/Qwen":     32768,
		"deepseek-ai/":  65536,
		"meta-llama/":   8192,
		"mistralai/":    32768,
		"Pro/deepseek-": 65536,
		"Pro/qwen-":     32768,
	} {
		if strings.HasPrefix(modelName, prefix) {
			return maxTokens
		}
	}

	return 0
}

func (a *App) TestAIConfigConnection(config data.AIConfig) map[string]any {
	baseURL := strings.TrimSpace(config.BaseUrl)
	if baseURL == "" {
		return map[string]any{
			"success": false,
			"message": "接口地址不能为空",
		}
	}
	if strings.TrimSpace(config.ApiKey) == "" {
		return map[string]any{
			"success": false,
			"message": "API Key 不能为空",
		}
	}

	modelsURL := strings.TrimRight(baseURL, "/")
	if !strings.HasSuffix(modelsURL, "/models") {
		modelsURL += "/models"
	}

	client := resty.New().SetTimeout(20 * time.Second)
	request := client.R().
		SetHeader("Authorization", "Bearer "+strings.TrimSpace(config.ApiKey)).
		SetHeader("Content-Type", "application/json")

	if config.HttpProxyEnabled && strings.TrimSpace(config.HttpProxy) != "" {
		client.SetProxy(strings.TrimSpace(config.HttpProxy))
	}

	response, err := request.Get(modelsURL)
	if err != nil {
		return map[string]any{
			"success":   false,
			"message":   err.Error(),
			"modelsUrl": modelsURL,
		}
	}

	if response.StatusCode() >= 400 {
		body := response.String()
		if len(body) > 240 {
			body = body[:240]
		}
		return map[string]any{
			"success":   false,
			"message":   fmt.Sprintf("请求失败: HTTP %d %s", response.StatusCode(), body),
			"modelsUrl": modelsURL,
		}
	}

	var payload struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(response.Body(), &payload); err != nil {
		return map[string]any{
			"success":   false,
			"message":   "模型列表解析失败: " + err.Error(),
			"modelsUrl": modelsURL,
		}
	}

	modelName := strings.TrimSpace(config.ModelName)
	modelExists := false
	models := make([]string, 0, 5)
	for _, item := range payload.Data {
		if item.ID == "" {
			continue
		}
		if modelName != "" && strings.EqualFold(strings.TrimSpace(item.ID), modelName) {
			modelExists = true
		}
		if len(models) < 5 {
			models = append(models, item.ID)
		}
	}

	message := fmt.Sprintf("连接成功，可用模型 %d 个", len(payload.Data))
	if len(models) > 0 {
		message += "，示例: " + strings.Join(models, ", ")
	}
	if modelName != "" {
		if modelExists {
			message += fmt.Sprintf("。当前模型 %s 在返回列表中，测试成功", modelName)
		} else {
			message += fmt.Sprintf("。但当前模型 %s 不在返回列表中，请检查模型名", modelName)
		}
	}

	return map[string]any{
		"success":     true,
		"message":     message,
		"modelsUrl":   modelsURL,
		"models":      models,
		"count":       len(payload.Data),
		"modelName":   modelName,
		"modelExists": modelExists,
	}
}

func (a *App) LongTigerRank(date string) *[]models.LongTigerRankData {
	return data.NewMarketNewsApi().LongTiger(date)
}
func (a *App) StockResearchReport(stockCode string) []any {
	return data.NewMarketNewsApi().StockResearchReport(stockCode, 7)
}
func (a *App) StockNotice(stockCode string) []any {
	return data.NewMarketNewsApi().StockNotice(stockCode)
}
func (a *App) IndustryResearchReport(industryCode string) []any {
	return data.NewMarketNewsApi().IndustryResearchReport(industryCode, 7)
}
func (a *App) EMDictCode(code string) []any { return data.NewMarketNewsApi().EMDictCode(code, a.cache) }
func (a *App) AnalyzeSentiment(text string) models.SentimentResult {
	return data.AnalyzeSentiment(text)
}
func (a *App) HotStock(marketType string) *[]models.HotItem {
	return data.NewMarketNewsApi().XUEQIUHotStock(100, marketType)
}
func (a *App) HotEvent(size int) *[]models.HotEvent {
	if size <= 0 {
		size = 10
	}
	return data.NewMarketNewsApi().HotEvent(size)
}
func (a *App) HotTopic(size int) []any {
	if size <= 0 {
		size = 10
	}
	return data.NewMarketNewsApi().HotTopic(size)
}
func (a *App) InvestCalendarTimeLine(yearMonth string) []any {
	return data.NewMarketNewsApi().InvestCalendar(yearMonth)
}
func (a *App) ClsCalendar() []any { return data.NewMarketNewsApi().ClsCalendar() }
func (a *App) SearchStock(words string) map[string]any {
	return data.NewSearchStockApi(words).SearchStock(5000)
}
func (a *App) GetHotStrategy() map[string]any { return data.NewSearchStockApi("").HotStrategy() }
func (a *App) GetFollowedFund() []data.FollowedFund {
	return data.NewFundApi().GetFollowedFund()
}
func (a *App) GetfundList(key string) []data.FundBasic {
	return data.NewFundApi().GetFundList(key)
}
func (a *App) FollowFund(fundCode string) string {
	return data.NewFundApi().FollowFund(fundCode)
}
func (a *App) UpdateFundWatchGroup(fundCode string, watchGroup string) string {
	return data.NewFundApi().UpdateFundWatchGroup(fundCode, watchGroup)
}
func (a *App) RenameFundWatchGroup(fromGroup string, toGroup string) string {
	return data.NewFundApi().RenameFundWatchGroup(fromGroup, toGroup)
}
func (a *App) DeleteFundWatchGroup(watchGroup string) string {
	return data.NewFundApi().DeleteFundWatchGroup(watchGroup)
}
func (a *App) UnFollowFund(fundCode string) string {
	return data.NewFundApi().UnFollowFund(fundCode)
}
func (a *App) GetAllStocks(page int, pageSize int, name string, technicalIndicators models.TechnicalIndicators) *models.AllStocksResp {
	return data.NewStockDataApi().GetAllStocks(page, pageSize, name, technicalIndicators)
}
func (a *App) ChatWithAgent(question string, aiConfigId int, sysPromptId *int) {
	ctx, cancel := context.WithCancel(a.ctx)
	cancelToken := a.setAgentCancel(cancel)
	defer a.clearAgentCancel(cancelToken)
	defer cancel()

	defer runtime.EventsEmit(a.ctx, "agent-message", map[string]any{
		"role":          "assistant",
		"response_meta": map[string]any{"finish_reason": "stop"},
	})

	if strings.TrimSpace(question) == "" {
		runtime.EventsEmit(a.ctx, "agent-message", map[string]any{
			"role":    "assistant",
			"content": "问题不能为空。",
		})
		return
	}

	configs := a.GetAiConfigs()
	if len(configs) == 0 {
		runtime.EventsEmit(a.ctx, "agent-message", map[string]any{
			"role":    "assistant",
			"content": "当前没有可用的 AI 配置，请先在设置中添加模型。",
		})
		return
	}

	resolvedConfigID := aiConfigId
	if resolvedConfigID == 0 {
		resolvedConfigID = int(configs[0].ID)
	}

	msgs := stockagent.NewStockAiAgentApi().ChatWithContext(ctx, question, resolvedConfigID, sysPromptId, true, 20, false, "")
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-msgs:
			if !ok {
				return
			}
			payload := map[string]any{
				"role": "assistant",
			}
			if msg.ReasoningContent != "" {
				payload["reasoning_content"] = msg.ReasoningContent
			}
			if msg.Content != "" {
				payload["content"] = msg.Content
			}
			if len(msg.ToolCalls) > 0 {
				payload["tool_calls"] = msg.ToolCalls
			}
			runtime.EventsEmit(a.ctx, "agent-message", payload)
		}
	}
}
func (a *App) AnalyzeSentimentWithFreqWeight(text string) map[string]any {
	result, cleanFrequencies := data.NewsAnalyze(text, false)
	return map[string]any{"result": result, "frequencies": cleanFrequencies}
}
func (a *App) GetAIResponseResultList(query models.AIResponseResultQuery) *models.AIResponseResultPageData {
	page, err := data.NewAIResponseResultService().GetAIResponseResultList(query)
	if err != nil {
		return &models.AIResponseResultPageData{}
	}
	return page
}
func (a *App) GetStockChanges(changeTypes []int, pageIndex, pageSize int) *data.StockChangesResponse {
	return data.NewStockChangesApi().GetStockChanges(changeTypes, pageIndex, pageSize)
}

func (a *App) GetAllStockChangesWithPaging(pageSize int) *data.StockChangesResponse {
	all := data.NewStockChangesApi().GetAllStockChangesWithPaging(pageSize)
	if all == nil {
		return &data.StockChangesResponse{}
	}
	historyService := data.NewStockChangeHistoryService()
	_, _ = historyService.SaveStockChangesWithDedup(all.Data)
	return all
}

func (a *App) GetStockChangeHistory(query models.StockChangeHistoryQuery) *models.StockChangeHistoryPageData {
	result, err := data.NewStockChangeHistoryService().GetHistoryList(query)
	if err != nil {
		return &models.StockChangeHistoryPageData{}
	}
	return result
}

func (a *App) SaveStockChangesToHistory(changeTypes []int) string {
	api := data.NewStockChangesApi()
	result := api.GetStockChanges(changeTypes, 0, 500)
	if result == nil || len(result.Data) == 0 {
		return "没有获取到异动数据"
	}

	err := data.NewStockChangeHistoryService().SaveStockChanges(result.Data)
	if err != nil {
		return "保存失败: " + err.Error()
	}
	return fmt.Sprintf("成功保存 %d 条异动数据", len(result.Data))
}

func (a *App) DeleteStockChangeHistory(days int) string {
	err := data.NewStockChangeHistoryService().DeleteOldData(days)
	if err != nil {
		return "删除失败: " + err.Error()
	}
	return fmt.Sprintf("已删除 %d 天前的历史数据", days)
}

func (a *App) DeleteAIResponseResult(id uint) string {
	if err := data.NewAIResponseResultService().DeleteAIResponseResult(id); err != nil {
		return "delete failed"
	}
	return "success"
}
func (a *App) GetAiRecommendStocksList(query models.AiRecommendStocksQuery) *models.AiRecommendStocksPageData {
	page, err := data.NewAiRecommendStocksService().GetAiRecommendStocksList(&query)
	if err != nil {
		return &models.AiRecommendStocksPageData{}
	}
	return page
}
func (a *App) DeleteAiRecommendStocks(id uint) string {
	if err := data.NewAiRecommendStocksService().DeleteAiRecommendStocks(id); err != nil {
		return "delete failed"
	}
	return "success"
}
func (a *App) GetPromptTemplateList(query models.PromptTemplateQuery) *models.PromptTemplatePageData {
	page, err := data.NewPromptTemplateApi().GetPromptTemplateList(&query)
	if err != nil {
		return &models.PromptTemplatePageData{}
	}
	return page
}
func (a *App) AddPromptTemplate(template models.PromptTemplate) string {
	return data.NewPromptTemplateApi().AddPrompt(template)
}

func (a *App) AddPrompt(template models.PromptTemplate) string {
	return data.NewPromptTemplateApi().AddPrompt(template)
}

func (a *App) UpdatePromptTemplate(template models.PromptTemplate) string {
	return data.NewPromptTemplateApi().AddPrompt(template)
}

func (a *App) DeletePromptTemplate(id uint) string { return data.NewPromptTemplateApi().DelPrompt(id) }

func (a *App) DelPrompt(id uint) string {
	return data.NewPromptTemplateApi().DelPrompt(id)
}
func (a *App) GetAllStockInfoList(query data.AllStockInfoQuery) *data.AllStockInfoPageData {
	page, err := data.NewStockDataApi().GetAllStockInfoList(&query)
	if err != nil {
		return &data.AllStockInfoPageData{}
	}
	return page
}
func (a *App) GetAllStockInfoById(id uint) *models.AllStockInfo {
	stock, err := data.NewStockDataApi().GetAllStockInfoById(id)
	if err != nil {
		return &models.AllStockInfo{}
	}
	return stock
}
func (a *App) AddAllStockInfo(stock models.AllStockInfo) string {
	if err := data.NewStockDataApi().AddAllStockInfo(stock); err != nil {
		return "failed: " + err.Error()
	}
	return "success"
}
func (a *App) DeleteAllStockInfo(id uint) string {
	if err := data.NewStockDataApi().DeleteAllStockInfo(id); err != nil {
		return "failed: " + err.Error()
	}
	return "success"
}
func (a *App) BatchDeleteAllStockInfo(ids []uint) string {
	if err := data.NewStockDataApi().BatchDeleteAllStockInfo(ids); err != nil {
		return "failed: " + err.Error()
	}
	return "success"
}
func (a *App) GetAllMarkets() []string {
	markets, err := data.NewStockDataApi().GetAllMarkets()
	if err != nil {
		return []string{}
	}
	return markets
}
func (a *App) GetAllIndustries() []string {
	industries, err := data.NewStockDataApi().GetAllIndustries()
	if err != nil {
		return []string{}
	}
	return industries
}
func (a *App) GetAllConcepts() []string {
	concepts, err := data.NewStockDataApi().GetAllConcepts()
	if err != nil {
		return []string{}
	}
	return concepts
}

func (a *App) SendFeishuMessage(message string, stockCode string) string {
	ttl, _ := a.cache.TTL([]byte(stockCode))
	if ttl > 0 {
		return ""
	}
	if err := a.cache.Set([]byte(stockCode), []byte("1"), 60*5); err != nil {
		logger.SugaredLogger.Errorf("set cache error:%s", err.Error())
		return ""
	}
	return data.NewFeishuAPI().SendFeishuMessage(message)
}

func (a *App) SendFeishuMessageByType(message string, stockCode string, msgType int) string {
	if strutil.HasPrefixAny(stockCode, []string{"SZ", "SH", "sh", "sz"}) && !isTradingTime(time.Now()) {
		return "非A股交易时间"
	}
	if strutil.HasPrefixAny(stockCode, []string{"hk", "HK"}) && !IsHKTradingTime(time.Now()) {
		return "非港股交易时间"
	}
	if strutil.HasPrefixAny(stockCode, []string{"us", "US", "gb_"}) && !IsUSTradingTime(time.Now()) {
		return "非美股交易时间"
	}
	ttl, _ := a.cache.TTL([]byte(stockCode))
	if ttl > 0 {
		return ""
	}
	if err := a.cache.Set([]byte(stockCode), []byte("1"), getMsgTypeTTL(msgType)); err != nil {
		logger.SugaredLogger.Errorf("set cache error:%s", err.Error())
		return ""
	}
	return data.NewFeishuAPI().SendFeishuMessage(message)
}

func (a *App) GetAiAssistantSession(sessionId string) (*models.AiAssistantSessionResp, error) {
	return data.GetAiAssistantSession(sessionId)
}

func (a *App) SaveAiAssistantSession(sessionId string, messages []models.AiAssistantMessage) error {
	return data.SaveAiAssistantSession(sessionId, messages)
}

func (a *App) FetchAndSaveMarketStatistic() {
	if !isTradingTime(time.Now()) {
		logger.SugaredLogger.Debugf("current time is outside trading hours, skip market statistic fetch")
		return
	}
	if err := data.NewMarketStatisticApi().FetchAndSave(); err != nil {
		logger.SugaredLogger.Errorf("fetch market statistic failed: %v", err)
	}
}

func (a *App) GetTodayMarketStatistic() []models.MarketStatistic {
	return data.NewMarketStatisticApi().GetTodayData()
}

func (a *App) GetMarketStatisticByDate(date string) []models.MarketStatistic {
	return data.NewMarketStatisticApi().GetByDate(date)
}

func (a *App) GetRecentDaysMarketStatistic(days int) []models.MarketStatistic {
	return data.NewMarketStatisticApi().GetRecentDaysData(days)
}

func (a *App) CreateMCPServer(server *models.MCPServer) string {
	if err := data.NewMCPServerApi().Create(server); err != nil {
		logger.SugaredLogger.Errorf("create MCP server failed: %v", err)
		return "创建失败: " + err.Error()
	}
	return "创建成功"
}

func (a *App) UpdateMCPServer(server *models.MCPServer) string {
	if err := data.NewMCPServerApi().Update(server); err != nil {
		logger.SugaredLogger.Errorf("update MCP server failed: %v", err)
		return "更新失败: " + err.Error()
	}
	return "更新成功"
}

func (a *App) DeleteMCPServer(id uint) string {
	if err := data.NewMCPServerApi().Delete(id); err != nil {
		logger.SugaredLogger.Errorf("delete MCP server failed: %v", err)
		return "删除失败: " + err.Error()
	}
	return "删除成功"
}

func (a *App) GetMCPServerByID(id uint) *models.MCPServer {
	server, err := data.NewMCPServerApi().GetByID(id)
	if err != nil {
		logger.SugaredLogger.Errorf("get MCP server failed: %v", err)
		return nil
	}
	return server
}

func (a *App) GetMCPServerList(query *models.MCPServerQuery) *models.MCPServerPageResp {
	return data.NewMCPServerApi().List(query)
}

func (a *App) EnableMCPServer(id uint, enable bool) string {
	if err := data.NewMCPServerApi().EnableServer(id, enable); err != nil {
		logger.SugaredLogger.Errorf("toggle MCP server failed: %v", err)
		return "操作失败: " + err.Error()
	}
	if enable {
		return "已启用"
	}
	return "已禁用"
}

func (a *App) TestMCPServer(id uint) string {
	result, err := data.NewMCPServerApi().TestConnection(id)
	if err != nil {
		logger.SugaredLogger.Errorf("test MCP server failed: %v", err)
		return "测试失败: " + err.Error()
	}
	return result
}

func (a *App) CreateSkill(skill *models.Skill) string {
	if err := data.NewSkillApi().Create(skill); err != nil {
		logger.SugaredLogger.Errorf("create skill failed: %v", err)
		return "创建失败: " + err.Error()
	}
	return "创建成功"
}

func (a *App) UpdateSkill(skill *models.Skill) string {
	if err := data.NewSkillApi().Update(skill); err != nil {
		logger.SugaredLogger.Errorf("update skill failed: %v", err)
		return "更新失败: " + err.Error()
	}
	return "更新成功"
}

func (a *App) DeleteSkill(id uint) string {
	if err := data.NewSkillApi().Delete(id); err != nil {
		logger.SugaredLogger.Errorf("delete skill failed: %v", err)
		return "删除失败: " + err.Error()
	}
	return "删除成功"
}

func (a *App) GetSkillByID(id uint) *models.Skill {
	skill, err := data.NewSkillApi().GetByID(id)
	if err != nil {
		logger.SugaredLogger.Errorf("get skill failed: %v", err)
		return nil
	}
	return skill
}

func (a *App) GetSkillList(query *models.SkillQuery) *models.SkillPageResp {
	return data.NewSkillApi().List(query)
}

func (a *App) EnableSkill(id uint, enable bool) string {
	if err := data.NewSkillApi().EnableSkill(id, enable); err != nil {
		logger.SugaredLogger.Errorf("toggle skill failed: %v", err)
		return "操作失败: " + err.Error()
	}
	if enable {
		return "已启用"
	}
	return "已禁用"
}

func (a *App) GetAllSkills() []models.Skill {
	return data.NewSkillApi().GetAll()
}

func (a *App) GetMCPToolsByServerID(serverID uint) []models.MCPServerTool {
	return data.NewMCPServerApi().GetToolsByServerID(serverID)
}

func (a *App) GetAllMCPTools() []models.MCPServerTool {
	return data.NewMCPServerApi().GetAllTools()
}

func refreshTelegraphList() *[]string {
	response, err := resty.New().R().
		SetHeader("Referer", "https://www.cls.cn/").
		SetHeader("User-Agent", "Mozilla/5.0").
		Get("https://www.cls.cn/telegraph")
	if err != nil {
		return &[]string{}
	}
	document, err := goquery.NewDocumentFromReader(strings.NewReader(string(response.Body())))
	if err != nil {
		return &[]string{}
	}
	var telegraph []string
	document.Find("div.telegraph-content-box").Each(func(_ int, selection *goquery.Selection) {
		telegraph = append(telegraph, selection.Text())
	})
	return &telegraph
}
