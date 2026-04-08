package data

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/duke-git/lancet/v2/convertor"
	"github.com/duke-git/lancet/v2/mathutil"
	"github.com/duke-git/lancet/v2/strutil"
	"github.com/go-resty/resty/v2"
	"go-stock/backend/db"
	"go-stock/backend/logger"
	"html"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"gorm.io/gorm"
)

type FundApi struct {
	client *resty.Client
	config *SettingConfig
}

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
	MaxDrawdown12     *float64 `json:"maxDrawdown12"`
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

func (f *FundApi) GetFundList(key string) []FundBasic {
	var funds []FundBasic
	db.Dao.Where("code like ? or name like ?", "%"+key+"%", "%"+key+"%").Limit(10).Find(&funds)
	return funds
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
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, "", nil, nil
	}

	tradeDate := day.Format("2006-01-02")
	var snapshots []FundEstimateSnapshot
	if err := db.Dao.Where("code = ? AND trade_date = ?", code, tradeDate).
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
			f.saveFundEstimateSnapshot(fund.Code, fund.Name, fund.NetEstimatedTime, fund.NetEstimatedUnit, fund.NetEstimatedRate)
		}
	}
}

// CrawlFundNetUnitValue fetches latest confirmed net unit value data.
func (f *FundApi) CrawlFundNetUnitValue(code string) {
	//	var fundNetUnitValue FundNetUnitValue
	url := fmt.Sprintf("http://hq.sinajs.cn/rn=%d&list=f_%s", time.Now().UnixMilli(), code)
	//logger.SugaredLogger.Infof("url:%s", url)
	response, err := f.client.SetTimeout(time.Duration(f.config.CrawlTimeOut)*time.Second).R().
		SetHeader("Host", "hq.sinajs.cn").
		SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36 Edg/119.0.0.0").
		SetHeader("Referer", "https://finance.sina.com.cn").
		Get(url)
	if err != nil {
		logger.SugaredLogger.Errorf("err:%s", err.Error())
		return
	}
	if response.StatusCode() == 200 {
		data := string(GB18030ToUTF8(response.Body()))
		//logger.SugaredLogger.Infof("data:%s", data)
		datas := strutil.SplitAndTrim(data, "=", "\"")
		if len(datas) >= 2 {
			//codex := strings.Split(datas[0], "hq_str_f_")[1]
			parts := strutil.SplitAndTrim(datas[1], ",", "\"")
			//logger.SugaredLogger.Infof("parts:%s", parts)
			val, err := convertor.ToFloat(parts[1])
			if err != nil {
				logger.SugaredLogger.Errorf("err:%s", err.Error())
				return
			}
			fund := &FollowedFund{
				Name:             parts[0],
				Code:             code,
				NetUnitValue:     &val,
				NetUnitValueDate: parts[4],
			}
			db.Dao.Model(fund).Where("code=?", fund.Code).Updates(fund)
		}

	}
}

func (f *FundApi) saveFundEstimateSnapshot(code string, name string, estimateTime string, estimatedUnit *float64, estimatedRate *float64) {
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
		Source:        "eastmoney",
	}

	var existing FundEstimateSnapshot
	err := db.Dao.Where("code = ? AND estimate_time = ?", snapshot.Code, snapshot.EstimateTime).First(&existing).Error
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
