package models

type SystemPolicy struct {
	ID                       uint `gorm:"primary_key" json:"id"`
	MaxFileSizeMB            int  `gorm:"default:50" json:"maxFileSizeMB"`
	MinValidityHours         int  `gorm:"default:1" json:"minValidityHours"`
	MaxValidityDays          int  `gorm:"default:30" json:"maxValidityDays"`
	DefaultValidityDays      int  `gorm:"default:7" json:"defaultValidityDays"`
	RequirePasswordMinLength int  `gorm:"default:6" json:"requirePasswordMinLength"`
}

func (SystemPolicy) TableName() string {
	return "system_policy"
}

