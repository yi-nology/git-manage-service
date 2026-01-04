package db

import (
	"github.com/yi-nology/git-manage-service/biz/model/po"
)

type SystemConfigDAO struct{}

func NewSystemConfigDAO() *SystemConfigDAO {
	return &SystemConfigDAO{}
}

func (dao *SystemConfigDAO) GetConfig(key string) (string, error) {
	var config po.SystemConfig
	err := DB.Where("key = ?", key).First(&config).Error
	if err != nil {
		return "", err
	}
	return config.Value, nil
}

func (dao *SystemConfigDAO) SetConfig(key, value string) error {
	config := po.SystemConfig{
		Key:   key,
		Value: value,
	}
	return DB.Save(&config).Error
}

func (dao *SystemConfigDAO) GetAll() (map[string]string, error) {
	var configs []po.SystemConfig
	err := DB.Find(&configs).Error
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, c := range configs {
		result[c.Key] = c.Value
	}
	return result, nil
}
