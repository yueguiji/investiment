package data

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/duke-git/lancet/v2/convertor"
	"github.com/duke-git/lancet/v2/mathutil"
	"go-stock/backend/logger"
)

type FundHoldingStock struct {
	Rank       int      `json:"rank"`
	StockCode  string   `json:"stockCode"`
	StockName  string   `json:"stockName"`
	Ratio      float64  `json:"ratio"`
	Shares     string   `json:"shares"`
	MarketCap  string   `json:"marketCap"`
	Quarter    string   `json:"quarter"`
	Price      *float64 `json:"price"`
	ChangeRate *float64 `json:"changeRate"`
	Market     string   `json:"market"`
}

func (f *FundApi) GetFundTop10Holdings(fundCode string) ([]FundHoldingStock, error) {
	defer func() {
		if r := recover(); r != nil {
			logger.SugaredLogger.Errorf("GetFundTop10Holdings panic: %v", r)
		}
	}()

	holdings, err := f.getTop10HoldingsViaHTML(fundCode)
	if err != nil {
		return nil, fmt.Errorf("get fund top holdings failed: %w", err)
	}
	f.fillHoldingStockQuotes(holdings)
	return holdings, nil
}

func (f *FundApi) getTop10HoldingsViaHTML(fundCode string) ([]FundHoldingStock, error) {
	url := fmt.Sprintf("https://fundf10.eastmoney.com/FundArchivesDatas.aspx?type=jjcc&code=%s&topline=10&year=&month=&rt=%f",
		fundCode, float64(time.Now().UnixMilli())/1000.0)

	resp, err := f.client.SetTimeout(time.Duration(f.config.CrawlTimeOut)*time.Second).R().
		SetHeader("User-Agent", fundDataUserAgent).
		SetHeader("Accept", "*/*").
		SetHeader("Referer", fmt.Sprintf("https://fundf10.eastmoney.com/ccmx_%s.html", fundCode)).
		Get(url)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("http status: %d", resp.StatusCode())
	}

	body := string(resp.Body())
	contentMatch := regexp.MustCompile(`(?s)content:"(.*?)",\w`).FindStringSubmatch(body)
	if len(contentMatch) <= 1 || contentMatch[1] == "" {
		return nil, fmt.Errorf("holdings content not found")
	}

	htmlContent := contentMatch[1]
	quarter := ""
	if quarterMatch := regexp.MustCompile(`(\d{4}).{1,3}(\d{1,2}).{1,3}(\d{1,2}).{1,3}`).FindStringSubmatch(htmlContent); len(quarterMatch) > 0 {
		quarter = quarterMatch[0]
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("parse holdings html failed: %w", err)
	}

	holdings := parseFundHoldingRows(doc, quarter, "table.tzpgtab tbody tr, table.tzxq tbody tr, table.tzpgtab tr, table.tzxq tr")
	if len(holdings) == 0 {
		holdings = parseFundHoldingRows(doc, quarter, "table tr")
	}
	if len(holdings) > 10 {
		holdings = holdings[:10]
	}
	return holdings, nil
}

func parseFundHoldingRows(doc *goquery.Document, quarter string, selector string) []FundHoldingStock {
	var holdings []FundHoldingStock
	doc.Find(selector).Each(func(i int, s *goquery.Selection) {
		if len(holdings) >= 10 {
			return
		}
		tds := s.Find("td")
		if tds.Length() < 4 {
			return
		}

		rank := parseFundHoldingRank(tds.Eq(0).Text(), len(holdings)+1)
		stockCode := strings.TrimSpace(tds.Eq(1).Text())
		stockName := strings.TrimSpace(tds.Eq(2).Text())
		if stockCode == "" && stockName == "" {
			return
		}

		ratioIdx := 6
		if tds.Length() <= ratioIdx {
			ratioIdx = tds.Length() - 1
		}
		ratio := parsePercentNumber(tds.Eq(ratioIdx).Text())
		href, _ := tds.Eq(1).Find("a").Attr("href")

		var price *float64
		var changeRate *float64
		if tds.Length() >= 5 {
			if p, ok := parsePositiveFloatPointer(tds.Eq(3).Text()); ok {
				price = p
			}
			if c, ok := parseFloatPointer(strings.TrimSuffix(strings.TrimSpace(tds.Eq(4).Text()), "%")); ok {
				changeRate = c
			}
		}

		shares := ""
		marketCap := ""
		if tds.Length() >= 9 {
			shares = strings.TrimSpace(tds.Eq(7).Text())
			marketCap = strings.TrimSpace(tds.Eq(8).Text())
		}

		holdings = append(holdings, FundHoldingStock{
			Rank:       rank,
			StockCode:  stockCode,
			StockName:  stockName,
			Ratio:      ratio,
			Shares:     shares,
			MarketCap:  marketCap,
			Quarter:    quarter,
			Price:      price,
			ChangeRate: changeRate,
			Market:     detectHoldingStockMarket(href, stockCode),
		})
	})
	return holdings
}

func parseFundHoldingRank(text string, fallback int) int {
	rank, err := strconv.Atoi(strings.TrimSpace(text))
	if err != nil || rank <= 0 {
		return fallback
	}
	return rank
}

func parsePercentNumber(text string) float64 {
	value := strings.TrimSpace(strings.TrimSuffix(text, "%"))
	parsed, _ := strconv.ParseFloat(value, 64)
	return parsed
}

func parseFloatPointer(text string) (*float64, bool) {
	parsed, err := strconv.ParseFloat(strings.TrimSpace(text), 64)
	if err != nil {
		return nil, false
	}
	return &parsed, true
}

func parsePositiveFloatPointer(text string) (*float64, bool) {
	parsed, ok := parseFloatPointer(text)
	if !ok || parsed == nil || *parsed <= 0 {
		return nil, false
	}
	return parsed, true
}

func holdingSinaPrefix(code string) string {
	for len(code) < 6 {
		code = "0" + code
	}
	if len(code) == 0 {
		return ""
	}
	switch code[0:1] {
	case "5", "6", "9":
		return "sh" + code
	case "0", "1", "2", "3":
		return "sz" + code
	case "4", "8":
		return "bj" + code
	default:
		return ""
	}
}

func detectHoldingStockMarket(href, code string) string {
	lowerHref := strings.ToLower(href)
	switch {
	case strings.Contains(lowerHref, "/hk/") || strings.Contains(lowerHref, "hk0"):
		return "HK"
	case strings.Contains(lowerHref, "/us/") || strings.Contains(lowerHref, "us0"):
		return "US"
	case strings.Contains(lowerHref, "/concept/") || strings.Contains(lowerHref, "/sh") || strings.Contains(lowerHref, "/sz") || strings.Contains(lowerHref, "/bj"):
		return "A"
	}

	if code == "" {
		return ""
	}
	for _, c := range code {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
			return "US"
		}
	}
	if holdingSinaPrefix(code) != "" {
		return "A"
	}
	if len(code) >= 4 && len(code) <= 6 {
		allDigit := true
		for _, c := range code {
			if c < '0' || c > '9' {
				allDigit = false
				break
			}
		}
		if allDigit {
			return "HK"
		}
	}
	return ""
}

func holdingQuoteCode(code, market string) string {
	padded := code
	switch market {
	case "A":
		return holdingSinaPrefix(padded)
	case "HK":
		for len(padded) < 5 {
			padded = "0" + padded
		}
		return "hk" + padded
	case "US":
		return "gb_" + strings.ToLower(padded)
	default:
		return ""
	}
}

type holdingQuoteData struct {
	price      float64
	changeRate float64
}

func (f *FundApi) fillHoldingStockQuotes(holdings []FundHoldingStock) {
	if len(holdings) == 0 {
		return
	}

	var codes []string
	for _, h := range holdings {
		if code := holdingQuoteCode(h.StockCode, h.Market); code != "" {
			codes = append(codes, code)
		}
	}
	if len(codes) == 0 {
		return
	}

	url := fmt.Sprintf("http://hq.sinajs.cn/rn=%d&list=%s", time.Now().UnixMilli(), strings.Join(codes, ","))
	resp, err := f.client.SetTimeout(time.Duration(f.config.CrawlTimeOut)*time.Second).R().
		SetHeader("Host", "hq.sinajs.cn").
		SetHeader("User-Agent", fundDataUserAgent).
		SetHeader("Referer", "https://finance.sina.com.cn").
		Get(url)
	if err != nil || resp.StatusCode() != 200 {
		if err != nil {
			logger.SugaredLogger.Warnf("fill holding stock quotes failed: %v", err)
		}
		return
	}

	quoteMap := parseSinaHoldingQuotes(string(resp.Body()))
	for i := range holdings {
		key := normalizeHoldingQuoteKey(holdings[i].StockCode, holdings[i].Market)
		if q, ok := quoteMap[key]; ok {
			holdings[i].Price = &q.price
			holdings[i].ChangeRate = &q.changeRate
		}
	}
}

func normalizeHoldingQuoteKey(code, market string) string {
	key := code
	switch market {
	case "A":
		for len(key) < 6 {
			key = "0" + key
		}
	case "HK":
		for len(key) < 5 {
			key = "0" + key
		}
	case "US":
		key = strings.ToLower(key)
	}
	return key
}

func parseSinaHoldingQuotes(body string) map[string]holdingQuoteData {
	quotes := make(map[string]holdingQuoteData)
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		eqParts := strings.SplitN(line, "=", 2)
		if len(eqParts) < 2 {
			continue
		}
		sinaCode := strings.TrimSpace(eqParts[0])
		content := strings.Trim(eqParts[1], " \"\n\r;")
		fields := strings.Split(content, ",")
		if len(fields) < 4 || content == "" {
			continue
		}
		if strings.Contains(sinaCode, "hq_str_hk") {
			parseSinaHKHoldingQuote(sinaCode, fields, quotes)
			continue
		}
		if strings.Contains(sinaCode, "hq_str_gb_") {
			parseSinaUSHoldingQuote(sinaCode, fields, quotes)
			continue
		}
		parseSinaASHoldingQuote(sinaCode, fields, quotes)
	}
	return quotes
}

func parseSinaASHoldingQuote(sinaCode string, fields []string, quotes map[string]holdingQuoteData) {
	if len(fields) < 4 {
		return
	}
	currentPrice, err1 := convertor.ToFloat(fields[3])
	yesterdayPrice, err2 := convertor.ToFloat(fields[2])
	if err1 != nil || err2 != nil || currentPrice == 0 || yesterdayPrice == 0 {
		return
	}
	pureCode := sinaCode
	if idx := strings.LastIndex(pureCode, "_"); idx >= 0 {
		pureCode = pureCode[idx+1:]
	}
	if len(pureCode) > 2 {
		pureCode = pureCode[2:]
	}
	quotes[pureCode] = holdingQuoteData{
		price:      currentPrice,
		changeRate: mathutil.RoundToFloat((currentPrice-yesterdayPrice)/yesterdayPrice*100, 2),
	}
}

func parseSinaHKHoldingQuote(sinaCode string, fields []string, quotes map[string]holdingQuoteData) {
	if len(fields) < 8 {
		return
	}
	currentPrice, err1 := convertor.ToFloat(fields[6])
	prevClose, err2 := convertor.ToFloat(fields[3])
	if err1 != nil || err2 != nil || currentPrice == 0 || prevClose == 0 {
		return
	}
	pureCode := sinaCode
	if idx := strings.LastIndex(pureCode, "_"); idx >= 0 {
		pureCode = pureCode[idx+1:]
	}
	pureCode = strings.TrimPrefix(pureCode, "hk")
	quotes[pureCode] = holdingQuoteData{
		price:      currentPrice,
		changeRate: mathutil.RoundToFloat((currentPrice-prevClose)/prevClose*100, 2),
	}
}

func parseSinaUSHoldingQuote(sinaCode string, fields []string, quotes map[string]holdingQuoteData) {
	if len(fields) < 2 {
		return
	}
	currentPrice, err := convertor.ToFloat(fields[1])
	if err != nil || currentPrice == 0 {
		return
	}
	changeRate := 0.0
	if len(fields) > 4 {
		changeRate, _ = convertor.ToFloat(strings.TrimSuffix(fields[4], "%"))
	}
	pureCode := sinaCode
	if idx := strings.LastIndex(pureCode, "gb_"); idx >= 0 {
		pureCode = pureCode[idx+3:]
	}
	quotes[strings.ToLower(pureCode)] = holdingQuoteData{
		price:      currentPrice,
		changeRate: mathutil.RoundToFloat(changeRate, 2),
	}
}
