package po

type SystemConfig struct {
	Key   string `gorm:"primaryKey" json:"key"`
	Value string `json:"value"`
}

func (SystemConfig) TableName() string {
	return "system_configs"
}
