package db

import (
	"github.com/yi-nology/git-manage-service/biz/model/po"
)

type ReviewCLIConfigDAO struct{}

func NewReviewCLIConfigDAO() *ReviewCLIConfigDAO { return &ReviewCLIConfigDAO{} }

func (d *ReviewCLIConfigDAO) Create(cfg *po.ReviewCLIConfig) error {
	return DB.Create(cfg).Error
}

func (d *ReviewCLIConfigDAO) FindAll() ([]po.ReviewCLIConfig, error) {
	var configs []po.ReviewCLIConfig
	err := DB.Find(&configs).Error
	return configs, err
}

func (d *ReviewCLIConfigDAO) FindByID(id uint) (*po.ReviewCLIConfig, error) {
	var cfg po.ReviewCLIConfig
	err := DB.First(&cfg, id).Error
	return &cfg, err
}

func (d *ReviewCLIConfigDAO) FindActive() ([]po.ReviewCLIConfig, error) {
	var configs []po.ReviewCLIConfig
	err := DB.Where("is_active = ?", true).Find(&configs).Error
	return configs, err
}

func (d *ReviewCLIConfigDAO) Save(cfg *po.ReviewCLIConfig) error {
	return DB.Save(cfg).Error
}

func (d *ReviewCLIConfigDAO) Delete(cfg *po.ReviewCLIConfig) error {
	return DB.Delete(cfg).Error
}
