package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/yi-nology/git-manage-service/biz/service/audit"
)

type AuditConfig struct {
	Action       string
	TargetType   string
	TargetSource string
}

var numericIDRe = regexp.MustCompile(`/\d+(/|$)`)
var namedParamPatterns = []struct {
	prefix string
	suffix string
	repl   string
}{
	{"/api/v1/repo/", "/author/", "/{repo_key}/author/"},
	{"/api/v1/repo/", "/maintenance/", "/{repo_key}/maintenance/"},
	{"/api/v1/reviews/config/", "", "/{repo_key}"},
	{"/api/v1/review/remote-config/", "", "/{provider_id}/{owner}/{repo}"},
	{"/api/v1/branch-rules/remote-config/", "", "/{provider_id}/{owner}/{repo}"},
	{"/api/webhooks/trigger/", "", "/{token}"},
	{"/api/v1/spec/commit/", "", "/{path}"},
	{"/api/v1/spec/content/", "", "/{path}"},
	{"/api/v1/spec/rules/", "", "/{id}"},
	{"/api/v1/settings/llm-providers/", "/default", "/{id}/default"},
	{"/api/v1/settings/llm-providers/", "/test", "/{id}/test"},
	{"/api/v1/settings/review-rules/", "", "/{rule_id}"},
}

var auditRoutes = map[string]AuditConfig{
	"POST:/api/v1/repo/create":         {Action: "CREATE", TargetType: "repo", TargetSource: "body:key"},
	"POST:/api/v1/repo/update":         {Action: "UPDATE", TargetType: "repo", TargetSource: "body:key"},
	"POST:/api/v1/repo/delete":         {Action: "DELETE", TargetType: "repo", TargetSource: "body:key"},
	"POST:/api/v1/repo/fetch":          {Action: "FETCH_REPO", TargetType: "repo", TargetSource: "body:repo_key"},
	"POST:/api/v1/repo/clone":          {Action: "CLONE_REPO", TargetType: "repo", TargetSource: "context"},
	"POST:/api/v1/repo/scan":           {Action: "SCAN_REPO", TargetType: "repo", TargetSource: "body:path"},
	"POST:/api/v1/repo/batch-create":   {Action: "BATCH_CREATE_REPO", TargetType: "repo", TargetSource: "body:repos"},
	"POST:/api/v1/repo/scan-directory": {Action: "SCAN_DIRECTORY", TargetType: "repo", TargetSource: "body:path"},

	"POST:/api/v1/tag/create": {Action: "TAG_CREATE", TargetType: "repo", TargetSource: "body:repo_key"},
	"POST:/api/v1/tag/delete": {Action: "TAG_DELETE", TargetType: "repo", TargetSource: "body:repo_key"},
	"POST:/api/v1/tag/push":   {Action: "TAG_PUSH", TargetType: "repo", TargetSource: "body:repo_key"},

	"POST:/api/v1/credentials/":          {Action: "CREDENTIAL_CREATE", TargetType: "credential", TargetSource: "response:id"},
	"PUT:/api/v1/credentials/{id}":       {Action: "CREDENTIAL_UPDATE", TargetType: "credential", TargetSource: "param:id"},
	"DELETE:/api/v1/credentials/{id}":    {Action: "CREDENTIAL_DELETE", TargetType: "credential", TargetSource: "param:id"},
	"POST:/api/v1/credentials/{id}/test": {Action: "CREDENTIAL_TEST", TargetType: "credential", TargetSource: "param:id"},

	"POST:/api/v1/system/db-ssh-keys":           {Action: "SSHKEY_CREATE", TargetType: "sshkey", TargetSource: "response:id"},
	"PUT:/api/v1/system/db-ssh-keys/{id}":       {Action: "SSHKEY_UPDATE", TargetType: "sshkey", TargetSource: "param:id"},
	"DELETE:/api/v1/system/db-ssh-keys/{id}":    {Action: "SSHKEY_DELETE", TargetType: "sshkey", TargetSource: "param:id"},
	"POST:/api/v1/system/db-ssh-keys/{id}/test": {Action: "SSHKEY_TEST", TargetType: "sshkey", TargetSource: "param:id"},

	"POST:/api/v1/workspace/stage":               {Action: "WORKSPACE_STAGE", TargetType: "repo", TargetSource: "body:repo_key"},
	"POST:/api/v1/workspace/unstage":             {Action: "WORKSPACE_UNSTAGE", TargetType: "repo", TargetSource: "body:repo_key"},
	"POST:/api/v1/workspace/commit":              {Action: "WORKSPACE_COMMIT", TargetType: "repo", TargetSource: "body:repo_key"},
	"POST:/api/v1/workspace/pull":                {Action: "WORKSPACE_PULL", TargetType: "repo", TargetSource: "body:repo_key"},
	"POST:/api/v1/workspace/push":                {Action: "WORKSPACE_PUSH", TargetType: "repo", TargetSource: "body:repo_key"},
	"POST:/api/v1/workspace/untrack":             {Action: "WORKSPACE_UNTRACK", TargetType: "repo", TargetSource: "body:repo_key"},
	"POST:/api/v1/workspace/gitignore":           {Action: "WORKSPACE_GITIGNORE", TargetType: "repo", TargetSource: "body:repo_key"},
	"POST:/api/v1/workspace/resolve":             {Action: "WORKSPACE_RESOLVE", TargetType: "repo", TargetSource: "body:repo_key"},
	"POST:/api/v1/workspace/ai-resolve":          {Action: "WORKSPACE_AI_RESOLVE", TargetType: "repo", TargetSource: "context"},
	"POST:/api/v1/workspace/generate-commit-msg": {Action: "WORKSPACE_AI_COMMIT_MSG", TargetType: "repo", TargetSource: "body:repo_key"},

	"POST:/api/v1/branch/create":          {Action: "CREATE_BRANCH", TargetType: "repo", TargetSource: "body:repo_key"},
	"POST:/api/v1/branch/delete":          {Action: "DELETE_BRANCH", TargetType: "repo", TargetSource: "body:repo_key"},
	"POST:/api/v1/branch/update":          {Action: "UPDATE_BRANCH", TargetType: "repo", TargetSource: "body:repo_key"},
	"POST:/api/v1/branch/checkout":        {Action: "CHECKOUT_BRANCH", TargetType: "repo", TargetSource: "body:repo_key"},
	"POST:/api/v1/branch/push":            {Action: "PUSH_BRANCH", TargetType: "repo", TargetSource: "body:repo_key"},
	"POST:/api/v1/branch/pull":            {Action: "PULL_BRANCH", TargetType: "repo", TargetSource: "body:repo_key"},
	"POST:/api/v1/branch/merge":           {Action: "MERGE", TargetType: "repo", TargetSource: "context"},
	"POST:/api/v1/branch/cherry-pick":     {Action: "CHERRY_PICK", TargetType: "repo", TargetSource: "body:repo_key"},
	"POST:/api/v1/branch/rebase":          {Action: "REBASE", TargetType: "repo", TargetSource: "body:repo_key"},
	"POST:/api/v1/branch/rebase/abort":    {Action: "REBASE_ABORT", TargetType: "repo", TargetSource: "body:repo_key"},
	"POST:/api/v1/branch/rebase/continue": {Action: "REBASE_CONTINUE", TargetType: "repo", TargetSource: "body:repo_key"},

	"POST:/api/v1/stash/save":  {Action: "STASH_SAVE", TargetType: "repo", TargetSource: "body:repo_key"},
	"POST:/api/v1/stash/apply": {Action: "STASH_APPLY", TargetType: "repo", TargetSource: "body:repo_key"},
	"POST:/api/v1/stash/pop":   {Action: "STASH_POP", TargetType: "repo", TargetSource: "body:repo_key"},
	"POST:/api/v1/stash/drop":  {Action: "STASH_DROP", TargetType: "repo", TargetSource: "body:repo_key"},
	"POST:/api/v1/stash/clear": {Action: "STASH_CLEAR", TargetType: "repo", TargetSource: "body:repo_key"},

	"POST:/api/v1/submodule/add":    {Action: "SUBMODULE_ADD", TargetType: "repo", TargetSource: "body:repo_key"},
	"POST:/api/v1/submodule/init":   {Action: "SUBMODULE_INIT", TargetType: "repo", TargetSource: "body:repo_key"},
	"POST:/api/v1/submodule/update": {Action: "SUBMODULE_UPDATE", TargetType: "repo", TargetSource: "body:repo_key"},
	"POST:/api/v1/submodule/sync":   {Action: "SUBMODULE_SYNC", TargetType: "repo", TargetSource: "body:repo_key"},
	"POST:/api/v1/submodule/remove": {Action: "SUBMODULE_REMOVE", TargetType: "repo", TargetSource: "body:repo_key"},

	"POST:/api/v1/sync/task/create": {Action: "SYNC_CREATE", TargetType: "task", TargetSource: "context"},
	"POST:/api/v1/sync/task/update": {Action: "SYNC_UPDATE", TargetType: "task", TargetSource: "context"},
	"POST:/api/v1/sync/task/delete": {Action: "SYNC_DELETE", TargetType: "task", TargetSource: "context"},
	"POST:/api/v1/sync/execute":     {Action: "SYNC_EXECUTE", TargetType: "task", TargetSource: "body:task_key"},
	"POST:/api/v1/sync/run":         {Action: "SYNC_RUN", TargetType: "task", TargetSource: "context"},
	"POST:/api/v1/sync/batch":       {Action: "SYNC_BATCH", TargetType: "task", TargetSource: "context"},
	"POST:/api/v1/sync/preview":     {Action: "SYNC_PREVIEW", TargetType: "task", TargetSource: "context"},

	"POST:/api/webhooks/task-sync":       {Action: "WEBHOOK_TRIGGER", TargetType: "task", TargetSource: "context"},
	"POST:/api/webhooks/trigger/{token}": {Action: "WEBHOOK_TRIGGER_BY_TOKEN", TargetType: "task", TargetSource: "context"},
	"POST:/api/webhooks/receive":         {Action: "WEBHOOK_RECEIVED", TargetType: "webhook", TargetSource: "context"},

	"POST:/api/v1/system/repo/submit":      {Action: "SUBMIT_CHANGES", TargetType: "repo", TargetSource: "context"},
	"POST:/api/v1/system/config":           {Action: "SYSTEM_CONFIG_UPDATE", TargetType: "settings", TargetSource: "context"},
	"POST:/api/v1/system/test-connection":  {Action: "SYSTEM_TEST_CONNECTION", TargetType: "system", TargetSource: "context"},
	"POST:/api/v1/system/select-directory": {Action: "SYSTEM_SELECT_DIR", TargetType: "system", TargetSource: "context"},

	"POST:/api/v1/settings/llm-providers":                                 {Action: "LLM_PROVIDER_CREATE", TargetType: "llm_provider", TargetSource: "response:id"},
	"PUT:/api/v1/settings/llm-providers/{id}":                             {Action: "LLM_PROVIDER_UPDATE", TargetType: "llm_provider", TargetSource: "param:id"},
	"DELETE:/api/v1/settings/llm-providers/{id}":                          {Action: "LLM_PROVIDER_DELETE", TargetType: "llm_provider", TargetSource: "param:id"},
	"POST:/api/v1/settings/llm-providers/{id}/default":                    {Action: "LLM_PROVIDER_SET_DEFAULT", TargetType: "llm_provider", TargetSource: "param:id"},
	"POST:/api/v1/settings/llm-providers/{id}/test":                       {Action: "LLM_PROVIDER_TEST", TargetType: "llm_provider", TargetSource: "param:id"},
	"PUT:/api/v1/settings/code-review":                                    {Action: "CODE_REVIEW_SETTINGS_UPDATE", TargetType: "settings", TargetSource: "context"},
	"PUT:/api/v1/settings/branch-rules":                                   {Action: "BRANCH_RULES_UPDATE", TargetType: "settings", TargetSource: "context"},
	"PUT:/api/v1/branch-rules/remote-config/{provider_id}/{owner}/{repo}": {Action: "REMOTE_BRANCH_RULES_UPDATE", TargetType: "settings", TargetSource: "context"},

	"POST:/api/v1/settings/review-rules":             {Action: "REVIEW_RULE_CREATE", TargetType: "review_rule", TargetSource: "response:id"},
	"PUT:/api/v1/settings/review-rules/batch":        {Action: "REVIEW_RULES_BATCH_UPDATE", TargetType: "settings", TargetSource: "context"},
	"PUT:/api/v1/settings/review-rules/{rule_id}":    {Action: "REVIEW_RULE_UPDATE", TargetType: "review_rule", TargetSource: "param:rule_id"},
	"DELETE:/api/v1/settings/review-rules/{rule_id}": {Action: "REVIEW_RULE_DELETE", TargetType: "review_rule", TargetSource: "param:rule_id"},

	"POST:/api/v1/providers":                 {Action: "PROVIDER_CREATE", TargetType: "provider", TargetSource: "response:id"},
	"PUT:/api/v1/providers/{id}":             {Action: "PROVIDER_UPDATE", TargetType: "provider", TargetSource: "param:id"},
	"DELETE:/api/v1/providers/{id}":          {Action: "PROVIDER_DELETE", TargetType: "provider", TargetSource: "param:id"},
	"POST:/api/v1/providers/{id}/test":       {Action: "PROVIDER_TEST", TargetType: "provider", TargetSource: "param:id"},
	"POST:/api/v1/providers/branches/create": {Action: "PROVIDER_REMOTE_BRANCH_CREATE", TargetType: "provider", TargetSource: "context"},
	"POST:/api/v1/providers/branches/delete": {Action: "PROVIDER_REMOTE_BRANCH_DELETE", TargetType: "provider", TargetSource: "context"},

	"POST:/api/v1/bindings":                  {Action: "BINDING_CREATE", TargetType: "binding", TargetSource: "response:id"},
	"PUT:/api/v1/bindings/{id}":              {Action: "BINDING_UPDATE", TargetType: "binding", TargetSource: "param:id"},
	"DELETE:/api/v1/bindings/{id}":           {Action: "BINDING_DELETE", TargetType: "binding", TargetSource: "param:id"},
	"POST:/api/v1/bindings/{id}/set-primary": {Action: "BINDING_SET_PRIMARY", TargetType: "binding", TargetSource: "param:id"},
	"POST:/api/v1/bindings/{id}/webhook":     {Action: "BINDING_REGISTER_WEBHOOK", TargetType: "binding", TargetSource: "param:id"},
	"DELETE:/api/v1/bindings/{id}/webhook":   {Action: "BINDING_DELETE_WEBHOOK", TargetType: "binding", TargetSource: "param:id"},

	"POST:/api/v1/cr/create":        {Action: "CR_CREATE", TargetType: "repo", TargetSource: "context"},
	"POST:/api/v1/cr/merge":         {Action: "CR_MERGE", TargetType: "repo", TargetSource: "context"},
	"POST:/api/v1/cr/close":         {Action: "CR_CLOSE", TargetType: "repo", TargetSource: "context"},
	"POST:/api/v1/cr/remote/create": {Action: "CR_CREATE_BY_PROVIDER", TargetType: "provider", TargetSource: "context"},
	"POST:/api/v1/cr/remote/merge":  {Action: "CR_MERGE_BY_PROVIDER", TargetType: "provider", TargetSource: "context"},
	"POST:/api/v1/cr/remote/close":  {Action: "CR_CLOSE_BY_PROVIDER", TargetType: "provider", TargetSource: "context"},
	"POST:/api/v1/cr/sync":          {Action: "CR_SYNC", TargetType: "repo", TargetSource: "context"},

	"POST:/api/v1/reviews/tasks":                                    {Action: "REVIEW_TASK_CREATE", TargetType: "repo", TargetSource: "context"},
	"POST:/api/v1/reviews/tasks/{id}/retry":                         {Action: "REVIEW_TASK_RETRY", TargetType: "review_task", TargetSource: "param:id"},
	"PUT:/api/v1/reviews/config/{repo_key}":                         {Action: "REVIEW_CONFIG_UPDATE", TargetType: "repo", TargetSource: "param:repo_key"},
	"PUT:/api/v1/review/remote-config/{provider_id}/{owner}/{repo}": {Action: "REMOTE_REVIEW_CONFIG_UPDATE", TargetType: "provider", TargetSource: "context"},

	"POST:/api/v1/author/identities":               {Action: "AUTHOR_IDENTITY_CREATE", TargetType: "author", TargetSource: "response:id"},
	"PUT:/api/v1/author/identities/{id}":           {Action: "AUTHOR_IDENTITY_UPDATE", TargetType: "author", TargetSource: "param:id"},
	"DELETE:/api/v1/author/identities/{id}":        {Action: "AUTHOR_IDENTITY_DELETE", TargetType: "author", TargetSource: "param:id"},
	"POST:/api/v1/author/identities/{id}/activate": {Action: "AUTHOR_IDENTITY_ACTIVATE", TargetType: "author", TargetSource: "param:id"},
	"PUT:/api/v1/repo/{repo_key}/author/config":    {Action: "AUTHOR_REPO_CONFIG_SET", TargetType: "repo", TargetSource: "param:repo_key"},
	"POST:/api/v1/repo/{repo_key}/author/fix":      {Action: "AUTHOR_FIX", TargetType: "repo", TargetSource: "param:repo_key"},
	"POST:/api/v1/repo/{repo_key}/author/fix-all":  {Action: "AUTHOR_FIX_ALL", TargetType: "repo", TargetSource: "param:repo_key"},

	"POST:/api/v1/notification/channel/create": {Action: "NOTIFICATION_CHANNEL_CREATE", TargetType: "channel", TargetSource: "response:id"},
	"POST:/api/v1/notification/channel/update": {Action: "NOTIFICATION_CHANNEL_UPDATE", TargetType: "channel", TargetSource: "body:id"},
	"POST:/api/v1/notification/channel/delete": {Action: "NOTIFICATION_CHANNEL_DELETE", TargetType: "channel", TargetSource: "body:id"},

	"POST:/api/v1/patch/save":   {Action: "PATCH_SAVE", TargetType: "repo", TargetSource: "body:repo_key"},
	"POST:/api/v1/patch/apply":  {Action: "PATCH_APPLY", TargetType: "repo", TargetSource: "body:repo_key"},
	"POST:/api/v1/patch/delete": {Action: "PATCH_DELETE", TargetType: "repo", TargetSource: "body:repo_key"},

	"POST:/api/v1/repo/{repo_key}/maintenance/slim":      {Action: "REPO_SLIM", TargetType: "repo", TargetSource: "param:repo_key"},
	"POST:/api/v1/repo/{repo_key}/maintenance/gc":        {Action: "REPO_GC", TargetType: "repo", TargetSource: "param:repo_key"},
	"POST:/api/v1/repo/{repo_key}/maintenance/gitignore": {Action: "REPO_GITIGNORE", TargetType: "repo", TargetSource: "param:repo_key"},

	"POST:/api/v1/spec/save":          {Action: "SAVE_SPEC", TargetType: "repo", TargetSource: "context"},
	"POST:/api/v1/spec/create":        {Action: "CREATE_SPEC", TargetType: "repo", TargetSource: "context"},
	"POST:/api/v1/spec/delete":        {Action: "DELETE_SPEC", TargetType: "repo", TargetSource: "context"},
	"PUT:/api/v1/spec/content/{path}": {Action: "SAVE_SPEC", TargetType: "repo", TargetSource: "context"},
	"POST:/api/v1/spec/commit/{path}": {Action: "COMMIT_SPEC", TargetType: "repo", TargetSource: "context"},
	"POST:/api/v1/spec/ai-assist":     {Action: "SPEC_AI_ASSIST", TargetType: "repo", TargetSource: "context"},
	"POST:/api/v1/spec/ai-fix":        {Action: "SPEC_AI_FIX", TargetType: "repo", TargetSource: "context"},
	"PUT:/api/v1/spec/config":         {Action: "SPEC_CONFIG_SAVE", TargetType: "settings", TargetSource: "context"},
	"POST:/api/v1/spec/rules":         {Action: "SPEC_LINT_RULE_CREATE", TargetType: "lint_rule", TargetSource: "response:id"},
	"PUT:/api/v1/spec/rules/{id}":     {Action: "SPEC_LINT_RULE_UPDATE", TargetType: "lint_rule", TargetSource: "param:id"},

	"POST:/api/v1/webhook/events/retry": {Action: "WEBHOOK_EVENT_RETRY", TargetType: "webhook_event", TargetSource: "context"},

	"POST:/api/v1/stats/lines/config": {Action: "STATS_CONFIG_SAVE", TargetType: "repo", TargetSource: "context"},
}

func AuditMiddleware() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		c.Next(ctx)

		if isReadMethod(c) {
			return
		}
		if c.Response.StatusCode() > 299 {
			return
		}
		if skip, _ := c.Get("audit_skip"); skip == true {
			return
		}

		cfg := matchRoute(string(c.Method()), string(c.Path()))
		if cfg == nil {
			return
		}

		target := extractTarget(c, cfg)
		details, _ := c.Get("audit_details")

		audit.AuditSvc.Log(c, cfg.Action, target, details)
	}
}

func isReadMethod(c *app.RequestContext) bool {
	m := string(c.Method())
	return m == "GET" || m == "HEAD" || m == "OPTIONS"
}

func matchRoute(method, path string) *AuditConfig {
	key := method + ":" + path
	if cfg, ok := auditRoutes[key]; ok {
		return &cfg
	}

	normalized := numericIDRe.ReplaceAllString(path, "/{id}$1")
	key = method + ":" + normalized
	if cfg, ok := auditRoutes[key]; ok {
		return &cfg
	}

	for _, p := range namedParamPatterns {
		if strings.HasPrefix(path, p.prefix) && (p.suffix == "" || strings.Contains(path, p.suffix)) {
			norm2 := applyNamedPattern(path, p.prefix, p.repl)
			key2 := method + ":" + norm2
			if cfg, ok := auditRoutes[key2]; ok {
				return &cfg
			}
		}
	}

	return nil
}

func applyNamedPattern(path, prefix, repl string) string {
	rest := strings.TrimPrefix(path, prefix)
	slashIdx := strings.Index(rest, "/")
	if slashIdx == -1 {
		return prefix + repl
	}
	subPath := rest[slashIdx:]
	return prefix + repl + subPath
}

func extractTarget(c *app.RequestContext, cfg *AuditConfig) string {
	if t, ok := c.Get("audit_target"); ok {
		if s, ok := t.(string); ok && s != "" {
			return s
		}
	}

	switch {
	case strings.HasPrefix(cfg.TargetSource, "body:"):
		field := strings.TrimPrefix(cfg.TargetSource, "body:")
		val := extractFromBody(c, field)
		if val != "" {
			return cfg.TargetType + ":" + val
		}

	case strings.HasPrefix(cfg.TargetSource, "query:"):
		field := strings.TrimPrefix(cfg.TargetSource, "query:")
		val := c.Query(field)
		if val != "" {
			return cfg.TargetType + ":" + val
		}

	case strings.HasPrefix(cfg.TargetSource, "param:"):
		field := strings.TrimPrefix(cfg.TargetSource, "param:")
		val := c.Param(field)
		if val != "" {
			return cfg.TargetType + ":" + val
		}

	case strings.HasPrefix(cfg.TargetSource, "response:"):
		field := strings.TrimPrefix(cfg.TargetSource, "response:")
		val := extractFromResponse(c, field)
		if val != "" {
			return cfg.TargetType + ":" + val
		}
	}

	return cfg.TargetType + ":unknown"
}

func extractFromBody(c *app.RequestContext, field string) string {
	body, _ := c.Body()
	if len(body) == 0 {
		return ""
	}
	var m map[string]interface{}
	if err := json.Unmarshal(body, &m); err != nil {
		return ""
	}
	if v, ok := m[field]; ok {
		if s, ok := v.(string); ok {
			return s
		}
		if f, ok := v.(float64); ok {
			return fmt.Sprintf("%.0f", f)
		}
	}
	return ""
}

func extractFromResponse(c *app.RequestContext, field string) string {
	body := c.Response.Body()
	if len(body) == 0 {
		return ""
	}
	var resp struct {
		Data interface{} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return ""
	}
	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		return ""
	}
	if v, ok := data[field]; ok {
		switch val := v.(type) {
		case string:
			return val
		case float64:
			return fmt.Sprintf("%.0f", val)
		}
	}
	return ""
}
