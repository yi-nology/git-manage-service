package spec

import (
	"context"
	"encoding/json"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/service/spec"
	"github.com/yi-nology/git-manage-service/pkg/response"
)

type SpecConfigResponse struct {
	DefaultTemplate string               `json:"defaultTemplate"`
	FormatOptions   *FormatOptionsConfig `json:"formatOptions"`
	AIConfig        *AIConfigDTO         `json:"aiConfig"`
}

type FormatOptionsConfig struct {
	Curlify         bool `json:"curlify"`
	RemoveClean     bool `json:"removeClean"`
	RemoveBuildRoot bool `json:"removeBuildRoot"`
	RemoveGroup     bool `json:"removeGroup"`
	LicenseSPDX     bool `json:"licenseSpdx"`
	SortDeps        bool `json:"sortDeps"`
	TabToSpaces     bool `json:"tabToSpaces"`
	IndentSize      int  `json:"indentSize"`
	PreambleOrder   bool `json:"preambleOrder"`
	AlignValues     bool `json:"alignValues"`
	PathMacros      bool `json:"pathMacros"`
	UtilMacros      bool `json:"utilMacros"`
	CommonCleanup   bool `json:"commonCleanup"`
	ConditionalTrim bool `json:"conditionalTrim"`
}

type AIConfigDTO struct {
	DefaultAction string `json:"defaultAction"`
	SystemPrompt  string `json:"systemPrompt"`
	AutoFix       bool   `json:"autoFix"`
}

type SaveSpecConfigReq struct {
	DefaultTemplate *string              `json:"defaultTemplate"`
	FormatOptions   *FormatOptionsConfig `json:"formatOptions"`
	AIConfig        *AIConfigDTO         `json:"aiConfig"`
}

func defaultFormatOptions() *FormatOptionsConfig {
	return &FormatOptionsConfig{
		Curlify:         true,
		RemoveClean:     true,
		RemoveBuildRoot: true,
		RemoveGroup:     false,
		LicenseSPDX:     true,
		SortDeps:        true,
		TabToSpaces:     true,
		IndentSize:      4,
		PreambleOrder:   true,
		AlignValues:     true,
		PathMacros:      true,
		UtilMacros:      true,
		CommonCleanup:   true,
		ConditionalTrim: true,
	}
}

func defaultAIConfig() *AIConfigDTO {
	return &AIConfigDTO{
		DefaultAction: "chat",
		SystemPrompt:  "",
		AutoFix:       false,
	}
}

func GetSpecConfig(ctx context.Context, c *app.RequestContext) {
	dao := db.NewSystemConfigDAO()
	result := SpecConfigResponse{
		FormatOptions: defaultFormatOptions(),
		AIConfig:      defaultAIConfig(),
	}

	if val, err := dao.GetConfig("spec_default_template"); err == nil && val != "" {
		result.DefaultTemplate = val
	} else {
		result.DefaultTemplate = spec.NewSpecService().GetSpecTemplate()
	}

	if val, err := dao.GetConfig("spec_format_options"); err == nil && val != "" {
		var opts FormatOptionsConfig
		if json.Unmarshal([]byte(val), &opts) == nil {
			result.FormatOptions = &opts
		}
	}

	if val, err := dao.GetConfig("spec_ai_config"); err == nil && val != "" {
		var aiCfg AIConfigDTO
		if json.Unmarshal([]byte(val), &aiCfg) == nil {
			result.AIConfig = &aiCfg
		}
	}

	response.Success(c, result)
}

func SaveSpecConfig(ctx context.Context, c *app.RequestContext) {
	var req SaveSpecConfigReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	dao := db.NewSystemConfigDAO()

	if req.DefaultTemplate != nil {
		if err := dao.SetConfig("spec_default_template", *req.DefaultTemplate); err != nil {
			response.InternalError(c, err)
			return
		}
	}

	if req.FormatOptions != nil {
		data, err := json.Marshal(req.FormatOptions)
		if err != nil {
			response.InternalError(c, err)
			return
		}
		if err := dao.SetConfig("spec_format_options", string(data)); err != nil {
			response.InternalError(c, err)
			return
		}
	}

	if req.AIConfig != nil {
		data, err := json.Marshal(req.AIConfig)
		if err != nil {
			response.InternalError(c, err)
			return
		}
		if err := dao.SetConfig("spec_ai_config", string(data)); err != nil {
			response.InternalError(c, err)
			return
		}
	}

	c.Set("audit_details", map[string]interface{}{"template": req.DefaultTemplate, "format": req.FormatOptions, "ai": req.AIConfig})
	response.Success(c, map[string]string{"message": "配置已保存"})
}
