package db

import (
	"github.com/yi-nology/git-manage-service/biz/model/po"
)

type SSHKeyDAO struct{ BaseDAO[po.SSHKey] }

func NewSSHKeyDAO() *SSHKeyDAO { return &SSHKeyDAO{} }

// FindByName 根据名称查询SSH密钥
func (d *SSHKeyDAO) FindByName(name string) (*po.SSHKey, error) {
	var key po.SSHKey
	err := DB.Where("name = ?", name).First(&key).Error
	return &key, err
}
