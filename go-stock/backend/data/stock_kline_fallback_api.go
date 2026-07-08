package data

import (
	"strings"
)

func (receiver StockDataApi) GetStockKLineWithFallback(stockCode string, kLineType string, days int64) *[]KLineData {
	result := &[]KLineData{}
	normalizedCode := normalizeStockKLineCode(stockCode)
	if normalizedCode == "" {
		return result
	}
	if kLineType == "" {
		kLineType = "day"
	}
	if days <= 0 {
		days = 365
	}

	if isHKOrUSKLineCode(normalizedCode) {
		result = receiver.GetHK_KLineData(normalizedCode, kLineType, days)
	} else {
		result = receiver.GetCommonKLineData(normalizedCode, kLineType, days)
	}
	if result != nil && len(*result) > 0 {
		return result
	}

	sinaType := kLineType
	if kLineType == "day" {
		sinaType = "240"
	}
	return receiver.GetKLineData(normalizedCode, sinaType, days)
}

func normalizeStockKLineCode(stockCode string) string {
	code := strings.TrimSpace(stockCode)
	if code == "" {
		return ""
	}
	if strings.Contains(code, ".") {
		return ConvertTushareCodeToStockCode(code)
	}
	return code
}

func isHKOrUSKLineCode(stockCode string) bool {
	code := strings.ToLower(strings.TrimSpace(stockCode))
	return strings.HasPrefix(code, "hk") || strings.HasPrefix(code, "us") || strings.HasPrefix(code, "gb_")
}
