// biz/dal/db/notification_channel_dao.go - 通知渠道DAO

package db

import (
	"github.com/yi-nology/git-manage-service/biz/model/po"
)

type NotificationChannelDAO struct {
	BaseDAO[po.NotificationChannel]
}

func NewNotificationChannelDAO() *NotificationChannelDAO {
	return &NotificationChannelDAO{}
}

// FindByType 按类型查询
func (d *NotificationChannelDAO) FindByType(channelType string) ([]po.NotificationChannel, error) {
	var channels []po.NotificationChannel
	err := DB.Where("type = ?", channelType).Find(&channels).Error
	return channels, err
}

// FindEnabled 查询已启用的渠道
func (d *NotificationChannelDAO) FindEnabled() ([]po.NotificationChannel, error) {
	var channels []po.NotificationChannel
	err := DB.Where("enabled = ?", true).Find(&channels).Error
	return channels, err
}

// Delete 覆盖基类：参数为对象指针（非 ID）
func (d *NotificationChannelDAO) Delete(channel *po.NotificationChannel) error {
	return DB.Delete(channel).Error
}
