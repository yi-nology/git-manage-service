package settings

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/yi-nology/git-manage-service/biz/service/llm"
	pkgresponse "github.com/yi-nology/git-manage-service/pkg/response"
)

func ListLLMPresets(ctx context.Context, c *app.RequestContext) {
	category := c.Query("category")
	if category != "" {
		pkgresponse.Success(c, llm.GetPresetsByCategory(category))
		return
	}
	pkgresponse.Success(c, llm.GetPresets())
}
