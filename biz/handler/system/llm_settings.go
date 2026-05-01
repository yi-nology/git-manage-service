package system

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/yi-nology/git-manage-service/biz/model/api"
	"github.com/yi-nology/git-manage-service/biz/service/llm"
	"github.com/yi-nology/git-manage-service/pkg/configs"
	pkgresponse "github.com/yi-nology/git-manage-service/pkg/response"
)

func ListLLMProviders(ctx context.Context, c *app.RequestContext) {
	providers, err := llm.ListProviders()
	if err != nil {
		pkgresponse.InternalServerError(c, err.Error())
		return
	}
	pkgresponse.Success(c, providers)
}

func GetLLMProvider(ctx context.Context, c *app.RequestContext) {
	id, err := parseIDParam(c.Param("id"))
	if err != nil {
		pkgresponse.BadRequest(c, "invalid id")
		return
	}
	provider, err := llm.GetProviderByID(id)
	if err != nil {
		pkgresponse.NotFound(c, err.Error())
		return
	}
	pkgresponse.Success(c, provider)
}

func CreateLLMProvider(ctx context.Context, c *app.RequestContext) {
	var req api.LLMProviderDTO
	if err := c.BindAndValidate(&req); err != nil {
		pkgresponse.BadRequest(c, err.Error())
		return
	}
	if req.Name == "" || req.Type == "" || req.Model == "" {
		pkgresponse.BadRequest(c, "name, type and model are required")
		return
	}
	if req.BaseURL == "" {
		pkgresponse.BadRequest(c, "base_url is required")
		return
	}
	result, err := llm.CreateProvider(req)
	if err != nil {
		pkgresponse.InternalServerError(c, err.Error())
		return
	}
	pkgresponse.Success(c, result)
}

func UpdateLLMProvider(ctx context.Context, c *app.RequestContext) {
	id, err := parseIDParam(c.Param("id"))
	if err != nil {
		pkgresponse.BadRequest(c, "invalid id")
		return
	}
	var req api.LLMProviderDTO
	if err := c.BindAndValidate(&req); err != nil {
		pkgresponse.BadRequest(c, err.Error())
		return
	}
	result, err := llm.UpdateProvider(id, req)
	if err != nil {
		pkgresponse.InternalServerError(c, err.Error())
		return
	}
	pkgresponse.Success(c, result)
}

func DeleteLLMProvider(ctx context.Context, c *app.RequestContext) {
	id, err := parseIDParam(c.Param("id"))
	if err != nil {
		pkgresponse.BadRequest(c, "invalid id")
		return
	}
	if err := llm.DeleteProvider(id); err != nil {
		pkgresponse.InternalServerError(c, err.Error())
		return
	}
	pkgresponse.Success(c, map[string]string{"status": "deleted"})
}

func SetDefaultLLMProvider(ctx context.Context, c *app.RequestContext) {
	id, err := parseIDParam(c.Param("id"))
	if err != nil {
		pkgresponse.BadRequest(c, "invalid id")
		return
	}
	if err := llm.SetDefaultProvider(id); err != nil {
		pkgresponse.InternalServerError(c, err.Error())
		return
	}
	pkgresponse.Success(c, map[string]string{"status": "ok"})
}

func TestLLMProvider(ctx context.Context, c *app.RequestContext) {
	id, err := parseIDParam(c.Param("id"))
	if err != nil {
		pkgresponse.BadRequest(c, "invalid id")
		return
	}
	if err := llm.TestProvider(id); err != nil {
		pkgresponse.InternalServerError(c, err.Error())
		return
	}
	pkgresponse.Success(c, map[string]string{"status": "ok", "message": "连接测试成功"})
}

func GetCodeReviewSettings(ctx context.Context, c *app.RequestContext) {
	cfg := configs.GlobalConfig.CodeReview
	pkgresponse.Success(c, api.CodeReviewGlobalSettingsDTO{
		Enabled:        cfg.Enabled,
		AutoReviewOnMR: cfg.AutoReviewOnMR,
		BlockOnHigh:    cfg.BlockOnHigh,
		MaxFiles:       cfg.MaxFiles,
		MaxDiffLines:   cfg.MaxDiffLines,
	})
}

func UpdateCodeReviewSettings(ctx context.Context, c *app.RequestContext) {
	var req api.CodeReviewGlobalSettingsDTO
	if err := c.BindAndValidate(&req); err != nil {
		pkgresponse.BadRequest(c, err.Error())
		return
	}
	cfg := &configs.GlobalConfig.CodeReview
	cfg.Enabled = req.Enabled
	cfg.AutoReviewOnMR = req.AutoReviewOnMR
	cfg.BlockOnHigh = req.BlockOnHigh
	if req.MaxFiles > 0 {
		cfg.MaxFiles = req.MaxFiles
	}
	if req.MaxDiffLines > 0 {
		cfg.MaxDiffLines = req.MaxDiffLines
	}
	pkgresponse.Success(c, req)
}

func parseIDParam(s string) (uint, error) {
	id, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}
