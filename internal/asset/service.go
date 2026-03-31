package asset

import (
	"go-stock/backend/db"
	"go-stock/backend/logger"
)

// Service 资产分析服务
type Service struct{}

// NewService 创建资产分析服务
func NewService() *Service {
	return &Service{}
}

// --- 资产类别 CRUD ---

// GetCategories 获取所有资产类别
func (s *Service) GetCategories() []AssetCategory {
	var categories []AssetCategory
	db.Dao.Order("sort_order asc").Find(&categories)
	return categories
}

// CreateCategory 创建资产类别
func (s *Service) CreateCategory(category AssetCategory) *AssetCategory {
	err := db.Dao.Create(&category).Error
	if err != nil {
		logger.SugaredLogger.Errorf("创建资产类别失败: %v", err)
		return nil
	}
	return &category
}

// --- 资产 CRUD ---

// GetAssets 获取所有资产
func (s *Service) GetAssets() []Asset {
	var assets []Asset
	db.Dao.Preload("Category").Find(&assets)
	return assets
}

// GetAssetsByType 按类型获取资产
func (s *Service) GetAssetsByType(assetType string) []Asset {
	var assets []Asset
	db.Dao.Preload("Category").Where("type = ?", assetType).Find(&assets)
	return assets
}

// CreateAsset 创建资产
func (s *Service) CreateAsset(a Asset) *Asset {
	err := db.Dao.Create(&a).Error
	if err != nil {
		logger.SugaredLogger.Errorf("创建资产失败: %v", err)
		return nil
	}
	return &a
}

// UpdateAsset 更新资产
func (s *Service) UpdateAsset(a Asset) *Asset {
	err := db.Dao.Save(&a).Error
	if err != nil {
		logger.SugaredLogger.Errorf("更新资产失败: %v", err)
		return nil
	}
	return &a
}

// DeleteAsset 删除资产
func (s *Service) DeleteAsset(id uint) bool {
	err := db.Dao.Delete(&Asset{}, id).Error
	if err != nil {
		logger.SugaredLogger.Errorf("删除资产失败: %v", err)
		return false
	}
	return true
}

// --- 汇总 ---

// GetAssetSummary 获取资产汇总
func (s *Service) GetAssetSummary() *AssetSummary {
	summary := &AssetSummary{}

	var assets []Asset
	db.Dao.Find(&assets)

	categoryMap := make(map[uint]*CategorySummary)

	for _, a := range assets {
		switch a.Type {
		case "liquid":
			summary.TotalLiquid += a.Amount
		case "fixed":
			summary.TotalFixed += a.Amount
		case "liability":
			summary.TotalLiability += a.Amount
		}

		if cs, ok := categoryMap[a.CategoryID]; ok {
			cs.TotalAmount += a.Amount
			cs.Count++
		} else {
			categoryMap[a.CategoryID] = &CategorySummary{
				CategoryID:  a.CategoryID,
				Type:        a.Type,
				TotalAmount: a.Amount,
				Count:       1,
			}
		}
	}

	summary.NetAsset = summary.TotalLiquid + summary.TotalFixed + summary.TotalLiability

	for _, cs := range categoryMap {
		summary.Categories = append(summary.Categories, *cs)
	}

	return summary
}

// InitDefaultCategories 初始化默认资产类别
func (s *Service) InitDefaultCategories() {
	defaults := []AssetCategory{
		{Name: "银行存款", Type: "liquid", Icon: "💰", SortOrder: 1, Description: "活期/定期存款"},
		{Name: "货币基金", Type: "liquid", Icon: "📈", SortOrder: 2, Description: "余额宝等货币基金"},
		{Name: "现金", Type: "liquid", Icon: "💵", SortOrder: 3, Description: "现金资产"},
		{Name: "投资账户", Type: "liquid", Icon: "📊", SortOrder: 4, Description: "股票/基金投资"},
		{Name: "房产", Type: "fixed", Icon: "🏠", SortOrder: 10, Description: "不动产"},
		{Name: "车辆", Type: "fixed", Icon: "🚗", SortOrder: 11, Description: "车辆资产"},
		{Name: "其他固定资产", Type: "fixed", Icon: "📦", SortOrder: 12, Description: "其他固定资产"},
		{Name: "房贷", Type: "liability", Icon: "🏦", SortOrder: 20, Description: "房屋贷款"},
		{Name: "车贷", Type: "liability", Icon: "🏧", SortOrder: 21, Description: "车辆贷款"},
		{Name: "信用卡", Type: "liability", Icon: "💳", SortOrder: 22, Description: "信用卡欠款"},
		{Name: "其他借贷", Type: "liability", Icon: "📝", SortOrder: 23, Description: "其他借贷"},
	}

	for _, d := range defaults {
		var count int64
		db.Dao.Model(&AssetCategory{}).Where("name = ?", d.Name).Count(&count)
		if count == 0 {
			db.Dao.Create(&d)
		}
	}
}
