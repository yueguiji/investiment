package data

import (
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/duke-git/lancet/v2/convertor"
	"go-stock/backend/logger"
)

type FundKLineApi struct {
	fundApi *FundApi
}

func NewFundKLineApi() *FundKLineApi {
	return &FundKLineApi{fundApi: NewFundApi()}
}

var (
	reNetWorthTrend = regexp.MustCompile(`var\s+Data_netWorthTrend\s*=\s*(\[.+?\]);`)
	reACWorthTrend  = regexp.MustCompile(`var\s+Data_ACWorthTrend\s*=\s*(\[.+?\]);`)
)

const fundDataUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36"

func (f *FundKLineApi) GetFundKLine(fundCode, klt string, limit int) *KLineSourceResult {
	return f.getOffExchangeFundKLine(fundCode, klt, limit)
}

func (f *FundKLineApi) GetFundKLineWithFallback(fundCode, klt string, limit int) *KLineSourceResult {
	result := f.GetFundKLine(fundCode, klt, limit)
	if result != nil {
		return result
	}
	empty := []KLineData{}
	return &KLineSourceResult{Data: &empty, Source: ""}
}

func (f *FundKLineApi) getOffExchangeFundKLine(fundCode, klt string, limit int) *KLineSourceResult {
	emResult := f.fetchOffExchangeFromEastMoney(fundCode, klt, limit)
	if emResult != nil && emResult.Data != nil && len(*emResult.Data) > 0 {
		emResult.Source = "eastmoney_nav"
		return emResult
	}

	pzResult := f.fetchOffExchangeFromPingZhongData(fundCode, klt, limit)
	if pzResult != nil && pzResult.Data != nil && len(*pzResult.Data) > 0 {
		pzResult.Source = "pingzhongdata_nav"
		return pzResult
	}

	empty := []KLineData{}
	return &KLineSourceResult{Data: &empty, Source: ""}
}

func (f *FundKLineApi) fetchOffExchangeFromEastMoney(fundCode, klt string, limit int) *KLineSourceResult {
	effectiveKlt := normalizeFundKLinePeriod(klt)
	days := expandFundKLineDays(effectiveKlt, limit)
	if days <= 0 {
		days = 120
	}

	endDate := time.Now().Format("2006-01-02")
	startDate := time.Now().AddDate(0, 0, -days).Format("2006-01-02")
	history, err := f.fundApi.GetFundHistoryNetValue(fundCode, 1, days, startDate, endDate)
	if err != nil {
		logger.SugaredLogger.Warnf("get fund history net value failed: code=%s err=%v", fundCode, err)
		empty := []KLineData{}
		return &KLineSourceResult{Data: &empty}
	}

	klines := make([]KLineData, 0, len(history))
	for i := len(history) - 1; i >= 0; i-- {
		item := history[i]
		value := fmt.Sprintf("%.4f", item.NetValue)
		klines = append(klines, KLineData{
			Day:    item.Date,
			Open:   value,
			Close:  value,
			High:   value,
			Low:    value,
			Volume: "0",
		})
	}

	klines = normalizeAndLimitFundKLines(klines, effectiveKlt, limit)
	return &KLineSourceResult{Data: &klines}
}

func (f *FundKLineApi) fetchOffExchangeFromPingZhongData(fundCode, klt string, limit int) *KLineSourceResult {
	effectiveKlt := normalizeFundKLinePeriod(klt)
	url := fmt.Sprintf("http://fund.eastmoney.com/pingzhongdata/%s.js", fundCode)
	resp, err := f.fundApi.client.R().
		SetHeader("User-Agent", fundDataUserAgent).
		SetHeader("Referer", fmt.Sprintf("http://fund.eastmoney.com/%s.html", fundCode)).
		Get(url)
	if err != nil || resp.StatusCode() != 200 {
		empty := []KLineData{}
		return &KLineSourceResult{Data: &empty}
	}

	body := string(resp.Body())
	klines := parseFundTrendKLines(body, reNetWorthTrend)
	if len(klines) == 0 {
		klines = parseFundTrendKLines(body, reACWorthTrend)
	}

	klines = normalizeAndLimitFundKLines(klines, effectiveKlt, limit)
	return &KLineSourceResult{Data: &klines}
}

func parseFundTrendKLines(body string, re *regexp.Regexp) []KLineData {
	if !re.MatchString(body) {
		return nil
	}
	match := re.FindStringSubmatch(body)
	if len(match) <= 1 {
		return nil
	}

	var dataItems [][]interface{}
	if err := json.Unmarshal([]byte(match[1]), &dataItems); err != nil {
		return nil
	}

	klines := make([]KLineData, 0, len(dataItems))
	for _, item := range dataItems {
		if len(item) < 2 {
			continue
		}
		timestamp, ok := item[0].(float64)
		if !ok {
			continue
		}
		value, ok := item[1].(float64)
		if !ok || value <= 0 {
			continue
		}
		netValue := fmt.Sprintf("%.4f", value)
		klines = append(klines, KLineData{
			Day:    time.Unix(int64(timestamp)/1000, 0).Format("2006-01-02"),
			Open:   netValue,
			Close:  netValue,
			High:   netValue,
			Low:    netValue,
			Volume: "0",
		})
	}
	return klines
}

func normalizeFundKLinePeriod(klt string) string {
	switch klt {
	case "101", "102", "103":
		return klt
	default:
		return "101"
	}
}

func expandFundKLineDays(klt string, limit int) int {
	if limit <= 0 {
		return 120
	}
	switch klt {
	case "102":
		return limit * 5
	case "103":
		return limit * 22
	default:
		return limit
	}
}

func normalizeAndLimitFundKLines(klines []KLineData, klt string, limit int) []KLineData {
	if len(klines) == 0 {
		return klines
	}
	switch klt {
	case "102":
		klines = aggregateFundKLinesToWeek(klines)
	case "103":
		klines = aggregateFundKLinesToMonth(klines)
	}
	if limit > 0 && len(klines) > limit {
		klines = klines[len(klines)-limit:]
	}
	return klines
}

func aggregateFundKLinesToWeek(dailyKlines []KLineData) []KLineData {
	var result []KLineData
	var weekData *KLineData
	currentWeek := ""

	for _, kd := range dailyKlines {
		t, err := time.Parse("2006-01-02", kd.Day)
		if err != nil {
			continue
		}
		_, week := t.ISOWeek()
		weekKey := fmt.Sprintf("%d-W%02d", t.Year(), week)
		if weekKey != currentWeek {
			if weekData != nil {
				result = append(result, *weekData)
			}
			copied := kd
			weekData = &copied
			currentWeek = weekKey
			continue
		}
		updateAggregateFundKLine(weekData, kd)
	}
	if weekData != nil {
		result = append(result, *weekData)
	}
	return result
}

func aggregateFundKLinesToMonth(dailyKlines []KLineData) []KLineData {
	var result []KLineData
	var monthData *KLineData
	currentMonth := ""

	for _, kd := range dailyKlines {
		t, err := time.Parse("2006-01-02", kd.Day)
		if err != nil {
			continue
		}
		monthKey := fmt.Sprintf("%d-%02d", t.Year(), t.Month())
		if monthKey != currentMonth {
			if monthData != nil {
				result = append(result, *monthData)
			}
			copied := kd
			monthData = &copied
			currentMonth = monthKey
			continue
		}
		updateAggregateFundKLine(monthData, kd)
	}
	if monthData != nil {
		result = append(result, *monthData)
	}
	return result
}

func updateAggregateFundKLine(target *KLineData, next KLineData) {
	if target == nil {
		return
	}
	target.Close = next.Close
	if compareFloatStr(next.High, target.High) > 0 {
		target.High = next.High
	}
	if compareFloatStr(next.Low, target.Low) < 0 && next.Low != "0" {
		target.Low = next.Low
	}
}

func compareFloatStr(a, b string) int {
	fa, errA := convertor.ToFloat(a)
	fb, errB := convertor.ToFloat(b)
	if errA != nil || errB != nil {
		return 0
	}
	if fa > fb {
		return 1
	}
	if fa < fb {
		return -1
	}
	return 0
}

type FundHistoryNetValue struct {
	Date        string  `json:"date"`
	NetValue    float64 `json:"netValue"`
	AccumValue  float64 `json:"accumValue"`
	DailyGrowth float64 `json:"dailyGrowth"`
	BuyStatus   string  `json:"buyStatus"`
	SellStatus  string  `json:"sellStatus"`
}

func (f *FundApi) GetFundHistoryNetValue(fundCode string, pageIndex, pageSize int, startDate, endDate string) ([]FundHistoryNetValue, error) {
	if pageIndex <= 0 {
		pageIndex = 1
	}
	if pageSize <= 0 {
		pageSize = 120
	}

	url := fmt.Sprintf("http://api.fund.eastmoney.com/f10/lsjz?fundCode=%s&pageIndex=%d&pageSize=%d&startDate=%s&endDate=%s&_%d",
		fundCode, pageIndex, pageSize, startDate, endDate, time.Now().UnixMilli())
	resp, err := f.client.SetTimeout(time.Duration(f.config.CrawlTimeOut)*time.Second).R().
		SetHeader("User-Agent", fundDataUserAgent).
		SetHeader("Referer", fmt.Sprintf("http://fundf10.eastmoney.com/jjjz_%s.html", fundCode)).
		Get(url)
	if err != nil {
		return nil, fmt.Errorf("request fund history net value failed: %w", err)
	}
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("fund history net value status: %d", resp.StatusCode())
	}

	var result struct {
		Data struct {
			LSJZList []struct {
				FSRQ  string `json:"FSRQ"`
				DWJZ  string `json:"DWJZ"`
				LJJZ  string `json:"LJJZ"`
				JZZZL string `json:"JZZZL"`
				SGZT  string `json:"SGZT"`
				SHZT  string `json:"SHZT"`
			} `json:"LSJZList"`
		} `json:"Data"`
	}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("parse fund history net value failed: %w", err)
	}

	values := make([]FundHistoryNetValue, 0, len(result.Data.LSJZList))
	for _, item := range result.Data.LSJZList {
		dwjz, _ := convertor.ToFloat(item.DWJZ)
		ljjz, _ := convertor.ToFloat(item.LJJZ)
		jzzzl, _ := convertor.ToFloat(item.JZZZL)
		values = append(values, FundHistoryNetValue{
			Date:        item.FSRQ,
			NetValue:    dwjz,
			AccumValue:  ljjz,
			DailyGrowth: jzzzl,
			BuyStatus:   item.SGZT,
			SellStatus:  item.SHZT,
		})
	}
	return values, nil
}
