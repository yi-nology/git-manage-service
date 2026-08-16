package db

import (
	"github.com/yi-nology/git-manage-service/biz/model/po"
)

type SSHKeyDAO struct{ BaseDAO[po.SSHKey] }

func NewSSHKeyDAO() *SSHKeyDAO { return &SSHKeyDAO{} }

// Update 更新 SSH 密钥（与 Save 相同，保持向后兼容）
func (d *SSHKeyDAO) Update(sshKey *po.SSHKey) error {
	return DB.Save(sshKey).Error
}

// FindByName 根据名称查询SSH密钥
func (d *SSHKeyDAO) FindByName(name string) (*po.SSHKey, error) {
	var key po.SSHKey
	err := DB.Where("name = ?", name).First(&key).Error
	return &key, err
}
