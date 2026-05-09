package spec

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/api"
	"github.com/yi-nology/git-manage-service/biz/model/po"
	spec "github.com/yi-nology/git-manage-service/biz/model/spec"
	gitSvc "github.com/yi-nology/git-manage-service/biz/service/git"
	lintSvc "github.com/yi-nology/git-manage-service/biz/service/lint"
	"github.com/yi-nology/git-manage-service/biz/service/llm"
	specService "github.com/yi-nology/git-manage-service/biz/service/spec"
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

// GetSpecTree .
// @router /api/v1/spec/tree [GET]
func GetSpecTree(ctx context.Context, c *app.RequestContext) {
	repoKey := c.Query("repo_key")
	if repoKey == "" {
		response.BadRequest(c, "repo_key is required")
		return
	}

	repo, err := db.NewRepoDAO().FindByKey(repoKey)
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	tree, err := buildSpecTree(repo.Path)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.Success(c, tree)
}

// ListSpecFiles .
// @router /api/v1/spec/list [GET]
func ListSpecFiles(ctx context.Context, c *app.RequestContext) {
	repoKey := c.Query("repo_key")
	if repoKey == "" {
		response.BadRequest(c, "repo_key is required")
		return
	}

	repo, err := db.NewRepoDAO().FindByKey(repoKey)
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	svc := specService.NewSpecService()
	files, err := svc.ListSpecFiles(repo.Path)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	if files == nil {
		files = []specService.SpecFileInfo{}
	}

	var dtos []api.SpecFileInfo
	for _, f := range files {
		dtos = append(dtos, api.SpecFileInfo{
			Name:    f.Name,
			Path:    f.Path,
			IsDir:   f.IsDir,
			Size:    f.Size,
			ModTime: f.ModTime,
		})
	}

	response.Success(c, dtos)
}

// GetSpecContent .
// @router /api/v1/spec/content [GET]
func GetSpecContent(ctx context.Context, c *app.RequestContext) {
	repoKey := c.Query("repo_key")
	path := c.Query("path")
	if repoKey == "" || path == "" {
		response.BadRequest(c, "repo_key and path are required")
		return
	}

	repo, err := db.NewRepoDAO().FindByKey(repoKey)
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	svc := specService.NewSpecService()
	content, err := svc.GetSpecContent(repo.Path, path)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.Success(c, api.SpecContentResponse{
		Content: content,
		Path:    path,
	})
}

// GetSpecContentByPath .
// @router /api/v1/spec/content/:path [GET]
func GetSpecContentByPath(ctx context.Context, c *app.RequestContext) {
	repoKey := c.Query("repo_key")
	path := c.Param("path")
	if repoKey == "" {
		response.BadRequest(c, "repo_key is required")
		return
	}
	if path == "" {
		path = c.Query("path")
	}

	repo, err := db.NewRepoDAO().FindByKey(repoKey)
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	svc := specService.NewSpecService()
	content, err := svc.GetSpecContent(repo.Path, path)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.Success(c, api.FileContent{
		Path:    path,
		Content: content,
	})
}

// SaveSpecContent .
// @router /api/v1/spec/save [POST]
func SaveSpecContent(ctx context.Context, c *app.RequestContext) {
	var req api.SaveSpecReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	repo, err := db.NewRepoDAO().FindByKey(req.RepoKey)
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	lintService := lintSvc.NewLintService()
	validationResult, err := lintService.Lint(req.Content, nil)
	if err == nil && validationResult != nil {
		for _, issue := range validationResult.Issues {
			if issue.Severity == "error" {
				response.BadRequest(c, "Spec validation failed: "+issue.Message)
				return
			}
		}
	}

	svc := specService.NewSpecService()

	err = svc.SaveSpecContent(repo.Path, req.Path, req.Content, req.CommitMessage)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	if req.CommitMessage != "" {
		gitService := gitSvc.NewGitService()
		if err := gitService.AddAndCommit(repo.Path, req.Path, req.CommitMessage); err != nil {
			response.InternalError(c, err)
			return
		}
	}

	c.Set("audit_target", "repo:"+repo.Key)
	c.Set("audit_details", map[string]string{"path": req.Path, "commit_message": req.CommitMessage})
	response.Success(c, api.SaveWithValidationResponse{
		Message:          "spec saved successfully",
		ValidationResult: validationResult,
	})
}

// SaveSpecContentByPath .
// @router /api/v1/spec/content/:path [PUT]
func SaveSpecContentByPath(ctx context.Context, c *app.RequestContext) {
	path := c.Param("path")
	if path == "" {
		response.BadRequest(c, "path is required")
		return
	}

	var req api.SaveSpecContentReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.Path == "" {
		req.Path = path
	}

	repo, err := db.NewRepoDAO().FindByKey(req.RepoKey)
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	svc := specService.NewSpecService()

	err = svc.SaveSpecContent(repo.Path, req.Path, req.Content, req.Message)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	if req.AutoCommit && req.Message != "" {
		gitService := gitSvc.NewGitService()
		if err := gitService.AddAndCommit(repo.Path, req.Path, req.Message); err != nil {
			response.InternalError(c, err)
			return
		}
	}

	c.Set("audit_target", "repo:"+repo.Key)
	c.Set("audit_details", map[string]string{"path": req.Path, "message": req.Message})
	response.Success(c, api.SaveSpecResponse{
		Message: "spec saved successfully",
		Path:    req.Path,
	})
}

// LintSpec .
// @router /api/v1/spec/lint [POST]
func LintSpec(ctx context.Context, c *app.RequestContext) {
	var req spec.LintRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	content := req.GetContent()
	if content == "" {
		response.BadRequest(c, "content is required")
		return
	}

	lintService := lintSvc.NewLintService()
	result, err := lintService.LintWithAI(ctx, content, req.GetRules(), req.GetMode())
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.Success(c, result)
}

// GetLintRules .
// @router /api/v1/spec/rules [GET]
func GetLintRules(ctx context.Context, c *app.RequestContext) {
	rules, err := db.NewLintRuleDAO().FindAll()
	if err != nil {
		response.InternalError(c, err)
		return
	}

	var dtos []api.LintRule
	for _, r := range rules {
		dtos = append(dtos, api.LintRule{
			ID:          r.ID,
			Name:        r.Name,
			Description: r.Description,
			Category:    r.Category,
			Severity:    r.Severity,
			Pattern:     r.Pattern,
			Enabled:     r.Enabled,
			Priority:    r.Priority,
			CreatedAt:   r.CreatedAt,
			UpdatedAt:   r.UpdatedAt,
		})
	}

	if dtos == nil {
		dtos = []api.LintRule{}
	}

	response.Success(c, dtos)
}

// UpdateLintRule .
// @router /api/v1/spec/rules/:id [PUT]
func UpdateLintRule(ctx context.Context, c *app.RequestContext) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "rule id is required")
		return
	}

	var req api.UpdateLintRuleReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	dao := db.NewLintRuleDAO()
	rule, err := dao.FindByID(id)
	if err != nil {
		response.NotFound(c, "rule not found")
		return
	}

	if req.Name != "" {
		rule.Name = req.Name
	}
	if req.Description != "" {
		rule.Description = req.Description
	}
	if req.Category != "" {
		rule.Category = req.Category
	}
	if req.Severity != "" {
		rule.Severity = req.Severity
	}
	if req.Pattern != "" {
		rule.Pattern = req.Pattern
	}
	if req.Enabled != nil {
		rule.Enabled = *req.Enabled
	}
	if req.Priority != nil {
		rule.Priority = *req.Priority
	}

	if err := dao.Save(rule); err != nil {
		response.InternalError(c, err)
		return
	}

	response.Success(c, api.LintRule{
		ID:          rule.ID,
		Name:        rule.Name,
		Description: rule.Description,
		Category:    rule.Category,
		Severity:    rule.Severity,
		Pattern:     rule.Pattern,
		Enabled:     rule.Enabled,
		Priority:    rule.Priority,
		CreatedAt:   rule.CreatedAt,
		UpdatedAt:   rule.UpdatedAt,
	})
}

// CreateLintRule .
// @router /api/v1/spec/rules [POST]
func CreateLintRule(ctx context.Context, c *app.RequestContext) {
	var req api.CreateLintRuleReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.ID == "" {
		response.BadRequest(c, "id is required")
		return
	}
	if req.Name == "" {
		response.BadRequest(c, "name is required")
		return
	}

	dao := db.NewLintRuleDAO()
	exists, _ := dao.ExistsByID(req.ID)
	if exists {
		response.BadRequest(c, "rule with this id already exists")
		return
	}

	rule := &po.LintRule{
		ID:          req.ID,
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		Severity:    req.Severity,
		Pattern:     req.Pattern,
		Enabled:     req.Enabled,
		Priority:    req.Priority,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if rule.Category == "" {
		rule.Category = "custom"
	}
	if rule.Severity == "" {
		rule.Severity = "warning"
	}

	if err := dao.Create(rule); err != nil {
		response.InternalError(c, err)
		return
	}

	response.Success(c, api.LintRule{
		ID:          rule.ID,
		Name:        rule.Name,
		Description: rule.Description,
		Category:    rule.Category,
		Severity:    rule.Severity,
		Pattern:     rule.Pattern,
		Enabled:     rule.Enabled,
		Priority:    rule.Priority,
		CreatedAt:   rule.CreatedAt,
		UpdatedAt:   rule.UpdatedAt,
	})
}

// CommitSpec .
// @router /api/v1/spec/commit/:path [POST]
func CommitSpec(ctx context.Context, c *app.RequestContext) {
	path := c.Param("path")
	if path == "" {
		response.BadRequest(c, "path is required")
		return
	}

	var req api.CommitSpecReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.Message == "" {
		response.BadRequest(c, "message is required")
		return
	}

	if req.Path == "" {
		req.Path = path
	}

	repo, err := db.NewRepoDAO().FindByKey(req.RepoKey)
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	if req.Content != "" {
		svc := specService.NewSpecService()
		if err := svc.SaveSpecContent(repo.Path, req.Path, req.Content, ""); err != nil {
			response.InternalError(c, err)
			return
		}
	}

	gitService := gitSvc.NewGitService()
	if err := gitService.AddAndCommit(repo.Path, req.Path, req.Message); err != nil {
		response.InternalError(c, err)
		return
	}

	c.Set("audit_target", "repo:"+repo.Key)
	c.Set("audit_details", map[string]string{"path": req.Path, "message": req.Message})
	response.Success(c, api.CommitResponse{
		Message: "committed successfully",
	})
}

// ValidateSpec .
// @router /api/v1/spec/validate [POST]
func ValidateSpec(ctx context.Context, c *app.RequestContext) {
	var req spec.ValidateSpecRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	svc := specService.NewSpecService()
	result := svc.ValidateSpec(req.GetContent())

	response.Success(c, result)
}

// CreateSpecFile .
// @router /api/v1/spec/create [POST]
func CreateSpecFile(ctx context.Context, c *app.RequestContext) {
	var req api.CreateSpecFileReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	repo, err := db.NewRepoDAO().FindByKey(req.RepoKey)
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	svc := specService.NewSpecService()
	path, err := svc.CreateSpecFileWithContent(repo.Path, req.Path, req.Name, req.Content)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	c.Set("audit_target", "repo:"+repo.Key)
	c.Set("audit_details", map[string]string{"path": path})
	response.Success(c, api.CreateFileResponse{
		Path:    path,
		Message: "Spec 文件创建成功",
	})
}

// DeleteSpecFile .
// @router /api/v1/spec/delete [POST]
func DeleteSpecFile(ctx context.Context, c *app.RequestContext) {
	var req api.DeleteSpecFileReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	repo, err := db.NewRepoDAO().FindByKey(req.RepoKey)
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	svc := specService.NewSpecService()
	err = svc.DeleteSpecFile(repo.Path, req.Path)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	if req.CommitMessage != "" {
		gitService := gitSvc.NewGitService()
		if err := gitService.RemoveAndCommit(repo.Path, req.Path, req.CommitMessage); err != nil {
			response.InternalError(c, err)
			return
		}
	}

	c.Set("audit_target", "repo:"+repo.Key)
	c.Set("audit_details", map[string]string{"path": req.Path})
	response.Success(c, api.MessageResponse{
		Message: "spec deleted",
	})
}

// AIAssistSpec .
// @router /api/v1/spec/ai-assist [POST]
func AIAssistSpec(ctx context.Context, c *app.RequestContext) {
	var req spec.AIAssistRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	content := req.GetContent()
	prompt := req.GetPrompt()
	if content == "" || prompt == "" {
		response.BadRequest(c, "content and prompt are required")
		return
	}

	var history []llm.ChatMessage
	for _, h := range req.GetHistory() {
		history = append(history, llm.ChatMessage{Role: h.GetRole(), Content: h.GetContent()})
	}

	svc := specService.NewSpecService()
	result, applyContent, err := svc.AIAssist(ctx, content, prompt, req.GetAction(), history)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	resp := api.AIAssistResponse{Result: result, ApplyContent: applyContent}
	c.Set("audit_target", "repo:spec")
	c.Set("audit_details", map[string]string{"action": req.GetAction()})
	response.Success(c, resp)
}

// AIFixSpec .
// @router /api/v1/spec/ai-fix [POST]
func AIFixSpec(ctx context.Context, c *app.RequestContext) {
	var req spec.AIFixRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	content := req.GetContent()
	issue := req.GetIssue()
	if content == "" || issue == "" {
		response.BadRequest(c, "content and issue are required")
		return
	}

	result, err := lintSvc.AIFix(ctx, content, issue, int(req.GetLine()), req.GetSeverity())
	if err != nil {
		response.InternalError(c, err)
		return
	}
	c.Set("audit_target", "repo:spec")
	c.Set("audit_details", map[string]string{"severity": req.GetSeverity()})
	response.Success(c, api.AIFixResponse{Content: result})
}

// FormatSpec .
// @router /api/v1/spec/format [POST]
func FormatSpec(ctx context.Context, c *app.RequestContext) {
	var err error
	var req spec.FormatSpecRequest
	err = c.BindAndValidate(&req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	content := req.GetContent()
	if content == "" {
		response.BadRequest(c, "content is required")
		return
	}

	opts := specService.FormatOptions{
		Curlify:         req.GetCurlify(),
		RemoveClean:     req.GetRemoveClean(),
		RemoveBuildRoot: req.GetRemoveBuildRoot(),
		RemoveGroup:     req.GetRemoveGroup(),
		LicenseSPDX:     req.GetLicenseSpdx(),
		SortDeps:        req.GetSortDeps(),
		TabToSpaces:     req.GetTabToSpaces(),
		IndentSize:      int(req.GetIndentSize()),
		PreambleOrder:   req.GetPreambleOrder(),
		AlignValues:     req.GetAlignValues(),
		PathMacros:      req.GetPathMacros(),
		UtilMacros:      req.GetUtilMacros(),
		CommonCleanup:   req.GetCommonCleanup(),
		ConditionalTrim: req.GetConditionalTrim(),
	}

	formatter := specService.NewSpecFormatter()
	formatted, changes, err := formatter.Format(content, opts)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	var dtos []api.FormatChangeDTO
	for _, ch := range changes {
		dtos = append(dtos, api.FormatChangeDTO{
			Line:   ch.Line,
			Type:   ch.Type,
			Before: ch.Before,
			After:  ch.After,
			Reason: ch.Reason,
		})
	}
	if dtos == nil {
		dtos = []api.FormatChangeDTO{}
	}

	response.Success(c, api.FormatResponse{
		Content: formatted,
		Changes: dtos,
	})
}

// GetSpecConfig .
// @router /api/v1/spec/config [GET]
func GetSpecConfig(ctx context.Context, c *app.RequestContext) {
	dao := db.NewSystemConfigDAO()
	result := SpecConfigResponse{
		FormatOptions: defaultFormatOptions(),
		AIConfig:      defaultAIConfig(),
	}

	if val, err := dao.GetConfig("spec_default_template"); err == nil && val != "" {
		result.DefaultTemplate = val
	} else {
		result.DefaultTemplate = specService.NewSpecService().GetSpecTemplate()
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

// SaveSpecConfig .
// @router /api/v1/spec/config [PUT]
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

func buildSpecTree(repoPath string) ([]*spec.SpecFile, error) {
	nodes := make(map[string]*spec.SpecFile)

	err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}

		relPath, _ := filepath.Rel(repoPath, path)

		if strings.HasSuffix(info.Name(), ".spec") || info.IsDir() {
			name := info.Name()
			p := relPath
			isDir := info.IsDir()
			size := info.Size()
			modTime := info.ModTime().Format("2006-01-02T15:04:05Z")
			nodes[relPath] = &spec.SpecFile{
				Name:    &name,
				Path:    &p,
				IsDir:   &isDir,
				Size:    &size,
				ModTime: &modTime,
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	childrenMap := make(map[string][]*spec.SpecFile)

	for path, node := range nodes {
		if path == "." {
			continue
		}

		parentPath := filepath.Dir(path)
		if parentPath == "" {
			parentPath = "."
		}

		childrenMap[parentPath] = append(childrenMap[parentPath], node)
	}

	var buildTree func(path string) *spec.SpecFile
	buildTree = func(path string) *spec.SpecFile {
		node := nodes[path]
		if node == nil {
			return nil
		}

		children := childrenMap[path]
		if len(children) > 0 {
			node.Children = children
			for _, child := range children {
				buildTree(*child.Path)
			}
		}

		return node
	}

	root := buildTree(".")
	if root == nil {
		return []*spec.SpecFile{}, nil
	}

	filterTree(root)

	if len(root.Children) > 0 {
		return root.Children, nil
	}

	return []*spec.SpecFile{}, nil
}

func createDirChain(pathMap map[string]*spec.SpecFile, path string, repoPath string) *spec.SpecFile {
	if path == "." || path == "" {
		return pathMap["."]
	}

	if dir, exists := pathMap[path]; exists {
		return dir
	}

	info, err := os.Stat(filepath.Join(repoPath, path))
	if err != nil {
		return nil
	}

	name := filepath.Base(path)
	p := path
	isDir := true
	modTime := info.ModTime().Format("2006-01-02T15:04:05Z")
	dir := &spec.SpecFile{
		Name:    &name,
		Path:    &p,
		IsDir:   &isDir,
		ModTime: &modTime,
	}
	pathMap[path] = dir

	parentPath := filepath.Dir(path)
	if parentPath == "" {
		parentPath = "."
	}

	parent := createDirChain(pathMap, parentPath, repoPath)
	if parent != nil {
		parent.Children = append(parent.Children, dir)
	}

	return dir
}

func filterTree(node *spec.SpecFile) bool {
	if !*node.IsDir {
		return strings.HasSuffix(*node.Name, ".spec")
	}

	var hasSpecFile bool
	var filteredChildren []*spec.SpecFile

	for _, child := range node.Children {
		if filterTree(child) {
			filteredChildren = append(filteredChildren, child)
			hasSpecFile = true
		}
	}

	node.Children = filteredChildren
	return hasSpecFile
}
