package query

import (
	"github.com/yi-nology/git-manage-service/biz/dal"
	"github.com/yi-nology/git-manage-service/biz/model"
)

type SystemConfigDAO struct{}

func NewSystemConfigDAO() *SystemConfigDAO {
	return &SystemConfigDAO{}
}

func (dao *SystemConfigDAO) GetConfig(key string) (string, error) {
	var config model.SystemConfig
	err := dal.DB.Where("key = ?", key).First(&config).Error
	if err != nil {
		return "", err
	}
	return config.Value, nil
}

func (dao *SystemConfigDAO) SetConfig(key, value string) error {
	config := model.SystemConfig{
		Key:   key,
		Value: value,
	}
	return dal.DB.Save(&config).Error
}

func (dao *SystemConfigDAO) GetAll() (map[string]string, error) {
	var configs []model.SystemConfig
	err := dal.DB.Find(&configs).Error
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, c := range configs {
		result[c.Key] = c.Value
	}
	return result, nil
}
