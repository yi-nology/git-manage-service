package handler

import (
	"context"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/yi-nology/git-manage-service/pkg/configs"
	"github.com/yi-nology/git-manage-service/biz/service"
)

type ConfigReq struct {
	DebugMode   bool   `json:"debug_mode"`
	AuthorName  string `json:"author_name"`
	AuthorEmail string `json:"author_email"`
}

// @Summary Get global configuration
// @Tags Config
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/config [get]
func GetConfig(ctx context.Context, c *app.RequestContext) {
	gitSvc := service.NewGitService()
	// Get global git config
	name, _ := gitSvc.RunCommand(".", "config", "--global", "user.name")
	email, _ := gitSvc.RunCommand(".", "config", "--global", "user.email")

	c.JSON(http.StatusOK, map[string]interface{}{
		"debug_mode":   configs.DebugMode,
		"author_name":  name,
		"author_email": email,
	})
}

// @Summary Update global configuration
// @Description Update global system configuration settings.
// @Tags Config
// @Accept json
// @Produce json
// @Param request body ConfigReq true "Config info"
// @Success 200 {object} map[string]interface{} "Updated config"
// @Failure 400 {object} map[string]string "Bad Request"
// @Router /api/config [post]
func UpdateConfig(ctx context.Context, c *app.RequestContext) {
	var req ConfigReq
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	configs.DebugMode = req.DebugMode

	// Update global git config
	gitSvc := service.NewGitService()
	if err := gitSvc.SetGlobalGitUser(req.AuthorName, req.AuthorEmail); err != nil {
		c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to set git config: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"debug_mode":   configs.DebugMode,
		"author_name":  req.AuthorName,
		"author_email": req.AuthorEmail,
	})
}
