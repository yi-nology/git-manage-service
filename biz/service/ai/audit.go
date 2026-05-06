package ai

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"github.com/yi-nology/git-manage-service/biz/service/llm"
)

func hashText(text string) string {
	sum := sha256.Sum256([]byte(text))
	return hex.EncodeToString(sum[:])
}

func serializeInput(systemPrompt string, messages []llm.ChatMessage) string {
	var b strings.Builder
	b.WriteString(systemPrompt)
	for _, msg := range messages {
		b.WriteString("\n\n")
		b.WriteString(msg.Role)
		b.WriteString(":")
		b.WriteString(msg.Content)
	}
	return b.String()
}

func recordInvocation(req TaskRequest, providerName, input, output, status, errMsg string, started time.Time) uint {
	if db.DB == nil {
		return 0
	}
	metadataJSON := ""
	if len(req.Metadata) > 0 {
		raw, err := json.Marshal(req.Metadata)
		if err == nil {
			metadataJSON = string(raw)
		}
	}
	record := &po.AIInvocation{
		TaskType:      string(req.Type),
		ProviderName:  providerName,
		PromptVersion: req.PromptVersion,
		RepoKey:       req.RepoKey,
		OperatorID:    req.OperatorID,
		RelatedID:     req.RelatedID,
		InputHash:     hashText(input),
		OutputHash:    hashText(output),
		Status:        status,
		ErrorMessage:  errMsg,
		LatencyMs:     time.Since(started).Milliseconds(),
		InputChars:    len(input),
		OutputChars:   len(output),
		MetadataJSON:  metadataJSON,
	}
	if err := db.DB.Create(record).Error; err != nil {
		log.Printf("[AI] failed to record invocation: %v", err)
		return 0
	}
	return record.ID
}

func RecordUserFeedback(invocationID uint, feedback string) error {
	if db.DB == nil {
		return nil
	}
	return db.DB.Model(&po.AIInvocation{}).Where("id = ?", invocationID).Update("user_feedback", feedback).Error
}

func GetInvocation(id uint) (*po.AIInvocation, error) {
	var inv po.AIInvocation
	err := db.DB.First(&inv, id).Error
	return &inv, err
}
