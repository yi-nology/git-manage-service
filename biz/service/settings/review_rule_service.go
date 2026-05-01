package settings

import (
	"fmt"
	"log"
	"time"

	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/po"
)

type ReviewRuleDTO struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
	Category    string `json:"category"`
	Enabled     bool   `json:"enabled"`
	SortOrder   int    `json:"sort_order"`
}

func ListReviewRules() ([]ReviewRuleDTO, error) {
	rules, err := db.NewReviewRuleDAO().FindAll()
	if err != nil {
		return nil, err
	}
	dtos := make([]ReviewRuleDTO, 0, len(rules))
	for _, r := range rules {
		dtos = append(dtos, ruleToDTO(r))
	}
	return dtos, nil
}

func GetReviewRule(id string) (*ReviewRuleDTO, error) {
	rule, err := db.NewReviewRuleDAO().FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("rule not found: %w", err)
	}
	dto := ruleToDTO(*rule)
	return &dto, nil
}

func CreateReviewRule(dto ReviewRuleDTO) (*ReviewRuleDTO, error) {
	if dto.ID == "" || dto.Name == "" {
		return nil, fmt.Errorf("id and name are required")
	}
	rule := &po.ReviewRule{
		ID:          dto.ID,
		Name:        dto.Name,
		Description: dto.Description,
		Severity:    dto.Severity,
		Category:    dto.Category,
		Enabled:     dto.Enabled,
		SortOrder:   dto.SortOrder,
	}
	if rule.Severity == "" {
		rule.Severity = "medium"
	}
	if err := db.NewReviewRuleDAO().Create(rule); err != nil {
		return nil, fmt.Errorf("failed to create rule: %w", err)
	}
	result := ruleToDTO(*rule)
	return &result, nil
}

func UpdateReviewRule(id string, dto ReviewRuleDTO) (*ReviewRuleDTO, error) {
	dao := db.NewReviewRuleDAO()
	rule, err := dao.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("rule not found: %w", err)
	}
	if dto.Name != "" {
		rule.Name = dto.Name
	}
	if dto.Description != "" {
		rule.Description = dto.Description
	}
	if dto.Severity != "" {
		rule.Severity = dto.Severity
	}
	if dto.Category != "" {
		rule.Category = dto.Category
	}
	rule.Enabled = dto.Enabled
	if dto.SortOrder > 0 {
		rule.SortOrder = dto.SortOrder
	}
	if err := dao.Save(rule); err != nil {
		return nil, fmt.Errorf("failed to save rule: %w", err)
	}
	result := ruleToDTO(*rule)
	return &result, nil
}

func DeleteReviewRule(id string) error {
	return db.NewReviewRuleDAO().Delete(id)
}

func BatchUpdateReviewRules(dtos []ReviewRuleDTO) error {
	dao := db.NewReviewRuleDAO()
	for _, dto := range dtos {
		rule, err := dao.FindByID(dto.ID)
		if err != nil {
			continue
		}
		rule.Enabled = dto.Enabled
		if dto.SortOrder > 0 {
			rule.SortOrder = dto.SortOrder
		}
		if err := dao.Save(rule); err != nil {
			return fmt.Errorf("failed to update rule %s: %w", dto.ID, err)
		}
	}
	return nil
}

func InitDefaultReviewRules() {
	dao := db.NewReviewRuleDAO()
	count, err := dao.Count()
	if err != nil {
		log.Printf("Warning: failed to count review rules: %v", err)
		return
	}
	if count > 0 {
		return
	}

	rules := getDefaultReviewRules()
	now := time.Now()
	for i := range rules {
		rules[i].CreatedAt = now
		rules[i].UpdatedAt = now
	}

	for _, r := range rules {
		if err := dao.Create(&r); err != nil {
			log.Printf("Warning: failed to insert default review rule %s: %v", r.ID, err)
		}
	}
}

func getDefaultReviewRules() []po.ReviewRule {
	return []po.ReviewRule{
		{ID: "secret", Name: "密钥泄露检测", Description: "检测硬编码密码、API Key、AWS Key、私钥、连接字符串等", Severity: "critical", Category: "security", Enabled: true, SortOrder: 1},
		{ID: "protected_file", Name: "保护文件检测", Description: "检测对 .env、CI 配置、Dockerfile、证书等关键文件的修改", Severity: "high", Category: "security", Enabled: true, SortOrder: 2},
		{ID: "diff_size", Name: "变更规模检测", Description: "检测超大 MR（>50文件、>500行/文件、>3000总行数）", Severity: "medium", Category: "quality", Enabled: true, SortOrder: 3},
		{ID: "migration", Name: "数据库迁移检测", Description: "检测破坏性 SQL（DROP）、缺少 FK 索引等问题", Severity: "high", Category: "database", Enabled: true, SortOrder: 4},
		{ID: "test_required", Name: "测试覆盖检测", Description: "检测源码变更是否缺少对应的测试文件", Severity: "low", Category: "quality", Enabled: true, SortOrder: 5},
	}
}

func ruleToDTO(r po.ReviewRule) ReviewRuleDTO {
	return ReviewRuleDTO{
		ID: r.ID, Name: r.Name, Description: r.Description,
		Severity: r.Severity, Category: r.Category,
		Enabled: r.Enabled, SortOrder: r.SortOrder,
	}
}
