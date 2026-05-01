package settings

import (
	"encoding/json"
	"log"

	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/api"
	"github.com/yi-nology/git-manage-service/pkg/configs"
)

const codeReviewSettingsKey = "code_review_settings"

func LoadCodeReviewSettingsFromDB() {
	dao := db.NewSystemConfigDAO()
	value, err := dao.GetConfig(codeReviewSettingsKey)
	if err != nil {
		log.Printf("[Settings] No saved code review settings in DB, using config.yaml defaults")
		return
	}
	var dto api.CodeReviewGlobalSettingsDTO
	if err := json.Unmarshal([]byte(value), &dto); err != nil {
		log.Printf("[Settings] Failed to parse saved code review settings: %v", err)
		return
	}
	cfg := &configs.GlobalConfig.CodeReview
	cfg.Enabled = dto.Enabled
	cfg.AutoReviewOnMR = dto.AutoReviewOnMR
	cfg.BlockOnHigh = dto.BlockOnHigh
	if dto.MaxFiles > 0 {
		cfg.MaxFiles = dto.MaxFiles
	}
	if dto.MaxDiffLines > 0 {
		cfg.MaxDiffLines = dto.MaxDiffLines
	}
	log.Printf("[Settings] Loaded code review settings from DB")
}

func SaveCodeReviewSettingsToDB(dto api.CodeReviewGlobalSettingsDTO) error {
	cfg := &configs.GlobalConfig.CodeReview
	cfg.Enabled = dto.Enabled
	cfg.AutoReviewOnMR = dto.AutoReviewOnMR
	cfg.BlockOnHigh = dto.BlockOnHigh
	if dto.MaxFiles > 0 {
		cfg.MaxFiles = dto.MaxFiles
	}
	if dto.MaxDiffLines > 0 {
		cfg.MaxDiffLines = dto.MaxDiffLines
	}
	data, err := json.Marshal(dto)
	if err != nil {
		return err
	}
	return db.NewSystemConfigDAO().SetConfig(codeReviewSettingsKey, string(data))
}
