package db

import (
	"fmt"

	"gorm.io/gorm"
)

// BaseDAO 提供通用 CRUD 方法，子 DAO 通过嵌入继承并可覆盖。
//
// 用法:
//
//	type CredentialDAO struct{ BaseDAO[po.Credential] }
//	func NewCredentialDAO() *CredentialDAO { return &CredentialDAO{} }
//	// 仅添加自定义方法，如 FindByName / FindMatchingURL
type BaseDAO[T any] struct{}

func (d *BaseDAO[T]) db() *gorm.DB { return DB }

func (d *BaseDAO[T]) Create(obj *T) error {
	return d.db().Create(obj).Error
}

func (d *BaseDAO[T]) Save(obj *T) error {
	return d.db().Save(obj).Error
}

func (d *BaseDAO[T]) FindByID(id uint) (*T, error) {
	var obj T
	if err := d.db().First(&obj, id).Error; err != nil {
		return nil, err
	}
	return &obj, nil
}

func (d *BaseDAO[T]) FindAll() ([]T, error) {
	var objs []T
	return objs, d.db().Find(&objs).Error
}

func (d *BaseDAO[T]) Delete(id uint) error {
	return d.db().Delete(new(T), id).Error
}

func (d *BaseDAO[T]) Count() (int64, error) {
	var count int64
	return count, d.db().Model(new(T)).Count(&count).Error
}

func (d *BaseDAO[T]) BatchCreate(items []T) error {
	return d.db().Create(&items).Error
}

// ExistsByField 检查指定字段是否存在匹配记录。
// 例: d.ExistsByField("name", "my-cred")
func (d *BaseDAO[T]) ExistsByField(field string, val any) (bool, error) {
	var count int64
	err := d.db().Model(new(T)).Where(fmt.Sprintf("%s = ?", field), val).Count(&count).Error
	return count > 0, err
}

// ExistsByFieldExcludeID 检查指定字段是否存在匹配记录（排除指定 ID）。
// 用于名称唯一性校验（更新时排除自身）。
func (d *BaseDAO[T]) ExistsByFieldExcludeID(field string, val any, excludeID uint) (bool, error) {
	var count int64
	err := d.db().Model(new(T)).Where(fmt.Sprintf("%s = ? AND id != ?", field), val, excludeID).Count(&count).Error
	return count > 0, err
}

// ExistsByName 是 ExistsByField("name", name) 的快捷方式。
func (d *BaseDAO[T]) ExistsByName(name string) (bool, error) {
	return d.ExistsByField("name", name)
}

// ExistsByNameExcludeID 是 ExistsByFieldExcludeID("name", name, excludeID) 的快捷方式。
func (d *BaseDAO[T]) ExistsByNameExcludeID(name string, excludeID uint) (bool, error) {
	return d.ExistsByFieldExcludeID("name", name, excludeID)
}
