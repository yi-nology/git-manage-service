package notification

import (
	"strings"
	"testing"

	"github.com/yi-nology/git-manage-service/biz/model/po"
)

func TestRenderTemplate_Success(t *testing.T) {
	tmpl := "任务: {{.TaskName}} 状态: {{.StatusText}}"
	data := &TemplateData{TaskName: "sync-1", StatusText: "成功"}
	result, err := RenderTemplate(tmpl, data)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, "sync-1") || !strings.Contains(result, "成功") {
		t.Errorf("unexpected result: %s", result)
	}
}

func TestRenderTemplate_Empty(t *testing.T) {
	result, err := RenderTemplate("", nil)
	if err != nil {
		t.Fatal(err)
	}
	if result != "" {
		t.Errorf("expected empty result, got %q", result)
	}
}

func TestRenderTemplate_InvalidSyntax(t *testing.T) {
	_, err := RenderTemplate("{{.TaskName", nil)
	if err == nil {
		t.Error("expected error for invalid template syntax")
	}
}

func TestRenderTemplate_DefaultFunc(t *testing.T) {
	tmpl := `{{default "fallback" .TaskName}}`
	data := &TemplateData{TaskName: ""}
	result, err := RenderTemplate(tmpl, data)
	if err != nil {
		t.Fatal(err)
	}
	if result != "fallback" {
		t.Errorf("expected fallback, got %q", result)
	}
}

func TestRenderTemplate_TruncateFunc(t *testing.T) {
	tmpl := `{{truncate 5 .TaskName}}`
	data := &TemplateData{TaskName: "very-long-task-name"}
	result, err := RenderTemplate(tmpl, data)
	if err != nil {
		t.Fatal(err)
	}
	if result != "very-..." {
		t.Errorf("expected truncation, got %q", result)
	}
}

func TestValidateTemplate_Valid(t *testing.T) {
	err := ValidateTemplate("{{.TaskName}}")
	if err != nil {
		t.Errorf("valid template should pass: %v", err)
	}
}

func TestValidateTemplate_Invalid(t *testing.T) {
	err := ValidateTemplate("{{.TaskName")
	if err == nil {
		t.Error("invalid template should fail validation")
	}
}

func TestValidateTemplate_Empty(t *testing.T) {
	err := ValidateTemplate("")
	if err != nil {
		t.Errorf("empty template should pass: %v", err)
	}
}

func TestRenderTitleAndContent_WithCustomTemplates(t *testing.T) {
	data := &TemplateData{
		TaskName:   "my-task",
		Status:     "success",
		EventType:  po.TriggerSyncSuccess,
		StatusText: "成功",
	}
	title, content := RenderTitleAndContent(
		"自定义标题: {{.TaskName}}",
		"自定义内容: {{.StatusText}}",
		data,
	)
	if !strings.Contains(title, "自定义标题") {
		t.Errorf("title: %s", title)
	}
	if !strings.Contains(content, "自定义内容") {
		t.Errorf("content: %s", content)
	}
}

func TestRenderTitleAndContent_FallbackDefaults(t *testing.T) {
	data := &TemplateData{
		TaskName:   "my-task",
		Status:     "success",
		EventType:  po.TriggerSyncSuccess,
		StatusText: "成功",
	}
	title, content := RenderTitleAndContent("", "", data)
	if title == "" {
		t.Error("expected non-empty default title")
	}
	if content == "" {
		t.Error("expected non-empty default content")
	}
}

func TestRenderTitleAndContent_NilData(t *testing.T) {
	title, content := RenderTitleAndContent("", "", nil)
	if title != "通知" {
		t.Errorf("expected fallback title, got %q", title)
	}
	if content != "" {
		t.Errorf("expected empty content for nil data, got %q", content)
	}
}

func TestGetDefaultTitleTemplate_KnownEvents(t *testing.T) {
	events := []string{
		po.TriggerSyncSuccess,
		po.TriggerSyncFailure,
		po.TriggerSyncConflict,
		po.TriggerWebhookReceived,
		po.TriggerBackupSuccess,
	}
	for _, ev := range events {
		tmpl := GetDefaultTitleTemplate(ev)
		if tmpl == "" {
			t.Errorf("expected non-empty title template for %s", ev)
		}
	}
}

func TestGetDefaultTitleTemplate_UnknownEvent(t *testing.T) {
	tmpl := GetDefaultTitleTemplate("unknown_event")
	if !strings.Contains(tmpl, "通知") {
		t.Errorf("expected fallback title, got %q", tmpl)
	}
}

func TestGetDefaultContentTemplate_KnownEvents(t *testing.T) {
	events := []string{
		po.TriggerSyncSuccess,
		po.TriggerSyncFailure,
		po.TriggerSyncConflict,
		po.TriggerCronTriggered,
	}
	for _, ev := range events {
		tmpl := GetDefaultContentTemplate(ev)
		if tmpl == "" {
			t.Errorf("expected non-empty content template for %s", ev)
		}
	}
}

func TestGetAvailableVariables_NonEmpty(t *testing.T) {
	vars := GetAvailableVariables()
	if len(vars) == 0 {
		t.Error("expected non-empty variable list")
	}
	found := false
	for _, v := range vars {
		if v.Name == "TaskName" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected TaskName in available variables")
	}
}

func TestFillDefaults_StatusText(t *testing.T) {
	data := &TemplateData{Status: "success"}
	fillDefaults(data)
	if data.StatusText != "成功" {
		t.Errorf("expected 成功, got %q", data.StatusText)
	}

	data2 := &TemplateData{Status: "failure"}
	fillDefaults(data2)
	if data2.StatusText != "失败" {
		t.Errorf("expected 失败, got %q", data2.StatusText)
	}

	data3 := &TemplateData{Status: "conflict"}
	fillDefaults(data3)
	if data3.StatusText != "冲突" {
		t.Errorf("expected 冲突, got %q", data3.StatusText)
	}
}

func TestFillDefaults_TaskNameFallback(t *testing.T) {
	data := &TemplateData{TaskKey: "task-123"}
	fillDefaults(data)
	if data.TaskName != "task-123" {
		t.Errorf("expected TaskName to fallback to TaskKey, got %q", data.TaskName)
	}
}

func TestFillDefaults_RepoNameFallback(t *testing.T) {
	data := &TemplateData{RepoKey: "repo-456"}
	fillDefaults(data)
	if data.RepoName != "repo-456" {
		t.Errorf("expected RepoName to fallback to RepoKey, got %q", data.RepoName)
	}
}

func TestFillDefaults_EventLabel(t *testing.T) {
	data := &TemplateData{EventType: po.TriggerSyncSuccess}
	fillDefaults(data)
	if data.EventLabel != "同步成功" {
		t.Errorf("expected 同步成功, got %q", data.EventLabel)
	}
}

func TestFillDefaults_TimestampAutoFilled(t *testing.T) {
	data := &TemplateData{Status: "success"}
	fillDefaults(data)
	if data.Timestamp == "" {
		t.Error("expected Timestamp to be auto-filled")
	}
}

func TestRenderTemplate_AllBranchSync(t *testing.T) {
	data := &TemplateData{
		TaskName:     "全分支同步",
		RepoName:     "my-repo",
		Status:       "success",
		StatusText:   "成功",
		EventType:    po.TriggerSyncSuccess,
		SyncMode:     "all-branch",
		SourceRemote: "origin",
		TargetRemote: "backup",
		BranchCount:  10,
		SuccessCount: 8,
		FailedCount:  2,
	}
	tmpl := GetDefaultContentTemplate(po.TriggerSyncSuccess)
	result, err := RenderTemplate(tmpl, data)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, "全分支同步") {
		t.Errorf("expected sync mode info, got: %s", result)
	}
}
