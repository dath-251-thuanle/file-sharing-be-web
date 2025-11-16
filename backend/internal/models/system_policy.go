package models

type SystemPolicy struct {
	ID                       uint `gorm:"primary_key" json:"id"`
	MaxFileSizeMB            int  `gorm:"default:50" json:"max_file_size_mb"`
	MinValidityHours         int  `gorm:"default:1" json:"min_validity_hours"`
	MaxValidityDays          int  `gorm:"default:30" json:"max_validity_days"`
	DefaultValidityDays      int  `gorm:"default:7" json:"default_validity_days"`
	RequirePasswordMinLength int  `gorm:"default:6" json:"require_password_min_length"`
}

func (SystemPolicy) TableName() string {
	return "system_policy"
}

