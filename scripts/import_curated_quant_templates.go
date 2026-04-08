//go:build tools

package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"investment-platform/internal/quant"

	"go-stock/backend/db"
)

type templateSeed struct {
	Name            string
	CategorySort    int
	Description     string
	Code            string
	BrokerPlatform  string
	StrategyType    string
	ScriptCategory  string
	StyleTags       string
	EmotionTags     string
	VolumeTags      string
	ScenarioTags    string
	CapitalTags     string
	FactorTags      string
	SearchKeywords  string
	SourcePlatforms string
	LinkedStocks    string
	Parameters      string
	Status          string
}

func main() {
	dbPath := strings.TrimSpace(os.Getenv("INVESTMENT_DB_PATH"))
	db.Init(dbPath)
	service := quant.NewService("data/quant_templates")
	service.InitDefaultCategories()

	categoryBySort := map[int]uint{}
	for _, item := range service.GetCategories() {
		categoryBySort[item.SortOrder] = item.ID
	}

	created := 0
	updated := 0
	for _, item := range curatedTemplates() {
		template := quant.Template{
			Name:            item.Name,
			CategoryID:      categoryBySort[item.CategorySort],
			Description:     item.Description,
			Language:        "python",
			Code:            strings.ReplaceAll(item.Code, "\"\"", "\""),
			BrokerPlatform:  item.BrokerPlatform,
			StrategyType:    item.StrategyType,
			ScriptCategory:  item.ScriptCategory,
			StyleTags:       item.StyleTags,
			EmotionTags:     item.EmotionTags,
			VolumeTags:      item.VolumeTags,
			ScenarioTags:    item.ScenarioTags,
			CapitalTags:     item.CapitalTags,
			FactorTags:      item.FactorTags,
			SearchKeywords:  item.SearchKeywords,
			SourcePlatforms: item.SourcePlatforms,
			LinkedStocks:    item.LinkedStocks,
			Parameters:      item.Parameters,
			Status:          defaultString(item.Status, "active"),
		}

		var current quant.Template
		if err := db.Dao.Where("name = ?", item.Name).First(&current).Error; err == nil && current.ID > 0 {
			template.ID = current.ID
			template.CreatedAt = current.CreatedAt
			template.UpdatedAt = time.Now()
			template.Version = current.Version
			template.LastUsedAt = current.LastUsedAt
			if service.UpdateTemplate(template) != nil {
				updated++
			}
			continue
		}

		if service.CreateTemplate(template) != nil {
			created++
		}
	}

	fmt.Printf("curated quant templates imported: created=%d updated=%d total=%d\n", created, updated, len(curatedTemplates()))
}

func defaultString(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func curatedTemplates() []templateSeed {
	return []templateSeed{
		{
			Name:            "宽基ETF多篮子轮动网格",
			CategorySort:    3,
			Description:     "参考聚宽 ETF 策略池、BigQuant ETF 轮动优化和国内宽基 ETF 常用交易框架整理而成。使用 8 个宽基/风格 ETF 组成轮动池，并在入选 ETF 上叠加 3% 网格分层执行。参考来源: https://www.joinquant.com/view/community/detail/90dc41ae172688f31e462ed3ede021ab?type=1 ; https://bigquant.com/codesharev3/68d0a1e5-ffc8-4c74-ab4b-add1104d7416",
			Code:            etfRotationGridCode([]string{"510300", "510050", "510500", "159915", "588000", "510880", "159949", "159845"}),
			BrokerPlatform:  "聚宽/BigQuant/通用",
			StrategyType:    "ETF轮动",
			ScriptCategory:  "策略主脚本",
			StyleTags:       "grid-trading,etf-rotation,market-timing",
			EmotionTags:     "normal-emotion,low-emotion,high-emotion",
			VolumeTags:      "incremental-volume,rising-volume,persistent-volume",
			ScenarioTags:    "sideways-market,fast-rotation,market-rally",
			CapitalTags:     "capital-50k-100k,capital-100k-300k,capital-above-300k",
			FactorTags:      "momentum,volatility,industry-heat,capital-flow",
			SearchKeywords:  "宽基 ETF 轮动 网格 510300 510050 510500 159915 588000 510880 159949 159845",
			SourcePlatforms: "joinquant,bigquant,github",
			LinkedStocks:    "510300,510050,510500,159915,588000,510880,159949,159845",
			Parameters:      "top_n=3; rebalance_days=5; grid_step=0.03; stop_loss=-0.06",
			Status:          "active",
		},
		etfGridSeed("沪深300ETF保守网格", "以沪深300ETF为核心，结合 10 日均线偏离做 2.5% 分层补仓和减仓，适合震荡偏多环境下的低频网格。参考来源: https://www.joinquant.com/view/community/detail/90dc41ae172688f31e462ed3ede021ab?type=1 ; https://bigquant.com/codesharev3/68d0a1e5-ffc8-4c74-ab4b-add1104d7416 ; https://github.com/yaaks7/grid-trading", "510300", "围绕沪深300ETF单品种做低频保守网格，优先控制回撤而不是追求高换手。", "joinquant,bigquant,github", "top_n=1; rebalance_days=5; grid_step=0.025; stop_loss=-0.05"),
		etfGridSeed("上证50ETF低波网格", "以上证50ETF为核心，侧重大盘蓝筹和低波动区间操作，适合资金偏稳健的长期底仓管理。参考来源: https://www.joinquant.com/view/community/detail/90dc41ae172688f31e462ed3ede021ab?type=1 ; https://github.com/yaaks7/grid-trading", "510050", "大盘价值风格网格，强调低频再平衡和小步长补仓。", "joinquant,github", "top_n=1; rebalance_days=5; grid_step=0.02; stop_loss=-0.045"),
		etfGridSeed("中证500ETF震荡网格", "用中证500ETF做中盘风格震荡网格，保留一定弹性但仍以区间交易为主。参考来源: https://bigquant.com/codesharev3/68d0a1e5-ffc8-4c74-ab4b-add1104d7416 ; https://github.com/yaaks7/grid-trading", "510500", "中盘ETF在震荡市里的偏保守网格版本，适合中等风险承受能力。", "bigquant,github", "top_n=1; rebalance_days=4; grid_step=0.03; stop_loss=-0.055"),
		etfGridSeed("红利ETF防御网格", "围绕红利ETF构建防御型网格，用分红和低波属性提升持有舒适度。参考来源: https://www.joinquant.com/view/community/detail/90dc41ae172688f31e462ed3ede021ab?type=1 ; https://bigquant.com/wiki/doc/ySSnil1iwd", "510880", "红利ETF低频网格，适合把高股息风格作为组合防守底仓。", "joinquant,bigquant", "top_n=1; rebalance_days=6; grid_step=0.02; stop_loss=-0.04"),
		etfGridSeed("双宽基ETF对称网格", "在沪深300ETF和上证50ETF之间做对称网格分配，适合只想管理核心宽基仓位的保守型账户。参考来源: https://www.joinquant.com/view/community/detail/90dc41ae172688f31e462ed3ede021ab?type=1 ; https://github.com/yaaks7/grid-trading", "510300,510050", "双宽基组合网格，强调核心资产之间的仓位平衡。", "joinquant,github", "top_n=2; rebalance_days=5; grid_step=0.022; stop_loss=-0.045"),
		etfGridSeed("宽基三篮子ETF均衡网格", "用沪深300、上证50和红利ETF组成三篮子保守网格，兼顾 Beta 和防御属性。参考来源: https://bigquant.com/wiki/doc/ySSnil1iwd ; https://github.com/yaaks7/grid-trading", "510300,510050,510880", "三篮子宽基+红利组合，适合账户的核心仓稳健运作。", "bigquant,github", "top_n=2; rebalance_days=5; grid_step=0.025; stop_loss=-0.05"),
		etfGridSeed("股债平衡ETF缓冲网格", "将沪深300、中证500和国债ETF放在同一低频网格池里，用债券仓位缓冲权益波动。参考来源: https://bigquant.com/wiki/doc/ySSnil1iwd ; https://github.com/yaaks7/grid-trading", "510300,510500,511010", "权益与国债ETF混合网格，适合回撤敏感型账户。", "bigquant,github", "top_n=2; rebalance_days=7; grid_step=0.025; stop_loss=-0.04"),
		etfGridSeed("红利低波+国债ETF低回撤网格", "将红利ETF、上证50ETF和国债ETF组合成偏低回撤网格，优先守住净值曲线稳定性。参考来源: https://bigquant.com/wiki/doc/ySSnil1iwd ; https://github.com/yaaks7/grid-trading", "510880,511010,510050", "红利与国债为主的保守型网格，适合存量资金管理。", "bigquant,github", "top_n=2; rebalance_days=7; grid_step=0.02; stop_loss=-0.035"),
		etfGridSeed("核心资产ETF回撤补仓网格", "围绕沪深300、上证50、红利ETF和国债ETF做回撤补仓网格，在下跌中逐级加仓、反弹中逐级收回。参考来源: https://www.joinquant.com/view/community/detail/90dc41ae172688f31e462ed3ede021ab?type=1 ; https://github.com/yaaks7/grid-trading", "510300,510050,510880,511010", "四资产低回撤网格，适合核心仓位的慢节奏管理。", "joinquant,github", "top_n=2; rebalance_days=6; grid_step=0.02; stop_loss=-0.04"),
		etfGridSeed("波动率自适应ETF网格", "以沪深300、中证500和红利ETF构成低频网格池，根据近期波动度动态控制加减仓触发。参考来源: https://bigquant.com/codesharev3/68d0a1e5-ffc8-4c74-ab4b-add1104d7416 ; https://github.com/yaaks7/grid-trading", "510300,510500,510880", "用波动率过滤信号的保守ETF网格版本。", "bigquant,github", "top_n=2; rebalance_days=5; grid_step=0.028; stop_loss=-0.05"),
		etfGridSeed("月度再平衡ETF稳健网格", "把月度再平衡思想与 ETF 网格结合，减少交易噪音，适合长期持有和定期维护。参考来源: https://bigquant.com/wiki/doc/ySSnil1iwd ; https://github.com/yaaks7/grid-trading", "510300,510500,510050,510880", "月频调仓的稳健型ETF网格，降低过度交易。", "bigquant,github", "top_n=2; rebalance_days=20; grid_step=0.03; stop_loss=-0.055"),
		etfGridSeed("现金管理增强ETF网格", "在沪深300、红利ETF和国债ETF之间做低频网格，把现金管理和权益增强结合起来。参考来源: https://bigquant.com/wiki/doc/ySSnil1iwd ; https://github.com/yaaks7/grid-trading", "510300,510880,511010", "更适合稳健自用账户的现金管理增强型ETF网格。", "bigquant,github", "top_n=2; rebalance_days=7; grid_step=0.02; stop_loss=-0.035"),
		etfSeed("核心资产ETF轮动", "参考聚宽社区“核心资产轮动”的思路，扩展成可运行的本地 Python 版。参考来源: https://www.joinquant.com/view/community/detail/d796ab61ca772cfd88d58786f8f9724f?type=1", "510300,159915,588000,510880,159949,512100", "以核心资产与风格 ETF 为池，按 20/60 日双动量和近 20 日波动率筛选前 2 名持有。", "joinquant"),
		etfSeed("二八ETF轮动策略", "参考 BigQuant 二八 ETF 轮动思路整理成标准模板。参考来源: https://bigquant.com/codesharev2/813ee05f-a7eb-45ad-aa46-51f634e19554", "510300,510050,159915,159949,510880", "二八轮动框架，风险偏好抬升时偏向成长和弹性 ETF，风险偏好回落时偏向大盘和红利。", "bigquant"),
		etfSeed("绝对动量ETF轮动", "参考 GitHub 上常见的 Dual Momentum / ETF Rotation 实现，结合 A 股宽基 ETF 池做成本地可运行版本。参考来源: https://github.com/mfrdixon/ETF-Portfolio-Management ; https://github.com/mhallsmoore/qstrader", "510300,510500,159915,588000,510880,159845", "绝对动量先过滤负收益资产，再在剩余 ETF 中持有前 2 名。", "github"),
		{
			Name:            "红利低波ETF防守轮动",
			CategorySort:    5,
			Description:     "围绕红利、低波和大盘风格做防守切换，适合风险偏好回落时段。参考来源: https://www.joinquant.com/view/community/detail/90dc41ae172688f31e462ed3ede021ab?type=1",
			Code:            defensiveEtfCode(),
			BrokerPlatform:  "聚宽/通用",
			StrategyType:    "ETF轮动",
			ScriptCategory:  "策略主脚本",
			StyleTags:       "etf-rotation,mean-reversion,market-timing",
			EmotionTags:     "low-emotion,panic,normal-emotion",
			VolumeTags:      "low-volume,incremental-volume",
			ScenarioTags:    "market-sharp-drop,sideways-market,high-divergence",
			CapitalTags:     "capital-30k-50k,capital-50k-100k,capital-100k-300k",
			FactorTags:      "volatility,quality,valuation",
			SearchKeywords:  "红利 低波 ETF 防守 510880 510300 大盘价值",
			SourcePlatforms: "joinquant",
			LinkedStocks:    "510880,510300,510050,510500",
			Parameters:      "max_drawdown=-0.05; rebalance_days=10",
			Status:          "active",
		},
		singleAssetSeed("沪深300价值增强择时", "沪深300 / 上证50 一侧为主仓，通过均线和成交波动确认切换时机。", "510300", "沪深300ETF", "market-timing,trend-following", "normal-emotion,low-emotion", "incremental-volume,rising-volume", "market-rally,high-divergence,market-sharp-drop", "momentum,volatility,capital-flow"),
		singleAssetSeed("中证500趋势跟随", "适合中盘风格抬升阶段的 ETF 趋势脚本，以中证500 ETF 为交易核心。", "510500", "中证500ETF", "trend-following,etf-rotation", "normal-emotion,high-emotion", "rising-volume,persistent-volume", "market-rally,bottom-rebound,fast-rotation", "momentum,volume-ratio,capital-flow"),
		singleAssetSeed("中证1000高弹性动量", "以中证1000 和小盘弹性风格为核心，适合风险偏好抬升且量能放大的环境。", "159845", "中证1000ETF", "trend-following,momentum-breakout,etf-rotation", "high-emotion,euphoria", "surge-volume,persistent-volume", "market-rally,fast-rotation,bottom-rebound", "momentum,volume-ratio,sentiment-strength"),
		{
			Name:            "创业板科创成长轮动",
			CategorySort:    5,
			Description:     "在创业板 ETF、科创 50 ETF 之间择强切换，适合科技成长主题共振阶段。",
			Code:            growthRotationCode(),
			BrokerPlatform:  "通用",
			StrategyType:    "成长轮动",
			ScriptCategory:  "策略主脚本",
			StyleTags:       "etf-rotation,trend-following,sector-rotation",
			EmotionTags:     "high-emotion,normal-emotion",
			VolumeTags:      "surge-volume,rising-volume",
			ScenarioTags:    "market-rally,fast-rotation,bottom-rebound",
			CapitalTags:     "capital-50k-100k,capital-100k-300k",
			FactorTags:      "momentum,industry-heat,theme-resonance",
			SearchKeywords:  "创业板 科创50 成长 科技 ETF 159915 588000",
			SourcePlatforms: "joinquant,github",
			LinkedStocks:    "159915,588000,159949",
			Parameters:      "top_n=1; rebalance_days=5; lookback=20",
			Status:          "active",
		},
		{
			Name:            "热门行业ETF轮动",
			CategorySort:    5,
			Description:     "综合行业热度、资金流和主题共振，在行业 ETF 池中做择强切换。适合联动推荐里捕捉热门行业变化。",
			Code:            industryRotationCode(),
			BrokerPlatform:  "通用",
			StrategyType:    "行业轮动",
			ScriptCategory:  "策略主脚本",
			StyleTags:       "sector-rotation,etf-rotation,market-timing",
			EmotionTags:     "normal-emotion,high-emotion",
			VolumeTags:      "rising-volume,persistent-volume,surge-volume",
			ScenarioTags:    "fast-rotation,sector-crowding,market-rally",
			CapitalTags:     "capital-50k-100k,capital-100k-300k,capital-above-300k",
			FactorTags:      "industry-heat,capital-flow,theme-resonance,momentum",
			SearchKeywords:  "行业 ETF 轮动 半导体 证券 消费 医药",
			SourcePlatforms: "joinquant,bigquant,github",
			LinkedStocks:    "512480,512010,512880,512690,515170",
			Parameters:      "top_n=2; lookback=15; rebalance_days=3",
			Status:          "active",
		},
		standardIndicatorSeed("双均线趋势跟随", "简单均线交叉是最标准的趋势入门策略之一，参考 GitHub 上大量公开实现整理成本地模板。参考来源: https://github.com/nikhil-adithyan/Algorithmic-Trading-with-Python", indicatorCode("ma_cross", "double_ma", "latest['short_ma'] > latest['long_ma']", "latest['short_ma'] < latest['long_ma']"), "trend-following,market-timing", "normal-emotion,high-emotion", "incremental-volume,rising-volume", "market-rally,bottom-rebound", "momentum,volume-ratio", "双均线 趋势 跟随 MA crossover", "github"),
		standardIndicatorSeed("MACD趋势确认", "MACD 是公开资料中最常见的趋势确认策略之一。参考来源: https://github.com/nikhil-adithyan/Algorithmic-Trading-with-Python", indicatorCode("macd", "macd_trend", "latest['macd'] > latest['signal']", "latest['macd'] < latest['signal']"), "trend-following,market-timing", "normal-emotion,high-emotion", "incremental-volume,rising-volume", "market-rally,fast-rotation", "momentum,capital-flow", "MACD 趋势 确认", "github"),
		standardIndicatorSeed("RSI均值回归", "RSI 低买高卖是经典均值回归脚本，适合震荡市。参考来源: https://github.com/nikhil-adithyan/Algorithmic-Trading-with-Python", indicatorCode("rsi", "rsi_reversion", "latest['rsi'] < 30", "latest['rsi'] > 70"), "mean-reversion", "normal-emotion,low-emotion", "low-volume,incremental-volume", "sideways-market,high-divergence", "volatility,momentum", "RSI 均值 回归", "github"),
		standardIndicatorSeed("布林带回归策略", "布林带触边回归是标准的区间交易模板。参考来源: https://github.com/nikhil-adithyan/Algorithmic-Trading-with-Python", indicatorCode("bollinger", "bollinger_reversion", "latest['close'] < latest['lower_band']", "latest['close'] > latest['upper_band']"), "mean-reversion,grid-trading", "normal-emotion,low-emotion", "low-volume,incremental-volume", "sideways-market,high-divergence", "volatility,volume-ratio", "布林带 回归 震荡", "github"),
		standardIndicatorSeed("唐奇安海龟突破", "海龟交易是标准突破策略，这里给出适合 A 股/ETF 的简化版本。参考来源: https://github.com/pplonski/turtle-trading-python", turtleCode(), "trend-following,momentum-breakout", "high-emotion,normal-emotion", "surge-volume,persistent-volume", "market-rally,bottom-rebound", "momentum,volatility", "海龟 唐奇安 突破 turtle trading", "github"),
		standardIndicatorSeed("SuperTrend趋势跟随", "SuperTrend 常被用作更平滑的趋势过滤器。参考来源: https://github.com/nikhil-adithyan/Algorithmic-Trading-with-Python", superTrendCode(), "trend-following,market-timing", "normal-emotion,high-emotion", "rising-volume,persistent-volume", "market-rally,bottom-rebound", "momentum,volatility", "SuperTrend 趋势 跟随", "github"),
		standardIndicatorSeed("CCI反转趋势策略", "CCI 结合趋势确认，适合做反转切入。参考来源: https://github.com/nikhil-adithyan/Algorithmic-Trading-with-Python", indicatorCode("cci", "cci_reversal", "latest['cci'] < -100", "latest['cci'] > 100"), "mean-reversion,trend-following", "normal-emotion,low-emotion", "incremental-volume", "sideways-market,bottom-rebound", "momentum,turnover", "CCI 反转 趋势", "github"),
		standardIndicatorSeed("随机指标KDJ反转", "随机指标策略适合短周期超买超卖判断。参考来源: https://github.com/nikhil-adithyan/Algorithmic-Trading-with-Python", indicatorCode("stoch", "stoch_reversal", "latest['k_value'] < 20", "latest['k_value'] > 80"), "mean-reversion,market-timing", "normal-emotion,low-emotion", "low-volume,incremental-volume", "sideways-market,high-divergence", "momentum,turnover", "KDJ 随机指标 反转", "github"),
		{
			Name:            "配对交易价差回归",
			CategorySort:    1,
			Description:     "参考公开 Pairs Trading 项目整理的价差回归模板，适合做同风格资产相对价值交易。参考来源: https://github.com/AJeanis/Pairs-Trading",
			Code:            pairsCode(),
			BrokerPlatform:  "GitHub/通用",
			StrategyType:    "套利",
			ScriptCategory:  "策略主脚本",
			StyleTags:       "arbitrage,mean-reversion",
			EmotionTags:     "normal-emotion,low-emotion",
			VolumeTags:      "incremental-volume",
			ScenarioTags:    "sideways-market,high-divergence",
			CapitalTags:     "capital-50k-100k,capital-100k-300k,capital-above-300k",
			FactorTags:      "valuation,volatility,quality",
			SearchKeywords:  "配对交易 pairs trading 价差 回归",
			SourcePlatforms: "github",
			LinkedStocks:    "pair_a,pair_b",
			Parameters:      "z_entry=2.0; z_exit=0.5; lookback=30",
			Status:          "active",
		},
		{
			Name:            "多因子质量动量选股",
			CategorySort:    2,
			Description:     "将质量、估值、动量三个常见因子标准化后打分选股，适合做程序库里的基础多因子模板。参考来源: https://github.com/stefan-jansen/machine-learning-for-trading ; https://github.com/nikhil-adithyan/Algorithmic-Trading-with-Python",
			Code:            multiFactorCode(),
			BrokerPlatform:  "GitHub/通用",
			StrategyType:    "多因子",
			ScriptCategory:  "因子计算",
			StyleTags:       "multi-factor,market-timing",
			EmotionTags:     "normal-emotion,low-emotion,high-emotion",
			VolumeTags:      "incremental-volume,rising-volume",
			ScenarioTags:    "market-rally,sideways-market,bottom-rebound",
			CapitalTags:     "capital-50k-100k,capital-100k-300k,capital-above-300k",
			FactorTags:      "momentum,quality,valuation,earnings-revision",
			SearchKeywords:  "多因子 质量 动量 估值 选股",
			SourcePlatforms: "github",
			LinkedStocks:    "000001.XSHG,000905.XSHG",
			Parameters:      "top_n=10; rebalance_days=20",
			Status:          "active",
		},
	}
}

func etfSeed(name string, description string, linked string, intro string, source string) templateSeed {
	return templateSeed{
		Name:            name,
		CategorySort:    5,
		Description:     description,
		Code:            etfMomentumCode(strings.Split(linked, ","), intro),
		BrokerPlatform:  "聚宽/BigQuant/通用",
		StrategyType:    "ETF轮动",
		ScriptCategory:  "策略主脚本",
		StyleTags:       "etf-rotation,market-timing,trend-following",
		EmotionTags:     "normal-emotion,high-emotion,low-emotion",
		VolumeTags:      "rising-volume,incremental-volume",
		ScenarioTags:    "market-rally,fast-rotation,sideways-market",
		CapitalTags:     "capital-50k-100k,capital-100k-300k,capital-above-300k",
		FactorTags:      "momentum,volatility,capital-flow",
		SearchKeywords:  strings.ReplaceAll(linked, ",", " "),
		SourcePlatforms: source,
		LinkedStocks:    linked,
		Parameters:      "top_n=2; rebalance_days=5; momentum_windows=20,60",
		Status:          "active",
	}
}

func etfGridSeed(name string, description string, linked string, intro string, source string, parameters string) templateSeed {
	_ = intro
	return templateSeed{
		Name:            name,
		CategorySort:    3,
		Description:     description,
		Code:            etfRotationGridCode(strings.Split(linked, ",")),
		BrokerPlatform:  "聚宽/BigQuant/GitHub/通用",
		StrategyType:    "ETF网格",
		ScriptCategory:  "策略主脚本",
		StyleTags:       "grid-trading,etf-rotation,mean-reversion",
		EmotionTags:     "low-emotion,normal-emotion",
		VolumeTags:      "low-volume,incremental-volume",
		ScenarioTags:    "sideways-market,high-divergence,market-sharp-drop",
		CapitalTags:     "capital-20k-30k,capital-30k-50k,capital-50k-100k,capital-100k-300k",
		FactorTags:      "volatility,mean-reversion,capital-flow",
		SearchKeywords:  fmt.Sprintf("%s %s ETF 网格 保守 低回撤", name, strings.ReplaceAll(linked, ",", " ")),
		SourcePlatforms: source,
		LinkedStocks:    linked,
		Parameters:      parameters,
		Status:          "active",
	}
}

func singleAssetSeed(name string, description string, symbol string, label string, style string, emotion string, volume string, scenario string, factor string) templateSeed {
	return templateSeed{
		Name:            name,
		CategorySort:    5,
		Description:     description,
		Code:            singleAssetTimingCode(symbol, label),
		BrokerPlatform:  "通用",
		StrategyType:    "市场择时",
		ScriptCategory:  "策略主脚本",
		StyleTags:       style,
		EmotionTags:     emotion,
		VolumeTags:      volume,
		ScenarioTags:    scenario,
		CapitalTags:     "capital-30k-50k,capital-50k-100k,capital-100k-300k",
		FactorTags:      factor,
		SearchKeywords:  fmt.Sprintf("%s %s", name, symbol),
		SourcePlatforms: "github,joinquant",
		LinkedStocks:    symbol,
		Parameters:      "fast_ma=10; slow_ma=30; volatility_lookback=20",
		Status:          "active",
	}
}

func init() {
	_ = standardIndicatorSeed
}

func standardIndicatorSeed(name string, description string, code string, style string, emotion string, volume string, scenario string, factor string, search string, source string) templateSeed {
	return templateSeed{
		Name:            name,
		CategorySort:    1,
		Description:     description,
		Code:            code,
		BrokerPlatform:  "GitHub/通用",
		StrategyType:    "技术指标",
		ScriptCategory:  "策略主脚本",
		StyleTags:       style,
		EmotionTags:     emotion,
		VolumeTags:      volume,
		ScenarioTags:    scenario,
		CapitalTags:     "capital-under-10k,capital-10k-20k,capital-20k-30k,capital-30k-50k",
		FactorTags:      factor,
		SearchKeywords:  search,
		SourcePlatforms: source,
		LinkedStocks:    "sample_asset",
		Parameters:      "lookback=20; risk_per_trade=0.02",
		Status:          "active",
	}
}

func etfRotationGridCode(universe []string) string {
	return fmt.Sprintf(`#!/usr/bin/env python
# -*- coding: utf-8 -*-
from dataclasses import dataclass
from typing import List, Dict
import pandas as pd
import numpy as np

ETF_UNIVERSE = %q

@dataclass
class Config:
    top_n: int = 3
    rebalance_days: int = 5
    grid_step: float = 0.03
    max_position_per_etf: float = 0.35

def build_sample_data(symbols: List[str], periods: int = 120) -> Dict[str, pd.DataFrame]:
    rng = np.random.default_rng(42)
    data = {}
    for idx, symbol in enumerate(symbols):
        base = 1.0 + idx * 0.05
        drift = np.linspace(0, 0.18, periods)
        noise = rng.normal(0, 0.015, periods).cumsum()
        close = np.maximum(base + drift + noise, 0.6)
        data[symbol] = pd.DataFrame({"close": close})
    return data

def calc_score(frame: pd.DataFrame) -> float:
    close = frame["close"]
    mom_20 = close.iloc[-1] / close.iloc[-20] - 1
    mom_60 = close.iloc[-1] / close.iloc[-60] - 1
    volatility = close.pct_change().tail(20).std()
    return mom_20 * 0.5 + mom_60 * 0.5 - volatility * 1.2

def build_positions(data: Dict[str, pd.DataFrame], cfg: Config) -> pd.DataFrame:
    rows = []
    for symbol, frame in data.items():
        score = calc_score(frame)
        close = frame["close"].iloc[-1]
        ma_10 = frame["close"].tail(10).mean()
        grid_bias = (close - ma_10) / ma_10
        rows.append({"symbol": symbol, "score": score, "close": close, "grid_bias": grid_bias})
    result = pd.DataFrame(rows).sort_values("score", ascending=False).reset_index(drop=True)
    result["selected"] = result.index < cfg.top_n
    result["target_weight"] = np.where(result["selected"], 1 / cfg.top_n, 0.0).clip(upper=cfg.max_position_per_etf)
    result["grid_action"] = np.where(result["grid_bias"] <= -cfg.grid_step, "add_grid", np.where(result["grid_bias"] >= cfg.grid_step, "reduce_grid", "hold"))
    return result

if __name__ == "__main__":
    print(build_positions(build_sample_data(ETF_UNIVERSE), Config()).to_string(index=False))
`, universe)
}

func etfMomentumCode(universe []string, intro string) string {
	return fmt.Sprintf(`#!/usr/bin/env python
# -*- coding: utf-8 -*-
import pandas as pd
import numpy as np

UNIVERSE = %q

def mock_prices(symbols, periods=100):
    rng = np.random.default_rng(7)
    rows = []
    for idx, symbol in enumerate(symbols):
        drift = np.linspace(0, 0.15-idx*0.01, periods)
        noise = rng.normal(0, 0.02, periods).cumsum()
        rows.append(pd.DataFrame({"symbol": symbol, "close": np.maximum(1 + drift + noise, 0.5)}))
    return pd.concat(rows, ignore_index=True)

def score(df):
    scores = []
    for symbol, frame in df.groupby("symbol"):
        frame = frame.reset_index(drop=True)
        mom_20 = frame["close"].iloc[-1] / frame["close"].iloc[-20] - 1
        mom_60 = frame["close"].iloc[-1] / frame["close"].iloc[-60] - 1
        vol = frame["close"].pct_change().tail(20).std()
        scores.append({"symbol": symbol, "score": mom_20 * 0.6 + mom_60 * 0.4 - vol})
    result = pd.DataFrame(scores).sort_values("score", ascending=False).reset_index(drop=True)
    result["target_weight"] = np.where(result.index < 2, 0.5, 0.0)
    return result

if __name__ == "__main__":
    print(%q)
    print(score(mock_prices(UNIVERSE)).to_string(index=False))
`, universe, intro)
}

func defensiveEtfCode() string {
	return `#!/usr/bin/env python
# -*- coding: utf-8 -*-
import pandas as pd
import numpy as np

SYMBOLS = ["510880", "510300", "510050", "510500"]

def mock_nav(symbols, periods=90):
    rng = np.random.default_rng(19)
    values = {}
    for idx, symbol in enumerate(symbols):
        close = 1 + np.linspace(0, 0.08-idx*0.01, periods) + rng.normal(0, 0.01, periods).cumsum()
        values[symbol] = np.maximum(0.7, close)
    return pd.DataFrame(values)

def choose_target(frame):
    returns_20 = frame.iloc[-1] / frame.iloc[-20] - 1
    drawdown = frame / frame.cummax() - 1
    score = returns_20 + drawdown.tail(20).mean().abs() * -0.5
    result = pd.DataFrame({"symbol": score.index, "score": score.values}).sort_values("score", ascending=False)
    result["target_weight"] = [0.7, 0.3, 0.0, 0.0]
    return result

if __name__ == "__main__":
    print(choose_target(mock_nav(SYMBOLS)).reset_index(drop=True).to_string(index=False))
`
}

func singleAssetTimingCode(symbol string, label string) string {
	return fmt.Sprintf(`#!/usr/bin/env python
# -*- coding: utf-8 -*-
import pandas as pd
import numpy as np

def sample_close(periods=100):
    rng = np.random.default_rng(21)
    close = 1 + np.linspace(0, 0.16, periods) + rng.normal(0, 0.015, periods).cumsum()
    return pd.Series(np.maximum(close, 0.6), name="close")

def evaluate(close):
    fast_ma = close.rolling(10).mean()
    slow_ma = close.rolling(30).mean()
    volatility = close.pct_change().rolling(20).std().iloc[-1]
    signal = "hold" if fast_ma.iloc[-1] > slow_ma.iloc[-1] else "watch"
    return {"symbol": %q, "name": %q, "signal": signal, "fast_ma": float(fast_ma.iloc[-1]), "slow_ma": float(slow_ma.iloc[-1]), "volatility": float(volatility)}

if __name__ == "__main__":
    print(evaluate(sample_close()))
`, symbol, label)
}

func growthRotationCode() string {
	return `#!/usr/bin/env python
# -*- coding: utf-8 -*-
import pandas as pd
import numpy as np

UNIVERSE = ["159915", "588000", "159949"]

def mock_prices():
    rng = np.random.default_rng(23)
    data = {}
    for idx, symbol in enumerate(UNIVERSE):
        close = 1 + np.linspace(0, 0.18-idx*0.02, 120) + rng.normal(0, 0.018, 120).cumsum()
        data[symbol] = np.maximum(close, 0.5)
    return pd.DataFrame(data)

def select(df):
    rows = []
    for column in df.columns:
        rows.append({"symbol": column, "score": df[column].iloc[-1] / df[column].iloc[-20] - 1})
    result = pd.DataFrame(rows).sort_values("score", ascending=False).reset_index(drop=True)
    result["target_weight"] = [1.0, 0.0, 0.0]
    return result

if __name__ == "__main__":
    print(select(mock_prices()).to_string(index=False))
`
}

func industryRotationCode() string {
	return `#!/usr/bin/env python
# -*- coding: utf-8 -*-
import pandas as pd

SECTORS = [
    {"symbol": "512480", "name": "SemiconductorETF", "heat": 9.2, "flow": 8.8, "momentum": 7.4},
    {"symbol": "512010", "name": "HealthcareETF", "heat": 6.0, "flow": 5.6, "momentum": 4.7},
    {"symbol": "512880", "name": "BrokerETF", "heat": 7.8, "flow": 7.5, "momentum": 6.9},
    {"symbol": "512690", "name": "LiquorETF", "heat": 5.9, "flow": 4.8, "momentum": 4.4},
    {"symbol": "515170", "name": "FoodETF", "heat": 6.6, "flow": 5.5, "momentum": 5.0},
]

def score(rows):
    frame = pd.DataFrame(rows)
    frame["score"] = frame["heat"] * 0.4 + frame["flow"] * 0.35 + frame["momentum"] * 0.25
    frame = frame.sort_values("score", ascending=False).reset_index(drop=True)
    frame["target_weight"] = [0.5, 0.5, 0.0, 0.0, 0.0]
    return frame

if __name__ == "__main__":
    print(score(SECTORS).to_string(index=False))
`
}

func indicatorCode(mode string, title string, buyExpr string, sellExpr string) string {
	return fmt.Sprintf(`#!/usr/bin/env python
# -*- coding: utf-8 -*-
import pandas as pd
import numpy as np

def build_sample_data(periods=160):
    rng = np.random.default_rng(31)
    close = 100 + np.linspace(0, 12, periods) + rng.normal(0, 1.8, periods).cumsum()
    frame = pd.DataFrame({"close": close})
    frame["high"] = frame["close"] * 1.01
    frame["low"] = frame["close"] * 0.99
    frame["short_ma"] = frame["close"].rolling(10).mean()
    frame["long_ma"] = frame["close"].rolling(30).mean()
    frame["ema12"] = frame["close"].ewm(span=12, adjust=False).mean()
    frame["ema26"] = frame["close"].ewm(span=26, adjust=False).mean()
    frame["macd"] = frame["ema12"] - frame["ema26"]
    frame["signal"] = frame["macd"].ewm(span=9, adjust=False).mean()
    delta = frame["close"].diff()
    gain = delta.clip(lower=0).rolling(14).mean()
    loss = -delta.clip(upper=0).rolling(14).mean()
    rs = gain / loss.replace(0, np.nan)
    frame["rsi"] = 100 - (100 / (1 + rs))
    frame["mid"] = frame["close"].rolling(20).mean()
    std = frame["close"].rolling(20).std()
    frame["upper_band"] = frame["mid"] + 2 * std
    frame["lower_band"] = frame["mid"] - 2 * std
    typical = (frame["high"] + frame["low"] + frame["close"]) / 3
    ma_typical = typical.rolling(20).mean()
    md = (typical - ma_typical).abs().rolling(20).mean()
    frame["cci"] = (typical - ma_typical) / (0.015 * md)
    low_min = frame["low"].rolling(14).min()
    high_max = frame["high"].rolling(14).max()
    frame["k_value"] = (frame["close"] - low_min) / (high_max - low_min) * 100
    return frame.dropna().reset_index(drop=True)

def generate_signal(frame):
    latest = frame.iloc[-1]
    if %s:
        return "buy"
    if %s:
        return "sell"
    return "hold"

if __name__ == "__main__":
    data = build_sample_data()
    print({"strategy": %q, "mode": %q, "signal": generate_signal(data), "close": float(data.iloc[-1]["close"])})
`, buyExpr, sellExpr, title, mode)
}

func turtleCode() string {
	return `#!/usr/bin/env python
# -*- coding: utf-8 -*-
import pandas as pd
import numpy as np

def sample(periods=180):
    rng = np.random.default_rng(41)
    close = 100 + np.linspace(0, 22, periods) + rng.normal(0, 2.0, periods).cumsum()
    frame = pd.DataFrame({"close": close})
    frame["high"] = frame["close"] * 1.01
    frame["low"] = frame["close"] * 0.99
    frame["entry_high"] = frame["high"].rolling(20).max()
    frame["exit_low"] = frame["low"].rolling(10).min()
    frame["atr"] = (frame["high"] - frame["low"]).rolling(14).mean()
    return frame.dropna().reset_index(drop=True)

if __name__ == "__main__":
    data = sample()
    latest = data.iloc[-1]
    signal = "buy_breakout" if latest["close"] >= latest["entry_high"] else "sell_exit" if latest["close"] <= latest["exit_low"] else "hold"
    print({"strategy": "donchian_turtle_breakout", "signal": signal, "atr": float(latest["atr"])})
`
}

func superTrendCode() string {
	return `#!/usr/bin/env python
# -*- coding: utf-8 -*-
import pandas as pd
import numpy as np

def sample(periods=150):
    rng = np.random.default_rng(43)
    close = 50 + np.linspace(0, 10, periods) + rng.normal(0, 1.2, periods).cumsum()
    frame = pd.DataFrame({"close": close})
    frame["high"] = frame["close"] * 1.01
    frame["low"] = frame["close"] * 0.99
    tr = (frame["high"] - frame["low"]).rolling(10).mean()
    hl2 = (frame["high"] + frame["low"]) / 2
    frame["upper"] = hl2 + 3 * tr
    frame["lower"] = hl2 - 3 * tr
    return frame.dropna().reset_index(drop=True)

if __name__ == "__main__":
    data = sample()
    latest = data.iloc[-1]
    signal = "buy" if latest["close"] > latest["upper"] else "hold" if latest["close"] > latest["lower"] else "sell"
    print({"strategy": "supertrend_following", "signal": signal, "close": float(latest["close"])})
`
}

func pairsCode() string {
	return `#!/usr/bin/env python
# -*- coding: utf-8 -*-
import pandas as pd
import numpy as np

def sample(periods=120):
    rng = np.random.default_rng(47)
    a = 100 + np.linspace(0, 6, periods) + rng.normal(0, 1.0, periods).cumsum()
    b = a * 0.92 + rng.normal(0, 1.5, periods)
    return pd.DataFrame({"pair_a": a, "pair_b": b})

def evaluate(frame):
    spread = frame["pair_a"] - frame["pair_b"]
    zscore = (spread - spread.rolling(30).mean()) / spread.rolling(30).std()
    latest = zscore.dropna().iloc[-1]
    if latest > 2:
        return "short_a_long_b"
    if latest < -2:
        return "long_a_short_b"
    return "hold"

if __name__ == "__main__":
    print({"strategy": "pairs_mean_reversion", "signal": evaluate(sample())})
`
}

func multiFactorCode() string {
	return `#!/usr/bin/env python
# -*- coding: utf-8 -*-
import pandas as pd

UNIVERSE = [
    {"symbol": "000001", "quality": 0.72, "valuation": 0.56, "momentum": 0.68},
    {"symbol": "000002", "quality": 0.65, "valuation": 0.78, "momentum": 0.51},
    {"symbol": "600000", "quality": 0.61, "valuation": 0.74, "momentum": 0.45},
    {"symbol": "600519", "quality": 0.88, "valuation": 0.34, "momentum": 0.62},
    {"symbol": "300750", "quality": 0.81, "valuation": 0.29, "momentum": 0.73},
]

def rank(universe):
    frame = pd.DataFrame(universe)
    frame["score"] = frame["quality"] * 0.4 + frame["momentum"] * 0.35 + (1 - frame["valuation"]) * 0.25
    frame = frame.sort_values("score", ascending=False).reset_index(drop=True)
    frame["selected"] = frame.index < 3
    return frame

if __name__ == "__main__":
    print(rank(UNIVERSE).to_string(index=False))
`
}
