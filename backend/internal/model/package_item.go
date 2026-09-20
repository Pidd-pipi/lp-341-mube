package model

// PackageItem 检查项目（隶属套餐）。
type PackageItem struct {
	ID             uint   `gorm:"primaryKey" json:"id"`
	PackageID      uint   `gorm:"index;not null" json:"package_id"`
	ItemName       string `gorm:"size:100;not null" json:"item_name"`
	ItemGroup      string `gorm:"size:50" json:"item_group"`
	RefValueRange  string `gorm:"size:100" json:"ref_value_range"`
	Department     string `gorm:"size:50" json:"department"`
	SortOrder      int    `json:"sort_order"`
}
