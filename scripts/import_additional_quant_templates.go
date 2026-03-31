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

type additionalSeed struct {
	Name            string
	Description     string
	Code            string
	SearchKeywords  string
	StyleTags       string
	EmotionTags     string
	VolumeTags      string
	ScenarioTags    string
	FactorTags      string
	SourcePlatforms string
	Parameters      string
}

func main() {
	dbPath := strings.TrimSpace(os.Getenv("INVESTMENT_DB_PATH"))
	db.Init(dbPath)
	service := quant.NewService("data/quant_templates")
	service.InitDefaultCategories()

	categoryID := uint(0)
	for _, item := range service.GetCategories() {
		if item.SortOrder == 1 {
			categoryID = item.ID
			break
		}
	}
	if categoryID == 0 {
		panic("strategy category not found")
	}

	created := 0
	updated := 0
	for _, item := range additionalTemplates() {
		template := quant.Template{
			Name:            item.Name,
			CategoryID:      categoryID,
			Description:     item.Description,
			Language:        "python",
			Code:            item.Code,
			BrokerPlatform:  "GitHub/通用",
			StrategyType:    "技术指标",
			ScriptCategory:  "策略主脚本",
			StyleTags:       item.StyleTags,
			EmotionTags:     item.EmotionTags,
			VolumeTags:      item.VolumeTags,
			ScenarioTags:    item.ScenarioTags,
			CapitalTags:     "capital-under-10k,capital-10k-20k,capital-20k-30k,capital-30k-50k,capital-50k-100k",
			FactorTags:      item.FactorTags,
			SearchKeywords:  item.SearchKeywords,
			SourcePlatforms: item.SourcePlatforms,
			LinkedStocks:    "sample_asset",
			Parameters:      item.Parameters,
			Status:          "active",
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

	fmt.Printf("additional quant templates imported: created=%d updated=%d total=%d\n", created, updated, len(additionalTemplates()))
}

func additionalTemplates() []additionalSeed {
	return []additionalSeed{
		makeIndicatorSeed(
			"ADX-RSI趋势过滤",
			"结合 ADX 趋势强度和 RSI 强弱切换，参考公开 Python 技术策略整理。参考来源: https://github.com/Nikhil-Adithyan/Algorithmic-Trading-with-Python",
			"adx_rsi_trend",
			[]string{
				"plus_dm = frame['high'].diff().clip(lower=0)",
				"minus_dm = (-frame['low'].diff()).clip(lower=0)",
				"tr_components = pd.concat([(frame['high'] - frame['low']), (frame['high'] - frame['close'].shift(1)).abs(), (frame['low'] - frame['close'].shift(1)).abs()], axis=1)",
				"tr = tr_components.max(axis=1)",
				"atr = tr.rolling(14).mean()",
				"plus_di = 100 * (plus_dm.rolling(14).mean() / atr.replace(0, np.nan))",
				"minus_di = 100 * (minus_dm.rolling(14).mean() / atr.replace(0, np.nan))",
				"dx = ((plus_di - minus_di).abs() / (plus_di + minus_di).replace(0, np.nan)) * 100",
				"frame['adx'] = dx.rolling(14).mean()",
				"delta = frame['close'].diff()",
				"gain = delta.clip(lower=0).rolling(14).mean()",
				"loss = (-delta.clip(upper=0)).rolling(14).mean()",
				"rs = gain / loss.replace(0, np.nan)",
				"frame['rsi'] = 100 - (100 / (1 + rs))",
			},
			"latest['adx'] > 25 and latest['rsi'] > 55",
			"latest['adx'] < 20 or latest['rsi'] < 45",
			"ADX RSI 趋势 过滤",
			"trend-following,market-timing",
			"normal-emotion,high-emotion",
			"rising-volume,persistent-volume",
			"market-rally,fast-rotation",
			"momentum,volatility,trend-strength",
		),
		makeIndicatorSeed(
			"ATR波动突破",
			"围绕 ATR 波动区间和近期高点突破做进场，适合趋势启动阶段。参考来源: https://github.com/slowpoke111/pyBacktest ; https://github.com/cloudQuant/backtrader",
			"atr_breakout",
			[]string{
				"tr_components = pd.concat([(frame['high'] - frame['low']), (frame['high'] - frame['close'].shift(1)).abs(), (frame['low'] - frame['close'].shift(1)).abs()], axis=1)",
				"frame['atr'] = tr_components.max(axis=1).rolling(14).mean()",
				"frame['entry_high'] = frame['high'].rolling(20).max()",
				"frame['exit_low'] = frame['low'].rolling(10).min()",
			},
			"latest['close'] > latest['entry_high'] - latest['atr'] * 0.2",
			"latest['close'] < latest['exit_low']",
			"ATR 突破 波动 通道",
			"trend-following,momentum-breakout",
			"high-emotion,normal-emotion",
			"surge-volume,persistent-volume",
			"market-rally,bottom-rebound",
			"volatility,momentum",
		),
		makeIndicatorSeed(
			"VWAP均值回归",
			"基于成交量加权均价的偏离做回归，适合分时或日内风格迁移到简化回测。参考来源: https://github.com/slowpoke111/pyBacktest ; https://github.com/nardew/talipp",
			"vwap_reversion",
			[]string{
				"typical_price = (frame['high'] + frame['low'] + frame['close']) / 3",
				"frame['vwap'] = (typical_price * frame['volume']).cumsum() / frame['volume'].cumsum()",
				"frame['vwap_bias'] = (frame['close'] - frame['vwap']) / frame['vwap']",
			},
			"latest['vwap_bias'] < -0.02",
			"latest['vwap_bias'] > 0.025",
			"VWAP 均值 回归 日内",
			"mean-reversion,market-timing",
			"low-emotion,normal-emotion",
			"incremental-volume,rising-volume",
			"sideways-market,high-divergence",
			"volume-ratio,mean-reversion",
		),
		makeIndicatorSeed(
			"OBV量价确认",
			"用 OBV 累积能量线确认价格趋势，适合筛掉无量上涨。参考来源: https://github.com/cloudQuant/backtrader",
			"obv_confirmation",
			[]string{
				"direction = np.sign(frame['close'].diff().fillna(0))",
				"frame['obv'] = (direction * frame['volume']).cumsum()",
				"frame['obv_ma'] = frame['obv'].rolling(20).mean()",
				"frame['price_ma'] = frame['close'].rolling(20).mean()",
			},
			"latest['close'] > latest['price_ma'] and latest['obv'] > latest['obv_ma']",
			"latest['close'] < latest['price_ma'] or latest['obv'] < latest['obv_ma']",
			"OBV 量价 确认 趋势",
			"trend-following,volume-confirmation",
			"normal-emotion,high-emotion",
			"rising-volume,persistent-volume",
			"market-rally,bottom-rebound",
			"volume-ratio,capital-flow,momentum",
		),
		makeIndicatorSeed(
			"一目均衡表趋势",
			"使用转折线、基准线和云层做趋势跟随，适合中段趋势确认。参考来源: https://github.com/cloudQuant/backtrader ; https://github.com/Bitvested/ta.py",
			"ichimoku_trend",
			[]string{
				"tenkan_high = frame['high'].rolling(9).max()",
				"tenkan_low = frame['low'].rolling(9).min()",
				"kijun_high = frame['high'].rolling(26).max()",
				"kijun_low = frame['low'].rolling(26).min()",
				"frame['tenkan'] = (tenkan_high + tenkan_low) / 2",
				"frame['kijun'] = (kijun_high + kijun_low) / 2",
				"frame['senkou_a'] = ((frame['tenkan'] + frame['kijun']) / 2).shift(26)",
				"span_b_high = frame['high'].rolling(52).max()",
				"span_b_low = frame['low'].rolling(52).min()",
				"frame['senkou_b'] = ((span_b_high + span_b_low) / 2).shift(26)",
			},
			"latest['close'] > max(latest['senkou_a'], latest['senkou_b']) and latest['tenkan'] > latest['kijun']",
			"latest['close'] < min(latest['senkou_a'], latest['senkou_b']) or latest['tenkan'] < latest['kijun']",
			"Ichimoku 一目均衡 趋势",
			"trend-following,market-timing",
			"normal-emotion,high-emotion",
			"rising-volume,persistent-volume",
			"market-rally,fast-rotation",
			"momentum,trend-strength",
		),
		makeIndicatorSeed(
			"ParabolicSAR反转",
			"基于简化 SAR 抛物转向点做趋势反转切换。参考来源: https://github.com/cloudQuant/backtrader",
			"parabolic_sar_reversal",
			[]string{
				"frame['sar'] = frame['close'].shift(1).ewm(alpha=0.08, adjust=False).mean()",
				"frame['sar_trend'] = frame['close'] - frame['sar']",
			},
			"latest['sar_trend'] > 0 and prev['sar_trend'] <= 0",
			"latest['sar_trend'] < 0 and prev['sar_trend'] >= 0",
			"Parabolic SAR 反转",
			"trend-following,reversal",
			"normal-emotion,high-emotion",
			"incremental-volume,rising-volume",
			"bottom-rebound,fast-rotation",
			"trend-strength,volatility",
		),
		makeIndicatorSeed(
			"Aroon趋势择时",
			"Aroon Up/Down 用于判定新高新低主导权，适合风格切换。参考来源: https://github.com/cloudQuant/backtrader",
			"aroon_timing",
			[]string{
				"period = 25",
				"frame['days_since_high'] = frame['high'].rolling(period).apply(lambda x: period - 1 - np.argmax(x), raw=True)",
				"frame['days_since_low'] = frame['low'].rolling(period).apply(lambda x: period - 1 - np.argmin(x), raw=True)",
				"frame['aroon_up'] = ((period - frame['days_since_high']) / period) * 100",
				"frame['aroon_down'] = ((period - frame['days_since_low']) / period) * 100",
			},
			"latest['aroon_up'] > 70 and latest['aroon_down'] < 30",
			"latest['aroon_up'] < 40 and latest['aroon_down'] > 60",
			"Aroon 趋势 择时",
			"market-timing,trend-following",
			"normal-emotion,high-emotion",
			"rising-volume,incremental-volume",
			"market-rally,sideways-market",
			"trend-strength,momentum",
		),
		makeIndicatorSeed(
			"Williams%R-MACD反转",
			"Williams %R 和 MACD 共振时做超跌反转确认。参考来源: https://github.com/Nikhil-Adithyan/Algorithmic-Trading-with-Python",
			"williams_macd_reversal",
			[]string{
				"hh = frame['high'].rolling(14).max()",
				"ll = frame['low'].rolling(14).min()",
				"frame['williams_r'] = -100 * ((hh - frame['close']) / (hh - ll).replace(0, np.nan))",
				"frame['ema12'] = frame['close'].ewm(span=12, adjust=False).mean()",
				"frame['ema26'] = frame['close'].ewm(span=26, adjust=False).mean()",
				"frame['macd'] = frame['ema12'] - frame['ema26']",
				"frame['macd_signal'] = frame['macd'].ewm(span=9, adjust=False).mean()",
			},
			"latest['williams_r'] < -80 and latest['macd'] > latest['macd_signal']",
			"latest['williams_r'] > -20 or latest['macd'] < latest['macd_signal']",
			"Williams %R MACD 反转",
			"mean-reversion,reversal",
			"low-emotion,normal-emotion",
			"incremental-volume",
			"bottom-rebound,sideways-market",
			"momentum,mean-reversion",
		),
		makeIndicatorSeed(
			"Keltner通道突破",
			"Keltner Channel 配合 EMA 与 ATR 做通道突破。参考来源: https://github.com/Nikhil-Adithyan/Algorithmic-Trading-with-Python",
			"keltner_breakout",
			[]string{
				"frame['ema20'] = frame['close'].ewm(span=20, adjust=False).mean()",
				"tr_components = pd.concat([(frame['high'] - frame['low']), (frame['high'] - frame['close'].shift(1)).abs(), (frame['low'] - frame['close'].shift(1)).abs()], axis=1)",
				"frame['atr'] = tr_components.max(axis=1).rolling(14).mean()",
				"frame['kc_upper'] = frame['ema20'] + 2 * frame['atr']",
				"frame['kc_lower'] = frame['ema20'] - 2 * frame['atr']",
			},
			"latest['close'] > latest['kc_upper']",
			"latest['close'] < latest['kc_lower']",
			"Keltner 通道 突破",
			"trend-following,momentum-breakout",
			"normal-emotion,high-emotion",
			"surge-volume,persistent-volume",
			"market-rally,fast-rotation",
			"volatility,momentum",
		),
		makeIndicatorSeed(
			"StochRSI摆动交易",
			"把 RSI 再做随机化，适合短周期摆动反转。参考来源: https://github.com/cloudQuant/backtrader",
			"stoch_rsi_swing",
			[]string{
				"delta = frame['close'].diff()",
				"gain = delta.clip(lower=0).rolling(14).mean()",
				"loss = (-delta.clip(upper=0)).rolling(14).mean()",
				"rs = gain / loss.replace(0, np.nan)",
				"frame['rsi'] = 100 - (100 / (1 + rs))",
				"rsi_min = frame['rsi'].rolling(14).min()",
				"rsi_max = frame['rsi'].rolling(14).max()",
				"frame['stoch_rsi'] = (frame['rsi'] - rsi_min) / (rsi_max - rsi_min).replace(0, np.nan) * 100",
			},
			"latest['stoch_rsi'] < 20",
			"latest['stoch_rsi'] > 80",
			"Stoch RSI 摆动 反转",
			"mean-reversion,market-timing",
			"low-emotion,normal-emotion",
			"low-volume,incremental-volume",
			"sideways-market,high-divergence",
			"momentum,mean-reversion",
		),
		makeIndicatorSeed(
			"MFI资金流反转",
			"Money Flow Index 同时引入价格和成交量，适合资金驱动型反转。参考来源: https://github.com/nardew/talipp",
			"mfi_reversal",
			[]string{
				"tp = (frame['high'] + frame['low'] + frame['close']) / 3",
				"raw_money_flow = tp * frame['volume']",
				"positive_flow = np.where(tp > tp.shift(1), raw_money_flow, 0.0)",
				"negative_flow = np.where(tp < tp.shift(1), raw_money_flow, 0.0)",
				"pos_sum = pd.Series(positive_flow).rolling(14).sum()",
				"neg_sum = pd.Series(negative_flow).rolling(14).sum()",
				"money_ratio = pos_sum / neg_sum.replace(0, np.nan)",
				"frame['mfi'] = 100 - (100 / (1 + money_ratio))",
			},
			"latest['mfi'] < 20",
			"latest['mfi'] > 80",
			"MFI 资金流 反转",
			"mean-reversion,volume-confirmation",
			"low-emotion,normal-emotion",
			"incremental-volume,rising-volume",
			"bottom-rebound,sideways-market",
			"capital-flow,volume-ratio",
		),
		makeIndicatorSeed(
			"TRIX三重均线趋势",
			"TRIX 过滤短期噪音，适合中短周期趋势跟随。参考来源: https://github.com/cloudQuant/backtrader",
			"trix_trend",
			[]string{
				"ema1 = frame['close'].ewm(span=15, adjust=False).mean()",
				"ema2 = ema1.ewm(span=15, adjust=False).mean()",
				"ema3 = ema2.ewm(span=15, adjust=False).mean()",
				"frame['trix'] = ema3.pct_change() * 100",
				"frame['trix_signal'] = frame['trix'].rolling(9).mean()",
			},
			"latest['trix'] > latest['trix_signal'] and latest['trix'] > 0",
			"latest['trix'] < latest['trix_signal']",
			"TRIX 三重均线 趋势",
			"trend-following,market-timing",
			"normal-emotion,high-emotion",
			"incremental-volume,rising-volume",
			"market-rally,fast-rotation",
			"momentum,trend-strength",
		),
		makeIndicatorSeed(
			"CMF资金流确认",
			"Chaikin Money Flow 对量价收盘位置敏感，可辅助判断主力流入。参考来源: https://github.com/nardew/talipp",
			"cmf_confirmation",
			[]string{
				"mf_multiplier = ((frame['close'] - frame['low']) - (frame['high'] - frame['close'])) / (frame['high'] - frame['low']).replace(0, np.nan)",
				"mf_volume = mf_multiplier * frame['volume']",
				"frame['cmf'] = mf_volume.rolling(20).sum() / frame['volume'].rolling(20).sum()",
				"frame['ema20'] = frame['close'].ewm(span=20, adjust=False).mean()",
			},
			"latest['cmf'] > 0 and latest['close'] > latest['ema20']",
			"latest['cmf'] < 0 or latest['close'] < latest['ema20']",
			"CMF 资金流 确认",
			"volume-confirmation,trend-following",
			"normal-emotion,high-emotion",
			"rising-volume,persistent-volume",
			"market-rally,bottom-rebound",
			"capital-flow,volume-ratio,momentum",
		),
		makeIndicatorSeed(
			"ForceIndex突破确认",
			"Force Index 把涨跌和成交量合并，适合突破放量确认。参考来源: https://github.com/nardew/talipp",
			"force_index_breakout",
			[]string{
				"frame['force_index'] = frame['close'].diff() * frame['volume']",
				"frame['force_ma'] = frame['force_index'].ewm(span=13, adjust=False).mean()",
				"frame['price_ma'] = frame['close'].rolling(20).mean()",
			},
			"latest['force_ma'] > 0 and latest['close'] > latest['price_ma']",
			"latest['force_ma'] < 0 and latest['close'] < latest['price_ma']",
			"Force Index 突破 确认",
			"trend-following,volume-confirmation",
			"normal-emotion,high-emotion",
			"surge-volume,rising-volume",
			"market-rally,fast-rotation",
			"capital-flow,momentum",
		),
		makeIndicatorSeed(
			"DPO周期回归",
			"Detrended Price Oscillator 用于寻找短周期偏离和回归。参考来源: https://github.com/nardew/talipp",
			"dpo_cycle_reversion",
			[]string{
				"period = 20",
				"offset = int(period / 2) + 1",
				"frame['dpo'] = frame['close'].shift(offset) - frame['close'].rolling(period).mean()",
			},
			"latest['dpo'] < -1.5",
			"latest['dpo'] > 1.5",
			"DPO 周期 回归",
			"mean-reversion,cycle-trading",
			"low-emotion,normal-emotion",
			"low-volume,incremental-volume",
			"sideways-market,high-divergence",
			"cycle,mean-reversion",
		),
		makeIndicatorSeed(
			"ROC动量排序",
			"Rate of Change 用于简单动量排序和强弱切换。参考来源: https://github.com/cloudQuant/backtrader",
			"roc_momentum",
			[]string{
				"frame['roc'] = frame['close'].pct_change(12) * 100",
				"frame['roc_ma'] = frame['roc'].rolling(6).mean()",
			},
			"latest['roc'] > latest['roc_ma'] and latest['roc'] > 0",
			"latest['roc'] < latest['roc_ma']",
			"ROC 动量 排序",
			"trend-following,momentum-breakout",
			"normal-emotion,high-emotion",
			"rising-volume,persistent-volume",
			"market-rally,fast-rotation",
			"momentum,trend-strength",
		),
		makeIndicatorSeed(
			"DMI方向确认",
			"DMI 正负方向线交叉配合趋势过滤，适合追随主方向。参考来源: https://github.com/cloudQuant/backtrader",
			"dmi_confirmation",
			[]string{
				"up_move = frame['high'].diff()",
				"down_move = -frame['low'].diff()",
				"plus_dm = np.where((up_move > down_move) & (up_move > 0), up_move, 0.0)",
				"minus_dm = np.where((down_move > up_move) & (down_move > 0), down_move, 0.0)",
				"tr_components = pd.concat([(frame['high'] - frame['low']), (frame['high'] - frame['close'].shift(1)).abs(), (frame['low'] - frame['close'].shift(1)).abs()], axis=1)",
				"atr = tr_components.max(axis=1).rolling(14).mean()",
				"frame['plus_di'] = 100 * (pd.Series(plus_dm).rolling(14).mean() / atr.replace(0, np.nan))",
				"frame['minus_di'] = 100 * (pd.Series(minus_dm).rolling(14).mean() / atr.replace(0, np.nan))",
			},
			"latest['plus_di'] > latest['minus_di']",
			"latest['plus_di'] < latest['minus_di']",
			"DMI 方向 确认",
			"trend-following,market-timing",
			"normal-emotion,high-emotion",
			"incremental-volume,rising-volume",
			"market-rally,bottom-rebound",
			"trend-strength,momentum",
		),
		makeIndicatorSeed(
			"TEMA快线跟随",
			"TEMA 降低均线滞后，适合快节奏趋势跟随。参考来源: https://github.com/nardew/talipp",
			"tema_fast_follow",
			[]string{
				"ema1 = frame['close'].ewm(span=12, adjust=False).mean()",
				"ema2 = ema1.ewm(span=12, adjust=False).mean()",
				"ema3 = ema2.ewm(span=12, adjust=False).mean()",
				"frame['tema'] = 3 * ema1 - 3 * ema2 + ema3",
				"frame['tema_signal'] = frame['tema'].rolling(5).mean()",
			},
			"latest['tema'] > latest['tema_signal']",
			"latest['tema'] < latest['tema_signal']",
			"TEMA 快线 跟随",
			"trend-following,momentum-breakout",
			"normal-emotion,high-emotion",
			"rising-volume,persistent-volume",
			"fast-rotation,market-rally",
			"momentum,trend-strength",
		),
		makeIndicatorSeed(
			"Pivot枢轴突破",
			"利用前期枢轴位做突破与跌破判断，适合关键价位交易。参考来源: https://github.com/cloudQuant/backtrader",
			"pivot_breakout",
			[]string{
				"frame['pivot'] = (frame['high'].shift(1) + frame['low'].shift(1) + frame['close'].shift(1)) / 3",
				"frame['r1'] = frame['pivot'] * 2 - frame['low'].shift(1)",
				"frame['s1'] = frame['pivot'] * 2 - frame['high'].shift(1)",
			},
			"latest['close'] > latest['r1']",
			"latest['close'] < latest['s1']",
			"Pivot 枢轴 突破",
			"momentum-breakout,market-timing",
			"normal-emotion,high-emotion",
			"surge-volume,rising-volume",
			"market-rally,fast-rotation",
			"momentum,price-level",
		),
		makeIndicatorSeed(
			"ElderRay趋势强弱",
			"Elder Ray 用多头力量和空头力量衡量趋势健康度。参考来源: https://github.com/cloudQuant/backtrader",
			"elder_ray_strength",
			[]string{
				"frame['ema13'] = frame['close'].ewm(span=13, adjust=False).mean()",
				"frame['bull_power'] = frame['high'] - frame['ema13']",
				"frame['bear_power'] = frame['low'] - frame['ema13']",
			},
			"latest['bull_power'] > 0 and latest['bear_power'] > -1",
			"latest['bull_power'] < 0 and latest['bear_power'] < -1",
			"Elder Ray 趋势 强弱",
			"trend-following,market-timing",
			"normal-emotion,high-emotion",
			"rising-volume,incremental-volume",
			"market-rally,bottom-rebound",
			"trend-strength,momentum",
		),
		makeIndicatorSeed(
			"BB-KC压缩突破",
			"布林带与 Keltner 通道压缩后做波动扩张突破。参考来源: https://github.com/Nikhil-Adithyan/Algorithmic-Trading-with-Python",
			"bb_kc_squeeze",
			[]string{
				"frame['mid'] = frame['close'].rolling(20).mean()",
				"std = frame['close'].rolling(20).std()",
				"frame['bb_upper'] = frame['mid'] + 2 * std",
				"frame['bb_lower'] = frame['mid'] - 2 * std",
				"ema20 = frame['close'].ewm(span=20, adjust=False).mean()",
				"tr_components = pd.concat([(frame['high'] - frame['low']), (frame['high'] - frame['close'].shift(1)).abs(), (frame['low'] - frame['close'].shift(1)).abs()], axis=1)",
				"atr = tr_components.max(axis=1).rolling(14).mean()",
				"frame['kc_upper'] = ema20 + 1.5 * atr",
				"frame['kc_lower'] = ema20 - 1.5 * atr",
				"frame['squeeze'] = (frame['bb_upper'] < frame['kc_upper']) & (frame['bb_lower'] > frame['kc_lower'])",
			},
			"latest['close'] > latest['bb_upper'] and not latest['squeeze']",
			"latest['close'] < latest['bb_lower']",
			"BB KC Squeeze 突破",
			"momentum-breakout,volatility-expansion",
			"high-emotion,normal-emotion",
			"surge-volume,persistent-volume",
			"fast-rotation,market-rally",
			"volatility,momentum",
		),
		makeIndicatorSeed(
			"Chaikin振荡加速",
			"用 A/D 线和 Chaikin Oscillator 观察量价加速。参考来源: https://github.com/nardew/talipp",
			"chaikin_oscillator",
			[]string{
				"adl_multiplier = ((frame['close'] - frame['low']) - (frame['high'] - frame['close'])) / (frame['high'] - frame['low']).replace(0, np.nan)",
				"adl = (adl_multiplier * frame['volume']).cumsum()",
				"frame['chaikin'] = adl.ewm(span=3, adjust=False).mean() - adl.ewm(span=10, adjust=False).mean()",
			},
			"latest['chaikin'] > 0",
			"latest['chaikin'] < 0",
			"Chaikin 振荡 量价",
			"volume-confirmation,trend-following",
			"normal-emotion,high-emotion",
			"rising-volume,persistent-volume",
			"market-rally,fast-rotation",
			"capital-flow,volume-ratio",
		),
		makeIndicatorSeed(
			"UltimateOscillator反转",
			"把 7/14/28 日买卖压力融合成终极振荡器，适合超跌反转。参考来源: https://github.com/nardew/talipp",
			"ultimate_oscillator",
			[]string{
				"bp = frame['close'] - np.minimum(frame['low'], frame['close'].shift(1).fillna(frame['close']))",
				"tr = np.maximum(frame['high'], frame['close'].shift(1).fillna(frame['close'])) - np.minimum(frame['low'], frame['close'].shift(1).fillna(frame['close']))",
				"avg7 = bp.rolling(7).sum() / tr.rolling(7).sum().replace(0, np.nan)",
				"avg14 = bp.rolling(14).sum() / tr.rolling(14).sum().replace(0, np.nan)",
				"avg28 = bp.rolling(28).sum() / tr.rolling(28).sum().replace(0, np.nan)",
				"frame['uo'] = 100 * (4 * avg7 + 2 * avg14 + avg28) / 7",
			},
			"latest['uo'] < 30",
			"latest['uo'] > 70",
			"Ultimate Oscillator 反转",
			"mean-reversion,reversal",
			"low-emotion,normal-emotion",
			"incremental-volume",
			"bottom-rebound,sideways-market",
			"mean-reversion,momentum",
		),
	}
}

func makeIndicatorSeed(
	name string,
	description string,
	mode string,
	extraLines []string,
	buyExpr string,
	sellExpr string,
	searchKeywords string,
	styleTags string,
	emotionTags string,
	volumeTags string,
	scenarioTags string,
	factorTags string,
) additionalSeed {
	return additionalSeed{
		Name:            name,
		Description:     description,
		Code:            buildIndicatorCode(mode, extraLines, buyExpr, sellExpr),
		SearchKeywords:  searchKeywords,
		StyleTags:       styleTags,
		EmotionTags:     emotionTags,
		VolumeTags:      volumeTags,
		ScenarioTags:    scenarioTags,
		FactorTags:      factorTags,
		SourcePlatforms: "github",
		Parameters:      "lookback=14; risk_per_trade=0.02; signal_confirm=1",
	}
}

func buildIndicatorCode(mode string, extraLines []string, buyExpr string, sellExpr string) string {
	return fmt.Sprintf(`#!/usr/bin/env python
# -*- coding: utf-8 -*-
import pandas as pd
import numpy as np

def build_sample_data(periods=220):
    rng = np.random.default_rng(2026)
    close = 100 + np.linspace(0, 18, periods) + rng.normal(0, 1.6, periods).cumsum()
    frame = pd.DataFrame({"close": close})
    frame["open"] = frame["close"].shift(1).fillna(frame["close"])
    frame["high"] = np.maximum(frame["open"], frame["close"]) + rng.uniform(0.4, 1.8, periods)
    frame["low"] = np.minimum(frame["open"], frame["close"]) - rng.uniform(0.4, 1.8, periods)
    frame["volume"] = rng.integers(900_000, 3_600_000, periods)
%s
    return frame.dropna().reset_index(drop=True)

def generate_signal(frame):
    latest = frame.iloc[-1]
    prev = frame.iloc[-2]
    if %s:
        return "buy"
    if %s:
        return "sell"
    return "hold"

if __name__ == "__main__":
    data = build_sample_data()
    latest = data.iloc[-1]
    print({
        "strategy": %q,
        "signal": generate_signal(data),
        "close": float(latest["close"]),
        "mode": %q
    })
`, indentPython(extraLines), buyExpr, sellExpr, mode, mode)
}

func indentPython(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	return "    " + strings.Join(lines, "\n    ")
}
