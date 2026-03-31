package quant

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

type TemplateCategory struct {
	gorm.Model
	Name        string                `json:"name" gorm:"uniqueIndex"`
	Description string                `json:"description"`
	SortOrder   int                   `json:"sortOrder"`
	IsDel       soft_delete.DeletedAt `gorm:"softDelete:flag" json:"-"`
}

func (t TemplateCategory) TableName() string {
	return "quant_template_categories"
}

type Template struct {
	gorm.Model
	Name            string                `json:"name" gorm:"index"`
	CategoryID      uint                  `json:"categoryId" gorm:"index"`
	Category        TemplateCategory      `json:"category" gorm:"foreignKey:CategoryID"`
	Description     string                `json:"description"`
	Language        string                `json:"language" gorm:"default:python"`
	Code            string                `json:"code" gorm:"type:text"`
	BrokerPlatform  string                `json:"brokerPlatform"`
	StrategyType    string                `json:"strategyType"`
	ScriptCategory  string                `json:"scriptCategory"`
	StyleTags       string                `json:"styleTags" gorm:"type:text"`
	EmotionTags     string                `json:"emotionTags" gorm:"type:text"`
	VolumeTags      string                `json:"volumeTags" gorm:"type:text"`
	ScenarioTags    string                `json:"scenarioTags" gorm:"type:text"`
	CapitalTags     string                `json:"capitalTags" gorm:"type:text"`
	FactorTags      string                `json:"factorTags" gorm:"type:text"`
	SearchKeywords  string                `json:"searchKeywords" gorm:"type:text"`
	SourcePlatforms string                `json:"sourcePlatforms" gorm:"type:text"`
	IsAIGenerated   bool                  `json:"isAiGenerated"`
	AIPrompt        string                `json:"aiPrompt"`
	AIModel         string                `json:"aiModel"`
	LinkedStocks    string                `json:"linkedStocks"`
	Parameters      string                `json:"parameters" gorm:"type:text"`
	BacktestResult  string                `json:"backtestResult" gorm:"type:text"`
	Version         int                   `json:"version" gorm:"default:1"`
	Status          string                `json:"status" gorm:"default:draft"`
	LastUsedAt      *time.Time            `json:"lastUsedAt"`
	IsDel           soft_delete.DeletedAt `gorm:"softDelete:flag" json:"-"`
}

func (t Template) TableName() string {
	return "quant_templates"
}

type GenerateRequest struct {
	StrategyDescription string  `json:"strategyDescription"`
	BrokerPlatform      string  `json:"brokerPlatform"`
	StrategyType        string  `json:"strategyType"`
	ScriptCategory      string  `json:"scriptCategory"`
	StockCodes          string  `json:"stockCodes"`
	RiskLevel           string  `json:"riskLevel"`
	Capital             float64 `json:"capital"`
	FactorTags          string  `json:"factorTags"`
	SceneTags           string  `json:"sceneTags"`
	AIModel             string  `json:"aiModel"`
	PromptTemplateID    int     `json:"promptTemplateId"`
	BaseCode            string  `json:"baseCode"`
	ExistingScriptName  string  `json:"existingScriptName"`
	ExistingDescription string  `json:"existingDescription"`
}

type GenerateResult struct {
	Code          string `json:"code"`
	Explanation   string `json:"explanation"`
	RiskWarning   string `json:"riskWarning"`
	SuggestedName string `json:"suggestedName"`
}

type TagOption struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type TagGroup struct {
	Key         string      `json:"key"`
	Label       string      `json:"label"`
	Description string      `json:"description"`
	Options     []TagOption `json:"options"`
}

type SearchLink struct {
	Name        string `json:"name"`
	URL         string `json:"url"`
	Description string `json:"description"`
}

type SearchHit struct {
	Source         string `json:"source"`
	Title          string `json:"title"`
	URL            string `json:"url"`
	Snippet        string `json:"snippet"`
	ContentSnippet string `json:"contentSnippet"`
	CodePreview    string `json:"codePreview"`
	CandidateType  string `json:"candidateType"`
	ImportHint     string `json:"importHint"`
	HasCode        bool   `json:"hasCode"`
	MatchScore     int    `json:"matchScore"`
}

type ScriptSearchRequest struct {
	Query          string   `json:"query"`
	Sources        []string `json:"sources"`
	ResultLimit    int      `json:"resultLimit"`
	RequirePython  bool     `json:"requirePython"`
	PreferPlatform string   `json:"preferPlatform"`
}

type LinkageAIRequest struct {
	Summary   string     `json:"summary"`
	Templates []Template `json:"templates"`
}
