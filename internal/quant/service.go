package quant

import (
	"encoding/xml"
	"fmt"
	"html"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"go-stock/backend/db"
	"go-stock/backend/logger"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-resty/resty/v2"
)

type Service struct {
	TemplateDir string
}

type bingRSS struct {
	Channel struct {
		Items []struct {
			Title       string `xml:"title"`
			Link        string `xml:"link"`
			Description string `xml:"description"`
		} `xml:"item"`
	} `xml:"channel"`
}

type searchSourceSpec struct {
	ID                  string
	DisplayName         string
	QueryHint           string
	AllowedHosts        []string
	AllowedPathPrefixes []string
}

func NewService(templateDir string) *Service {
	os.MkdirAll(templateDir, os.ModePerm)
	return &Service{TemplateDir: templateDir}
}

func normalizeSearchSourceID(source string) string {
	normalized := strings.ToLower(strings.TrimSpace(source))
	switch {
	case normalized == "myquant" || strings.Contains(normalized, "myquant") || strings.Contains(source, "掘金"):
		return "myquant"
	case normalized == "joinquant" || strings.Contains(normalized, "joinquant") || strings.Contains(source, "聚宽"):
		return "joinquant"
	case normalized == "ricequant" || strings.Contains(normalized, "ricequant"):
		return "ricequant"
	case normalized == "bigquant" || strings.Contains(normalized, "bigquant"):
		return "bigquant"
	case normalized == "github" || strings.Contains(normalized, "github"):
		return "github"
	case normalized == "papers" || strings.Contains(normalized, "arxiv") || strings.Contains(normalized, "ssrn") || strings.Contains(source, "论文"):
		return "papers"
	default:
		return normalized
	}
}

func getSearchSourceSpec(source string, requirePython bool, preferPlatform string) searchSourceSpec {
	suffixParts := make([]string, 0, 2)
	if requirePython {
		suffixParts = append(suffixParts, "Python")
	}
	if strings.TrimSpace(preferPlatform) != "" {
		suffixParts = append(suffixParts, strings.TrimSpace(preferPlatform))
	}
	suffix := strings.Join(suffixParts, " ")

	switch normalizeSearchSourceID(source) {
	case "myquant":
		return searchSourceSpec{
			ID:                  "myquant",
			DisplayName:         "掘金量化",
			QueryHint:           strings.TrimSpace("策略 示例 文档 " + suffix),
			AllowedHosts:        []string{"myquant.cn", "www.myquant.cn"},
			AllowedPathPrefixes: []string{"/docs"},
		}
	case "joinquant":
		return searchSourceSpec{
			ID:                  "joinquant",
			DisplayName:         "聚宽",
			QueryHint:           strings.TrimSpace("策略 社区 帖子 因子 " + suffix),
			AllowedHosts:        []string{"joinquant.com", "www.joinquant.com"},
			AllowedPathPrefixes: []string{"/community/post/detailMobile", "/community/post/detail", "/post/", "/view/community/detail"},
		}
	case "ricequant":
		return searchSourceSpec{
			ID:                  "ricequant",
			DisplayName:         "Ricequant",
			QueryHint:           strings.TrimSpace("研究 文档 因子 策略 " + suffix),
			AllowedHosts:        []string{"ricequant.com", "www.ricequant.com"},
			AllowedPathPrefixes: []string{"/doc"},
		}
	case "bigquant":
		return searchSourceSpec{
			ID:                  "bigquant",
			DisplayName:         "BigQuant",
			QueryHint:           strings.TrimSpace("策略 社区 因子 模型 " + suffix),
			AllowedHosts:        []string{"bigquant.com", "www.bigquant.com"},
			AllowedPathPrefixes: []string{"/wiki/collection", "/wiki/doc", "/codesharev2", "/codesharev3"},
		}
	case "github":
		return searchSourceSpec{
			ID:                  "github",
			DisplayName:         "GitHub",
			QueryHint:           strings.TrimSpace("quant strategy python " + suffix),
			AllowedHosts:        []string{"github.com", "www.github.com"},
			AllowedPathPrefixes: []string{"/"},
		}
	case "papers":
		return searchSourceSpec{
			ID:           "papers",
			DisplayName:  "论文/研究",
			QueryHint:    strings.TrimSpace("quant strategy factor research " + suffix),
			AllowedHosts: []string{"arxiv.org", "www.arxiv.org", "ssrn.com", "www.ssrn.com"},
		}
	default:
		return searchSourceSpec{
			ID:                  normalizeSearchSourceID(source),
			DisplayName:         strings.TrimSpace(source),
			QueryHint:           strings.TrimSpace(suffix),
			AllowedHosts:        nil,
			AllowedPathPrefixes: nil,
		}
	}
}

func (s *Service) GetCategories() []TemplateCategory {
	var categories []TemplateCategory
	db.Dao.Order("sort_order asc").Find(&categories)
	return categories
}

func (s *Service) CreateCategory(c TemplateCategory) *TemplateCategory {
	if err := db.Dao.Create(&c).Error; err != nil {
		logger.SugaredLogger.Errorf("create quant category failed: %v", err)
		return nil
	}
	return &c
}

func (s *Service) InitDefaultCategories() {
	defaults := []TemplateCategory{
		{Name: "策略框架", Description: "可直接运行的主策略脚本", SortOrder: 1},
		{Name: "多因子", Description: "适合因子筛选和打分组合的策略", SortOrder: 2},
		{Name: "网格交易", Description: "适合震荡区间和仓位分层的策略", SortOrder: 3},
		{Name: "情绪交易", Description: "适合结合市场情绪与热点轮动的策略", SortOrder: 4},
		{Name: "择时轮动", Description: "适合行业轮动和大盘风格切换", SortOrder: 5},
		{Name: "AI生成", Description: "由 AI 生成和修订的策略脚本", SortOrder: 6},
		{Name: "自定义", Description: "用户自由维护的量化脚本", SortOrder: 99},
	}
	for _, item := range defaults {
		var count int64
		db.Dao.Model(&TemplateCategory{}).Where("name = ?", item.Name).Count(&count)
		if count == 0 {
			db.Dao.Create(&item)
		}
	}
}

func (s *Service) SanitizeStoredTemplates() {
	var templates []Template
	if err := db.Dao.Find(&templates).Error; err != nil {
		logger.SugaredLogger.Errorf("load quant templates for sanitize failed: %v", err)
		return
	}

	for _, item := range templates {
		updates := map[string]any{}

		sanitizedCode := sanitizeTemplateCode(item.Code)
		if sanitizedCode != "" && sanitizedCode != item.Code {
			updates["code"] = sanitizedCode
		}

		sanitizedName := sanitizeTemplateName(item.Name, item.Description, item.ID)
		if sanitizedName != "" && sanitizedName != item.Name {
			updates["name"] = sanitizedName
		}

		sanitizedDescription := sanitizeTemplateDescription(item.Description)
		if sanitizedDescription != item.Description {
			updates["description"] = sanitizedDescription
		}

		if len(updates) == 0 {
			continue
		}
		if err := db.Dao.Model(&Template{}).Where("id = ?", item.ID).Updates(updates).Error; err != nil {
			logger.SugaredLogger.Errorf("sanitize quant template %d failed: %v", item.ID, err)
			continue
		}
		logger.SugaredLogger.Infof("sanitized quant template record: %d", item.ID)
	}
}

func (s *Service) GetTemplates(categoryId uint, status string, page, pageSize int) ([]Template, int64) {
	var templates []Template
	var total int64

	query := db.Dao.Model(&Template{}).Preload("Category")
	if categoryId > 0 {
		query = query.Where("category_id = ?", categoryId)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	query.Count(&total)
	query.Order("updated_at desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&templates)
	return templates, total
}

func (s *Service) GetTemplate(id uint) *Template {
	var t Template
	if err := db.Dao.Preload("Category").First(&t, id).Error; err != nil {
		return nil
	}
	return &t
}

func normalizeTemplate(t *Template) {
	t.Name = sanitizeTemplateName(t.Name, t.Description, t.ID)
	t.Description = sanitizeTemplateDescription(t.Description)
	t.Code = sanitizeTemplateCode(t.Code)
	if t.Language == "" {
		t.Language = "python"
	}
	if t.ScriptCategory == "" {
		t.ScriptCategory = "策略框架"
	}
	if t.Status == "" {
		t.Status = "draft"
	}
	if t.Version <= 0 {
		t.Version = 1
	}
}

func sanitizeTemplateName(name string, description string, id uint) string {
	cleaned := strings.TrimSpace(name)
	cleaned = strings.ReplaceAll(cleaned, "\r", " ")
	cleaned = strings.ReplaceAll(cleaned, "\n", " ")
	cleaned = regexp.MustCompile(`\s+`).ReplaceAllString(cleaned, " ")

	lower := strings.ToLower(cleaned)
	isBroken := cleaned == "" ||
		len([]rune(cleaned)) > 48 ||
		strings.Contains(lower, "```") ||
		strings.Contains(lower, "<think>") ||
		strings.Contains(lower, "import ") ||
		strings.Contains(lower, "def ") ||
		strings.Contains(lower, "class ") ||
		strings.Contains(lower, "#!/usr/bin") ||
		strings.Contains(lower, "请根据") ||
		strings.Contains(lower, "用户需求") ||
		strings.Contains(lower, "策略说明") ||
		strings.Contains(lower, "风险提示") ||
		strings.Contains(lower, "建议名称")

	if !isBroken {
		return truncateRunes(cleaned, 32)
	}

	desc := strings.TrimSpace(description)
	desc = strings.ReplaceAll(desc, "\r", " ")
	desc = strings.ReplaceAll(desc, "\n", " ")
	desc = regexp.MustCompile(`\s+`).ReplaceAllString(desc, " ")
	if desc != "" {
		return truncateRunes(desc, 24)
	}
	if id > 0 {
		return fmt.Sprintf("量化脚本_%d", id)
	}
	return "量化脚本"
}

func sanitizeTemplateDescription(description string) string {
	cleaned := strings.TrimSpace(description)
	if cleaned == "" {
		return ""
	}
	cleaned = strings.ReplaceAll(cleaned, "\r\n", "\n")
	cleaned = regexp.MustCompile(`\n{3,}`).ReplaceAllString(cleaned, "\n\n")
	return truncateRunes(cleaned, 400)
}

func truncateRunes(value string, limit int) string {
	runes := []rune(strings.TrimSpace(value))
	if len(runes) <= limit {
		return string(runes)
	}
	return strings.TrimSpace(string(runes[:limit]))
}

func (s *Service) CreateTemplate(t Template) *Template {
	normalizeTemplate(&t)
	t.Version = 1
	if err := db.Dao.Create(&t).Error; err != nil {
		logger.SugaredLogger.Errorf("create quant template failed: %v", err)
		return nil
	}
	return &t
}

func (s *Service) UpdateTemplate(t Template) *Template {
	var current Template
	if err := db.Dao.First(&current, t.ID).Error; err == nil {
		t.Version = current.Version + 1
	}
	normalizeTemplate(&t)
	if err := db.Dao.Save(&t).Error; err != nil {
		logger.SugaredLogger.Errorf("update quant template failed: %v", err)
		return nil
	}
	return &t
}

func (s *Service) DeleteTemplate(id uint) bool {
	return db.Dao.Delete(&Template{}, id).Error == nil
}

func (s *Service) ActivateTemplate(id uint) bool {
	now := time.Now()
	return db.Dao.Model(&Template{}).Where("id = ?", id).
		Updates(map[string]any{"status": "active", "last_used_at": &now}).Error == nil
}

func (s *Service) ExportTemplate(id uint) (string, error) {
	t := s.GetTemplate(id)
	if t == nil {
		return "", fmt.Errorf("template not found")
	}
	filename := fmt.Sprintf("%s_v%d.py", t.Name, t.Version)
	targetPath := filepath.Join(s.TemplateDir, filename)
	if err := os.WriteFile(targetPath, []byte(t.Code), 0644); err != nil {
		return "", fmt.Errorf("export template failed: %v", err)
	}
	return targetPath, nil
}

func (s *Service) GetTagTaxonomy() []TagGroup {
	return []TagGroup{
		{
			Key:         "style",
			Label:       "策略类型",
			Description: "策略主风格与执行方式",
			Options: []TagOption{
				{Label: "多因子", Value: "multi-factor"},
				{Label: "网格交易", Value: "grid-trading"},
				{Label: "趋势跟随", Value: "trend-following"},
				{Label: "均值回归", Value: "mean-reversion"},
				{Label: "事件驱动", Value: "event-driven"},
				{Label: "高频微结构", Value: "micro-structure"},
				{Label: "行业轮动", Value: "sector-rotation"},
				{Label: "择时", Value: "market-timing"},
				{Label: "套利", Value: "arbitrage"},
				{Label: "打板/强势股", Value: "momentum-breakout"},
				{Label: "ETF轮动", Value: "etf-rotation"},
				{Label: "可转债", Value: "convertible-bond"},
			},
		},
		{
			Key:         "emotion",
			Label:       "情绪适配",
			Description: "适合当前市场情绪环境的脚本",
			Options: []TagOption{
				{Label: "标准情绪", Value: "normal-emotion"},
				{Label: "低情绪", Value: "low-emotion"},
				{Label: "高情绪", Value: "high-emotion"},
				{Label: "极端恐慌", Value: "panic"},
				{Label: "亢奋追涨", Value: "euphoria"},
			},
		},
		{
			Key:         "volume",
			Label:       "量能适配",
			Description: "适合不同成交量环境的脚本",
			Options: []TagOption{
				{Label: "缩量", Value: "low-volume"},
				{Label: "温和放量", Value: "rising-volume"},
				{Label: "增量", Value: "incremental-volume"},
				{Label: "爆量", Value: "surge-volume"},
				{Label: "持续放量", Value: "persistent-volume"},
			},
		},
		{
			Key:         "scenario",
			Label:       "适配场景",
			Description: "适合特定市场场景的脚本",
			Options: []TagOption{
				{Label: "大盘大幅下跌", Value: "market-sharp-drop"},
				{Label: "大盘单边上涨", Value: "market-rally"},
				{Label: "震荡箱体", Value: "sideways-market"},
				{Label: "热点轮动加快", Value: "fast-rotation"},
				{Label: "板块抱团", Value: "sector-crowding"},
				{Label: "高位分歧", Value: "high-divergence"},
				{Label: "低位修复", Value: "bottom-rebound"},
			},
		},
		{
			Key:         "capital",
			Label:       "资金适配",
			Description: "适合不同资金规模的脚本",
			Options: []TagOption{
				{Label: "1万以下", Value: "capital-under-10k"},
				{Label: "1-2万", Value: "capital-10k-20k"},
				{Label: "2-3万", Value: "capital-20k-30k"},
				{Label: "3-5万", Value: "capital-30k-50k"},
				{Label: "5-10万", Value: "capital-50k-100k"},
				{Label: "10-30万", Value: "capital-100k-300k"},
				{Label: "30万以上", Value: "capital-above-300k"},
			},
		},
		{
			Key:         "factor",
			Label:       "特征因子",
			Description: "可用于因子筛选和脚本检索的维度",
			Options: []TagOption{
				{Label: "动量", Value: "momentum"},
				{Label: "波动率", Value: "volatility"},
				{Label: "换手率", Value: "turnover"},
				{Label: "量比", Value: "volume-ratio"},
				{Label: "资金流", Value: "capital-flow"},
				{Label: "情绪强度", Value: "sentiment-strength"},
				{Label: "行业热度", Value: "industry-heat"},
				{Label: "热词共振", Value: "theme-resonance"},
				{Label: "收益质量", Value: "quality"},
				{Label: "估值", Value: "valuation"},
				{Label: "盈利修正", Value: "earnings-revision"},
				{Label: "北向资金", Value: "northbound-flow"},
			},
		},
		{
			Key:         "script-category",
			Label:       "脚本分类",
			Description: "脚本在库中的管理类别",
			Options: []TagOption{
				{Label: "策略主脚本", Value: "strategy-main"},
				{Label: "选股器", Value: "stock-screener"},
				{Label: "因子计算", Value: "factor-engine"},
				{Label: "风控模块", Value: "risk-control"},
				{Label: "回测脚本", Value: "backtest"},
				{Label: "执行器", Value: "execution"},
				{Label: "监控告警", Value: "monitor-alert"},
				{Label: "工具函数", Value: "utility"},
			},
		},
	}
}

func (s *Service) BuildScriptSearchLinks(query string) []SearchLink {
	return []SearchLink{
		{
			Name:        "MyQuant Docs",
			URL:         "https://www.bing.com/search?q=" + url.QueryEscape("site:myquant.cn/docs "+query+" Python strategy"),
			Description: "Target public MyQuant docs, examples, and strategy notes",
		},
		{
			Name:        "JoinQuant Community",
			URL:         "https://www.bing.com/search?q=" + url.QueryEscape("(site:joinquant.com/view/community/detail OR site:joinquant.com/post OR site:joinquant.com/community/post/detailMobile) "+query+" strategy Python"),
			Description: "Target JoinQuant community posts and strategy research",
		},
		{
			Name:        "Ricequant",
			URL:         "https://www.bing.com/search?q=" + url.QueryEscape("site:ricequant.com/doc "+query+" strategy factor"),
			Description: "Target Ricequant docs, research, and factor articles",
		},
		{
			Name:        "BigQuant",
			URL:         "https://www.bing.com/search?q=" + url.QueryEscape("(site:bigquant.com/wiki/doc OR site:bigquant.com/wiki/collection OR site:bigquant.com/codesharev2 OR site:bigquant.com/codesharev3) "+query+" strategy factor"),
			Description: "Target BigQuant community, factor pages, and code share pages",
		},
		{
			Name:        "GitHub",
			URL:         "https://github.com/search?q=" + url.QueryEscape(query+" quant python strategy") + "&type=repositories",
			Description: "Search public Python quant strategy repositories",
		},
		{
			Name:        "Papers/Research",
			URL:         "https://www.bing.com/search?q=" + url.QueryEscape("(site:arxiv.org OR site:ssrn.com) "+query+" quant strategy"),
			Description: "Search quant research papers and related studies",
		},
	}
}

func (s *Service) BuildLinkageAIPrompt(summary string, templates []Template) string {
	var builder strings.Builder
	builder.WriteString("你是一名负责量化脚本切换建议的研究助理。\n")
	builder.WriteString("请根据当前市场状态，从候选量化脚本中选出最值得切换的一只，并给出专业、简洁、可执行的理由。\n")
	builder.WriteString("输出必须是 JSON，对象结构固定如下：\n")
	builder.WriteString("{\"scriptName\":\"\",\"action\":\"\",\"reason\":\"\",\"riskHint\":\"\",\"confidence\":0}\n")
	builder.WriteString("要求：\n")
	builder.WriteString("1. scriptName 必须严格使用候选脚本中的原始名称。\n")
	builder.WriteString("2. action 只能是：优先切换 / 观察等待 / 暂不切换。\n")
	builder.WriteString("3. reason 需要包含市场驱动因素和脚本匹配点。\n")
	builder.WriteString("4. confidence 取值范围 0 到 10。\n")
	builder.WriteString("5. 不要输出 JSON 之外的任何说明。\n\n")
	builder.WriteString("当前市场摘要：\n")
	builder.WriteString(strings.TrimSpace(summary))
	builder.WriteString("\n\n候选脚本：\n")
	for idx, item := range templates {
		tags := collectTemplateTagTexts(item)
		builder.WriteString(fmt.Sprintf("%d. 名称: %s\n", idx+1, strings.TrimSpace(item.Name)))
		builder.WriteString(fmt.Sprintf("   状态: %s\n", strings.TrimSpace(item.Status)))
		builder.WriteString(fmt.Sprintf("   分类: %s / %s / %s\n", strings.TrimSpace(item.ScriptCategory), strings.TrimSpace(item.StrategyType), strings.TrimSpace(item.BrokerPlatform)))
		builder.WriteString(fmt.Sprintf("   标签: %s\n", strings.Join(tags, ", ")))
		builder.WriteString(fmt.Sprintf("   说明: %s\n", strings.TrimSpace(item.Description)))
		builder.WriteString(fmt.Sprintf("   关键词: %s\n", strings.TrimSpace(item.SearchKeywords)))
	}
	return builder.String()
}

func collectTemplateTagTexts(item Template) []string {
	parts := []string{
		item.StyleTags,
		item.EmotionTags,
		item.VolumeTags,
		item.ScenarioTags,
		item.CapitalTags,
		item.FactorTags,
		item.LinkedStocks,
	}
	seen := map[string]struct{}{}
	result := make([]string, 0, 16)
	for _, block := range parts {
		for _, raw := range strings.Split(block, ",") {
			value := strings.TrimSpace(raw)
			if value == "" {
				continue
			}
			if _, ok := seen[value]; ok {
				continue
			}
			seen[value] = struct{}{}
			result = append(result, value)
		}
	}
	return result
}

func (s *Service) SearchScriptSources(req ScriptSearchRequest) ([]SearchHit, error) {
	query := strings.TrimSpace(req.Query)
	if query == "" {
		return []SearchHit{}, nil
	}

	sources := req.Sources
	if len(sources) == 0 {
		sources = []string{"myquant", "joinquant", "ricequant", "bigquant", "github"}
	}

	limit := req.ResultLimit
	if limit <= 0 {
		limit = 18
	}

	perSource := 3
	if limit < len(sources) {
		perSource = 1
	}

	allHits := make([]SearchHit, 0, limit)
	seen := map[string]bool{}
	for _, source := range sources {
		queryVariants := s.buildSearchQueriesForSource(source, query, req.RequirePython, req.PreferPlatform)
		hits := make([]SearchHit, 0, perSource)
		for _, searchQuery := range queryVariants {
			found, err := s.searchBingRSS(source, searchQuery, perSource)
			if err != nil {
				logger.SugaredLogger.Warnf("search source %s failed: %v", source, err)
				continue
			}
			hits = appendDedupHits(hits, found, perSource*2)
			if len(hits) >= perSource {
				break
			}
		}
		if len(hits) == 0 {
			fallbackHits, fallbackErr := s.searchSourceFallback(source, query, perSource)
			if fallbackErr != nil {
				logger.SugaredLogger.Warnf("fallback search source %s failed: %v", source, fallbackErr)
			} else {
				hits = fallbackHits
			}
		}
		for _, hit := range hits {
			if seen[hit.URL] {
				continue
			}
			seen[hit.URL] = true
			allHits = append(allHits, hit)
			if len(allHits) >= limit {
				break
			}
		}
		if len(allHits) >= limit {
			break
		}
	}

	sort.SliceStable(allHits, func(i, j int) bool {
		return scoreSearchHit(allHits[i], query) > scoreSearchHit(allHits[j], query)
	})

	allHits = s.enrichSearchHits(query, allHits)
	sort.SliceStable(allHits, func(i, j int) bool {
		return allHits[i].MatchScore > allHits[j].MatchScore
	})

	return allHits, nil
}

func (s *Service) BuildGeneratePrompt(req GenerateRequest) string {
	return fmt.Sprintf(`你是一位资深量化交易开发工程师。请根据以下需求生成一个完整、可执行、结构清晰的 Python 量化脚本。

## 需求概览
- 策略描述: %s
- 平台适配: %s
- 策略类型: %s
- 脚本分类: %s
- 关联标的: %s
- 风险等级: %s
- 资金规模: %.2f 元
- 重点因子: %s
- 重点场景: %s

## 输出要求
1. 输出完整 Python 代码，包含必要导入、参数区、策略类或主函数、信号逻辑、风控逻辑和日志。
2. 如果适合，请将策略拆成“因子计算 / 信号生成 / 仓位管理 / 执行”几个模块。
3. 注释清晰，便于直接存入脚本库复用。
4. 风控必须包含止损、仓位上限、异常行情保护。
5. 代码应优先兼容常见量化环境，尽量使用 pandas / numpy。
6. 最后附一段“参数建议”和“适用行情提示”。
`,
		req.StrategyDescription,
		req.BrokerPlatform,
		req.StrategyType,
		req.ScriptCategory,
		req.StockCodes,
		req.RiskLevel,
		req.Capital,
		req.FactorTags,
		req.SceneTags,
	)
}

func (s *Service) BuildSearchAgentPrompt(req ScriptSearchRequest, hits []SearchHit) string {
	lines := []string{
		"你是一名量化研究助理和脚本检索 Agent。",
		"请根据用户检索意图，从候选结果中筛选最值得看的量化脚本来源，并给出精炼建议。",
		"",
		"## 用户需求",
		fmt.Sprintf("- 检索关键词: %s", strings.TrimSpace(req.Query)),
		fmt.Sprintf("- 偏好平台: %s", strings.TrimSpace(req.PreferPlatform)),
		fmt.Sprintf("- 必须偏向 Python: %t", req.RequirePython),
		"",
		"## 候选搜索结果",
	}

	for index, hit := range hits {
		lines = append(lines,
			fmt.Sprintf("%d. 来源: %s", index+1, hit.Source),
			fmt.Sprintf("   标题: %s", hit.Title),
			fmt.Sprintf("   链接: %s", hit.URL),
			fmt.Sprintf("   摘要: %s", hit.Snippet),
		)
	}

	lines = append(lines,
		"",
		"## 输出要求",
		"请只按以下结构输出：",
		"## 检索结论",
		"用 3 到 6 行总结最值得看的结果与原因。",
		"## 优先查看",
		"列出 3 到 5 条，格式为“1. 来源 - 标题 - 理由”。",
		"## 筛选建议",
		"给出 3 到 5 条后续筛选建议，聚焦平台、因子、场景和可运行性。",
		"## 可直接入库",
		"指出哪些结果更适合直接沉淀到本地量化程序库，并说明原因。",
	)

	return strings.Join(lines, "\n")
}

func (s *Service) BuildGeneratePromptWithContext(req GenerateRequest) string {
	var builder strings.Builder
	builder.WriteString("你是一位资深量化交易开发工程师。请根据以下需求生成一个完整、可执行、结构清晰的 Python 量化脚本。\n\n")
	builder.WriteString("## 需求概览\n")
	builder.WriteString(fmt.Sprintf("- 策略描述: %s\n", strings.TrimSpace(req.StrategyDescription)))
	builder.WriteString(fmt.Sprintf("- 平台适配: %s\n", strings.TrimSpace(req.BrokerPlatform)))
	builder.WriteString(fmt.Sprintf("- 策略类型: %s\n", strings.TrimSpace(req.StrategyType)))
	builder.WriteString(fmt.Sprintf("- 脚本分类: %s\n", strings.TrimSpace(req.ScriptCategory)))
	builder.WriteString(fmt.Sprintf("- 关联标的: %s\n", strings.TrimSpace(req.StockCodes)))
	builder.WriteString(fmt.Sprintf("- 风险等级: %s\n", strings.TrimSpace(req.RiskLevel)))
	builder.WriteString(fmt.Sprintf("- 资金规模: %.2f 元\n", req.Capital))
	builder.WriteString(fmt.Sprintf("- 重点因子: %s\n", strings.TrimSpace(req.FactorTags)))
	builder.WriteString(fmt.Sprintf("- 重点场景: %s\n", strings.TrimSpace(req.SceneTags)))

	if strings.TrimSpace(req.ExistingScriptName) != "" || strings.TrimSpace(req.ExistingDescription) != "" || strings.TrimSpace(req.BaseCode) != "" {
		builder.WriteString("\n## 增量迭代上下文\n")
		if strings.TrimSpace(req.ExistingScriptName) != "" {
			builder.WriteString(fmt.Sprintf("- 原脚本名称: %s\n", strings.TrimSpace(req.ExistingScriptName)))
		}
		if strings.TrimSpace(req.ExistingDescription) != "" {
			builder.WriteString(fmt.Sprintf("- 原脚本说明: %s\n", strings.TrimSpace(req.ExistingDescription)))
		}
		builder.WriteString("- 这是在已有脚本基础上的继续修改，请优先复用原有合理结构，并按本次要求升级、修复或重构。\n")
		builder.WriteString("- 如果原脚本有明显 bug、风控不足或结构混乱，可以直接修正，但不要偏离原策略意图。\n")
		if strings.TrimSpace(req.BaseCode) != "" {
			builder.WriteString("\n## 原脚本代码\n```python\n")
			builder.WriteString(strings.TrimSpace(req.BaseCode))
			builder.WriteString("\n```\n")
		}
	}

	builder.WriteString(`

## 输出要求
1. 输出完整 Python 代码，包含必要导入、参数区、策略类或主函数、信号逻辑、风控逻辑和日志。
2. 如适合，请将策略拆成“因子计算 / 信号生成 / 仓位管理 / 执行”几个模块。
3. 注释清晰，便于直接存入脚本库复用。
4. 风控必须包含止损、仓位上限、异常行情保护。
5. 代码应优先兼容常见量化环境，尽量使用 pandas / numpy。
6. 最后附一段“参数建议”和“适用行情提示”。
`)
	return builder.String()
}

func sanitizeTemplateCode(code string) string {
	raw := strings.TrimSpace(code)
	if raw == "" {
		return ""
	}

	if strings.Contains(raw, "```") {
		re := regexp.MustCompile("(?s)```(?:python)?\\s*(.*?)```")
		match := re.FindStringSubmatch(raw)
		if len(match) >= 2 {
			return strings.TrimSpace(match[1])
		}
	}

	if strings.Contains(strings.ToLower(raw), "<think>") {
		re := regexp.MustCompile("(?is)<think>.*?</think>")
		cleaned := strings.TrimSpace(re.ReplaceAllString(raw, ""))
		return cleaned
	}

	return raw
}

func (s *Service) buildSearchQueryForSource(source string, query string, requirePython bool, preferPlatform string) string {
	spec := getSearchSourceSpec(source, requirePython, preferPlatform)
	switch spec.ID {
	case "myquant":
		return strings.TrimSpace("site:myquant.cn/docs (" + buildKeywordExpansion(query) + ") " + spec.QueryHint)
	case "joinquant":
		return strings.TrimSpace("(site:joinquant.com/view/community/detail OR site:joinquant.com/post OR site:joinquant.com/community/post/detailMobile OR site:joinquant.com/community/post/detail) (" + buildKeywordExpansion(query) + ") " + spec.QueryHint)
	case "ricequant":
		return strings.TrimSpace("(site:ricequant.com/doc OR site:ricequant.com/community) (" + buildKeywordExpansion(query) + ") " + spec.QueryHint)
	case "bigquant":
		return strings.TrimSpace("(site:bigquant.com/codesharev2 OR site:bigquant.com/codesharev3 OR site:bigquant.com/wiki/doc) (" + buildKeywordExpansion(query) + ") " + spec.QueryHint)
	case "github":
		return strings.TrimSpace("site:github.com (" + buildKeywordExpansion(query) + ") " + spec.QueryHint)
	case "papers":
		return strings.TrimSpace("(site:arxiv.org OR site:ssrn.com) (" + buildKeywordExpansion(query) + ") " + spec.QueryHint)
	default:
		return strings.TrimSpace("(" + buildKeywordExpansion(query) + ") " + spec.QueryHint)
	}
}

func (s *Service) buildSearchQueriesForSource(source string, query string, requirePython bool, preferPlatform string) []string {
	queries := []string{s.buildSearchQueryForSource(source, query, requirePython, preferPlatform)}
	expanded := buildKeywordExpansion(query)
	if expanded != strings.TrimSpace(query) {
		spec := getSearchSourceSpec(source, requirePython, preferPlatform)
		queries = append(queries, strings.TrimSpace(expanded+" "+spec.QueryHint))
	}
	return dedupeStrings(queries)
}

func (s *Service) searchBingRSS(source string, query string, limit int) ([]SearchHit, error) {
	spec := getSearchSourceSpec(source, strings.Contains(strings.ToLower(query), "python"), "")
	client := resty.New().
		SetTimeout(20*time.Second).
		SetHeader("User-Agent", "Mozilla/5.0 investment-platform quant-search")

	resp, err := client.R().
		SetQueryParams(map[string]string{
			"q":      query,
			"format": "rss",
			"mkt":    "zh-CN",
		}).
		Get("https://www.bing.com/search")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() >= 400 {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode())
	}

	var feed bingRSS
	if err := xml.Unmarshal(resp.Body(), &feed); err != nil {
		return nil, err
	}

	hits := make([]SearchHit, 0, limit)
	for _, item := range feed.Channel.Items {
		if strings.TrimSpace(item.Link) == "" {
			continue
		}
		if !searchHitMatchesAllowedHosts(item.Link, spec.AllowedHosts) {
			continue
		}
		if !searchHitMatchesAllowedPaths(item.Link, spec.AllowedPathPrefixes) {
			continue
		}
		hits = append(hits, SearchHit{
			Source:  spec.DisplayName,
			Title:   strings.TrimSpace(htmlStrip(item.Title)),
			URL:     strings.TrimSpace(item.Link),
			Snippet: strings.TrimSpace(htmlStrip(item.Description)),
		})
		if len(hits) >= limit {
			break
		}
	}
	return hits, nil
}

func (s *Service) searchSourceFallback(source string, query string, limit int) ([]SearchHit, error) {
	spec := getSearchSourceSpec(source, strings.Contains(strings.ToLower(query), "python"), "")
	switch spec.ID {
	case "github":
		return s.searchGitHubRepos(spec, query, limit)
	case "joinquant":
		return s.scrapePlatformLinks(spec, "https://www.joinquant.com/", query, limit, []string{"/post/", "/community/post/detailMobile", "/community/post/detail", "/view/community/detail"})
	case "bigquant":
		return s.scrapePlatformLinks(spec, "https://bigquant.com/wiki/home", query, limit, []string{"/codesharev2", "/codesharev3", "/wiki/doc"})
	case "myquant":
		return s.scrapePlatformLinks(spec, "https://www.myquant.cn/docs/python", query, limit, spec.AllowedPathPrefixes)
	case "ricequant":
		return s.scrapePlatformLinks(spec, "https://www.ricequant.com/doc/", query, limit, []string{"/doc"})
	default:
		return []SearchHit{}, nil
	}
}

func (s *Service) searchGitHubRepos(spec searchSourceSpec, query string, limit int) ([]SearchHit, error) {
	client := resty.New().
		SetTimeout(20*time.Second).
		SetHeader("User-Agent", "investment-platform quant-search").
		SetHeader("Accept", "application/vnd.github+json")

	var repoPayload struct {
		Items []struct {
			FullName    string `json:"full_name"`
			HTMLURL     string `json:"html_url"`
			Description string `json:"description"`
		} `json:"items"`
	}

	repoQuery := buildGitHubSearchQuery(query)
	resp, err := client.R().
		SetResult(&repoPayload).
		SetQueryParams(map[string]string{
			"q":        repoQuery,
			"sort":     "stars",
			"order":    "desc",
			"per_page": fmt.Sprintf("%d", limit),
		}).
		Get("https://api.github.com/search/repositories")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() >= 400 {
		return nil, fmt.Errorf("github search HTTP %d", resp.StatusCode())
	}

	hits := make([]SearchHit, 0, limit)
	for _, item := range repoPayload.Items {
		hits = append(hits, SearchHit{
			Source:  spec.DisplayName,
			Title:   strings.TrimSpace(item.FullName),
			URL:     strings.TrimSpace(item.HTMLURL),
			Snippet: strings.TrimSpace(item.Description),
		})
		if len(hits) >= limit {
			break
		}
	}

	var codePayload struct {
		Items []struct {
			Name       string `json:"name"`
			HTMLURL    string `json:"html_url"`
			Path       string `json:"path"`
			Repository struct {
				FullName string `json:"full_name"`
			} `json:"repository"`
		} `json:"items"`
	}

	codeResp, codeErr := client.R().
		SetResult(&codePayload).
		SetQueryParams(map[string]string{
			"q":        repoQuery + " extension:py",
			"per_page": fmt.Sprintf("%d", limit),
		}).
		Get("https://api.github.com/search/code")
	if codeErr == nil && codeResp.StatusCode() < 400 {
		for _, item := range codePayload.Items {
			hits = appendDedupHits(hits, []SearchHit{
				{
					Source:  spec.DisplayName,
					Title:   strings.TrimSpace(item.Repository.FullName + " / " + item.Path),
					URL:     strings.TrimSpace(item.HTMLURL),
					Snippet: "GitHub Python code result",
				},
			}, limit)
			if len(hits) >= limit {
				break
			}
		}
	}

	sort.SliceStable(hits, func(i, j int) bool {
		return scoreSearchHit(hits[i], query) > scoreSearchHit(hits[j], query)
	})
	if len(hits) > limit {
		hits = hits[:limit]
	}
	return hits, nil
}

func (s *Service) scrapePlatformLinks(spec searchSourceSpec, pageURL string, query string, limit int, allowedPrefixes []string) ([]SearchHit, error) {
	client := resty.New().
		SetTimeout(20*time.Second).
		SetHeader("User-Agent", "Mozilla/5.0 investment-platform quant-search")

	resp, err := client.R().Get(pageURL)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() >= 400 {
		return nil, fmt.Errorf("fetch page HTTP %d", resp.StatusCode())
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(resp.String()))
	if err != nil {
		return nil, err
	}

	tokens := tokenizeSearchQuery(query)
	candidates := make([]SearchHit, 0, limit*2)
	seen := map[string]bool{}

	doc.Find("a[href]").Each(func(_ int, sel *goquery.Selection) {
		if len(candidates) >= limit*4 {
			return
		}
		href, ok := sel.Attr("href")
		if !ok {
			return
		}
		link := absoluteSearchURL(pageURL, href)
		if link == "" || seen[link] {
			return
		}
		if !searchHitMatchesAllowedHosts(link, spec.AllowedHosts) {
			return
		}
		if !searchHitMatchesAllowedPaths(link, allowedPrefixes) {
			return
		}

		title := strings.TrimSpace(htmlStrip(sel.Text()))
		if title == "" {
			return
		}

		textForMatch := strings.ToLower(title + " " + link)
		score := 0
		for _, token := range tokens {
			if strings.Contains(textForMatch, token) {
				score += 2
			}
		}
		if len(tokens) > 0 && score == 0 && len(candidates) >= limit {
			return
		}

		seen[link] = true
		candidates = append(candidates, SearchHit{
			Source:  spec.DisplayName,
			Title:   title,
			URL:     link,
			Snippet: fmt.Sprintf("Curated from %s public page", spec.DisplayName),
		})
	})

	sort.SliceStable(candidates, func(i, j int) bool {
		return scoreSearchHit(candidates[i], query) > scoreSearchHit(candidates[j], query)
	})
	if len(candidates) > limit {
		candidates = candidates[:limit]
	}
	return candidates, nil
}

func searchHitMatchesAllowedHosts(rawURL string, allowedHosts []string) bool {
	if len(allowedHosts) == 0 {
		return true
	}

	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return false
	}

	host := strings.ToLower(parsed.Hostname())
	for _, allowed := range allowedHosts {
		allowed = strings.ToLower(strings.TrimSpace(allowed))
		if allowed == "" {
			continue
		}
		if host == allowed || strings.HasSuffix(host, "."+allowed) {
			return true
		}
	}
	return false
}

func searchHitMatchesAllowedPaths(rawURL string, allowedPathPrefixes []string) bool {
	if len(allowedPathPrefixes) == 0 {
		return true
	}

	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return false
	}

	pathValue := parsed.EscapedPath()
	if pathValue == "" {
		pathValue = parsed.Path
	}
	if pathValue == "" {
		return false
	}

	for _, prefix := range allowedPathPrefixes {
		prefix = strings.TrimSpace(prefix)
		if prefix == "" {
			continue
		}
		if prefix == "/" || strings.HasPrefix(pathValue, prefix) {
			return true
		}
	}
	return false
}

func absoluteSearchURL(baseURL string, href string) string {
	href = strings.TrimSpace(href)
	if href == "" || strings.HasPrefix(href, "#") || strings.HasPrefix(strings.ToLower(href), "javascript:") {
		return ""
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		return ""
	}
	ref, err := url.Parse(href)
	if err != nil {
		return ""
	}
	return base.ResolveReference(ref).String()
}

func tokenizeSearchQuery(query string) []string {
	rawTokens := strings.Fields(strings.ToLower(buildKeywordExpansion(query)))
	tokens := make([]string, 0, len(rawTokens))
	for _, token := range rawTokens {
		token = strings.TrimSpace(token)
		if len([]rune(token)) < 2 {
			continue
		}
		switch token {
		case "python", "strategy", "qmt", "etf", "script":
			tokens = append(tokens, token)
		default:
			tokens = append(tokens, token)
		}
	}
	return tokens
}

func htmlStrip(value string) string {
	re := regexp.MustCompile("<[^>]+>")
	cleaned := re.ReplaceAllString(value, " ")
	cleaned = strings.ReplaceAll(cleaned, "&quot;", "\"")
	cleaned = strings.ReplaceAll(cleaned, "&amp;", "&")
	cleaned = strings.ReplaceAll(cleaned, "&#39;", "'")
	return strings.Join(strings.Fields(cleaned), " ")
}

func scoreSearchHit(hit SearchHit, query string) int {
	score := 0
	text := strings.ToLower(hit.Title + " " + hit.Snippet + " " + hit.ContentSnippet + " " + hit.ImportHint + " " + hit.Source)
	for _, token := range tokenizeSearchQuery(query) {
		if strings.Contains(text, token) {
			score += 3
		}
	}
	if strings.Contains(text, "python") {
		score += 2
	}
	if strings.Contains(text, "??") || strings.Contains(text, "strategy") {
		score += 2
	}
	if strings.Contains(text, "??") || strings.Contains(text, "backtest") {
		score += 1
	}
	urlLower := strings.ToLower(hit.URL)
	if strings.Contains(urlLower, "joinquant.com/community/post/detailmobile") {
		score += 4
	}
	if strings.Contains(urlLower, "joinquant.com/view/community/detail") {
		score += 6
	}
	if strings.Contains(urlLower, "bigquant.com/codesharev") || strings.Contains(urlLower, "github.com") {
		score += 6
	}
	if strings.Contains(urlLower, "bigquant.com/wiki/doc") || strings.Contains(urlLower, "joinquant.com/post/") {
		score += 4
	}
	if strings.Contains(urlLower, "myquant.cn/docs") || strings.Contains(urlLower, "ricequant.com/doc") {
		score += 2
	}
	if strings.Contains(text, "etf") {
		score += 3
	}
	if strings.Contains(text, "网格") || strings.Contains(text, "grid") {
		score += 3
	}
	if strings.Contains(text, "源码") || strings.Contains(text, "策略") || strings.Contains(text, "回测") || strings.Contains(text, "codeshare") {
		score += 2
	}
	if strings.Contains(text, "文档") || strings.Contains(text, "api文档") || strings.Contains(text, "快速上手") || strings.Contains(text, "精华帖子") || strings.Contains(text, "私享会") {
		score -= 4
	}
	if hit.HasCode {
		score += 8
	}
	if strings.TrimSpace(hit.CodePreview) != "" {
		score += 4
	}
	switch hit.CandidateType {
	case "github-code":
		score += 8
	case "codeshare":
		score += 7
	case "github-repo":
		score += 5
	case "community-post":
		score += 2
	case "docs":
		score += 1
	}
	if strings.Contains(urlLower, "zhidao.baidu.com") || strings.Contains(urlLower, "jingyan.baidu.com") {
		score -= 10
	}
	return score
}

func buildKeywordExpansion(query string) string {
	parts := []string{strings.TrimSpace(query)}
	lower := strings.ToLower(query)

	if strings.Contains(lower, "宽基") || strings.Contains(lower, "etf") {
		parts = append(parts, "ETF 宽基 指数基金 index etf")
	}
	if strings.Contains(lower, "网格") {
		parts = append(parts, "网格交易 grid trading")
	}
	if strings.Contains(lower, "交易") || strings.Contains(lower, "策略") {
		parts = append(parts, "策略 脚本 源码 回测 strategy script backtest")
	}
	if !strings.Contains(lower, "python") {
		parts = append(parts, "Python")
	}

	return strings.Join(dedupeStrings(parts), " ")
}

func buildGitHubSearchQuery(query string) string {
	expanded := strings.ToLower(buildKeywordExpansion(query))
	parts := []string{"python quant strategy etf"}
	if strings.Contains(expanded, "grid") || strings.Contains(expanded, "网格") {
		parts = append(parts, "\"grid trading\"")
	}
	if strings.Contains(expanded, "etf") || strings.Contains(expanded, "宽基") {
		parts = append(parts, "etf")
	}
	return strings.Join(dedupeStrings(parts), " ")
}

func dedupeStrings(values []string) []string {
	result := make([]string, 0, len(values))
	seen := map[string]bool{}
	for _, value := range values {
		clean := strings.TrimSpace(value)
		if clean == "" || seen[clean] {
			continue
		}
		seen[clean] = true
		result = append(result, clean)
	}
	return result
}

func appendDedupHits(existing []SearchHit, incoming []SearchHit, limit int) []SearchHit {
	seen := map[string]bool{}
	for _, item := range existing {
		seen[item.URL] = true
	}
	for _, item := range incoming {
		if strings.TrimSpace(item.URL) == "" || seen[item.URL] {
			continue
		}
		seen[item.URL] = true
		existing = append(existing, item)
		if limit > 0 && len(existing) >= limit {
			break
		}
	}
	return existing
}

func (s *Service) BuildSearchAgentCandidatePrompt(req ScriptSearchRequest, hits []SearchHit) string {
	lines := []string{
		"You are a quant script discovery and curation assistant.",
		"Act like a human researcher who is trying to find real script candidates worth saving into a local template library.",
		"",
		"## User Need",
		fmt.Sprintf("- Query: %s", strings.TrimSpace(req.Query)),
		fmt.Sprintf("- Preferred Platform: %s", strings.TrimSpace(req.PreferPlatform)),
		fmt.Sprintf("- Prefer Python: %t", req.RequirePython),
		"",
		"## Candidate Results",
	}

	for index, hit := range hits {
		lines = append(lines,
			fmt.Sprintf("%d. Source: %s", index+1, hit.Source),
			fmt.Sprintf("   Title: %s", hit.Title),
			fmt.Sprintf("   URL: %s", hit.URL),
			fmt.Sprintf("   Candidate Type: %s", strings.TrimSpace(hit.CandidateType)),
			fmt.Sprintf("   Search Snippet: %s", strings.TrimSpace(hit.Snippet)),
			fmt.Sprintf("   Content Snippet: %s", strings.TrimSpace(hit.ContentSnippet)),
			fmt.Sprintf("   Import Hint: %s", strings.TrimSpace(hit.ImportHint)),
			fmt.Sprintf("   Has Code: %t", hit.HasCode),
			fmt.Sprintf("   Match Score: %d", hit.MatchScore),
		)
		if strings.TrimSpace(hit.CodePreview) != "" {
			lines = append(lines, fmt.Sprintf("   Code Preview:\n```python\n%s\n```", strings.TrimSpace(hit.CodePreview)))
		}
	}

	lines = append(lines,
		"",
		"## Output Format",
		"Reply in Chinese only.",
		"Use exactly the following sections:",
		"## 检索结论",
		"Summarize which results look like real reusable scripts, and which are only docs, discussions, or weak matches.",
		"## 优先查看",
		"List 3 to 5 items in the format: 1. 来源 - 标题 - 理由",
		"## 入库建议",
		"List 1 to 3 items that are the best candidates to save into the local template library, with reasons.",
		"## 后续筛选建议",
		"Give 3 concrete next steps for validating runnability and platform adaptation.",
	)

	return strings.Join(lines, "\n")
}

func (s *Service) enrichSearchHits(query string, hits []SearchHit) []SearchHit {
	if len(hits) == 0 {
		return hits
	}

	maxEnriched := len(hits)
	if maxEnriched > 8 {
		maxEnriched = 8
	}

	for i := range hits {
		hits[i].CandidateType = classifySearchHitType(hits[i])
		hits[i].MatchScore = scoreSearchHit(hits[i], query)
		if i >= maxEnriched {
			hits[i].ImportHint = buildImportHint(hits[i])
			continue
		}

		contentSnippet, codePreview, err := s.fetchSearchHitDetails(hits[i])
		if err != nil {
			logger.SugaredLogger.Warnf("enrich search hit failed: %s %v", hits[i].URL, err)
			hits[i].ImportHint = buildImportHint(hits[i])
			continue
		}

		hits[i].ContentSnippet = contentSnippet
		hits[i].CodePreview = codePreview
		hits[i].HasCode = strings.TrimSpace(codePreview) != ""
		hits[i].ImportHint = buildImportHint(hits[i])
		hits[i].MatchScore = scoreSearchHit(hits[i], query)
	}

	return hits
}

func (s *Service) fetchSearchHitDetails(hit SearchHit) (string, string, error) {
	client := resty.New().
		SetTimeout(8*time.Second).
		SetHeader("User-Agent", "Mozilla/5.0 investment-platform quant-search-enricher")

	if rawURL := githubRawContentURL(hit.URL); rawURL != "" {
		resp, err := client.R().Get(rawURL)
		if err == nil && resp.StatusCode() < 400 {
			body := strings.TrimSpace(resp.String())
			if looksLikePythonCode(body) {
				return buildContentSnippet(body), truncateRunes(body, 900), nil
			}
		}
	}

	resp, err := client.R().Get(hit.URL)
	if err != nil {
		return "", "", err
	}
	if resp.StatusCode() >= 400 {
		return "", "", fmt.Errorf("HTTP %d", resp.StatusCode())
	}

	body := strings.TrimSpace(resp.String())
	contentType := strings.ToLower(strings.TrimSpace(resp.Header().Get("Content-Type")))
	if shouldTreatBodyAsRawCode(hit.URL, contentType) && looksLikePythonCode(body) {
		return buildContentSnippet(body), truncateRunes(body, 900), nil
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		return "", "", err
	}

	return extractMainTextSnippet(doc), extractPageCodePreview(doc), nil
}

func classifySearchHitType(hit SearchHit) string {
	urlLower := strings.ToLower(strings.TrimSpace(hit.URL))
	switch {
	case strings.Contains(urlLower, "github.com") && strings.Contains(urlLower, "/blob/"):
		return "github-code"
	case strings.Contains(urlLower, "github.com"):
		return "github-repo"
	case strings.Contains(urlLower, "/codeshare"):
		return "codeshare"
	case strings.Contains(urlLower, "/doc"):
		return "docs"
	case strings.Contains(urlLower, "/community") || strings.Contains(urlLower, "/post") || strings.Contains(urlLower, "/detail"):
		return "community-post"
	case strings.Contains(urlLower, "arxiv.org") || strings.Contains(urlLower, "ssrn.com"):
		return "research"
	default:
		return "web-page"
	}
}

func githubRawContentURL(rawURL string) string {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return ""
	}
	if !strings.EqualFold(parsed.Hostname(), "github.com") && !strings.EqualFold(parsed.Hostname(), "www.github.com") {
		return ""
	}

	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) < 5 || parts[2] != "blob" {
		return ""
	}

	return fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s",
		parts[0],
		parts[1],
		parts[3],
		strings.Join(parts[4:], "/"),
	)
}

func extractMainTextSnippet(doc *goquery.Document) string {
	selectors := []string{
		"article",
		"main",
		".markdown-body",
		".post-content",
		".post-detail",
		".content",
		".article-content",
		"body",
	}

	for _, selector := range selectors {
		text := normalizeScrapedText(doc.Find(selector).First().Text())
		if len([]rune(text)) >= 80 {
			return truncateRunes(text, 320)
		}
	}
	return ""
}

func extractPageCodePreview(doc *goquery.Document) string {
	selectors := []string{
		"pre",
		"code",
		".highlight pre",
		".blob-code-inner",
		".codehilite pre",
	}

	var blocks []string
	for _, selector := range selectors {
		doc.Find(selector).Each(func(_ int, sel *goquery.Selection) {
			if len(blocks) >= 3 {
				return
			}
			text := strings.TrimSpace(html.UnescapeString(sel.Text()))
			text = normalizeCodeWhitespace(text)
			if !looksLikePythonCode(text) {
				return
			}
			blocks = append(blocks, truncateRunes(text, 420))
		})
		if len(blocks) > 0 {
			break
		}
	}

	if len(blocks) == 0 {
		return ""
	}
	return strings.Join(blocks, "\n\n")
}

func buildContentSnippet(text string) string {
	return truncateRunes(normalizeScrapedText(text), 320)
}

func normalizeScrapedText(text string) string {
	text = html.UnescapeString(strings.TrimSpace(text))
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	text = regexp.MustCompile(`\n{2,}`).ReplaceAllString(text, "\n")
	text = regexp.MustCompile(`[ \t]+`).ReplaceAllString(text, " ")
	return strings.TrimSpace(text)
}

func normalizeCodeWhitespace(text string) string {
	text = html.UnescapeString(strings.TrimSpace(text))
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	lines := strings.Split(text, "\n")
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		cleaned := strings.TrimRight(line, " \t")
		if strings.TrimSpace(cleaned) == "" && len(result) > 0 && result[len(result)-1] == "" {
			continue
		}
		result = append(result, cleaned)
	}
	return strings.TrimSpace(strings.Join(result, "\n"))
}

func buildImportHint(hit SearchHit) string {
	switch {
	case hit.HasCode && strings.Contains(hit.CandidateType, "github"):
		return "可直接作为代码候选，优先检查依赖和平台适配后入库。"
	case hit.HasCode:
		return "页面里已经带有代码片段，适合作为脚本原型沉淀到模板库。"
	case hit.CandidateType == "codeshare":
		return "像是策略分享页，建议继续打开核对是否有完整代码。"
	case hit.CandidateType == "community-post":
		return "更像策略讨论或研究帖，适合作为灵感来源，不一定能直接入库。"
	case hit.CandidateType == "docs":
		return "更像文档或教程，适合作为平台适配参考。"
	default:
		return "先看页面正文和是否包含代码，再决定是否沉淀入库。"
	}
}

func looksLikePythonCode(text string) bool {
	lower := strings.ToLower(strings.TrimSpace(text))
	if lower == "" {
		return false
	}
	if strings.HasPrefix(lower, "<!doctype html") || strings.HasPrefix(lower, "<html") {
		return false
	}

	signals := []string{
		"import ",
		"from ",
		"def ",
		"class ",
		"if __name__ ==",
		"pd.",
		"np.",
		"return ",
		"for ",
	}
	score := 0
	for _, signal := range signals {
		if strings.Contains(lower, signal) {
			score++
		}
	}
	return score >= 2 || (strings.Count(text, "\n") >= 6 && strings.Contains(text, ":"))
}

func shouldTreatBodyAsRawCode(rawURL string, contentType string) bool {
	urlLower := strings.ToLower(strings.TrimSpace(rawURL))
	if strings.HasSuffix(urlLower, ".py") || strings.Contains(urlLower, "raw.githubusercontent.com") {
		return true
	}
	if strings.Contains(contentType, "text/plain") || strings.Contains(contentType, "application/octet-stream") {
		return true
	}
	return false
}
