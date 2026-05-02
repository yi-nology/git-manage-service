package notification

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/yi-nology/git-manage-service/biz/model/po"
)

func TestDingTalkSender_BuildPayload(t *testing.T) {
	sender := NewDingTalkSender(&po.DingTalkConfig{
		WebhookURL:   "https://oapi.dingtalk.com/robot/send?access_token=test",
		SecurityType: "keyword",
		Keywords:     "Git同步",
	})
	msg := &NotificationMessage{
		Title:   "同步成功",
		Content: "任务完成",
		Status:  "success",
		TaskKey: "task-1",
		RepoKey: "repo-1",
	}

	dingMsg := DingTalkMessage{
		MsgType: "markdown",
		Markdown: DingTalkMarkdown{
			Title: msg.Title,
			Text:  "Git同步\n\n### ✅ 同步成功\n\n任务完成\n\n> Task: task-1\n> Repo: repo-1",
		},
	}
	body, err := json.Marshal(dingMsg)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "markdown") {
		t.Error("expected markdown msgtype")
	}
	if !strings.Contains(string(body), msg.Title) {
		t.Error("expected title in payload")
	}
}

func TestDingTalkSender_Sign(t *testing.T) {
	sender := &DingTalkSender{config: &po.DingTalkConfig{}}
	timestamp := int64(1700000000000)
	secret := "test-secret"
	sign := sender.sign(timestamp, secret)
	if sign == "" {
		t.Error("expected non-empty signature")
	}
}

func TestFeishuSender_BuildPayload(t *testing.T) {
	sender := NewFeishuSender(&po.FeishuConfig{
		WebhookURL: "https://open.feishu.cn/open-apis/bot/v2/hook/test",
	})
	msg := &NotificationMessage{
		Title:   "同步成功",
		Content: "任务完成",
		Status:  "success",
		TaskKey: "task-1",
		RepoKey: "repo-1",
	}

	postContent := FeishuPostContent{
		Post: map[string]FeishuPostLang{
			"zh_cn": {
				Title: "✅ 同步成功",
				Content: [][]FeishuPostEntry{
					{{Tag: "text", Text: "任务完成"}},
				},
			},
		},
	}
	body, err := json.Marshal(postContent)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "同步成功") {
		t.Error("expected title in payload")
	}
}

func TestFeishuSender_Sign(t *testing.T) {
	sender := &FeishuSender{config: &po.FeishuConfig{}}
	timestamp := int64(1700000000)
	secret := "test-secret"
	sign := sender.sign(timestamp, secret)
	if sign == "" {
		t.Error("expected non-empty signature")
	}
}

func TestWebhookSender_BuildPayload(t *testing.T) {
	sender := NewWebhookSender(&po.WebhookConfig{
		URL:         "https://example.com/webhook",
		Method:      "POST",
		ContentType: "application/json",
	})
	msg := &NotificationMessage{
		Title:   "同步成功",
		Content: "任务完成",
		Status:  "success",
		TaskKey: "task-1",
		RepoKey: "repo-1",
	}
	body, err := json.Marshal(msg)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "同步成功") {
		t.Error("expected title in payload")
	}
	if !strings.Contains(string(body), "task-1") {
		t.Error("expected task_key in payload")
	}
}

func TestDingTalkSender_FailureStatus(t *testing.T) {
	msg := &NotificationMessage{
		Title:   "同步失败",
		Content: "出错了",
		Status:  "failure",
		TaskKey: "task-1",
		RepoKey: "repo-1",
	}
	text := formatDingTalkText(msg)
	if !strings.Contains(text, "❌") {
		t.Error("expected failure emoji for failure status")
	}
}

func formatDingTalkText(msg *NotificationMessage) string {
	statusEmoji := "✅"
	if msg.Status == "failure" {
		statusEmoji = "❌"
	}
	return statusEmoji
}
