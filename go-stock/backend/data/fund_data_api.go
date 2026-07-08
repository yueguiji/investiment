package data

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/duke-git/lancet/v2/convertor"
	"github.com/duke-git/lancet/v2/mathutil"
	"github.com/duke-git/lancet/v2/strutil"
	"github.com/go-resty/resty/v2"
	"go-stock/backend/db"
	"go-stock/backend/logger"
	"golang.org/x/net/html/charset"
	"html"
	"io"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"gorm.io/gorm"
)

type FundApi struct {
	client *resty.Client
	config *SettingConfig
}

type FundCatalogItem struct {
	Code string `json:"code"`
	Name string `json:"name"`
	Type string `json:"type"`
}

var fundCatalogCache = struct {
	mu       sync.RWMutex
	items    []FundCatalogItem
	loadedAt time.Time
}{}

func NewFundApi() *FundApi {
	return &FundApi{
		client: resty.New(),
		config: GetSettingConfig(),
	}
}

type FollowedFund struct {
	gorm.Model
	Code             string   `json:"code" gorm:"index"`
	Name             string   `json:"name"`
	IsWatchlist      bool     `json:"isWatchlist" gorm:"index;default:false"`
	WatchGroup       string   `json:"watchGroup" gorm:"index"`
	NetUnitValue     *float64 `json:"netUnitValue"`
	NetUnitValueDate string   `json:"netUnitValueDate"`
	NetEstimatedUnit *float64 `json:"netEstimatedUnit"`
	NetEstimatedTime string   `json:"netEstimatedUnitTime"`
	NetAccumulated   *float64 `json:"netAccumulated"`
	NetEstimatedRate *float64 `json:"netEstimatedRate"`

	FundBasic FundBasic `json:"fundBasic" gorm:"foreignKey:Code;references:Code"`
}

func (FollowedFund) TableName() string {
	return "followed_fund"
}

type FundEstimateSnapshot struct {
	gorm.Model
	Code          string   `json:"code" gorm:"index:idx_fund_estimate_snapshots_code_time,priority:1;index"`
	Name          string   `json:"name"`
	TradeDate     string   `json:"tradeDate" gorm:"index"`
	EstimateTime  string   `json:"estimateTime" gorm:"index:idx_fund_estimate_snapshots_code_time,priority:2"`
	EstimatedUnit float64  `json:"estimatedUnit"`
	EstimatedRate *float64 `json:"estimatedRate"`
	Source        string   `json:"source"`
}

func (FundEstimateSnapshot) TableName() string {
	return "fund_estimate_snapshots"
}

type FundBasic struct {
	gorm.Model
	Code              string   `json:"code" gorm:"index"`
	Name              string   `json:"name"`
	FullName          string   `json:"fullName"`
	Type              string   `json:"type"`
	Establishment     string   `json:"establishment"`
	Scale             string   `json:"scale"`
	Company           string   `json:"company"`
	Manager           string   `json:"manager"`
	Rating            string   `json:"rating"`
	TrackingTarget    string   `json:"trackingTarget"`
	NetUnitValue      *float64 `json:"netUnitValue"`
	NetUnitValueDate  string   `json:"netUnitValueDate"`
	NetEstimatedUnit  *float64 `json:"netEstimatedUnit"`
	NetEstimatedTime  string   `json:"netEstimatedUnitTime"`
	NetAccumulated    *float64 `json:"netAccumulated"`
	NetGrowth1        *float64 `json:"netGrowth1"`
	NetGrowth3        *float64 `json:"netGrowth3"`
	NetGrowth6        *float64 `json:"netGrowth6"`
	NetGrowth12       *float64 `json:"netGrowth12"`
	NetGrowth36       *float64 `json:"netGrowth36"`
	NetGrowth60       *float64 `json:"netGrowth60"`
	NetGrowthYTD      *float64 `json:"netGrowthYTD"`
	NetGrowthAll      *float64 `json:"netGrowthAll"`
	NetGrowth7        *float64 `json:"netGrowth7"`
	MaxDrawdown1      *float64 `json:"maxDrawdown1"`
	MaxDrawdown3      *float64 `json:"maxDrawdown3"`
	MaxDrawdown6      *float64 `json:"maxDrawdown6"`
	MaxDrawdown12     *float64 `json:"maxDrawdown12"`
	Volatility12      *float64 `json:"volatility12"`
	Sharpe12          *float64 `json:"sharpe12"`
	Calmar12          *float64 `json:"calmar12"`
	StageRank1M       int      `json:"stageRank1m"`
	StageRank1MTotal  int      `json:"stageRank1mTotal"`
	StageRank3M       int      `json:"stageRank3m"`
	StageRank3MTotal  int      `json:"stageRank3mTotal"`
	StageRank6M       int      `json:"stageRank6m"`
	StageRank6MTotal  int      `json:"stageRank6mTotal"`
	StageRank12M      int      `json:"stageRank12m"`
	StageRank12MTotal int      `json:"stageRank12mTotal"`
	RedeemFeeFreeDays int      `json:"redeemFeeFreeDays"`
	TopIndustry       string   `json:"topIndustry"`
	TopIndustryWeight *float64 `json:"topIndustryWeight"`
	TopIndustryDate   string   `json:"topIndustryDate"`
	ScreenUpdatedAt   string   `json:"screenUpdatedAt"`
}

func (FundBasic) TableName() string {
	return "fund_basic"
}

func (f *FundApi) CrawlFundBasic(fundCode string) (*FundBasic, error) {
	defer func() {
		if r := recover(); r != nil {
			logger.SugaredLogger.Errorf("CrawlFundBasic panic: %v", r)
		}
	}()

	if f.client != nil {
		return f.crawlFundBasicHTTP(fundCode)
	}

	crawler := CrawlerApi{
		crawlerBaseInfo: CrawlerBaseInfo{
			Name:    "澶╁ぉ鍩洪噾",
			BaseUrl: "http://fund.eastmoney.com",
			Headers: map[string]string{"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36"},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(f.config.CrawlTimeOut)*time.Second)
	defer cancel()

	crawler = crawler.NewCrawler(ctx, crawler.crawlerBaseInfo)
	url := fmt.Sprintf("%s/%s.html", crawler.crawlerBaseInfo.BaseUrl, fundCode)
	htmlContent, ok := crawler.GetHtml(url, ".merchandiseDetail", true)
	if !ok {
		return nil, fmt.Errorf("fund page parse failed")
	}

	fund := &FundBasic{Code: fundCode}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, err
	}

	name := strings.TrimSpace(doc.Find(".merchandiseDetail .fundDetail-tit").First().Text())
	name = strings.ReplaceAll(name, "查看相关ETF>", "")
	fund.Name = strings.TrimSpace(name)

	fullName := strings.TrimSpace(doc.Find(".merchandiseDetail .infoOfFund .fundDetail-footer li").First().Text())
	if strings.Contains(fullName, "基金全称") {
		fund.FullName = strings.TrimSpace(extractInfoValue(fullName))
	}

	doc.Find(".infoOfFund table td").Each(func(_ int, s *goquery.Selection) {
		text := strings.TrimSpace(strutil.RemoveWhiteSpace(s.Text(), true))
		if text == "" {
			return
		}
		value := extractInfoValue(text)
		switch {
		case strings.Contains(text, "基金类型") || strings.Contains(text, "类型"):
			fund.Type = value
		case strings.Contains(text, "成立日期"):
			fund.Establishment = value
		case strings.Contains(text, "基金规模") || strings.Contains(text, "规模"):
			fund.Scale = value
		case strings.Contains(text, "管理人") || strings.Contains(text, "基金公司"):
			fund.Company = value
		case strings.Contains(text, "基金经理") || strings.Contains(text, "经理人"):
			fund.Manager = value
		case strings.Contains(text, "基金评级") || strings.Contains(text, "评级"):
			fund.Rating = value
		case strings.Contains(text, "跟踪标的") || strings.Contains(text, "标的"):
			fund.TrackingTarget = value
		}
	})

	parseGrowthMetrics(doc, fund)
	f.applyFundFeeInfo(fundCode, fund)

	count := int64(0)
	db.Dao.Model(fund).Where("code=?", fund.Code).Count(&count)
	if count == 0 {
		db.Dao.Create(fund)
	} else {
		db.Dao.Model(fund).Where("code=?", fund.Code).Updates(fund)
	}

	return fund, nil
}

func extractInfoValue(text string) string {
	text = strings.TrimSpace(text)
	for _, sep := range []string{"：", ":", "（", "("} {
		if strings.Contains(text, sep) {
			parts := strings.SplitN(text, sep, 2)
			if len(parts) == 2 {
				value := strings.TrimSpace(strings.Trim(parts[1], "）)"))
				if value != "" {
					return value
				}
			}
		}
	}
	return text
}

func parseGrowthMetrics(doc *goquery.Document, fund *FundBasic) {
	assignGrowth := func(label string, value *float64) {
		if value == nil {
			return
		}
		switch label {
		case "近1月":
			fund.NetGrowth1 = value
		case "近3月":
			fund.NetGrowth3 = value
		case "近6月":
			fund.NetGrowth6 = value
		case "近1年":
			fund.NetGrowth12 = value
		case "近3年":
			fund.NetGrowth36 = value
		case "近5年":
			fund.NetGrowth60 = value
		case "今年来":
			fund.NetGrowthYTD = value
		case "成立来":
			fund.NetGrowthAll = value
		}
	}

	doc.Find(".dataOfFund dl").Each(func(_ int, s *goquery.Selection) {
		label := strings.TrimSpace(strutil.RemoveWhiteSpace(s.Find("dt").First().Text(), true))
		valueText := strings.TrimSpace(strutil.RemoveWhiteSpace(s.Find("dd").First().Text(), true))
		value := parsePercentPointer(valueText)
		assignGrowth(label, value)
	})

	doc.Find(".dataOfFund dl > dd").Each(func(_ int, s *goquery.Selection) {
		text := strings.TrimSpace(strutil.RemoveWhiteSpace(s.Text(), true))
		if text == "" || !strings.Contains(text, "%") {
			return
		}
		matches := regexp.MustCompile(`(近1月|近3月|近6月|近1年|近3年|近5年|今年来|成立来)`).FindStringSubmatch(text)
		if len(matches) == 0 {
			return
		}
		assignGrowth(matches[1], parsePercentPointer(text))
	})
}

func parsePercentPointer(text string) *float64 {
	re := regexp.MustCompile(`[-+]?\d+(\.\d+)?`)
	match := re.FindString(strings.ReplaceAll(text, ",", ""))
	if match == "" {
		return nil
	}
	value, err := strconv.ParseFloat(match, 64)
	if err != nil {
		return nil
	}
	return &value
}

func (f *FundApi) crawlFundBasicHTTP(fundCode string) (*FundBasic, error) {
	fundCode = strings.TrimSpace(fundCode)
	if fundCode == "" {
		return nil, fmt.Errorf("fund code is empty")
	}

	htmlContent, err := f.fetchFundBasicPageHTML(fundCode)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, err
	}

	fund := &FundBasic{Code: fundCode}
	parseFundBasicHTTPDocument(doc, fund)

	pingzhongData, pingzhongErr := f.fetchFundPingzhongData(fundCode)
	if pingzhongErr == nil {
		parseFundBasicHTTPPingzhongData(pingzhongData, fund)
	} else {
		logger.SugaredLogger.Warnf("fetchFundPingzhongData failed for %s: %v", fundCode, pingzhongErr)
	}
	f.applyFundFeeInfo(fundCode, fund)

	if strings.TrimSpace(fund.Name) == "" {
		return nil, fmt.Errorf("fund page parse failed")
	}
	if strings.TrimSpace(fund.FullName) == "" {
		fund.FullName = fund.Name
	}

	count := int64(0)
	db.Dao.Model(fund).Where("code=?", fund.Code).Count(&count)
	if count == 0 {
		db.Dao.Create(fund)
	} else {
		db.Dao.Model(fund).Where("code=?", fund.Code).Updates(fund)
	}

	return fund, nil
}

func (f *FundApi) fetchFundBasicPageHTML(fundCode string) (string, error) {
	response, err := f.client.SetTimeout(time.Duration(f.config.CrawlTimeOut)*time.Second).R().
		SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36").
		SetHeader("Referer", "https://fund.eastmoney.com/").
		Get(fmt.Sprintf("https://fund.eastmoney.com/%s.html", fundCode))
	if err != nil {
		return "", err
	}
	if response.StatusCode() != 200 {
		return "", fmt.Errorf("unexpected status code: %d", response.StatusCode())
	}
	return decodeHTTPDocument(response.Body(), response.Header().Get("Content-Type"))
}

func (f *FundApi) fetchFundPingzhongData(code string) (string, error) {
	response, err := f.client.SetTimeout(time.Duration(f.config.CrawlTimeOut)*time.Second).R().
		SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36").
		SetHeader("Referer", fmt.Sprintf("https://fund.eastmoney.com/%s.html", code)).
		Get(fmt.Sprintf("https://fund.eastmoney.com/pingzhongdata/%s.js?v=%d", code, time.Now().UnixMilli()))
	if err != nil {
		return "", err
	}
	if response.StatusCode() != 200 {
		return "", fmt.Errorf("unexpected status code: %d", response.StatusCode())
	}
	return string(response.Body()), nil
}

func (f *FundApi) fetchFundFeePageHTML(code string) (string, error) {
	client := f.client
	if client == nil {
		client = resty.New()
	}
	response, err := client.SetTimeout(time.Duration(f.config.CrawlTimeOut)*time.Second).R().
		SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36").
		SetHeader("Referer", fmt.Sprintf("https://fund.eastmoney.com/%s.html", code)).
		Get(fmt.Sprintf("https://fundf10.eastmoney.com/jjfl_%s.html", code))
	if err != nil {
		return "", err
	}
	if response.StatusCode() != 200 {
		return "", fmt.Errorf("unexpected status code: %d", response.StatusCode())
	}
	return decodeHTTPDocument(response.Body(), response.Header().Get("Content-Type"))
}

func (f *FundApi) applyFundFeeInfo(code string, fund *FundBasic) {
	if fund == nil {
		return
	}
	feeHTML, err := f.fetchFundFeePageHTML(code)
	if err != nil {
		logger.SugaredLogger.Warnf("fetchFundFeePageHTML failed for %s: %v", code, err)
		return
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(feeHTML))
	if err != nil {
		logger.SugaredLogger.Warnf("parse fund fee document failed for %s: %v", code, err)
		return
	}
	parseFundFeeHTTPDocument(doc, fund)
}

func decodeHTTPDocument(body []byte, contentType string) (string, error) {
	reader, err := charset.NewReader(bytes.NewReader(body), contentType)
	if err != nil {
		return "", err
	}
	decoded, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

func parseFundBasicHTTPDocument(doc *goquery.Document, fund *FundBasic) {
	name := strings.TrimSpace(doc.Find(".merchandiseDetail .fundDetail-tit").First().Text())
	name = strings.ReplaceAll(name, "查看相关ETF>", "")
	name = regexp.MustCompile(`\s*\([0-9A-Za-z]+\)\s*$`).ReplaceAllString(name, "")
	if name != "" {
		fund.Name = strings.TrimSpace(name)
	}

	doc.Find(".infoOfFund table td").Each(func(_ int, s *goquery.Selection) {
		text := normalizeFundHTTPInfoText(s.Text())
		if text == "" {
			return
		}
		value := extractFundHTTPInfoValue(text)
		if containsFundHTTPText(text, "基金全称") {
			fund.FullName = value
		}
		if containsFundHTTPText(text, "基金类型", "类型：", "类型:") {
			fund.Type = value
		}
		if containsFundHTTPText(text, "成立日", "成立日期") {
			fund.Establishment = value
		}
		if containsFundHTTPText(text, "基金规模", "规模：", "规模:") {
			fund.Scale = value
		}
		if containsFundHTTPText(text, "管理人", "基金公司") {
			fund.Company = value
		}
		if containsFundHTTPText(text, "基金经理", "经理：", "经理:") {
			fund.Manager = value
		}
		if containsFundHTTPText(text, "基金评级", "评级") {
			fund.Rating = value
		}
		if containsFundHTTPText(text, "跟踪标的") {
			fund.TrackingTarget = value
		}
		switch {
		case strings.Contains(text, "基金全称"):
			fund.FullName = value
		case strings.Contains(text, "基金类型") || strings.HasPrefix(text, "类型："):
			fund.Type = value
		case strings.Contains(text, "成立日") || strings.Contains(text, "成立日期"):
			fund.Establishment = value
		case strings.Contains(text, "基金规模") || strings.HasPrefix(text, "规模："):
			fund.Scale = value
		case strings.Contains(text, "管理人") || strings.Contains(text, "基金公司"):
			fund.Company = value
		case strings.Contains(text, "基金经理") || strings.HasPrefix(text, "经理："):
			fund.Manager = value
		case strings.Contains(text, "基金评级") || strings.Contains(text, "评级"):
			fund.Rating = value
		case strings.Contains(text, "跟踪标的"):
			fund.TrackingTarget = value
		}
	})
}

func parseFundBasicHTTPPingzhongData(content string, fund *FundBasic) {
	if fund == nil || strings.TrimSpace(content) == "" {
		return
	}
	if value := extractFundHTTPJSStringVar(content, "fS_name"); value != "" && strings.TrimSpace(fund.Name) == "" {
		fund.Name = strings.TrimSpace(value)
	}
	assignFundHTTPJSPercentVar(content, "syl_1y", &fund.NetGrowth1)
	assignFundHTTPJSPercentVar(content, "syl_3y", &fund.NetGrowth3)
	assignFundHTTPJSPercentVar(content, "syl_6y", &fund.NetGrowth6)
	assignFundHTTPJSPercentVar(content, "syl_1n", &fund.NetGrowth12)
}

func normalizeFundHTTPInfoText(text string) string {
	text = html.UnescapeString(strings.TrimSpace(text))
	text = strings.ReplaceAll(text, "\u00a0", "")
	text = strings.ReplaceAll(text, " ", "")
	text = strings.ReplaceAll(text, " ", "")
	return text
}

func containsFundHTTPText(text string, keywords ...string) bool {
	for _, keyword := range keywords {
		if keyword != "" && strings.Contains(text, keyword) {
			return true
		}
	}
	return false
}

func looksLikeFundZeroRedeemFee(text string) bool {
	value := parsePercentPointer(text)
	return value != nil && *value == 0
}

func extractFundFeeFreeDays(text string) (int, bool) {
	if text == "" {
		return 0, false
	}
	normalized := normalizeFundHTTPInfoText(text)
	if containsFundHTTPText(normalized, "天", "日", "周", "月", "年") {
		matches := regexp.MustCompile(`\d+`).FindAllString(normalized, -1)
		minDays := 0
		for _, match := range matches {
			days, err := strconv.Atoi(match)
			if err != nil || days <= 0 {
				continue
			}
			if minDays == 0 || days < minDays {
				minDays = days
			}
		}
		if minDays > 0 {
			return minDays, true
		}
	}
	if !strings.Contains(normalized, "天") && !strings.Contains(normalized, "日") && !strings.Contains(normalized, "澶") {
		return 0, false
	}
	match := regexp.MustCompile(`\d+`).FindString(normalized)
	if match == "" {
		return 0, false
	}
	days, err := strconv.Atoi(match)
	if err != nil || days <= 0 {
		return 0, false
	}
	return days, true
}

func parseFundFeeHTTPDocument(doc *goquery.Document, fund *FundBasic) {
	if doc == nil || fund == nil {
		return
	}

	minDays := 0
	doc.Find("table tr").Each(func(_ int, row *goquery.Selection) {
		cells := row.Find("th,td")
		if cells.Length() < 2 {
			return
		}
		condition := normalizeFundHTTPInfoText(cells.First().Text())
		feeText := normalizeFundHTTPInfoText(cells.Last().Text())
		if containsFundHTTPText(condition, "适用期限", "赎回费率", "适用金额", "费率") {
			return
		}
		if !looksLikeFundZeroRedeemFee(feeText) {
			return
		}
		days, ok := extractFundFeeFreeDays(condition)
		if !ok {
			return
		}
		if minDays == 0 || days < minDays {
			minDays = days
		}
	})
	if minDays > 0 {
		fund.RedeemFeeFreeDays = minDays
		return
	}

	table := doc.Find(`a[name="shfl"]`).First().ParentsFiltered(".boxitem").First().Find("table").First()
	if table.Length() == 0 {
		return
	}

	minDays = 0
	table.Find("tbody tr").Each(func(_ int, row *goquery.Selection) {
		cells := row.Find("td")
		if cells.Length() < 2 {
			return
		}
		condition := normalizeFundHTTPInfoText(cells.First().Text())
		feeText := normalizeFundHTTPInfoText(cells.Eq(1).Text())
		if !looksLikeFundZeroRedeemFee(feeText) {
			return
		}
		days, ok := extractFundFeeFreeDays(condition)
		if !ok {
			return
		}
		if minDays == 0 || days < minDays {
			minDays = days
		}
	})

	if minDays > 0 {
		fund.RedeemFeeFreeDays = minDays
	}
}

func extractFundHTTPInfoValue(text string) string {
	for _, sep := range []string{"：", ":"} {
		if strings.Contains(text, sep) {
			parts := strings.SplitN(text, sep, 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	if strings.Contains(text, "：") {
		parts := strings.SplitN(text, "：", 2)
		if len(parts) == 2 {
			return strings.TrimSpace(parts[1])
		}
	}
	return strings.TrimSpace(text)
}

func extractFundHTTPJSStringVar(content string, varName string) string {
	pattern := fmt.Sprintf(`var\s+%s\s*=\s*"([^"]*)"`, regexp.QuoteMeta(varName))
	match := regexp.MustCompile(pattern).FindStringSubmatch(content)
	if len(match) < 2 {
		return ""
	}
	return html.UnescapeString(strings.TrimSpace(match[1]))
}

func extractFundHTTPJSPercentVar(content string, varName string) *float64 {
	pattern := fmt.Sprintf(`var\s+%s\s*=\s*"([^"]*)"`, regexp.QuoteMeta(varName))
	match := regexp.MustCompile(pattern).FindStringSubmatch(content)
	if len(match) < 2 {
		return nil
	}
	return parsePercentPointer(match[1])
}

func assignFundHTTPJSPercentVar(content string, varName string, target **float64) {
	if target == nil {
		return
	}
	if value := extractFundHTTPJSPercentVar(content, varName); value != nil {
		*target = value
	}
}

func (f *FundApi) GetFundList(key string) []FundBasic {
	var funds []FundBasic
	db.Dao.Where("code like ? or name like ?", "%"+key+"%", "%"+key+"%").Limit(10).Find(&funds)
	return funds
}

func (f *FundApi) GetEastmoneyFundCatalog(forceRefresh bool) ([]FundCatalogItem, error) {
	fundCatalogCache.mu.RLock()
	if !forceRefresh && len(fundCatalogCache.items) > 0 && time.Since(fundCatalogCache.loadedAt) < 12*time.Hour {
		items := append([]FundCatalogItem(nil), fundCatalogCache.items...)
		fundCatalogCache.mu.RUnlock()
		return items, nil
	}
	fundCatalogCache.mu.RUnlock()

	response, err := f.client.SetTimeout(time.Duration(f.config.CrawlTimeOut)*time.Second).R().
		SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36").
		SetHeader("Referer", "https://fund.eastmoney.com/").
		Get("https://fund.eastmoney.com/js/fundcode_search.js")
	if err != nil {
		return nil, err
	}
	if response.StatusCode() != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", response.StatusCode())
	}

	match := regexp.MustCompile(`var\s+r\s*=\s*(\[[\s\S]*\])\s*;?\s*$`).FindStringSubmatch(string(response.Body()))
	if len(match) < 2 {
		return nil, fmt.Errorf("fund catalog payload not found")
	}

	var raw [][]string
	if err := json.Unmarshal([]byte(match[1]), &raw); err != nil {
		return nil, err
	}

	items := make([]FundCatalogItem, 0, len(raw))
	for _, row := range raw {
		if len(row) < 4 {
			continue
		}
		code := strings.TrimSpace(row[0])
		name := strings.TrimSpace(row[2])
		fundType := strings.TrimSpace(row[3])
		if code == "" || name == "" || fundType == "" {
			continue
		}
		items = append(items, FundCatalogItem{
			Code: code,
			Name: name,
			Type: fundType,
		})
	}

	fundCatalogCache.mu.Lock()
	fundCatalogCache.items = append([]FundCatalogItem(nil), items...)
	fundCatalogCache.loadedAt = time.Now()
	fundCatalogCache.mu.Unlock()
	return items, nil
}

func (f *FundApi) GetFollowedFund() []FollowedFund {
	var funds []FollowedFund
	db.Dao.Preload("FundBasic").
		Where("is_watchlist = ?", true).
		Order("CASE WHEN COALESCE(watch_group, '') = '' THEN 1 ELSE 0 END").
		Order("watch_group asc").
		Order("code asc").
		Find(&funds)
	for i, fund := range funds {
		if fund.NetUnitValue != nil && fund.NetEstimatedUnit != nil && *fund.NetUnitValue > 0 {
			netEstimatedRate := (*(funds[i].NetEstimatedUnit) - *(funds[i].NetUnitValue)) / *(fund.NetUnitValue) * 100
			netEstimatedRate = mathutil.RoundToFloat(netEstimatedRate, 2)
			funds[i].NetEstimatedRate = &netEstimatedRate
		}

	}
	return funds
}

type FollowedFundPagedResult struct {
	Items      []FollowedFund `json:"items"`
	TotalCount int64          `json:"totalCount"`
	PageIndex  int            `json:"pageIndex"`
	PageSize   int            `json:"pageSize"`
	TotalPages int            `json:"totalPages"`
}

func (f *FundApi) GetFollowedFundPaged(pageIndex, pageSize int, keyword string) *FollowedFundPagedResult {
	if pageIndex <= 0 {
		pageIndex = 1
	}
	if pageSize <= 0 {
		pageSize = 4
	}

	all := f.GetFollowedFund()
	keyword = strings.TrimSpace(keyword)
	filtered := make([]FollowedFund, 0, len(all))
	for _, fund := range all {
		if keyword == "" ||
			strings.Contains(fund.Code, keyword) ||
			strings.Contains(fund.Name, keyword) ||
			strings.Contains(fund.FundBasic.FullName, keyword) {
			filtered = append(filtered, fund)
		}
	}

	totalCount := int64(len(filtered))
	totalPages := int(math.Ceil(float64(totalCount) / float64(pageSize)))
	if totalPages == 0 {
		totalPages = 1
	}
	start := (pageIndex - 1) * pageSize
	if start > len(filtered) {
		start = len(filtered)
	}
	end := start + pageSize
	if end > len(filtered) {
		end = len(filtered)
	}

	return &FollowedFundPagedResult{
		Items:      filtered[start:end],
		TotalCount: totalCount,
		PageIndex:  pageIndex,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
}

type FundRankingItem struct {
	Code             string   `json:"code"`
	Name             string   `json:"name"`
	Pinyin           string   `json:"pinyin"`
	NetValueDate     string   `json:"netValueDate"`
	NetUnitValue     *float64 `json:"netUnitValue"`
	NetAccumulated   *float64 `json:"netAccumulated"`
	DailyGrowth      *float64 `json:"dailyGrowth"`
	WeekGrowth       *float64 `json:"weekGrowth"`
	MonthGrowth      *float64 `json:"monthGrowth"`
	ThreeMonthGrowth *float64 `json:"threeMonthGrowth"`
	SixMonthGrowth   *float64 `json:"sixMonthGrowth"`
	YearGrowth       *float64 `json:"yearGrowth"`
	TwoYearGrowth    *float64 `json:"twoYearGrowth"`
	ThreeYearGrowth  *float64 `json:"threeYearGrowth"`
	YTDGrowth        *float64 `json:"ytdGrowth"`
	SinceInception   *float64 `json:"sinceInception"`
	EstablishDate    string   `json:"establishDate"`
	Purchasable      bool     `json:"purchasable"`
	Scale            *float64 `json:"scale"`
	PurchaseRate     *float64 `json:"purchaseRate"`
	DiscountRate     *float64 `json:"discountRate"`
	FundTypeDetail   string   `json:"fundTypeDetail"`
}

type FundRankingResult struct {
	Items      []FundRankingItem `json:"items"`
	TotalCount int               `json:"totalCount"`
	PageIndex  int               `json:"pageIndex"`
	PageSize   int               `json:"pageSize"`
	TotalPages int               `json:"totalPages"`
}

type FundSearchItem struct {
	Code string `json:"code"`
	Name string `json:"name"`
	Type string `json:"type"`
}

func fundParseFloatPtr(s string) *float64 {
	s = strings.TrimSpace(s)
	if s == "" || s == "-" {
		return nil
	}
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil
	}
	return &val
}

func (f *FundApi) SearchFundCodes(keyword string) []FundSearchItem {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return []FundSearchItem{}
	}
	url := fmt.Sprintf("https://fundsuggest.eastmoney.com/FundSearch/api/FundSearchAPI.ashx?callback=&m=1&key=%s", keyword)
	resp, err := f.client.SetTimeout(time.Duration(f.config.CrawlTimeOut)*time.Second).R().
		SetHeader("User-Agent", getRandomUA()).
		SetHeader("Referer", "https://fund.eastmoney.com/").
		Get(url)
	if err != nil || resp.StatusCode() != 200 {
		return []FundSearchItem{}
	}

	var result struct {
		Datas []struct {
			Code         string `json:"CODE"`
			CodeLower    string `json:"Code"`
			Name         string `json:"NAME"`
			NameLower    string `json:"Name"`
			Type         string `json:"CATEGORYDESC"`
			FundBaseInfo *struct {
				FCODE     string `json:"FCODE"`
				SHORTNAME string `json:"SHORTNAME"`
				FTYPE     string `json:"FTYPE"`
			} `json:"FundBaseInfo"`
		} `json:"Datas"`
	}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return []FundSearchItem{}
	}

	items := make([]FundSearchItem, 0, len(result.Datas))
	for _, item := range result.Datas {
		code := strings.TrimSpace(item.Code)
		if code == "" {
			code = strings.TrimSpace(item.CodeLower)
		}
		name := strings.TrimSpace(item.Name)
		if name == "" {
			name = strings.TrimSpace(item.NameLower)
		}
		fundType := strings.TrimSpace(item.Type)
		if item.FundBaseInfo != nil {
			if code == "" {
				code = strings.TrimSpace(item.FundBaseInfo.FCODE)
			}
			if name == "" {
				name = strings.TrimSpace(item.FundBaseInfo.SHORTNAME)
			}
			if fundType == "" {
				fundType = strings.TrimSpace(item.FundBaseInfo.FTYPE)
			}
		}
		if code == "" || name == "" {
			continue
		}
		items = append(items, FundSearchItem{Code: code, Name: name, Type: fundType})

		var count int64
		db.Dao.Model(&FundBasic{}).Where("code = ?", code).Count(&count)
		if count == 0 {
			db.Dao.Create(&FundBasic{Code: code, Name: name, Type: fundType})
		} else {
			db.Dao.Model(&FundBasic{}).Where("code = ?", code).Updates(map[string]any{"name": name, "type": fundType})
		}
	}
	return items
}

func (f *FundApi) GetFundRanking(marketType, fundType, sortField, sortOrder string, pageIndex, pageSize int) (*FundRankingResult, error) {
	defer func() {
		if r := recover(); r != nil {
			logger.SugaredLogger.Errorf("GetFundRanking panic: %v", r)
		}
	}()

	if marketType == "" {
		marketType = "kf"
	}
	if fundType == "" {
		fundType = "all"
	}
	if sortField == "" {
		sortField = "jnzf"
	}
	if sortOrder == "" {
		sortOrder = "desc"
	}
	if pageIndex <= 0 {
		pageIndex = 1
	}
	if pageSize <= 0 {
		pageSize = 50
	}

	referer := "https://fund.eastmoney.com/data/fundranking.html"
	if marketType == "fb" {
		referer = "https://fund.eastmoney.com/data/fbsfundranking.html"
		if fundType == "all" || fundType == "gp" || fundType == "hh" || fundType == "zq" || fundType == "zs" || fundType == "qdii" || fundType == "fof" {
			fundType = "ct"
		}
	}

	queryParams := map[string]string{
		"op": "ph",
		"dt": marketType,
		"ft": fundType,
		"rs": "",
		"gs": "0",
		"sc": sortField,
		"st": sortOrder,
		"sd": "",
		"ed": "",
		"pi": strconv.Itoa(pageIndex),
		"pn": strconv.Itoa(pageSize),
		"v":  strconv.FormatInt(time.Now().UnixMilli(), 10),
	}
	if marketType == "kf" {
		queryParams["qdii"] = ""
		queryParams["tabSubtype"] = ",,,,"
		queryParams["dx"] = "1"
	}

	resp, err := f.client.SetTimeout(time.Duration(f.config.CrawlTimeOut)*time.Second).R().
		SetHeader("User-Agent", getRandomUA()).
		SetHeader("Referer", referer).
		SetQueryParams(queryParams).
		Get("https://fund.eastmoney.com/data/rankhandler.aspx")
	if err != nil {
		return nil, fmt.Errorf("request fund ranking API failed: %w", err)
	}
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("fund ranking HTTP status: %d", resp.StatusCode())
	}

	body := string(resp.Body())
	startIdx := strings.Index(body, "datas:[")
	if startIdx == -1 {
		return nil, fmt.Errorf("fund ranking datas not found")
	}
	startIdx += len("datas:[")
	endIdx := strings.Index(body[startIdx:], "]")
	if endIdx == -1 {
		return nil, fmt.Errorf("fund ranking data format invalid")
	}
	datasContent := body[startIdx : startIdx+endIdx]

	totalCount := 0
	if match := regexp.MustCompile(`allRecords:(\d+)`).FindStringSubmatch(body); len(match) > 1 {
		totalCount, _ = strconv.Atoi(match[1])
	}
	totalPages := 0
	if match := regexp.MustCompile(`allPages:(\d+)`).FindStringSubmatch(body); len(match) > 1 {
		totalPages, _ = strconv.Atoi(match[1])
	}

	records := regexp.MustCompile(`"([^"]*)"`).FindAllStringSubmatch(datasContent, -1)
	items := make([]FundRankingItem, 0, len(records))
	for _, record := range records {
		if len(record) < 2 {
			continue
		}
		fields := strings.Split(record[1], ",")
		if len(fields) < 17 {
			continue
		}
		item := FundRankingItem{
			Code:             fields[0],
			Name:             fields[1],
			Pinyin:           fields[2],
			NetValueDate:     fields[3],
			NetUnitValue:     fundParseFloatPtr(fields[4]),
			NetAccumulated:   fundParseFloatPtr(fields[5]),
			DailyGrowth:      fundParseFloatPtr(fields[6]),
			WeekGrowth:       fundParseFloatPtr(fields[7]),
			MonthGrowth:      fundParseFloatPtr(fields[8]),
			ThreeMonthGrowth: fundParseFloatPtr(fields[9]),
			SixMonthGrowth:   fundParseFloatPtr(fields[10]),
			YearGrowth:       fundParseFloatPtr(fields[11]),
			TwoYearGrowth:    fundParseFloatPtr(fields[12]),
			ThreeYearGrowth:  fundParseFloatPtr(fields[13]),
			YTDGrowth:        fundParseFloatPtr(fields[14]),
			SinceInception:   fundParseFloatPtr(fields[15]),
			EstablishDate:    fields[16],
		}
		if marketType == "kf" && len(fields) >= 21 {
			item.Purchasable = fields[17] == "1"
			item.Scale = fundParseFloatPtr(fields[18])
			item.PurchaseRate = fundParseFloatPtr(fields[19])
			item.DiscountRate = fundParseFloatPtr(fields[20])
		} else if marketType == "fb" && len(fields) >= 23 {
			item.FundTypeDetail = fields[21]
			item.Scale = fundParseFloatPtr(fields[22])
		}
		items = append(items, item)
	}

	return &FundRankingResult{
		Items:      items,
		TotalCount: totalCount,
		PageIndex:  pageIndex,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (f *FundApi) FollowFund(fundCode string) string {
	var fund FundBasic
	db.Dao.Where("code=?", fundCode).First(&fund)
	if fund.Code != "" {
		var follow FollowedFund
		err := db.Dao.Where("code = ?", fundCode).First(&follow).Error
		if err == nil {
			saveErr := db.Dao.Model(&follow).Updates(map[string]any{
				"name":         fund.Name,
				"is_watchlist": true,
			}).Error
			if saveErr != nil {
				return "鍏虫敞澶辫触"
			}
			return "鍏虫敞鎴愬姛"
		}
		if err != nil && err != gorm.ErrRecordNotFound {
			return "鍏虫敞澶辫触"
		}

		follow = FollowedFund{
			Code:        fundCode,
			Name:        fund.Name,
			IsWatchlist: true,
		}
		if createErr := db.Dao.Create(&follow).Error; createErr != nil {
			return "鍏虫敞澶辫触"
		}
		return "鍏虫敞鎴愬姛"
	} else {
		return "基金信息不存在"
	}
}

func (f *FundApi) UpdateFundWatchGroup(fundCode string, watchGroup string) string {
	fundCode = strings.TrimSpace(fundCode)
	watchGroup = strings.TrimSpace(watchGroup)
	if fundCode == "" {
		return "基金代码不能为空"
	}

	var fund FollowedFund
	err := db.Dao.Where("code = ?", fundCode).First(&fund).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "基金自选不存在"
		}
		return "分组保存失败"
	}

	if updateErr := db.Dao.Model(&fund).Updates(map[string]any{"watch_group": watchGroup}).Error; updateErr != nil {
		return "分组保存失败"
	}
	if watchGroup == "" {
		return "已移出分组"
	}
	return "分组已更新"
}

func (f *FundApi) RenameFundWatchGroup(fromGroup string, toGroup string) string {
	fromGroup = strings.TrimSpace(fromGroup)
	toGroup = strings.TrimSpace(toGroup)
	if fromGroup == "" || toGroup == "" {
		return "分组名称不能为空"
	}
	if fromGroup == toGroup {
		return "分组名称未变化"
	}
	if err := db.Dao.Model(&FollowedFund{}).
		Where("is_watchlist = ? AND watch_group = ?", true, fromGroup).
		Update("watch_group", toGroup).Error; err != nil {
		return "分组重命名失败"
	}
	return "分组已重命名"
}

func (f *FundApi) DeleteFundWatchGroup(watchGroup string) string {
	watchGroup = strings.TrimSpace(watchGroup)
	if watchGroup == "" {
		return "分组名称不能为空"
	}
	if err := db.Dao.Model(&FollowedFund{}).
		Where("is_watchlist = ? AND watch_group = ?", true, watchGroup).
		Update("watch_group", "").Error; err != nil {
		return "分组删除失败"
	}
	return "分组已删除，基金已移入未分组"
}

func (f *FundApi) UnFollowFund(fundCode string) string {
	var fund FollowedFund
	db.Dao.Where("code=?", fundCode).First(&fund)
	if fund.Code != "" {
		err := db.Dao.Model(&fund).Updates(map[string]any{
			"is_watchlist": false,
			"watch_group":  "",
		}).Error
		if err != nil {
			return "鍙栨秷鍏虫敞澶辫触"
		}
		return "鍙栨秷鍏虫敞鎴愬姛"
	} else {
		return "基金信息不存在"
	}
}

func (f *FundApi) AllFund() {
	defer func() {
		if r := recover(); r != nil {
			//logger.SugaredLogger.Errorf("AllFund panic: %v", r)
		}
	}()

	response, err := f.client.SetTimeout(time.Duration(f.config.CrawlTimeOut)*time.Second).R().
		SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36").
		Get("https://fund.eastmoney.com/allfund.html")
	if err != nil {
		return
	}
	//涓枃缂栫爜
	htmlContent := GB18030ToUTF8(response.Body())

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	cnt := 0
	doc.Find("ul.num_right li").Each(func(i int, s *goquery.Selection) {
		text := strutil.SplitEx(s.Text(), "|", true)
		if len(text) > 0 {
			cnt++
			name := text[0]
			str := regexp.MustCompile(`[()（）]`).Split(name, -1)
			if len(str) < 2 {
				return
			}
			//logger.SugaredLogger.Infof("%d,鍩洪噾淇℃伅 code:%s,name:%s", cnt, str[0], str[1])
			//go f.CrawlFundBasic(str[0])
			fund := &FundBasic{
				Code: strings.TrimSpace(str[0]),
				Name: strings.TrimSpace(str[1]),
			}
			count := int64(0)
			db.Dao.Model(fund).Where("code=?", fund.Code).Count(&count)
			if count == 0 {
				db.Dao.Create(fund)
			}

		}
	})

}

type FundNetUnitValue struct {
	Fundcode string `json:"fundcode"`
	Name     string `json:"name"`
	Jzrq     string `json:"jzrq"`
	Dwjz     string `json:"dwjz"`
	Gsz      string `json:"gsz"`
	Gszzl    string `json:"gszzl"`
	Gztime   string `json:"gztime"`
}

type fundMobileEstimateResponse struct {
	Datas []struct {
		Code             string `json:"FCODE"`
		Name             string `json:"SHORTNAME"`
		NetEstimatedUnit string `json:"GSZ"`
		NetEstimatedRate string `json:"GSZZL"`
		NetEstimatedTime string `json:"GZTIME"`
	} `json:"Datas"`
	Success bool `json:"Success"`
}

type danjuanFundDetailResponse struct {
	Data struct {
		FundPosition struct {
			StockList []struct {
				Percent          float64 `json:"percent"`
				ChangePercentage float64 `json:"change_percentage"`
			} `json:"stock_list"`
		} `json:"fund_position"`
	} `json:"data"`
	ResultCode int `json:"result_code"`
}

type FundTrendPoint struct {
	Timestamp   int64    `json:"timestamp"`
	Date        string   `json:"date"`
	Value       float64  `json:"value"`
	DailyReturn *float64 `json:"dailyReturn,omitempty"`
}

type FundEstimatePoint struct {
	Timestamp     int64    `json:"timestamp"`
	Time          string   `json:"time"`
	EstimatedUnit float64  `json:"estimatedUnit"`
	EstimatedRate *float64 `json:"estimatedRate"`
}

type FundConfirmedNetUnit struct {
	Name         string
	Code         string
	UnitValue    float64
	PreviousUnit float64
	Date         string
	DailyReturn  *float64
}

type FundStageRanking struct {
	Period             string   `json:"period"`
	ReturnRate         *float64 `json:"returnRate"`
	SimilarAverageRate *float64 `json:"similarAverageRate"`
	BenchmarkLabel     string   `json:"benchmarkLabel"`
	BenchmarkRate      *float64 `json:"benchmarkRate"`
	Rank               int      `json:"rank"`
	RankTotal          int      `json:"rankTotal"`
	RankPercentile     *float64 `json:"rankPercentile"`
	RankDelta          int      `json:"rankDelta"`
	RankDeltaDirection string   `json:"rankDeltaDirection"`
	Quartile           string   `json:"quartile"`
}

type FundIndustryInfo struct {
	Industry   string   `json:"industry"`
	Weight     *float64 `json:"weight"`
	ReportDate string   `json:"reportDate"`
}

type eastmoneyTrendPoint struct {
	X            int64    `json:"x"`
	Y            float64  `json:"y"`
	EquityReturn *float64 `json:"equityReturn"`
}

func (f *FundApi) GetFundTrend(code string) ([]FundTrendPoint, string, *float64, error) {
	response, err := f.client.SetTimeout(time.Duration(f.config.CrawlTimeOut)*time.Second).R().
		SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36").
		SetHeader("Referer", "https://fund.eastmoney.com/").
		Get(fmt.Sprintf("https://fund.eastmoney.com/pingzhongdata/%s.js?v=%d", code, time.Now().UnixMilli()))
	if err != nil {
		return nil, "", nil, err
	}
	if response.StatusCode() != 200 {
		return nil, "", nil, fmt.Errorf("unexpected status code: %d", response.StatusCode())
	}

	content := string(response.Body())
	updatedAt := extractFundTrendUpdatedAt(content)
	match := regexp.MustCompile(`(?s)var\s+Data_netWorthTrend\s*=\s*(\[.*?\]);`).FindStringSubmatch(content)
	if len(match) < 2 {
		return nil, updatedAt, nil, fmt.Errorf("net worth trend not found")
	}

	var raw []eastmoneyTrendPoint
	if err := json.Unmarshal([]byte(match[1]), &raw); err != nil {
		return nil, updatedAt, nil, err
	}

	points := make([]FundTrendPoint, 0, len(raw))
	var latestReturn *float64
	for _, item := range raw {
		day := time.UnixMilli(item.X).Format("2006-01-02")
		points = append(points, FundTrendPoint{
			Timestamp:   item.X,
			Date:        day,
			Value:       item.Y,
			DailyReturn: item.EquityReturn,
		})
		if item.EquityReturn != nil {
			val := *item.EquityReturn
			latestReturn = &val
		}
	}

	return points, updatedAt, latestReturn, nil
}

func (f *FundApi) GetFundEstimatedTrend(code string, day time.Time) ([]FundEstimatePoint, string, *float64, error) {
	return f.GetFundEstimatedTrendBySource(code, day, GetSettingConfig().FundEstimateSource)
}

func (f *FundApi) GetFundEstimatedTrendBySource(code string, day time.Time, source string) ([]FundEstimatePoint, string, *float64, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, "", nil, nil
	}

	tradeDate := day.Format("2006-01-02")
	var snapshots []FundEstimateSnapshot
	if err := db.Dao.Where("code = ? AND trade_date = ? AND source = ?", code, tradeDate, NormalizeFundEstimateSource(source)).
		Order("estimate_time asc").
		Find(&snapshots).Error; err != nil {
		return nil, "", nil, err
	}

	points := make([]FundEstimatePoint, 0, len(snapshots))
	var updatedAt string
	var latestRate *float64
	for _, item := range snapshots {
		points = append(points, FundEstimatePoint{
			Timestamp:     parseFundEstimateTimestamp(item.EstimateTime),
			Time:          item.EstimateTime,
			EstimatedUnit: item.EstimatedUnit,
			EstimatedRate: item.EstimatedRate,
		})
		updatedAt = item.EstimateTime
		if item.EstimatedRate != nil {
			value := *item.EstimatedRate
			latestRate = &value
		}
	}

	return points, updatedAt, latestRate, nil
}

func extractFundTrendUpdatedAt(content string) string {
	match := regexp.MustCompile(`/\*(\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2}:\d{2})\*/`).FindStringSubmatch(content)
	if len(match) < 2 {
		return ""
	}
	return match[1]
}

func (f *FundApi) GetFundStageRankings(code string) ([]FundStageRanking, error) {
	response, err := f.client.SetTimeout(time.Duration(f.config.CrawlTimeOut)*time.Second).R().
		SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36").
		SetHeader("Referer", "https://fundf10.eastmoney.com/").
		Get(fmt.Sprintf("https://fundf10.eastmoney.com/FundArchivesDatas.aspx?type=jdzf&code=%s", code))
	if err != nil {
		return nil, err
	}
	if response.StatusCode() != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", response.StatusCode())
	}

	contentHTML, err := extractFundArchivesContent(string(response.Body()))
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(contentHTML))
	if err != nil {
		return nil, err
	}

	benchmarkLabel := strings.TrimSpace(doc.Find(".jdzfnew ul.fcol li").Eq(3).Text())
	if benchmarkLabel == "" {
		benchmarkLabel = "娌繁300"
	}

	rankings := make([]FundStageRanking, 0, 8)
	doc.Find(".jdzfnew ul").Each(func(i int, row *goquery.Selection) {
		if row.HasClass("fcol") {
			return
		}

		cells := row.ChildrenFiltered("li")
		if cells.Length() < 7 {
			return
		}

		period := strings.TrimSpace(cells.Eq(0).Text())
		if period == "" {
			return
		}

		rank, total := parseFundRankPair(cells.Eq(4).Text())
		percentile := calcFundRankPercentile(rank, total)
		delta, direction := parseFundRankDelta(cells.Eq(5).Text())
		quartile := strings.TrimSpace(cells.Eq(6).Text())
		if quartile == "" {
			quartile = strings.TrimSpace(cells.Eq(6).Find("p.sifen").Text())
		}

		rankings = append(rankings, FundStageRanking{
			Period:             period,
			ReturnRate:         parseFundPercentValue(cells.Eq(1).Text()),
			SimilarAverageRate: parseFundPercentValue(cells.Eq(2).Text()),
			BenchmarkLabel:     benchmarkLabel,
			BenchmarkRate:      parseFundPercentValue(cells.Eq(3).Text()),
			Rank:               rank,
			RankTotal:          total,
			RankPercentile:     percentile,
			RankDelta:          delta,
			RankDeltaDirection: direction,
			Quartile:           quartile,
		})
	})

	return rankings, nil
}

func (f *FundApi) GetFundTopIndustry(code string) (*FundIndustryInfo, error) {
	apiResp, err := f.client.SetTimeout(time.Duration(f.config.CrawlTimeOut)*time.Second).R().
		SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36").
		SetHeader("Referer", "https://fundf10.eastmoney.com/").
		SetHeader("Accept", "application/json, text/javascript, */*; q=0.01").
		Get(fmt.Sprintf("https://api.fund.eastmoney.com/f10/HYPZ/?fundCode=%s&year=&callback=?", code))
	if err == nil && apiResp.StatusCode() == 200 {
		payload := strings.TrimSpace(string(apiResp.Body()))
		payload = strings.TrimPrefix(payload, "?(")
		payload = strings.TrimSuffix(payload, ")")
		if payload != "" {
			var result struct {
				ErrCode int `json:"ErrCode"`
				Data    struct {
					QuarterInfos []struct {
						JZRQ     string `json:"JZRQ"`
						HYPZInfo []struct {
							HYMC      string `json:"HYMC"`
							ZJZBL     string `json:"ZJZBL"`
							ZJZBLDesc string `json:"ZJZBLDesc"`
						} `json:"HYPZInfo"`
					} `json:"QuarterInfos"`
				} `json:"Data"`
			}
			if unmarshalErr := json.Unmarshal([]byte(payload), &result); unmarshalErr == nil && result.ErrCode == 0 {
				for _, quarter := range result.Data.QuarterInfos {
					for _, item := range quarter.HYPZInfo {
						industry := strings.TrimSpace(html.UnescapeString(item.HYMC))
						if industry == "" {
							continue
						}

						rawWeight := strings.TrimSpace(item.ZJZBL)
						if rawWeight == "" {
							rawWeight = strings.TrimSpace(strings.TrimSuffix(item.ZJZBLDesc, "%"))
						}
						weight, parseErr := strconv.ParseFloat(rawWeight, 64)
						if parseErr != nil {
							return nil, parseErr
						}

						return &FundIndustryInfo{
							Industry:   industry,
							Weight:     &weight,
							ReportDate: strings.TrimSpace(quarter.JZRQ),
						}, nil
					}
				}
			}
		}
	}

	return nil, fmt.Errorf("top industry data not found")

	response, err := f.client.SetTimeout(time.Duration(f.config.CrawlTimeOut)*time.Second).R().
		SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36").
		SetHeader("Referer", "https://fundf10.eastmoney.com/").
		Get(fmt.Sprintf("https://fundf10.eastmoney.com/hytz_%s.html", code))
	if err != nil {
		return nil, err
	}
	if response.StatusCode() != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", response.StatusCode())
	}

	content := string(response.Body())
	summaryPattern := regexp.MustCompile(`鎴\s*([0-9]{4}-[0-9]{2}-[0-9]{2})锛?.+?)鍗犲噣鍊兼瘮涓?[0-9.]+)%[^銆俔*鎺掑悕绗竴`)
	match := summaryPattern.FindStringSubmatch(content)
	if len(match) < 4 {
		return nil, fmt.Errorf("top industry summary not found")
	}

	weight, err := strconv.ParseFloat(strings.TrimSpace(match[3]), 64)
	if err != nil {
		return nil, err
	}

	return &FundIndustryInfo{
		Industry:   strings.TrimSpace(html.UnescapeString(match[2])),
		Weight:     &weight,
		ReportDate: strings.TrimSpace(match[1]),
	}, nil
}

func extractFundArchivesContent(body string) (string, error) {
	match := regexp.MustCompile(`content:"((?:\\.|[^"\\])*)"`).FindStringSubmatch(body)
	if len(match) < 2 {
		return "", fmt.Errorf("fund archives content not found")
	}

	unquoted, err := strconv.Unquote(`"` + match[1] + `"`)
	if err != nil {
		return "", err
	}
	return html.UnescapeString(unquoted), nil
}

func parseFundPercentValue(raw string) *float64 {
	cleaned := strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(raw, "%", ""), "＋", "+"))
	if cleaned == "" || cleaned == "---" {
		return nil
	}

	value, err := strconv.ParseFloat(cleaned, 64)
	if err != nil {
		return nil
	}
	return &value
}

func parseFundRankPair(raw string) (int, int) {
	cleaned := strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(raw, "\n", ""), "\t", ""))
	if cleaned == "" || cleaned == "---" {
		return 0, 0
	}

	parts := strings.Split(cleaned, "|")
	if len(parts) != 2 {
		return 0, 0
	}

	rank, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
	total, _ := strconv.Atoi(strings.TrimSpace(parts[1]))
	return rank, total
}

func calcFundRankPercentile(rank int, total int) *float64 {
	if rank <= 0 || total <= 0 {
		return nil
	}

	value := (float64(total-rank+1) / float64(total)) * 100
	value = mathutil.RoundToFloat(value, 2)
	return &value
}

func parseFundRankDelta(raw string) (int, string) {
	cleaned := strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(raw, "\n", ""), "\t", ""))
	if cleaned == "" || cleaned == "---" {
		return 0, ""
	}

	direction := "flat"
	switch {
	case strings.ContainsAny(cleaned, "↑↗"):
		direction = "up"
	case strings.ContainsAny(cleaned, "↓↘"):
		direction = "down"
	case strings.Contains(cleaned, "-"):
		direction = "down"
	}

	numberOnly := strings.NewReplacer("↑", "", "↗", "", "↓", "", "↘", "", "-", "", "+", "", " ", "").Replace(cleaned)
	value, err := strconv.Atoi(numberOnly)
	if err != nil {
		return 0, direction
	}
	return value, direction
}

// CrawlFundNetEstimatedUnit fetches intraday estimated unit value data.
func (f *FundApi) CrawlFundNetEstimatedUnit(code string) {
	f.CrawlFundNetEstimatedUnitFromSource(code, "eastmoney_js")
}

func (f *FundApi) CrawlFundNetEstimatedUnitFromSource(code string, source string) {
	switch NormalizeFundEstimateSource(source) {
	case "eastmoney_mobile":
		f.crawlFundNetEstimatedUnitFromEastmoneyMobile(code)
	case "danjuan_position":
		f.crawlFundNetEstimatedUnitFromDanjuanPosition(code)
	case "ai_corrected":
		f.crawlFundNetEstimatedUnitFromAICorrected(code)
	default:
		f.crawlFundNetEstimatedUnitFromEastmoneyJS(code)
	}
}

type fundEstimateSourceScore struct {
	Bias      float64
	MAE       float64
	Samples   int
	Intercept float64
	Slope     float64
	LinearMAE float64
}

func (f *FundApi) crawlFundNetEstimatedUnitFromAICorrected(code string) {
	code = strings.TrimSpace(code)
	if code == "" {
		return
	}

	baseSources := aiCorrectedCurrentSources()
	for _, source := range baseSources {
		f.CrawlFundNetEstimatedUnitFromSource(code, source)
	}

	var followed FollowedFund
	if err := db.Dao.Preload("FundBasic").Where("code = ?", code).First(&followed).Error; err != nil {
		return
	}
	if followed.NetUnitValue == nil || *followed.NetUnitValue <= 0 {
		return
	}

	latestSnapshots := f.latestTodayFundEstimateSnapshots(code, baseSources)
	if len(latestSnapshots) == 0 {
		return
	}

	actualRates := f.recentActualFundDailyReturns(code)
	category := inferAICorrectedFundCategory(defaultLabelForFundEstimate(followed.FundBasic.Type, defaultLabelForFundEstimate(followed.FundBasic.Name, followed.Name)))
	name := defaultLabelForFundEstimate(followed.Name, followed.FundBasic.Name)
	today := time.Now().Format("2006-01-02")
	historySnapshots := f.latestFundEstimateSnapshotsSince(code, time.Now().AddDate(0, 0, -75).Format("2006-01-02"), aiCorrectedHistorySources())
	for source, snapshot := range latestSnapshots {
		historySnapshots[today+"|"+canonicalAIEstimateSource(source)] = snapshot
	}
	corrected, ok := buildAICorrectedSnapshotForDate(code, name, today, baseSources, historySnapshots, actualRates, category, followed.NetUnitValue)
	if !ok || corrected.EstimatedRate == nil {
		return
	}

	fund := &FollowedFund{
		Code:             code,
		Name:             name,
		NetEstimatedTime: corrected.EstimateTime,
		NetEstimatedUnit: &corrected.EstimatedUnit,
		NetEstimatedRate: corrected.EstimatedRate,
	}
	db.Dao.Model(fund).Where("code=?", fund.Code).Updates(fund)
	f.saveFundEstimateSnapshotWithSource(fund.Code, fund.Name, fund.NetEstimatedTime, fund.NetEstimatedUnit, fund.NetEstimatedRate, "ai_corrected")
	f.BackfillAICorrectedFundEstimates(code, 45)
}

func (f *FundApi) BackfillAICorrectedFundEstimates(code string, days int) int {
	code = strings.TrimSpace(code)
	if code == "" {
		return 0
	}
	if days <= 0 {
		days = 45
	}

	baseSources := aiCorrectedCurrentSources()
	startDate := time.Now().AddDate(0, 0, -days).Format("2006-01-02")
	scoreStartDate := time.Now().AddDate(0, 0, -75).Format("2006-01-02")
	today := time.Now().Format("2006-01-02")

	actualRates := f.recentActualFundDailyReturnsSince(code, scoreStartDate)
	if len(actualRates) == 0 {
		return 0
	}

	var followed FollowedFund
	_ = db.Dao.Preload("FundBasic").Where("code = ?", code).First(&followed).Error
	category := inferAICorrectedFundCategory(defaultLabelForFundEstimate(followed.FundBasic.Type, defaultLabelForFundEstimate(followed.FundBasic.Name, followed.Name)))
	name := defaultLabelForFundEstimate(followed.Name, followed.FundBasic.Name)

	var snapshots []FundEstimateSnapshot
	if err := db.Dao.Where("code = ? AND trade_date >= ? AND trade_date < ? AND source IN ? AND estimated_rate IS NOT NULL AND estimated_unit > 0", code, startDate, today, aiCorrectedHistorySources()).
		Order("trade_date asc, source asc, estimate_time asc").
		Find(&snapshots).Error; err != nil {
		logger.SugaredLogger.Warnf("query fund estimate snapshots for ai backfill failed for %s: %v", code, err)
		return 0
	}
	historySnapshots := f.latestFundEstimateSnapshotsSince(code, scoreStartDate, aiCorrectedHistorySources())
	if len(snapshots) == 0 && len(historySnapshots) == 0 {
		return 0
	}

	dateSet := make(map[string]bool)
	for _, snapshot := range snapshots {
		if snapshot.EstimatedRate == nil || strings.TrimSpace(snapshot.TradeDate) == "" {
			continue
		}
		dateSet[snapshot.TradeDate] = true
	}

	created := 0
	for tradeDate := range dateSet {
		if _, ok := actualRates[tradeDate]; !ok {
			continue
		}
		var existing FundEstimateSnapshot
		err := db.Dao.Where("code = ? AND trade_date = ? AND source = ?", code, tradeDate, "ai_corrected").First(&existing).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			continue
		}

		corrected, ok := buildAICorrectedSnapshotForDate(code, name, tradeDate, baseSources, historySnapshots, actualRates, category, nil)
		if !ok {
			continue
		}
		if err == nil {
			if updateErr := db.Dao.Model(&existing).Updates(map[string]interface{}{
				"name":           corrected.Name,
				"estimate_time":  corrected.EstimateTime,
				"estimated_unit": corrected.EstimatedUnit,
				"estimated_rate": corrected.EstimatedRate,
			}).Error; updateErr == nil {
				created++
			}
			continue
		}
		if err := db.Dao.Create(&corrected).Error; err == nil {
			created++
		}
	}
	return created
}

func (f *FundApi) latestTodayFundEstimateSnapshots(code string, sources []string) map[string]FundEstimateSnapshot {
	result := make(map[string]FundEstimateSnapshot)
	today := time.Now().Format("2006-01-02")
	for _, source := range sources {
		var snapshot FundEstimateSnapshot
		err := db.Dao.Where("code = ? AND trade_date = ? AND source = ? AND estimated_unit > 0 AND estimated_rate IS NOT NULL", code, today, source).
			Order("estimate_time desc").
			First(&snapshot).Error
		if err == nil {
			result[source] = snapshot
		}
	}
	return result
}

func (f *FundApi) recentActualFundDailyReturns(code string) map[string]float64 {
	return f.recentActualFundDailyReturnsSince(code, time.Now().AddDate(0, 0, -75).Format("2006-01-02"))
}

func (f *FundApi) recentActualFundDailyReturnsSince(code string, startDate string) map[string]float64 {
	result := make(map[string]float64)
	trend, _, _, err := f.GetFundTrend(code)
	if err != nil {
		return result
	}
	for _, point := range trend {
		if point.DailyReturn == nil || strings.TrimSpace(point.Date) == "" || point.Date < startDate {
			continue
		}
		result[point.Date] = *point.DailyReturn
	}
	return result
}

func (f *FundApi) estimateSourceScores(code string, sources []string, actualRates map[string]float64) map[string]fundEstimateSourceScore {
	return f.estimateSourceScoresUntil(code, sources, actualRates, "")
}

func (f *FundApi) estimateSourceScoresUntil(code string, sources []string, actualRates map[string]float64, beforeDate string) map[string]fundEstimateSourceScore {
	if len(actualRates) == 0 {
		return defaultFundEstimateSourceScores(sources)
	}

	startDate := time.Now().AddDate(0, 0, -75).Format("2006-01-02")
	latestByDateSource := f.latestFundEstimateSnapshotsSince(code, startDate, aiCorrectedHistorySources())
	return estimateSourceScoresFromSnapshots(sources, actualRates, latestByDateSource, beforeDate)
}

func defaultFundEstimateSourceScores(sources []string) map[string]fundEstimateSourceScore {
	result := make(map[string]fundEstimateSourceScore, len(sources))
	for _, source := range sources {
		result[canonicalAIEstimateSource(source)] = fundEstimateSourceScore{MAE: 0.45}
	}
	return result
}

func estimateSourceScoresFromSnapshots(sources []string, actualRates map[string]float64, latestByDateSource map[string]FundEstimateSnapshot, beforeDate string) map[string]fundEstimateSourceScore {
	result := defaultFundEstimateSourceScores(sources)
	if len(actualRates) == 0 || len(latestByDateSource) == 0 {
		return result
	}
	type acc struct {
		errSum float64
		absSum float64
		count  int
		xSum   float64
		ySum   float64
		xxSum  float64
		xySum  float64
		xs     []float64
		ys     []float64
	}
	accBySource := make(map[string]acc)
	for _, snapshot := range latestByDateSource {
		if beforeDate != "" && snapshot.TradeDate >= beforeDate {
			continue
		}
		actualRate, ok := actualRates[snapshot.TradeDate]
		if !ok || snapshot.EstimatedRate == nil {
			continue
		}
		source := canonicalAIEstimateSource(snapshot.Source)
		errorValue := *snapshot.EstimatedRate - actualRate
		item := accBySource[source]
		item.errSum += errorValue
		item.absSum += absFloat64(errorValue)
		item.count++
		item.xSum += *snapshot.EstimatedRate
		item.ySum += actualRate
		item.xxSum += *snapshot.EstimatedRate * *snapshot.EstimatedRate
		item.xySum += *snapshot.EstimatedRate * actualRate
		item.xs = append(item.xs, *snapshot.EstimatedRate)
		item.ys = append(item.ys, actualRate)
		accBySource[source] = item
	}

	for _, source := range sources {
		canonicalSource := canonicalAIEstimateSource(source)
		item := accBySource[canonicalSource]
		if item.count == 0 {
			continue
		}
		score := fundEstimateSourceScore{
			Bias:    item.errSum / float64(item.count),
			MAE:     item.absSum / float64(item.count),
			Samples: item.count,
		}
		denominator := float64(item.count)*item.xxSum - item.xSum*item.xSum
		if item.count >= 5 && absFloat64(denominator) > 0.000001 {
			score.Slope = (float64(item.count)*item.xySum - item.xSum*item.ySum) / denominator
			score.Intercept = (item.ySum - score.Slope*item.xSum) / float64(item.count)
		} else {
			score.Slope = 1
			score.Intercept = -score.Bias
		}
		score.LinearMAE = leaveOneOutLinearMAE(item.xs, item.ys)
		result[canonicalSource] = score
	}
	return result
}

func leaveOneOutLinearMAE(xs []float64, ys []float64) float64 {
	if len(xs) < 8 || len(xs) != len(ys) {
		return 999
	}
	total := 0.0
	count := 0
	for excluded := range xs {
		intercept, slope, ok := fitLinearEstimateExcluding(xs, ys, excluded)
		if !ok || absFloat64(slope) > 4 {
			continue
		}
		predicted := intercept + slope*xs[excluded]
		total += absFloat64(predicted - ys[excluded])
		count++
	}
	if count == 0 {
		return 999
	}
	return total / float64(count)
}

func fitLinearEstimateExcluding(xs []float64, ys []float64, excluded int) (float64, float64, bool) {
	count := 0
	xSum := 0.0
	ySum := 0.0
	xxSum := 0.0
	xySum := 0.0
	for i := range xs {
		if i == excluded {
			continue
		}
		count++
		xSum += xs[i]
		ySum += ys[i]
		xxSum += xs[i] * xs[i]
		xySum += xs[i] * ys[i]
	}
	denominator := float64(count)*xxSum - xSum*xSum
	if count < 5 || absFloat64(denominator) < 0.000001 {
		return 0, 0, false
	}
	slope := (float64(count)*xySum - xSum*ySum) / denominator
	intercept := (ySum - slope*xSum) / float64(count)
	return intercept, slope, true
}

func (f *FundApi) latestFundEstimateSnapshotsSince(code string, startDate string, sources []string) map[string]FundEstimateSnapshot {
	result := make(map[string]FundEstimateSnapshot)
	var snapshots []FundEstimateSnapshot
	if err := db.Dao.Where("code = ? AND trade_date >= ? AND source IN ? AND estimated_rate IS NOT NULL AND estimated_unit > 0", code, startDate, sources).
		Order("trade_date asc, source asc, estimate_time asc").
		Find(&snapshots).Error; err != nil {
		return result
	}
	for _, snapshot := range snapshots {
		if snapshot.EstimatedRate == nil || strings.TrimSpace(snapshot.TradeDate) == "" {
			continue
		}
		source := canonicalAIEstimateSource(snapshot.Source)
		key := snapshot.TradeDate + "|" + source
		existing, ok := result[key]
		if !ok || snapshot.EstimateTime >= existing.EstimateTime || (snapshot.Source == "eastmoney_js" && existing.Source == "eastmoney") {
			snapshot.Source = source
			result[key] = snapshot
		}
	}
	return result
}

type aiEstimateCandidate struct {
	Name         string
	Source       string
	Rate         float64
	Unit         float64
	EstimateTime string
	Raw          bool
}

type aiEstimateModelScore struct {
	MAE     float64
	Samples int
}

func buildAICorrectedSnapshotForDate(code string, name string, tradeDate string, sources []string, latestByDateSource map[string]FundEstimateSnapshot, actualRates map[string]float64, category string, baseNetUnitValue *float64) (FundEstimateSnapshot, bool) {
	candidate, ok := selectAICorrectedCandidate(tradeDate, sources, latestByDateSource, actualRates, category)
	if !ok {
		return FundEstimateSnapshot{}, false
	}
	estimatedRate := mathutil.RoundToFloat(candidate.Rate, 2)
	estimatedUnit := candidate.Unit
	if baseNetUnitValue != nil && *baseNetUnitValue > 0 {
		estimatedUnit = mathutil.RoundToFloat(*baseNetUnitValue*(1+estimatedRate/100), 4)
	} else {
		estimatedUnit = mathutil.RoundToFloat(estimatedUnit, 4)
	}
	if estimatedUnit <= 0 {
		return FundEstimateSnapshot{}, false
	}

	return FundEstimateSnapshot{
		Code:          code,
		Name:          strings.TrimSpace(name),
		TradeDate:     tradeDate,
		EstimateTime:  candidate.EstimateTime,
		EstimatedUnit: estimatedUnit,
		EstimatedRate: &estimatedRate,
		Source:        "ai_corrected",
	}, true
}

func selectAICorrectedCandidate(tradeDate string, sources []string, latestByDateSource map[string]FundEstimateSnapshot, actualRates map[string]float64, category string) (aiEstimateCandidate, bool) {
	sourceScores := estimateSourceScoresFromSnapshots(sources, actualRates, latestByDateSource, "")
	targetCandidates := buildAIEstimateCandidatesForDate(tradeDate, sources, latestByDateSource, sourceScores, category)
	if len(targetCandidates) == 0 {
		return aiEstimateCandidate{}, false
	}

	walkScores := walkForwardAIEstimateModelScores(tradeDate, sources, latestByDateSource, actualRates, category)
	bestRaw, bestRawScore, hasRaw := bestRawAIEstimateCandidate(sources, targetCandidates, walkScores, sourceScores)
	if !hasRaw {
		return aiEstimateCandidate{}, false
	}

	selected := bestRaw
	selectedScore := bestRawScore
	minImprovement := 0.015
	if bestRawScore.MAE > 0.4 {
		minImprovement = bestRawScore.MAE * 0.08
	}
	for name, candidate := range targetCandidates {
		if candidate.Raw {
			continue
		}
		if strings.HasPrefix(name, "linear:") {
			sourceScore := sourceScores[candidate.Source]
			if sourceScore.Samples >= 10 && sourceScore.LinearMAE+minImprovement < bestRawScore.MAE {
				if selected.Raw || sourceScore.LinearMAE < selectedScore.MAE {
					selected = candidate
					selectedScore = aiEstimateModelScore{MAE: sourceScore.LinearMAE, Samples: sourceScore.Samples}
				}
			}
			continue
		}
		score, ok := walkScores[name]
		if !ok || score.Samples < 4 {
			continue
		}
		if bestRawScore.Samples > 0 && score.MAE+minImprovement >= bestRawScore.MAE {
			continue
		}
		if selected.Raw || selectedScore.Samples == 0 || score.MAE < selectedScore.MAE {
			selected = candidate
			selectedScore = score
		}
	}
	return selected, true
}

func buildAIEstimateCandidatesForDate(tradeDate string, sources []string, latestByDateSource map[string]FundEstimateSnapshot, sourceScores map[string]fundEstimateSourceScore, category string) map[string]aiEstimateCandidate {
	result := make(map[string]aiEstimateCandidate)
	for _, source := range sources {
		source = canonicalAIEstimateSource(source)
		snapshot, ok := latestByDateSource[tradeDate+"|"+source]
		if !ok || snapshot.EstimatedRate == nil || snapshot.EstimatedUnit <= 0 {
			continue
		}
		rawRate := *snapshot.EstimatedRate
		result["raw:"+source] = aiEstimateCandidate{
			Name:         "raw:" + source,
			Source:       source,
			Rate:         rawRate,
			Unit:         snapshot.EstimatedUnit,
			EstimateTime: snapshot.EstimateTime,
			Raw:          true,
		}

		score := sourceScores[source]
		if score.Samples >= 3 {
			correctedRate := rawRate - shrunkenAIEstimateBias(score, category)
			maxCorrection := aiCorrectedMaxCorrection(category, score)
			correctedRate = clampFundEstimateRate(correctedRate, rawRate-maxCorrection, rawRate+maxCorrection)
			result["corrected:"+source] = aiEstimateCandidate{
				Name:         "corrected:" + source,
				Source:       source,
				Rate:         correctedRate,
				Unit:         snapshot.EstimatedUnit,
				EstimateTime: snapshot.EstimateTime,
			}
		}
		if score.Samples >= 10 && score.MAE >= 0.25 && score.LinearMAE < score.MAE*0.9 && score.Slope != 0 && absFloat64(score.Slope) <= 4 {
			linearRate := score.Intercept + score.Slope*rawRate
			maxCorrection := aiCorrectedMaxCorrection(category, score)
			linearRate = clampFundEstimateRate(linearRate, rawRate-maxCorrection, rawRate+maxCorrection)
			result["linear:"+source] = aiEstimateCandidate{
				Name:         "linear:" + source,
				Source:       source,
				Rate:         linearRate,
				Unit:         snapshot.EstimatedUnit,
				EstimateTime: snapshot.EstimateTime,
			}
		}
	}
	if ensemble, ok := buildAIEnsembleCandidateForDate(tradeDate, sources, latestByDateSource, sourceScores, category); ok {
		result[ensemble.Name] = ensemble
	}
	return result
}

func buildAIEnsembleCandidateForDate(tradeDate string, sources []string, latestByDateSource map[string]FundEstimateSnapshot, sourceScores map[string]fundEstimateSourceScore, category string) (aiEstimateCandidate, bool) {
	totalWeight := 0.0
	weightedRate := 0.0
	weightedUnit := 0.0
	latestEstimateTime := ""
	rawMin := 0.0
	rawMax := 0.0
	hasRawRange := false

	for _, source := range sources {
		source = canonicalAIEstimateSource(source)
		snapshot, ok := latestByDateSource[tradeDate+"|"+source]
		if !ok || snapshot.EstimatedRate == nil || snapshot.EstimatedUnit <= 0 {
			continue
		}
		score := sourceScores[source]
		if score.Samples < 3 {
			continue
		}
		weight := aiCorrectedSourceWeight(source, category, score)
		if weight <= 0 {
			continue
		}
		rawRate := *snapshot.EstimatedRate
		correctedRate := rawRate - shrunkenAIEstimateBias(score, category)
		weightedRate += correctedRate * weight
		weightedUnit += snapshot.EstimatedUnit * weight
		totalWeight += weight
		if latestEstimateTime == "" || snapshot.EstimateTime > latestEstimateTime {
			latestEstimateTime = snapshot.EstimateTime
		}
		if !hasRawRange {
			rawMin = rawRate
			rawMax = rawRate
			hasRawRange = true
		} else {
			if rawRate < rawMin {
				rawMin = rawRate
			}
			if rawRate > rawMax {
				rawMax = rawRate
			}
		}
	}

	if totalWeight <= 0 || latestEstimateTime == "" {
		return aiEstimateCandidate{}, false
	}
	estimatedRate := weightedRate / totalWeight
	if hasRawRange {
		maxCorrection := aiCorrectedMaxCorrection(category, fundEstimateSourceScore{MAE: absFloat64(rawMax - rawMin)})
		estimatedRate = clampFundEstimateRate(estimatedRate, rawMin-maxCorrection, rawMax+maxCorrection)
	}
	return aiEstimateCandidate{
		Name:         "ensemble",
		Source:       "ensemble",
		Rate:         estimatedRate,
		Unit:         weightedUnit / totalWeight,
		EstimateTime: latestEstimateTime,
	}, true
}

func walkForwardAIEstimateModelScores(targetDate string, sources []string, latestByDateSource map[string]FundEstimateSnapshot, actualRates map[string]float64, category string) map[string]aiEstimateModelScore {
	type acc struct {
		sum   float64
		count int
	}
	accByModel := make(map[string]acc)
	dates := make([]string, 0, len(actualRates))
	for date := range actualRates {
		if targetDate != "" && date >= targetDate {
			continue
		}
		dates = append(dates, date)
	}
	sort.Strings(dates)
	for _, date := range dates {
		actualRate, ok := actualRates[date]
		if !ok {
			continue
		}
		sourceScores := estimateSourceScoresFromSnapshots(sources, actualRates, latestByDateSource, date)
		candidates := buildAIEstimateCandidatesForDate(date, sources, latestByDateSource, sourceScores, category)
		for name, candidate := range candidates {
			item := accByModel[name]
			item.sum += absFloat64(candidate.Rate - actualRate)
			item.count++
			accByModel[name] = item
		}
	}
	result := make(map[string]aiEstimateModelScore, len(accByModel))
	for name, item := range accByModel {
		if item.count == 0 {
			continue
		}
		result[name] = aiEstimateModelScore{
			MAE:     item.sum / float64(item.count),
			Samples: item.count,
		}
	}
	return result
}

func bestRawAIEstimateCandidate(sources []string, candidates map[string]aiEstimateCandidate, walkScores map[string]aiEstimateModelScore, sourceScores map[string]fundEstimateSourceScore) (aiEstimateCandidate, aiEstimateModelScore, bool) {
	var best aiEstimateCandidate
	bestScore := aiEstimateModelScore{MAE: 999, Samples: 0}
	hasBest := false
	for _, source := range sources {
		source = canonicalAIEstimateSource(source)
		candidate, ok := candidates["raw:"+source]
		if !ok {
			continue
		}
		score, ok := walkScores[candidate.Name]
		if !ok {
			rawScore := sourceScores[source]
			score = aiEstimateModelScore{MAE: rawScore.MAE, Samples: rawScore.Samples}
		}
		if !hasBest || score.Samples > 0 && (bestScore.Samples == 0 || score.MAE < bestScore.MAE) {
			best = candidate
			bestScore = score
			hasBest = true
		}
	}
	if hasBest {
		return best, bestScore, true
	}
	for _, source := range sources {
		source = canonicalAIEstimateSource(source)
		if candidate, ok := candidates["raw:"+source]; ok {
			return candidate, aiEstimateModelScore{}, true
		}
	}
	return aiEstimateCandidate{}, aiEstimateModelScore{}, false
}

func aiCorrectedCurrentSources() []string {
	return []string{"eastmoney_js", "eastmoney_mobile", "danjuan_position"}
}

func aiCorrectedHistorySources() []string {
	return []string{"eastmoney_js", "eastmoney_mobile", "danjuan_position"}
}

func canonicalAIEstimateSource(source string) string {
	source = NormalizeFundEstimateSource(source)
	if source == "eastmoney" {
		return "eastmoney_js"
	}
	return source
}

func shrunkenAIEstimateBias(score fundEstimateSourceScore, category string) float64 {
	if score.Samples < 3 {
		return 0
	}
	shrink := float64(score.Samples) / float64(score.Samples+4)
	maxCorrection := aiCorrectedMaxCorrection(category, score)
	return clampFundEstimateRate(score.Bias*shrink, -maxCorrection, maxCorrection)
}

func aiCorrectedMaxCorrection(category string, score fundEstimateSourceScore) float64 {
	if score.MAE > 0.45 {
		return mathutil.RoundToFloat(clampFundEstimateRate(score.MAE*2.2, 0.8, 3.5), 2)
	}
	switch category {
	case "stable":
		return 0.12
	case "equity":
		return 0.8
	default:
		return 0.45
	}
}

func aiCorrectedSourceWeight(source string, category string, score fundEstimateSourceScore) float64 {
	mae := score.MAE
	if mae <= 0 {
		mae = 0.08
	}
	weight := 1 / (mae + 0.08)
	if score.Samples == 0 {
		weight *= 0.35
	} else if score.Samples < 3 {
		weight *= 0.65
	}

	switch category {
	case "stable":
		if source == "eastmoney_js" || source == "eastmoney_mobile" {
			weight *= 1.25
		}
		if source == "danjuan_position" {
			weight *= 0.55
		}
	case "equity":
		if source == "danjuan_position" {
			weight *= 1.18
		}
	}
	return weight
}

func inferAICorrectedFundCategory(text string) string {
	normalized := strings.ToUpper(strings.TrimSpace(text))
	switch {
	case strings.Contains(normalized, "货币"), strings.Contains(normalized, "现金"), strings.Contains(normalized, "同业存单"),
		strings.Contains(normalized, "债"), strings.Contains(normalized, "BOND"), strings.Contains(normalized, "CASH"):
		return "stable"
	case strings.Contains(normalized, "股票"), strings.Contains(normalized, "混合"), strings.Contains(normalized, "指数"),
		strings.Contains(normalized, "ETF"), strings.Contains(normalized, "QDII"), strings.Contains(normalized, "FOF"),
		strings.Contains(normalized, "STOCK"), strings.Contains(normalized, "EQUITY"):
		return "equity"
	default:
		return "other"
	}
}

func defaultLabelForFundEstimate(primary string, fallback string) string {
	primary = strings.TrimSpace(primary)
	if primary != "" {
		return primary
	}
	return strings.TrimSpace(fallback)
}

func clampFundEstimateRate(value float64, low float64, high float64) float64 {
	if low > high {
		return value
	}
	if value < low {
		return low
	}
	if value > high {
		return high
	}
	return value
}

func absFloat64(value float64) float64 {
	if value < 0 {
		return -value
	}
	return value
}

func (f *FundApi) crawlFundNetEstimatedUnitFromEastmoneyJS(code string) {
	var fundNetUnitValue FundNetUnitValue
	response, err := f.client.SetTimeout(time.Duration(f.config.CrawlTimeOut)*time.Second).R().
		SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36").
		SetHeader("Referer", "https://fund.eastmoney.com/").
		SetQueryParams(map[string]string{"rt": strconv.FormatInt(time.Now().UnixMilli(), 10)}).
		Get(fmt.Sprintf("https://fundgz.1234567.com.cn/js/%s.js", code))
	if err != nil {
		logger.SugaredLogger.Errorf("err:%s", err.Error())
		return
	}
	if response.StatusCode() == 200 {
		htmlContent := string(response.Body())
		//logger.SugaredLogger.Infof("htmlContent:%s", htmlContent)
		if strings.Contains(htmlContent, "jsonpgz") {
			htmlContent = strutil.Trim(htmlContent, "jsonpgz(", ");")
			htmlContent = strutil.Trim(htmlContent, ");")
			//logger.SugaredLogger.Infof("鍩洪噾鍑€鍊间俊鎭?%s", htmlContent)
			err := json.Unmarshal([]byte(htmlContent), &fundNetUnitValue)
			if err != nil {
				//logger.SugaredLogger.Errorf("json.Unmarshal error:%s", err.Error())
				return
			}
			fund := &FollowedFund{
				Code:             fundNetUnitValue.Fundcode,
				Name:             fundNetUnitValue.Name,
				NetEstimatedTime: fundNetUnitValue.Gztime,
			}
			netEstimatedUnit, err := convertor.ToFloat(fundNetUnitValue.Gsz)
			if err == nil {
				fund.NetEstimatedUnit = &netEstimatedUnit
			}
			netEstimatedRate, err := convertor.ToFloat(fundNetUnitValue.Gszzl)
			if err == nil {
				fund.NetEstimatedRate = &netEstimatedRate
			}
			db.Dao.Model(fund).Where("code=?", fund.Code).Updates(fund)
			f.saveFundEstimateSnapshotWithSource(fund.Code, fund.Name, fund.NetEstimatedTime, fund.NetEstimatedUnit, fund.NetEstimatedRate, "eastmoney_js")
		}
	}
}

func (f *FundApi) crawlFundNetEstimatedUnitFromEastmoneyMobile(code string) {
	code = strings.TrimSpace(code)
	if code == "" {
		return
	}
	var mobileResp fundMobileEstimateResponse
	response, err := f.client.SetTimeout(time.Duration(f.config.CrawlTimeOut)*time.Second).R().
		SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36").
		SetHeader("Referer", "https://fund.eastmoney.com/").
		SetFormData(map[string]string{
			"pageIndex": "1",
			"pageSize":  "1",
			"appType":   "ttjj",
			"product":   "EFund",
			"plat":      "Android",
			"Version":   "6.2.4",
			"deviceid":  "9e16077fca2fcr78ep0ltn98",
			"Fcodes":    code,
		}).
		Post("https://fundmobapi.eastmoney.com/FundMNewApi/FundMNFInfo")
	if err != nil {
		logger.SugaredLogger.Errorf("crawl mobile fund estimate failed for %s: %v", code, err)
		return
	}
	if response.StatusCode() != 200 {
		return
	}
	if err := json.Unmarshal(response.Body(), &mobileResp); err != nil || len(mobileResp.Datas) == 0 {
		return
	}
	item := mobileResp.Datas[0]
	itemCode := strings.TrimSpace(item.Code)
	if itemCode == "" {
		itemCode = code
	}
	fund := &FollowedFund{
		Code:             itemCode,
		Name:             item.Name,
		NetEstimatedTime: item.NetEstimatedTime,
	}
	netEstimatedUnit, err := convertor.ToFloat(item.NetEstimatedUnit)
	if err == nil && netEstimatedUnit > 0 {
		fund.NetEstimatedUnit = &netEstimatedUnit
	}
	netEstimatedRate, err := convertor.ToFloat(item.NetEstimatedRate)
	if err == nil {
		fund.NetEstimatedRate = &netEstimatedRate
	}
	if fund.NetEstimatedUnit == nil || strings.TrimSpace(fund.NetEstimatedTime) == "" {
		logger.SugaredLogger.Warnf("mobile fund estimate is empty for %s", code)
		return
	}
	db.Dao.Model(fund).Where("code=?", fund.Code).Updates(fund)
	f.saveFundEstimateSnapshotWithSource(fund.Code, fund.Name, fund.NetEstimatedTime, fund.NetEstimatedUnit, fund.NetEstimatedRate, "eastmoney_mobile")
}

func (f *FundApi) crawlFundNetEstimatedUnitFromDanjuanPosition(code string) {
	code = strings.TrimSpace(code)
	if code == "" {
		return
	}
	response, err := f.client.SetTimeout(time.Duration(f.config.CrawlTimeOut)*time.Second).R().
		SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36").
		SetHeader("Referer", "https://danjuanapp.com/").
		Get(fmt.Sprintf("https://danjuanapp.com/djapi/fund/detail/%s", code))
	if err != nil {
		logger.SugaredLogger.Errorf("crawl danjuan fund detail failed for %s: %v", code, err)
		return
	}
	if response.StatusCode() != 200 {
		return
	}
	var detail danjuanFundDetailResponse
	if err := json.Unmarshal(response.Body(), &detail); err != nil || len(detail.Data.FundPosition.StockList) == 0 {
		return
	}
	estimatedRate := 0.0
	for _, stock := range detail.Data.FundPosition.StockList {
		estimatedRate += stock.Percent * stock.ChangePercentage / 100
	}

	var followed FollowedFund
	if err := db.Dao.Where("code = ?", code).First(&followed).Error; err != nil {
		return
	}
	if followed.NetUnitValue == nil || *followed.NetUnitValue <= 0 {
		return
	}
	estimatedUnit := mathutil.RoundToFloat(*followed.NetUnitValue*(1+estimatedRate/100), 4)
	estimatedRate = mathutil.RoundToFloat(estimatedRate, 2)
	estimateTime := time.Now().Format("2006-01-02 15:04")
	fund := &FollowedFund{
		Code:             code,
		Name:             followed.Name,
		NetEstimatedTime: estimateTime,
		NetEstimatedUnit: &estimatedUnit,
		NetEstimatedRate: &estimatedRate,
	}
	db.Dao.Model(fund).Where("code=?", fund.Code).Updates(fund)
	f.saveFundEstimateSnapshotWithSource(fund.Code, fund.Name, fund.NetEstimatedTime, fund.NetEstimatedUnit, fund.NetEstimatedRate, "danjuan_position")
}

// CrawlFundNetUnitValue fetches latest confirmed net unit value data.
func (f *FundApi) CrawlFundNetUnitValue(code string) {
	confirmed, err := f.GetFundConfirmedNetUnitValue(code)
	if err != nil {
		logger.SugaredLogger.Errorf("err:%s", err.Error())
		return
	}
	if confirmed == nil {
		return
	}
	fund := &FollowedFund{
		Name:             confirmed.Name,
		Code:             confirmed.Code,
		NetUnitValue:     &confirmed.UnitValue,
		NetUnitValueDate: confirmed.Date,
	}
	db.Dao.Model(fund).Where("code=?", fund.Code).Updates(fund)
}

func (f *FundApi) GetFundConfirmedNetUnitValue(code string) (*FundConfirmedNetUnit, error) {
	url := fmt.Sprintf("http://hq.sinajs.cn/rn=%d&list=f_%s", time.Now().UnixMilli(), code)
	response, err := f.client.SetTimeout(time.Duration(f.config.CrawlTimeOut)*time.Second).R().
		SetHeader("Host", "hq.sinajs.cn").
		SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36 Edg/119.0.0.0").
		SetHeader("Referer", "https://finance.sina.com.cn").
		Get(url)
	if err != nil {
		return nil, err
	}
	if response.StatusCode() != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", response.StatusCode())
	}
	raw := string(GB18030ToUTF8(response.Body()))
	datas := strutil.SplitAndTrim(raw, "=", "\"")
	if len(datas) < 2 {
		return nil, fmt.Errorf("invalid sina fund quote for %s", code)
	}
	parts := strutil.SplitAndTrim(datas[1], ",", "\"")
	if len(parts) < 5 {
		return nil, fmt.Errorf("invalid sina fund quote parts for %s", code)
	}
	unitValue, err := convertor.ToFloat(parts[1])
	if err != nil {
		return nil, err
	}
	previousUnit, _ := convertor.ToFloat(parts[3])
	var dailyReturn *float64
	if previousUnit > 0 {
		value := mathutil.RoundToFloat((unitValue/previousUnit-1)*100, 2)
		dailyReturn = &value
	}
	return &FundConfirmedNetUnit{
		Name:         parts[0],
		Code:         code,
		UnitValue:    unitValue,
		PreviousUnit: previousUnit,
		Date:         parts[4],
		DailyReturn:  dailyReturn,
	}, nil
}

func (f *FundApi) saveFundEstimateSnapshot(code string, name string, estimateTime string, estimatedUnit *float64, estimatedRate *float64) {
	f.saveFundEstimateSnapshotWithSource(code, name, estimateTime, estimatedUnit, estimatedRate, "eastmoney")
}

func (f *FundApi) saveFundEstimateSnapshotWithSource(code string, name string, estimateTime string, estimatedUnit *float64, estimatedRate *float64, source string) {
	code = strings.TrimSpace(code)
	estimateTime = strings.TrimSpace(estimateTime)
	if code == "" || estimateTime == "" || estimatedUnit == nil {
		return
	}

	snapshot := FundEstimateSnapshot{
		Code:          code,
		Name:          strings.TrimSpace(name),
		TradeDate:     parseFundEstimateTradeDate(estimateTime),
		EstimateTime:  estimateTime,
		EstimatedUnit: *estimatedUnit,
		EstimatedRate: estimatedRate,
		Source:        strings.TrimSpace(source),
	}
	if snapshot.Source == "" {
		snapshot.Source = "eastmoney"
	}

	var existing FundEstimateSnapshot
	err := db.Dao.Where("code = ? AND estimate_time = ? AND source = ?", snapshot.Code, snapshot.EstimateTime, snapshot.Source).First(&existing).Error
	switch err {
	case nil:
		db.Dao.Model(&existing).Updates(snapshot)
	case gorm.ErrRecordNotFound:
		db.Dao.Create(&snapshot)
	default:
		logger.SugaredLogger.Warnf("save fund estimate snapshot failed for %s: %v", snapshot.Code, err)
		return
	}

	expireBefore := time.Now().AddDate(0, 0, -14).Format("2006-01-02")
	db.Dao.Where("trade_date <> '' AND trade_date < ?", expireBefore).Delete(&FundEstimateSnapshot{})
}

func parseFundEstimateTradeDate(value string) string {
	value = strings.TrimSpace(value)
	if len(value) >= 10 {
		return value[:10]
	}
	return ""
}

func parseFundEstimateTimestamp(value string) int64 {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}

	location := time.Now().Location()
	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		time.RFC3339,
	}
	for _, layout := range layouts {
		if parsed, err := time.ParseInLocation(layout, value, location); err == nil {
			return parsed.UnixMilli()
		}
	}
	return 0
}
